const serverUrlEl = document.getElementById('serverUrl');
const apiTokenEl = document.getElementById('apiToken');
const savedEl = document.getElementById('saved');

chrome.storage.local.get(['serverUrl', 'apiToken'], ({ serverUrl, apiToken }) => {
  serverUrlEl.value = serverUrl || '';
  apiTokenEl.value = apiToken || '';
});

document.getElementById('save').addEventListener('click', async () => {
  await chrome.storage.local.set({
    serverUrl: serverUrlEl.value.trim(),
    apiToken: apiTokenEl.value.trim(),
  });
  savedEl.textContent = 'Saved.';
  setTimeout(() => (savedEl.textContent = ''), 2000);
});
