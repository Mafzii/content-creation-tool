export const API = '';

export let topicsCache = [];
export let stylesCache = [];
export let sourcesCache = [];

export let variantTexts = ['', '', ''];
export let selectedVariantIdx = 0;
export let variantHistory = [[], [], []];

export function setTopicsCache(v) { topicsCache = v; }
export function setStylesCache(v) { stylesCache = v; }
export function setSourcesCache(v) { sourcesCache = v; }
export function setVariantTexts(v) { variantTexts = v; }
export function setSelectedVariantIdx(v) { selectedVariantIdx = v; }
export function resetVariantHistory() { variantHistory = [[], [], []]; }
export function pushVariantVersion(idx, text, label) {
  variantHistory[idx].push({ text, label, timestamp: Date.now() });
}
export function revertVariant(variantIdx, historyIdx) {
  const entry = variantHistory[variantIdx][historyIdx];
  if (!entry) return;
  variantTexts[variantIdx] = entry.text;
  variantHistory[variantIdx] = variantHistory[variantIdx].slice(0, historyIdx + 1);
}
