export const API = '';

export let topicsCache = [];
export let stylesCache = [];
export let sourcesCache = [];

export let variantTexts = ['', '', ''];
export let selectedVariantIdx = 0;

export function setTopicsCache(v) { topicsCache = v; }
export function setStylesCache(v) { stylesCache = v; }
export function setSourcesCache(v) { sourcesCache = v; }
export function setVariantTexts(v) { variantTexts = v; }
export function setSelectedVariantIdx(v) { selectedVariantIdx = v; }
