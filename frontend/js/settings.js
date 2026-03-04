import { apiFetch } from './api.js';
import { showNotice, clearNotice } from './ui.js';

export async function loadSettings() {
  try {
    const data = await apiFetch('/settings');
    if (data['llm_provider']) document.getElementById('setting-provider').value = data['llm_provider'];
    if (data['llm_model']) document.getElementById('setting-model').value = data['llm_model'];
    if (data['llm_api_key']) document.getElementById('setting-apikey').value = data['llm_api_key'];
  } catch (e) { showNotice('settings', e.message); }
}

async function saveSetting(key, value) {
  await apiFetch('/settings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key, value })
  });
}

export function initSettingsForm() {
  document.getElementById('btn-save-settings').addEventListener('click', async () => {
    clearNotice('settings');
    try {
      await saveSetting('llm_provider', document.getElementById('setting-provider').value);
      await saveSetting('llm_model', document.getElementById('setting-model').value);
      await saveSetting('llm_api_key', document.getElementById('setting-apikey').value);
      showNotice('settings', 'Settings saved.', 'success');
    } catch (e) { showNotice('settings', e.message); }
  });
}
