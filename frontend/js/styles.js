import { apiFetch } from './api.js';
import { showNotice, makeItem, renderList } from './ui.js';
import * as state from './state.js';
import { populateStyleSelects } from './dropdowns.js';
import { deleteEntity } from './forms.js';
import { openEditModal } from './edit-modal.js';

export async function loadStyles() {
  try {
    const data = await apiFetch('/styles');
    state.setStylesCache(data || []);
    const items = state.stylesCache.map(s => makeItem(
      { primary: s.name, secondary: (s.tone ? s.tone + ' · ' : '') + (s.prompt ? s.prompt.slice(0, 60) + (s.prompt.length > 60 ? '…' : '') : '') },
      () => deleteEntity('styles', s.id, loadStyles),
      () => openEditModal('styles', s)
    ));
    renderList('list-styles', items);
    populateStyleSelects();
  } catch (e) { showNotice('styles', e.message); }
}
