import * as state from './state.js';
import { apiFetch } from './api.js';
import { showNotice, clearNotice, esc, makeItem, renderList } from './ui.js';
import { populateDraftSourceList } from './dropdowns.js';
import { deleteEntity } from './forms.js';
import { openEditModal } from './edit-modal.js';

export async function loadSources() {
  try {
    const data = await apiFetch('/sources');
    state.setSourcesCache(data || []);
    const items = state.sourcesCache.map(s => {
      let secondary = s.type;
      if (s.type === 'url' && s.raw) secondary += ': ' + s.raw.slice(0, 60);
      else if (s.type === 'file' && s.raw) secondary += ': ' + s.raw;
      else if (s.raw) secondary += ': ' + s.raw.slice(0, 60);

      if (s.extract_mode === 'ai') secondary += ' [AI]';

      const item = makeItem(
        { primary: s.name, secondary, badge: s.status || 'ready' },
        () => deleteEntity('sources', s.id, loadSources),
        () => openEditModal('sources', s)
      );

      if (s.type === 'url') {
        const refetchBtn = document.createElement('button');
        refetchBtn.className = 'refetch-btn';
        refetchBtn.textContent = 're-fetch';
        refetchBtn.onclick = async () => {
          refetchBtn.disabled = true;
          refetchBtn.textContent = 'fetching…';
          try {
            await apiFetch(`/sources/${s.id}/fetch`, { method: 'POST' });
            pollSourceStatus(s.id, () => loadSources());
          } catch (err) {
            showNotice('sources', 'Re-fetch failed: ' + err.message);
            refetchBtn.disabled = false;
            refetchBtn.textContent = 're-fetch';
          }
        };
        const deleteBtn = item.querySelector('.delete-btn');
        item.insertBefore(refetchBtn, deleteBtn);
      }

      if (s.type === 'url' && s.raw) {
        const metaEl = item.querySelector('.item-meta');
        if (metaEl) {
          metaEl.innerHTML = `url: <a href="${esc(s.raw)}" target="_blank" rel="noopener" style="color:var(--accent);text-decoration:underline">${esc(s.raw.slice(0, 50))}</a>`;
          if (s.extract_mode === 'ai') metaEl.innerHTML += ' <span style="color:var(--muted)">[AI]</span>';
        }
      }

      return item;
    });
    renderList('list-sources', items);
    populateDraftSourceList();
  } catch (e) { showNotice('sources', e.message); }
}

export function pollSourceStatus(sourceId, onReady) {
  const interval = setInterval(async () => {
    try {
      const data = await apiFetch(`/sources/${sourceId}/status`);
      if (data.status === 'ready' || data.status === 'error' || data.status === 'partial') {
        clearInterval(interval);
        onReady();
      }
    } catch {
      clearInterval(interval);
    }
  }, 2000);
}

export function updateSourceRawLabel() {
  const type = document.getElementById('source-type-select').value;
  const lbl = document.getElementById('source-raw-label');
  const inp = document.getElementById('source-raw-input');
  const rawField = document.getElementById('source-raw-field');
  const fileField = document.getElementById('source-file-field');
  const contentField = document.getElementById('source-content-field');
  const extractModeField = document.getElementById('source-extract-mode-field');
  const topicField = document.getElementById('source-topic-field');

  if (type === 'url') {
    rawField.style.display = '';
    fileField.style.display = 'none';
    contentField.style.display = 'none';
    extractModeField.style.display = '';
    lbl.textContent = 'URL';
    inp.placeholder = 'https://...';
    inp.style.minHeight = '38px';
    inp.rows = 1;
    updateTopicFieldVisibility();
  } else if (type === 'file') {
    rawField.style.display = 'none';
    fileField.style.display = '';
    contentField.style.display = 'none';
    extractModeField.style.display = 'none';
    topicField.style.display = 'none';
  } else {
    rawField.style.display = '';
    fileField.style.display = 'none';
    contentField.style.display = '';
    extractModeField.style.display = 'none';
    topicField.style.display = 'none';
    lbl.textContent = 'Content';
    inp.placeholder = 'Paste your source text here…';
    inp.style.minHeight = '100px';
    inp.rows = null;
  }
}

function updateTopicFieldVisibility() {
  const topicField = document.getElementById('source-topic-field');
  const selectedMode = document.querySelector('input[name="extract_mode"]:checked');
  if (selectedMode && selectedMode.value === 'ai') {
    topicField.style.display = '';
  } else {
    topicField.style.display = 'none';
  }
}

export function populateSourceTopicSelect() {
  const sel = document.getElementById('source-topic-select');
  const current = sel.value;
  sel.innerHTML = '<option value="0">-- none --</option>';
  state.topicsCache.forEach(t => {
    const opt = document.createElement('option');
    opt.value = t.id;
    opt.textContent = t.name;
    sel.appendChild(opt);
  });
  if (current) sel.value = current;
}

export function initSourceForm() {
  document.getElementById('source-type-select').addEventListener('change', updateSourceRawLabel);

  // Listen for extract mode radio changes
  document.querySelectorAll('input[name="extract_mode"]').forEach(radio => {
    radio.addEventListener('change', updateTopicFieldVisibility);
  });

  document.getElementById('form-sources').addEventListener('submit', async e => {
    e.preventDefault();
    clearNotice('sources');
    const form = e.target;
    const fd = new FormData(form);
    const type = fd.get('type') || 'text';

    try {
      let res;
      if (type === 'file') {
        const uploadFd = new FormData();
        uploadFd.append('name', fd.get('name') || '');
        uploadFd.append('type', 'file');
        const fileInput = document.getElementById('source-file-input');
        if (!fileInput.files.length) {
          showNotice('sources', 'Please select a file.');
          return;
        }
        uploadFd.append('file', fileInput.files[0]);
        res = await fetch(state.API + '/sources', { method: 'POST', body: uploadFd });
      } else {
        const raw = fd.get('raw') || '';
        let content = fd.get('content') || '';
        if (type === 'text' && !content) content = raw;
        const body = { name: fd.get('name') || '', type, raw, content };

        if (type === 'url') {
          body.extract_mode = fd.get('extract_mode') || 'standard';
          body.topic_id = parseInt(fd.get('source_topic_id') || '0', 10);
        }

        res = await fetch(state.API + '/sources', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body)
        });
      }

      if (res.status === 201) {
        const created = await res.json();
        form.reset();
        updateSourceRawLabel();
        await loadSources();

        if (type === 'url') {
          const modeLabel = created.extract_mode === 'ai' ? 'AI extracting' : 'Fetching';
          showNotice('sources', `Source created. ${modeLabel} URL content…`, 'success');
          const statusEl = document.getElementById('source-fetch-status');
          statusEl.style.display = 'block';
          statusEl.textContent = `${modeLabel} URL content…`;
          pollSourceStatus(created.id, async () => {
            statusEl.style.display = 'none';
            await loadSources();
            const src = state.sourcesCache.find(s => s.id === created.id);
            if (src && src.status === 'partial') {
              showNotice('sources', 'URL fetched but AI extraction failed. Raw content was used instead.', 'error');
            } else {
              showNotice('sources', 'URL content fetched.', 'success');
            }
          });
        } else {
          showNotice('sources', 'Created successfully.', 'success');
        }
      } else {
        const text = await res.text();
        throw new Error(text || `HTTP ${res.status}`);
      }
    } catch (err) { showNotice('sources', err.message); }
  });
}
