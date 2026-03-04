import { apiFetch } from './api.js';
import { showNotice, clearNotice } from './ui.js';
import * as state from './state.js';

function openGenerateModal() {
  document.getElementById('modal-generate').classList.add('open');
  renderVariantsInModal();
}

function closeGenerateModal() {
  document.getElementById('modal-generate').classList.remove('open');
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
    preview.className = 'variant-preview collapsed';
    preview.textContent = text;

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
