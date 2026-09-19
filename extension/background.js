// Service worker (Chrome) / event page (Firefox): orchestrates a capture -
// asks content.js for page data and (via content.js) the PDF bytes, then
// uploads the result to the self-hosted backend. Talks to popup.js for the
// user-initiated action.
//
// Firefox's native `browser.*` API is promise-based; its `chrome.*` shim
// only supports the callback style, not promises. `chrome.*` itself is
// promise-based in Chrome. Using `browser` when it exists (Firefox) and
// falling back to `chrome` (Chrome/Edge) keeps every call below
// promise-based in both.
const browserAPI = typeof browser !== 'undefined' ? browser : chrome;

// Neither a URL string nor a Content-Type header can be trusted to prove a
// response is really a PDF (confirmed: a ScienceDirect pdfft URL ends in
// "...-main.pdf" purely because of a query parameter, regardless of what
// actually comes back). A real PDF's first bytes are always the literal
// signature "%PDF-" - check that directly instead.
function looksLikePdfBytes(buffer) {
  const header = new TextDecoder().decode(new Uint8Array(buffer.slice(0, 5)));
  return header === '%PDF-';
}

async function getSettings() {
  const { serverUrl, apiToken } = await browserAPI.storage.local.get(['serverUrl', 'apiToken']);
  return { serverUrl: serverUrl || '', apiToken: apiToken || '' };
}

async function collectPageInfo(tabId) {
  return browserAPI.tabs.sendMessage(tabId, { type: 'RIGORMINING_COLLECT_PAGE_INFO' });
}

// Derives a fallback title from a PDF URL's filename, for the "no content
// script available" path below, where there's no page <title> to use.
function filenameFromUrl(url) {
  try {
    const last = new URL(url).pathname.split('/').filter(Boolean).pop();
    return last ? decodeURIComponent(last.replace(/\.pdf$/i, '')) : url;
  } catch (e) {
    return url;
  }
}

// Fetches a PDF directly from the background context, for tabs where no
// content script is running at all - see the "no content script" branch in
// capture() below for when this applies. Because this isn't running in the
// page's own origin, it does NOT reliably carry that page's session
// cookies (institutional login, proxy, etc.) the way content.js's fetch
// does - a PDF that needed a login to view may still fail here even though
// it displayed fine in the tab. That's a real, disclosed limitation, not a
// bug: there's no way to run script in the page in this situation at all
// (see the comment in capture()), so there's no context left that could
// carry the session correctly.
//
// credentials: 'omit', deliberately: a background script's origin is never
// the same as any real website's, so cookies would never actually be sent
// here regardless. Explicitly omitting (rather than 'include', which tries
// anyway) avoids the exact CORS failure diagnosed for MDPI: a publisher
// that redirects to a separate host serving `Access-Control-Allow-Origin:
// *` rejects the request outright if it's credentialed, since that
// wildcard is spec-incompatible with sending credentials.
async function fetchPdfDirect(url) {
  console.log('[Rigormining] (background) attempting direct PDF fetch:', url);
  try {
    const res = await fetch(url, { credentials: 'omit' });
    console.log('[Rigormining] (background) direct fetch response:', {
      status: res.status,
      type: res.type,
      finalUrl: res.url,
      contentType: res.headers.get('content-type'),
    });
    if (!res.ok) {
      return { blob: null, reason: `PDF fetch failed (HTTP ${res.status}) - this can happen if the page needs a login this fetch can't carry` };
    }
    const buffer = await res.arrayBuffer();
    if (!looksLikePdfBytes(buffer)) {
      console.warn('[Rigormining] (background) fetched content is not a real PDF');
      const contentType = res.headers.get('content-type') || '';
      return {
        blob: null,
        reason: `the response wasn't actually a PDF (content-type: ${contentType || 'unknown'}) - likely a login or interstitial page`,
      };
    }
    return { blob: new Blob([buffer], { type: 'application/pdf' }), reason: null };
  } catch (e) {
    console.error('[Rigormining] (background) direct PDF fetch threw:', { name: e.name, message: e.message, e });
    return { blob: null, reason: `PDF fetch failed (${e.message})` };
  }
}

