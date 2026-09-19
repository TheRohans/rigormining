// Runs on every page. Only ever *collects* raw data on request - all
// citation/DOI interpretation happens server-side (see backend
// internals/handlers/capture.go), so this script can stay tiny and stable.
//
// PDF fetching also happens here rather than in the background script,
// deliberately: a fetch() made from this content script is same-origin
// with the page itself, so it automatically carries whatever session the
// user already has on that site - campus IP access, an EZproxy-rewritten
// institutional session, a Shibboleth/SSO cookie, etc. A fetch from the
// background service worker would be cross-origin (chrome-extension://...)
// and would NOT carry those cookies by default, so it would just hit a
// login/paywall page instead of the real PDF.

// A handful of publishers serve their PDF through a JS-driven inline
// viewer with no citation_pdf_url meta tag and no plain link at all - the
// actual file exists at a real, fetchable URL, it's just not advertised on
// the viewer page itself. These are hand-picked URL-rewrite rules for
// cases we've hit, the same idea as the per-site "translators" reference
// managers like Zotero maintain. A wrong or stale guess just surfaces as a
// clear fetch-failed reason in the popup - no harm done - so add to this
// list as new cases turn up rather than trying to solve it generically.
function guessPublisherPdfUrl(pageUrl) {
  const url = new URL(pageUrl);

  // ACM Digital Library: /doi/epdf/<doi> is the inline JS viewer;
  // /doi/pdf/<doi> is the same paper's direct, fetchable PDF.
  if (url.hostname === 'dl.acm.org') {
    const match = url.pathname.match(/^\/doi\/epdf\/(.+)$/);
    if (match) {
      return `${url.origin}/doi/pdf/${match[1]}`;
    }
  }

  return '';
}

