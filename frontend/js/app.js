import { loadTopics } from './topics.js';
import { loadSources, initSourceForm } from './sources.js';
import { loadStyles } from './styles.js';
import { loadDrafts, initDraftForm } from './drafts.js';
import { loadSettings, initSettingsForm } from './settings.js';
import { initEditModal } from './edit-modal.js';
import { initSimpleForms } from './forms.js';
import { initGenerateModal } from './generate-modal.js';
import { initTabs } from './tabs.js';

// Resolve circular deps: pass loaders to edit modal
initEditModal({
  topics: loadTopics,
  sources: loadSources,
  styles: loadStyles,
  drafts: loadDrafts,
});

// Wire up forms
initSimpleForms(loadTopics, loadStyles);
initSourceForm();
initDraftForm();
initGenerateModal();
initSettingsForm();

// Wire up tabs
initTabs({
  topics: loadTopics,
  sources: loadSources,
  styles: loadStyles,
  drafts: loadDrafts,
  settings: loadSettings,
});

// Init
document.getElementById('tab-settings').style.display = 'none';
document.getElementById('tab-settings').classList.remove('active');
document.getElementById('tab-topics').classList.add('active');

await Promise.all([loadTopics(), loadStyles()]);
await Promise.all([loadSources(), loadDrafts()]);
