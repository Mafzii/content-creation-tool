const STORAGE_KEY = 'cct_form_drafts';
const SAVE_DEBOUNCE_MS = 300;

let trackedForms = [];

function isChanged(el) {
  if (el.type === 'checkbox' || el.type === 'radio') return el.checked !== el.defaultChecked;
  return el.value !== el.defaultValue;
}

function debounce(fn, delay) {
  let timer = null;
  return function (...args) {
    const context = this;
    clearTimeout(timer);
    timer = setTimeout(() => fn.apply(context, args), delay);
  };
}

function getAllFormData() {
  const saved = {};
  for (const { id, form } of trackedForms) {
    const data = {};
    let hasValue = false;
    for (const el of form.elements) {
      if (!el.name) continue;
      if (el.type === 'checkbox') {
        // collect checked checkboxes into array
        if (!data[el.name]) data[el.name] = [];
        if (el.checked) data[el.name].push(el.value);
        if (isChanged(el)) hasValue = true;
      } else if (el.type === 'file') {
        continue; // can't persist file inputs
      } else {
        data[el.name] = el.value;
        if (isChanged(el)) hasValue = true;
      }
    }
    if (hasValue) saved[id] = data;
  }
  return saved;
}

function saveAll() {
  const data = getAllFormData();
  if (Object.keys(data).length) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  } else {
    localStorage.removeItem(STORAGE_KEY);
  }
}

const debouncedSaveAll = debounce(saveAll, SAVE_DEBOUNCE_MS);

function restoreForm(id, form) {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return;
  let all;
  try { all = JSON.parse(raw); } catch { return; }
  const data = all[id];
  if (!data) return;
  for (const el of form.elements) {
    if (!el.name || !(el.name in data)) continue;
    if (el.type === 'checkbox') {
      const vals = data[el.name];
      el.checked = Array.isArray(vals) && vals.includes(el.value);
    } else if (el.type === 'file') {
      continue;
    } else {
      el.value = data[el.name];
    }
  }
}

function clearForm(id) {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return;
  let all;
  try { all = JSON.parse(raw); } catch { return; }
  delete all[id];
  if (Object.keys(all).length) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(all));
  } else {
    localStorage.removeItem(STORAGE_KEY);
  }
}

export function hasUnsavedForms() {
  for (const { form } of trackedForms) {
    for (const el of form.elements) {
      if (!el.name || el.type === 'file') continue;
      if (isChanged(el)) return true;
    }
  }
  return false;
}

export function trackForm(id, form) {
  if (!form) return;
  trackedForms.push({ id, form });
  restoreForm(id, form);

  form.addEventListener('input', debouncedSaveAll);
  form.addEventListener('change', saveAll);

  const origReset = form.reset.bind(form);
  form.reset = () => {
    origReset();
    clearForm(id);
  };
}

export function initBeforeUnloadWarning() {
  window.addEventListener('beforeunload', e => {
    if (hasUnsavedForms()) {
      e.preventDefault();
    }
  });
}
