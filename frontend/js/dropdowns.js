import { topicsCache, stylesCache, sourcesCache } from './state.js';

export function populateSelect(selectEl, items, valueKey, labelKey) {
  const current = selectEl.value;
  selectEl.innerHTML = '<option value="">— select —</option>';
  items.forEach(item => {
    const opt = document.createElement('option');
    opt.value = item[valueKey];
    opt.textContent = item[labelKey];
    selectEl.appendChild(opt);
  });
  if (current) selectEl.value = current;
}

export function populateTopicSelects() {
  document.querySelectorAll('select[name="topic_id"]').forEach(sel => {
    populateSelect(sel, topicsCache, 'id', 'name');
  });
}

export function populateStyleSelects() {
  document.querySelectorAll('select[name="style_id"]').forEach(sel => {
    populateSelect(sel, stylesCache, 'id', 'name');
  });
}

export function populateDraftSourceList() {
  const wrap = document.getElementById('draft-source-list');
  wrap.innerHTML = '';
  if (!sourcesCache.length) {
    wrap.innerHTML = '<div style="padding:0.6rem 0.75rem;font-size:0.78rem;color:var(--muted)">No sources available.</div>';
    return;
  }
  sourcesCache.forEach(s => {
    const lbl = document.createElement('label');
    const cb = document.createElement('input');
    cb.type = 'checkbox';
    cb.value = s.id;
    cb.name = 'source_ids';
    lbl.appendChild(cb);
    lbl.appendChild(document.createTextNode(' ' + s.name));
    wrap.appendChild(lbl);
  });
}
