import { apiFetch } from './api.js';
import { showNotice, clearNotice, renderMarkdown } from './ui.js';
import * as state from './state.js';

function openGenerateModal() {
  document.getElementById('modal-generate').classList.add('open');
  renderVariantsInModal();
}

function closeGenerateModal() {
  document.getElementById('modal-generate').classList.remove('open');
}

function timeAgo(ts) {
  const diff = Math.floor((Date.now() - ts) / 1000);
  if (diff < 60) return 'just now';
  if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
  if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
  return Math.floor(diff / 86400) + 'd ago';
}

function renderVariantsInModal() {
  const container = document.getElementById('modal-variant-cards');
  container.innerHTML = '';

  state.variantTexts.forEach((text, i) => {
    const card = document.createElement('div');
    card.className = 'variant-card' + (i === state.selectedVariantIdx ? ' selected' : '');
    card.dataset.idx = i;

    const header = document.createElement('div');
    header.className = 'variant-header';

    const radio = document.createElement('input');
    radio.type = 'radio';
    radio.name = 'variant_radio';
    radio.value = i;
    radio.checked = i === state.selectedVariantIdx;
    radio.addEventListener('change', () => {
      state.setSelectedVariantIdx(i);
      container.querySelectorAll('.variant-card').forEach((c, ci) => {
        c.classList.toggle('selected', ci === i);
      });
    });

    const lbl = document.createElement('span');
    lbl.className = 'variant-label';
    lbl.textContent = `Variant ${i + 1}`;

    const expandBtn = document.createElement('button');
    expandBtn.className = 'variant-expand';
    expandBtn.textContent = 'expand';
    expandBtn.type = 'button';

    const preview = document.createElement('div');
    preview.className = 'variant-preview md-content collapsed';
    preview.innerHTML = renderMarkdown(text);

    expandBtn.addEventListener('click', e => {
      e.stopPropagation();
      const collapsed = preview.classList.toggle('collapsed');
      expandBtn.textContent = collapsed ? 'expand' : 'collapse';
    });

    card.addEventListener('click', () => {
      radio.checked = true;
      radio.dispatchEvent(new Event('change'));
    });

    header.appendChild(radio);
    header.appendChild(lbl);
    header.appendChild(expandBtn);
    card.appendChild(header);
    card.appendChild(preview);

    // Version history
    const history = state.variantHistory[i];
    if (history && history.length > 1) {
      const historyWrap = document.createElement('div');
      historyWrap.className = 'version-history';

      const toggleBtn = document.createElement('button');
      toggleBtn.className = 'version-toggle';
      toggleBtn.type = 'button';
      toggleBtn.textContent = `${history.length} versions`;

      const list = document.createElement('div');
      list.className = 'version-list';
      list.style.display = 'none';

      toggleBtn.addEventListener('click', e => {
        e.stopPropagation();
        const hidden = list.style.display === 'none';
        list.style.display = hidden ? 'block' : 'none';
        toggleBtn.textContent = hidden
          ? `${history.length} versions (hide)`
          : `${history.length} versions`;
      });

      history.forEach((entry, hi) => {
        const item = document.createElement('div');
        item.className = 'version-item';

        const label = document.createElement('span');
        label.className = 'version-label';
        label.textContent = `v${hi + 1}: ${entry.label}`;

        const time = document.createElement('span');
        time.className = 'version-time';
        time.textContent = timeAgo(entry.timestamp);

        item.appendChild(label);
        item.appendChild(time);

        if (hi === history.length - 1) {
          const tag = document.createElement('span');
          tag.className = 'version-current-tag';
          tag.textContent = 'current';
          item.appendChild(tag);
        } else {
          const restoreBtn = document.createElement('button');
          restoreBtn.className = 'version-revert';
          restoreBtn.type = 'button';
          restoreBtn.textContent = 'restore';
          restoreBtn.addEventListener('click', ev => {
            ev.stopPropagation();
            state.revertVariant(i, hi);
            renderVariantsInModal();
          });
          item.appendChild(restoreBtn);
        }

        list.appendChild(item);
      });

      historyWrap.appendChild(toggleBtn);
      historyWrap.appendChild(list);
      card.appendChild(historyWrap);
    }

    container.appendChild(card);
  });
}

export function initGenerateModal() {
  document.getElementById('modal-generate-close').addEventListener('click', closeGenerateModal);
  document.getElementById('modal-generate-cancel').addEventListener('click', closeGenerateModal);
  document.getElementById('modal-generate').addEventListener('click', e => {
    if (e.target === document.getElementById('modal-generate')) closeGenerateModal();
  });

  document.getElementById('btn-generate').addEventListener('click', async () => {
    const form = document.getElementById('form-drafts');
    const fd = new FormData(form);
    const topicId = parseInt(fd.get('topic_id'), 10);
    const styleId = parseInt(fd.get('style_id'), 10);
    if (!topicId || !styleId) {
      showNotice('drafts', 'Please select a topic and style before generating.');
      return;
    }
    const sourceIds = fd.getAll('source_ids').map(v => parseInt(v, 10)).filter(Boolean);
    const notes = fd.get('notes') || '';

    const btn = document.getElementById('btn-generate');
    btn.disabled = true;
    btn.textContent = 'Generating…';
    clearNotice('drafts');

    try {
      const data = await apiFetch('/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ topic_id: topicId, style_id: styleId, source_ids: sourceIds, notes })
      });
      state.setVariantTexts(data.variants || ['', '', '']);
      state.setSelectedVariantIdx(0);

      // Initialize version history
      state.resetVariantHistory();
      state.variantTexts.forEach((text, idx) => {
        state.pushVariantVersion(idx, text, 'Original');
      });

      openGenerateModal();
    } catch (err) {
      showNotice('drafts', 'Generation failed: ' + err.message);
    } finally {
      btn.disabled = false;
      btn.textContent = 'Generate 3 Drafts';
    }
  });

  document.getElementById('modal-btn-use-variant').addEventListener('click', () => {
    document.getElementById('draft-content').value = state.variantTexts[state.selectedVariantIdx];
    closeGenerateModal();
  });

  document.getElementById('modal-btn-tweak').addEventListener('click', async () => {
    const instruction = document.getElementById('modal-tweak-instruction').value.trim();
    if (!instruction) return;
    const content = state.variantTexts[state.selectedVariantIdx];
    if (!content) return;

    const btn = document.getElementById('modal-btn-tweak');
    btn.disabled = true;
    btn.textContent = 'Tweaking…';

    try {
      const data = await apiFetch('/tweak', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content, instruction })
      });
      state.variantTexts[state.selectedVariantIdx] = data.content;

      // Push to version history
      const label = instruction.length > 40 ? instruction.slice(0, 40) + '…' : instruction;
      state.pushVariantVersion(state.selectedVariantIdx, data.content, label);

      renderVariantsInModal();
      document.getElementById('modal-tweak-instruction').value = '';
    } catch (err) {
      showNotice('drafts', 'Tweak failed: ' + err.message);
    } finally {
      btn.disabled = false;
      btn.textContent = 'Apply Tweak';
    }
  });
}
