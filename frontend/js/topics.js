import { apiFetch } from './api.js';
import { showNotice, makeItem, renderList } from './ui.js';
import * as state from './state.js';
import { populateTopicSelects } from './dropdowns.js';
import { populateSourceTopicSelect } from './sources.js';
import { deleteEntity } from './forms.js';
import { openEditModal } from './edit-modal.js';

export async function loadTopics() {
  try {
    const data = await apiFetch('/topics');
    state.setTopicsCache(data || []);
    const items = state.topicsCache.map(t => makeItem(
      { primary: t.name, secondary: t.keywords || t.description || '' },
      () => deleteEntity('topics', t.id, loadTopics),
      () => openEditModal('topics', t)
    ));
    renderList('list-topics', items);
    populateTopicSelects();
    populateSourceTopicSelect();
  } catch (e) { showNotice('topics', e.message); }
}