// General-purpose fallback for sites with no usable citation_pdf_url meta
// tag at all (confirmed: ScienceDirect, even for a fully open-access,
// CC-licensed article - Elsevier just doesn't publish one): find a "View
// PDF"/"Download PDF"-looking link actually present in the page's DOM.
//
// One real wrinkle this was built around: such a page can have MORE than
// one PDF-looking link - e.g. ScienceDirect's "Recommended articles"
// sidebar lists "View PDF" links for OTHER papers too. Blindly taking the
// first match found 14 candidates on one page and only 1 of them was for
// the actual article being viewed. So a candidate only counts if the
// page's own identifier (its URL's last path segment - a PII, DOI suffix,
// slug, etc.) also appears in that candidate's href, which is what
// actually distinguishes "this article's PDF" from "a different one
// recommended alongside it."
function findPdfLinkInDom() {
  const pathSegments = window.location.pathname.split('/').filter(Boolean);
  const pageId = pathSegments[pathSegments.length - 1] || '';

  const candidates = Array.from(document.querySelectorAll('a[href]'))
    .filter((el) => {
      const href = el.href; // the resolved absolute URL, not the raw attribute
      const text = (el.textContent || '').trim();
      const looksLikePdf = /\.pdf(\?|#|$)|pdfft|\/pdf\b/i.test(href) || /view pdf|download pdf|full[- ]text pdf/i.test(text);
      return looksLikePdf && (!pageId || href.includes(pageId));
    })
    .map((el) => el.href);

  console.log('[Rigormining] pdf-like DOM links matching this page\'s own id:', candidates);
  return candidates[0] || '';
}

// Diagnostic logging: given how many one-off per-publisher quirks this
// extension has already hit (JS viewers, redirect-to-a-different-CORS-
// origin, etc.) and will keep hitting, cheap visibility into exactly what
// URL was tried and why it failed is worth keeping permanently rather than
// re-adding it every time. Catches CSP violations directly rather than
// relying on eyeballing the browser's own separate console message, so we
// can say for certain whether that's the cause.
document.addEventListener('securitypolicyviolation', (e) => {
  console.warn('[Rigormining] CSP violation:', {
    violatedDirective: e.violatedDirective,
    blockedURI: e.blockedURI,
    documentURI: e.documentURI,
    originalPolicy: e.originalPolicy,
  });
});

chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message.type === 'RIGORMINING_COLLECT_PAGE_INFO') {
    const meta = Array.from(document.querySelectorAll('meta[name], meta[property]')).map((el) => ({
      name: el.getAttribute('name') || el.getAttribute('property') || '',
      content: el.getAttribute('content') || '',
    }));

    const pdfMeta = meta.find((m) => m.name.toLowerCase() === 'citation_pdf_url');
    let pdfUrl = pdfMeta ? pdfMeta.content : guessPublisherPdfUrl(window.location.href);
    if (!pdfUrl) {
      pdfUrl = findPdfLinkInDom();
    }
    if (!pdfUrl && /\.pdf(\?|#|$)/i.test(window.location.href)) {
      pdfUrl = window.location.href;
    }

    console.log('[Rigormining] page url:', window.location.href);
    console.log(
      '[Rigormining] citation/pdf-related meta tags (name: content):',
      meta.filter((m) => /citation|pdf/i.test(m.name)).map((m) => `${m.name}: ${m.content}`),
    );
    console.log('[Rigormining] chosen pdfUrl:', pdfUrl || '(none found)');

    sendResponse({
      url: window.location.href,
      title: document.title,
      meta,
      pdfUrl,
    });
    return true;
  }

  if (message.type === 'RIGORMINING_FETCH_PDF') {
    console.log('[Rigormining] attempting PDF fetch:', message.url);
    // 'same-origin' (not 'include'): still sends cookies for a same-origin
    // request - which covers institutional/proxy/SSO sessions fine, since
    // those live on the same origin as the page - but doesn't force
    // credentials onto a DIFFERENT origin a redirect might land on. Some
    // publishers (confirmed: MDPI) redirect their PDF URL to a separate
    // static-file host that serves with `Access-Control-Allow-Origin: *`;
    // that wildcard is explicitly incompatible with a credentialed request
    // (the CORS spec forbids it, since a "readable by anyone" response
    // combined with "also send this browser's cookies" would be a
    // cross-site data leak) - so 'include' hard-fails there with a
    // "NetworkError", while 'same-origin' just omits credentials for that
    // cross-origin leg and lets the otherwise-public fetch succeed.
    fetch(message.url, { credentials: 'same-origin' })
      .then(async (res) => {
        console.log('[Rigormining] PDF fetch response:', {
          status: res.status,
          type: res.type,
          finalUrl: res.url,
          contentType: res.headers.get('content-type'),
        });
        if (!res.ok) {
          sendResponse({ ok: false, error: `HTTP ${res.status}` });
          return;
        }
        const contentType = res.headers.get('content-type') || '';
        const buffer = await res.arrayBuffer();

        // Trust neither the URL string nor the Content-Type header - both
        // can be misleading (confirmed: a ScienceDirect pdfft URL ends in
        // "...-main.pdf" only because of a query parameter, not because the
        // endpoint reliably returns a raw file; an interstitial/error page
        // it might return instead could still be served as text/html or
        // something else entirely). A real PDF's first bytes are always
        // the literal signature "%PDF-" - check that directly instead.
        const header = new TextDecoder().decode(new Uint8Array(buffer.slice(0, 5)));
        if (header !== '%PDF-') {
          console.warn('[Rigormining] fetched content is not a real PDF - first bytes were:', header);
          sendResponse({
            ok: false,
            error: `the response wasn't actually a PDF (content-type: ${contentType || 'unknown'}) - likely a login or interstitial page`,
          });
          return;
        }

        sendResponse({ ok: true, contentType, buffer });
      })
      .catch((err) => {
        console.error('[Rigormining] PDF fetch threw:', { name: err.name, message: err.message, err });
        sendResponse({ ok: false, error: err.message });
      });
    return true; // keep the message channel open for the async response
  }

  return undefined;
});
