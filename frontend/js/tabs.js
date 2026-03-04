export function initTabs(tabLoaders) {
  document.getElementById('nav').addEventListener('click', e => {
    const btn = e.target.closest('button[data-tab]');
    if (!btn) return;
    document.querySelectorAll('nav button').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-panel').forEach(p => { p.classList.remove('active'); p.style.display = ''; });
    btn.classList.add('active');
    const panel = document.getElementById(`tab-${btn.dataset.tab}`);
    panel.classList.add('active');
    tabLoaders[btn.dataset.tab]();
  });
}
