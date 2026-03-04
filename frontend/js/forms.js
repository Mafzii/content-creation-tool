import { API } from './state.js';
import { apiFetch } from './api.js';
import { showNotice, clearNotice } from './ui.js';

export function formData(form) {
  const data = {};
  new FormData(form).forEach((v, k) => { data[k] = v; });
  return data;
}

export async function handleSubmit(entity, form, reload) {
  clearNotice(entity);
  const body = formData(form);
  try {
    const res = await fetch(API + `/${entity}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    if (res.status === 201) {
      form.reset();
      await reload();
      showNotice(entity, 'Created successfully.', 'success');
    } else {
      const text = await res.text();
      throw new Error(text || `HTTP ${res.status}`);
    }
  } catch (e) { showNotice(entity, e.message); }
}

export async function deleteEntity(entity, id, reload) {
  if (!confirm(`Delete this ${entity.slice(0, -1)}?`)) return;
  try {
    await apiFetch(`/${entity}/${id}`, { method: 'DELETE' });
    await reload();
  } catch (e) { showNotice(entity, e.message); }
}

export function initSimpleForms(loadTopics, loadStyles) {
  document.getElementById('form-topics').addEventListener('submit', e => {
    e.preventDefault();
    handleSubmit('topics', e.target, loadTopics);
  });
  document.getElementById('form-styles').addEventListener('submit', e => {
    e.preventDefault();
    handleSubmit('styles', e.target, loadStyles);
  });
}
