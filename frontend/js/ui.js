export function showNotice(tab, msg, type = 'error') {
  const el = document.getElementById(`notice-${tab}`);
  el.textContent = msg;
  el.className = `notice ${type}`;
  if (type === 'success') setTimeout(() => { el.className = 'notice'; el.textContent = ''; }, 3000);
}

export function clearNotice(tab) {
  const el = document.getElementById(`notice-${tab}`);
  el.className = 'notice'; el.textContent = '';
}

export function esc(str) {
  return (str || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

export function makeItem(fields, onDelete, onEdit) {
  const div = document.createElement('div');
  div.className = 'list-item';
  const info = document.createElement('div');
  info.className = 'item-info';

  const name = document.createElement('div');
  name.className = 'item-name';
  name.textContent = fields.primary;
  info.appendChild(name);

  if (fields.secondary) {
    const meta = document.createElement('div');
    meta.className = 'item-meta';
    meta.textContent = fields.secondary;
    info.appendChild(meta);
  }
  div.appendChild(info);

  if (fields.badge) {
    const badge = document.createElement('span');
    badge.className = 'item-badge' + (fields.badge === 'published' ? ' published' : fields.badge === 'ready' ? ' ready' : fields.badge === 'pending' ? ' pending' : fields.badge === 'error' ? ' error' : '');
    badge.textContent = fields.badge;
    div.appendChild(badge);
  }

  if (onEdit) {
    const editBtn = document.createElement('button');
    editBtn.className = 'edit-btn';
    editBtn.textContent = 'edit';
    editBtn.onclick = onEdit;
    div.appendChild(editBtn);
  }

  const btn = document.createElement('button');
  btn.className = 'delete-btn';
  btn.textContent = 'delete';
  btn.onclick = onDelete;
  div.appendChild(btn);
  return div;
}

export function renderList(containerId, items) {
  const el = document.getElementById(containerId);
  el.innerHTML = '';
  if (!items.length) {
    el.innerHTML = '<div class="list-empty">Nothing here yet.</div>';
    return;
  }
  items.forEach(item => el.appendChild(item));
}
