import { apiFetch } from './api.js';
import { esc, renderMarkdown } from './ui.js';
import { topicsCache, stylesCache, sourcesCache } from './state.js';

let loaders = {};

export function initEditModal(entityLoaders) {
  loaders = entityLoaders;

  document.getElementById('modal-edit-close').addEventListener('click', closeEditModal);
  document.getElementById('modal-edit-cancel').addEventListener('click', closeEditModal);
  document.getElementById('modal-edit').addEventListener('click', e => {
    if (e.target === document.getElementById('modal-edit')) closeEditModal();
  });
}

function closeEditModal() {
  document.getElementById('modal-edit').classList.remove('open');
}

function field(label, html) {
  return `<div class="field"><label>${label}</label>${html}</div>`;
}

export function openEditModal(entity, item) {
  const titles = { topics: 'Edit Topic', sources: 'Edit Source', styles: 'Edit Style', drafts: 'Edit Draft' };
  document.getElementById('modal-edit-title').textContent = titles[entity] || 'Edit';
  document.getElementById('notice-edit').className = 'notice';
  document.getElementById('notice-edit').textContent = '';

  const fieldsEl = document.getElementById('modal-edit-fields');
  fieldsEl.innerHTML = renderEditFields(entity, item);

  // Wire up Write/Preview tabs for drafts
  if (entity === 'drafts') {
    fieldsEl.querySelectorAll('.editor-tab').forEach(tab => {
      tab.addEventListener('click', () => {
        fieldsEl.querySelectorAll('.editor-tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        const writePane = fieldsEl.querySelector('.editor-write');
        const previewPane = fieldsEl.querySelector('.editor-preview');
        if (tab.dataset.target === 'preview') {
          writePane.style.display = 'none';
          previewPane.style.display = 'block';
          previewPane.innerHTML = renderMarkdown(fieldsEl.querySelector('#ef-content').value);
        } else {
          writePane.style.display = 'block';
          previewPane.style.display = 'none';
        }
      });
    });
  }

  const saveBtn = document.getElementById('modal-edit-save');
  saveBtn.onclick = async () => {
    const body = collectEditBody(entity, fieldsEl, item);
    try {
      saveBtn.disabled = true;
      await apiFetch(`/${entity}/${item.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });
      closeEditModal();
      await loaders[entity]();
    } catch (e) {
      const n = document.getElementById('notice-edit');
      n.textContent = e.message;
      n.className = 'notice error';
    } finally {
      saveBtn.disabled = false;
    }
  };

  document.getElementById('modal-edit').classList.add('open');
}

function renderEditFields(entity, item) {
  if (entity === 'topics') {
    return field('Name', `<input type="text" id="ef-name" value="${esc(item.name)}" required>`) +
      field('Description', `<textarea id="ef-description" style="min-height:70px">${esc(item.description)}</textarea>`) +
      field('Keywords', `<input type="text" id="ef-keywords" value="${esc(item.keywords)}">`);
  }
  if (entity === 'sources') {
    return field('Name', `<input type="text" id="ef-name" value="${esc(item.name)}" required>`) +
      field('Type', `<select id="ef-type"><option value="text"${item.type==='text'?' selected':''}>Text</option><option value="url"${item.type==='url'?' selected':''}>URL</option><option value="file"${item.type==='file'?' selected':''}>File</option></select>`) +
      field('Raw', `<textarea id="ef-raw" style="min-height:80px">${esc(item.raw)}</textarea>`) +
      field('Content', `<textarea id="ef-content" style="min-height:80px">${esc(item.content)}</textarea>`) +
      field('Status', `<select id="ef-status"><option value="ready"${item.status==='ready'?' selected':''}>ready</option><option value="pending"${item.status==='pending'?' selected':''}>pending</option><option value="error"${item.status==='error'?' selected':''}>error</option></select>`);
  }
  if (entity === 'styles') {
    return field('Name', `<input type="text" id="ef-name" value="${esc(item.name)}" required>`) +
      field('Tone', `<input type="text" id="ef-tone" value="${esc(item.tone)}">`) +
      field('Prompt', `<textarea id="ef-prompt" style="min-height:90px">${esc(item.prompt)}</textarea>`) +
      field('Example', `<textarea id="ef-example" style="min-height:70px">${esc(item.example)}</textarea>`);
  }
  if (entity === 'drafts') {
    const topicOpts = topicsCache.map(t => `<option value="${t.id}"${t.id===item.topic_id?' selected':''}>${esc(t.name)}</option>`).join('');
    const styleOpts = stylesCache.map(s => `<option value="${s.id}"${s.id===item.style_id?' selected':''}>${esc(s.name)}</option>`).join('');
    const sourceChecks = sourcesCache.map(s =>
      `<label><input type="checkbox" name="source_ids" value="${s.id}"${(item.source_ids||[]).includes(s.id)?' checked':''}> ${esc(s.name)}</label>`
    ).join('');
    return field('Title', `<input type="text" id="ef-title" value="${esc(item.title)}" required>`) +
      field('Topic', `<select id="ef-topic_id"><option value="">— select —</option>${topicOpts}</select>`) +
      field('Style', `<select id="ef-style_id"><option value="">— select —</option>${styleOpts}</select>`) +
      field('Sources', `<div class="checkbox-list" id="ef-sources">${sourceChecks || '<div style="padding:0.6rem 0.75rem;font-size:0.78rem;color:var(--muted)">No sources.</div>'}</div>`) +
      field('Notes', `<textarea id="ef-notes" style="min-height:60px">${esc(item.notes)}</textarea>`) +
      `<div class="field"><label>Content</label>
        <div class="editor-tabs">
          <button type="button" class="editor-tab active" data-target="write">Write</button>
          <button type="button" class="editor-tab" data-target="preview">Preview</button>
        </div>
        <div class="editor-write"><textarea id="ef-content" style="min-height:120px">${esc(item.content)}</textarea></div>
        <div class="editor-preview md-content" style="display:none"></div>
      </div>` +
      field('Status', `<select id="ef-status"><option value="draft"${item.status==='draft'?' selected':''}>draft</option><option value="published"${item.status==='published'?' selected':''}>published</option></select>`);
  }
  return '';
}

function collectEditBody(entity, fieldsEl, item) {
  const v = id => { const el = fieldsEl.querySelector('#' + id); return el ? el.value : ''; };
  if (entity === 'topics') return { name: v('ef-name'), description: v('ef-description'), keywords: v('ef-keywords') };
  if (entity === 'sources') return { name: v('ef-name'), type: v('ef-type'), raw: v('ef-raw'), content: v('ef-content'), status: v('ef-status') };
  if (entity === 'styles') return { name: v('ef-name'), tone: v('ef-tone'), prompt: v('ef-prompt'), example: v('ef-example') };
  if (entity === 'drafts') {
    const srcIds = [...fieldsEl.querySelectorAll('input[name="source_ids"]:checked')].map(cb => parseInt(cb.value, 10));
    return {
      title: v('ef-title'),
      topic_id: parseInt(v('ef-topic_id'), 10) || 0,
      style_id: parseInt(v('ef-style_id'), 10) || 0,
      source_ids: srcIds,
      notes: v('ef-notes'),
      content: v('ef-content'),
      status: v('ef-status'),
    };
  }
  return {};
}