// Delegates the actual fetch to content.js, which runs in the page's own
// origin and therefore carries whatever session/cookies the user already
// has there (institutional proxy, SSO, campus IP access, etc.) - see the
// comment in content.js for why this can't be done from here directly.
// Still fails for PDFs that aren't at a directly-fetchable URL at all (some
// publishers serve the PDF through a JS-driven viewer rather than a plain
// link) - that's an inherent limit, not something either context can fix.
//
// A missing file is a first-class supported *outcome* (the item is still
// created, to be attached later from the library UI) - but silently
// swallowing *why* it's missing makes a real failure indistinguishable
// from "this page just has no PDF." So this always returns a reason
// alongside the blob: null when there was nothing to try, a message
// otherwise, which the popup surfaces to the user.
async function tryFetchPdf(tabId, pdfUrl) {
  if (!pdfUrl) {
    // Distinct from every other branch below returning null here: this
    // means nothing was even attempted (no citation_pdf_url meta tag, no
    // known publisher rewrite rule, URL doesn't look like a PDF itself) -
    // as opposed to an attempt that failed. Collapsing those into the same
    // silent "no file" was exactly the ambiguity that made a real gap (no
    // PDF link on the page at all) indistinguishable from a working
    // capture, from the popup alone.
    return { blob: null, reason: 'no PDF link found on this page' };
  }

  let result;
  try {
    result = await browserAPI.tabs.sendMessage(tabId, { type: 'RIGORMINING_FETCH_PDF', url: pdfUrl });
  } catch (e) {
    return { blob: null, reason: `could not reach the page to fetch the PDF (${e.message})` };
  }

  if (!result || !result.ok) {
    // content.js already does the real-PDF-bytes check itself (see
    // looksLikePdfBytes there) before ever reporting ok: true, so a "not
    // actually a PDF" failure surfaces here as result.error already -
    // nothing left to re-validate on this side.
    return { blob: null, reason: `PDF fetch failed (${result ? result.error : 'no response'})` };
  }
  return { blob: new Blob([result.buffer], { type: 'application/pdf' }), reason: null };
}

async function capture(tabId) {
  const { serverUrl, apiToken } = await getSettings();
  if (!serverUrl || !apiToken) {
    throw new Error('Set your Rigormining server URL and API token in the extension options first.');
  }

  const tab = await browserAPI.tabs.get(tabId);

  // A local file isn't something to "capture from the web" at all - and
  // browsers deliberately block extensions from fetching file:// URLs from
  // a background script regardless of permissions (letting extensions
  // silently read arbitrary files off disk is exactly what that blocks).
  // There's no workaround, and none needed: the right tool for a file
  // that's already on disk is the library's own drag-and-drop upload.
  if (tab.url && tab.url.startsWith('file://')) {
    throw new Error(
      "This is a local file - browsers don't allow extensions to read files from disk. Open your Rigormining library and drag the file into the upload area instead.",
    );
  }

  // Firefox refuses to run content scripts inside its own built-in PDF
  // viewer (deliberately - the viewer becomes a privileged, browser-owned
  // page once a PDF opens directly, not an ordinary web page). When that's
  // what's open, collectPageInfo()'s sendMessage rejects with "receiving
  // end does not exist" - there's no page DOM to scrape metadata from
  // anyway, so fall back to the tab's own URL/title and fetch the PDF from
  // here instead of through content.js.
  let pageInfo;
  let pdfResult;
  try {
    pageInfo = await collectPageInfo(tabId);
    pdfResult = await tryFetchPdf(tabId, pageInfo.pdfUrl);
  } catch (e) {
    pageInfo = { url: tab.url, title: filenameFromUrl(tab.url), meta: [] };
    pdfResult = await fetchPdfDirect(tab.url);
  }

  const { blob: pdfBlob, reason: fileWarning } = pdfResult;

  const form = new FormData();
  form.set('url', pageInfo.url);
  form.set('title', pageInfo.title);
  form.set('meta', JSON.stringify(pageInfo.meta));
  if (pdfBlob) {
    form.set('file', pdfBlob, 'capture.pdf');
  }

  const res = await fetch(`${serverUrl.replace(/\/$/, '')}/api/v1/capture`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${apiToken}` },
    body: form,
  });

  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`Server rejected capture (${res.status}): ${text}`);
  }

  const item = await res.json();
  return { item, fileWarning };
}

browserAPI.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message.type !== 'RIGORMINING_CAPTURE') return undefined;

  capture(message.tabId)
    .then(({ item, fileWarning }) => sendResponse({ ok: true, item, fileWarning }))
    .catch((err) => sendResponse({ ok: false, error: err.message }));

  return true; // keep the message channel open for the async response
});
