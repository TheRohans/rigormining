const statusEl = document.getElementById('status');
const button = document.getElementById('capture');

document.getElementById('open-settings').addEventListener('click', () => {
  chrome.runtime.openOptionsPage();
});

button.addEventListener('click', async () => {
  button.disabled = true;
  statusEl.style.color = '';
  statusEl.textContent = 'Capturing...';

  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  chrome.runtime.sendMessage({ type: 'RIGORMINING_CAPTURE', tabId: tab.id }, (response) => {
    button.disabled = false;
    if (!response) {
      statusEl.style.color = '#b91c1c';
      statusEl.textContent = 'No response from the extension background.';
      return;
    }
    if (!response.ok) {
      statusEl.style.color = '#b91c1c';
      if (response.error && response.error.includes('options')) {
        statusEl.innerHTML = `${response.error} <a href="#" id="open-options">Open options</a>`;
        document.getElementById('open-options').addEventListener('click', (e) => {
          e.preventDefault();
          chrome.runtime.openOptionsPage();
        });
        return;
      }
      statusEl.textContent = `Error: ${response.error}`;
      return;
    }

    if (response.fileWarning) {
      // Item was still created - a missing file is a supported outcome -
      // but the reason it's missing should be visible, not silent.
      statusEl.style.color = '#b45309';
      statusEl.textContent = `Added "${response.item.title}" to your library, but no PDF was attached: ${response.fileWarning}. `;
    } else {
      statusEl.style.color = '#16a34a';
      statusEl.textContent = `Added "${response.item.title}" to your library${response.item.file_type ? ' with its PDF' : ''}. `;
    }

    // The server fills in whatever citation meta tags the page had, but
    // that's frequently incomplete (or the page had none at all) - link
    // straight to the item's own edit page (which now has every
    // BibTeX-relevant field) rather than duplicating that whole editable
    // form inside this tiny, ephemeral popup.
    chrome.storage.local.get(['serverUrl'], ({ serverUrl }) => {
      if (!serverUrl || !response.item?.id) return;
      const editLink = document.createElement('a');
      editLink.href = `${serverUrl.replace(/\/$/, '')}/library/${response.item.id}`;
      editLink.target = '_blank';
      editLink.rel = 'noopener noreferrer';
      editLink.textContent = 'Edit details';
      statusEl.appendChild(editLink);
    });
  });
});
