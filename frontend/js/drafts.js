import * as state from './state.js';
import { apiFetch } from './api.js';
import { showNotice, clearNotice, makeItem, renderList } from './ui.js';
import { deleteEntity } from './forms.js';
import { openEditModal } from './edit-modal.js';

export async function loadDrafts() {
  try {
    const data = await apiFetch('/drafts');
    const items = (data || []).map(d => {
      const topic = state.topicsCache.find(t => t.id === d.topic_id);
      const style = state.stylesCache.find(s => s.id === d.style_id);
      return makeItem(
        {
          primary: d.title,
          secondary: `${topic ? topic.name : d.topic_id} · ${style ? style.name : d.style_id}`,
          badge: d.status
        },
        () => deleteEntity('drafts', d.id, loadDrafts),
        () => openEditModal('drafts', d)
      );
    });
    renderList('list-drafts', items);
  } catch (e) { showNotice('drafts', e.message); }
}

export function initDraftForm() {
  document.getElementById('form-drafts').addEventListener('submit', async e => {
    e.preventDefault();
    clearNotice('drafts');
    const form = e.target;
    const fd = new FormData(form);
    const body = {
      title: fd.get('title') || '',
      content: fd.get('content') || '',
      topic_id: parseInt(fd.get('topic_id'), 10) || 0,
      style_id: parseInt(fd.get('style_id'), 10) || 0,
      status: fd.get('status') || 'draft',
      notes: fd.get('notes') || '',
      source_ids: fd.getAll('source_ids').map(v => parseInt(v, 10)).filter(Boolean),
    };
    try {
      const res = await fetch(state.API + '/drafts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      if (res.status === 201) {
        form.reset();
        await loadDrafts();
        showNotice('drafts', 'Created successfully.', 'success');
      } else {
        const text = await res.text();
        throw new Error(text || `HTTP ${res.status}`);
      }
    } catch (err) { showNotice('drafts', err.message); }
  });
}
