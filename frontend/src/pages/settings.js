// settings.js — 设置页面 (with i18n support)

App.registerPage('settings', function(container) {
    const isSetupMode = !App.llmConfigured;

    container.innerHTML = `
        ${isSetupMode ? `
        <div class="setup-banner">
            <span class="setup-banner-icon">🔧</span>
            <div class="setup-banner-text">${I18n.t('settings.setup_banner')}</div>
        </div>` : ''}
        <h2 class="page-header">${I18n.t('settings.title')}</h2>
        <p class="page-desc">${I18n.t('settings.desc')}</p>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                ${I18n.t('settings.llm_config')}
            </div>
            <div class="form-group">
                <label for="llm-base-url">${I18n.t('settings.base_url')}</label>
                <input type="text" id="llm-base-url" placeholder="${I18n.t('settings.base_url_placeholder')}">
            </div>
            <div class="form-group">
                <label for="llm-api-key">${I18n.t('settings.api_key')}</label>
                <input type="password" id="llm-api-key" placeholder="${I18n.t('settings.api_key_placeholder')}">
            </div>
            <div class="form-group">
                <label for="llm-model">${I18n.t('settings.model')}</label>
                <input type="text" id="llm-model" placeholder="${I18n.t('settings.model_placeholder')}">
            </div>
            <div class="btn-group">
                <button class="btn btn-primary" id="test-llm-btn">${I18n.t('settings.test_connection')}</button>
                <button class="btn btn-default" id="save-settings-btn">${I18n.t('settings.save_settings')}</button>
            </div>
            <div id="llm-test-result" class="mt-12"></div>
        </div>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                ${I18n.t('settings.default_dirs')}
            </div>
            <div class="form-group">
                <label for="default-input-dir">${I18n.t('settings.default_input_dir')}</label>
                <div class="input-with-btn">
                    <input type="text" id="default-input-dir" placeholder="${I18n.t('settings.default_input_placeholder')}" readonly>
                    <button class="btn btn-default btn-sm" id="browse-default-input-btn">${I18n.t('common.browse')}</button>
                </div>
            </div>
            <div class="form-group">
                <label for="default-output-dir">${I18n.t('settings.default_output_dir')}</label>
                <div class="input-with-btn">
                    <input type="text" id="default-output-dir" placeholder="${I18n.t('settings.default_output_placeholder')}" readonly>
                    <button class="btn btn-default btn-sm" id="browse-default-output-btn">${I18n.t('common.browse')}</button>
                </div>
            </div>
        </div>
        <div id="settings-message" class="mt-12"></div>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                ${I18n.t('settings.other')}
            </div>
            <div class="form-group">
                <label for="sample-lines">${I18n.t('settings.sample_lines')}</label>
                <input type="number" id="sample-lines" min="1" max="1000" placeholder="${I18n.t('settings.sample_lines_placeholder')}">
            </div>
            <div class="form-group">
                <label for="language-select">${I18n.t('settings.language')}</label>
                <select id="language-select" class="form-select">
                    <option value="zh-CN">简体中文</option>
                    <option value="en">English</option>
                </select>
            </div>
            <label class="wizard-checkbox" style="margin-bottom:0">
                <input type="checkbox" id="show-wizard-toggle">
                <span>${I18n.t('settings.show_wizard')}</span>
            </label>
        </div>
    `;

    const fields = {
        baseUrl: document.getElementById('llm-base-url'),
        apiKey: document.getElementById('llm-api-key'),
        model: document.getElementById('llm-model'),
        inputDir: document.getElementById('default-input-dir'),
        outputDir: document.getElementById('default-output-dir'),
        sampleLines: document.getElementById('sample-lines'),
        language: document.getElementById('language-select'),
    };
    const msgEl = document.getElementById('settings-message');
    const testResultEl = document.getElementById('llm-test-result');
    const wizardToggle = document.getElementById('show-wizard-toggle');

    // Cache loaded settings so we can preserve fields not shown in the UI (e.g. uv_path)
    let loadedSettings = null;

    // Load current settings
    (async () => {
        try {
            const s = await window.go.main.App.GetSettings();
            loadedSettings = s;
            fields.baseUrl.value = s.llm.base_url || '';
            fields.apiKey.value = s.llm.api_key || '';
            fields.model.value = s.llm.model_name || '';
            fields.inputDir.value = s.default_input_dir || '';
            fields.outputDir.value = s.default_output_dir || '';
            fields.sampleLines.value = s.sample_lines || 5;
            fields.language.value = s.language || I18n.currentLang;
        } catch (err) {
            msgEl.innerHTML = '<div class="alert alert-error">' + I18n.t('settings.load_failed') + ': ' + escapeHtml(String(err)) + '</div>';
        }
        try {
            const show = await window.go.main.App.GetShowWizard();
            wizardToggle.checked = show;
        } catch (_) {
            wizardToggle.checked = true;
        }
    })();

    wizardToggle.addEventListener('change', async () => {
        try {
            await window.go.main.App.SetShowWizard(wizardToggle.checked);
        } catch (_) { /* ignore */ }
    });

    // Language change handler
    fields.language.addEventListener('change', async () => {
        const newLang = fields.language.value;
        I18n.setLanguage(newLang);
        
        // Save language setting
        try {
            const settings = gatherSettings();
            await window.go.main.App.SaveSettings(settings);
        } catch (_) { /* ignore */ }
        
        // Reload current page to apply translations
        App.navigate(App.currentPage);
    });

    function gatherSettings() {
        return {
            llm: {
                base_url: fields.baseUrl.value.trim(),
                api_key: fields.apiKey.value.trim(),
                model_name: fields.model.value.trim(),
            },
            uv_path: (loadedSettings && loadedSettings.uv_path) ? loadedSettings.uv_path : 'uv',
            default_input_dir: fields.inputDir.value.trim(),
            default_output_dir: fields.outputDir.value.trim(),
            sample_lines: parseInt(fields.sampleLines.value, 10) || 5,
            language: fields.language.value,
        };
    }

    // Directory browse buttons
    document.getElementById('browse-default-input-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory(I18n.t('dir.select_default_input'));
            if (dir) fields.inputDir.value = dir;
        } catch (_) {}
    });

    document.getElementById('browse-default-output-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory(I18n.t('dir.select_default_output'));
            if (dir) fields.outputDir.value = dir;
        } catch (_) {}
    });

    // Save settings
    document.getElementById('save-settings-btn').addEventListener('click', async () => {
        try {
            await window.go.main.App.SaveSettings(gatherSettings());
            msgEl.innerHTML = '<div class="alert alert-success">' + I18n.t('settings.saved') + '</div>';
            setTimeout(() => { msgEl.innerHTML = ''; }, 3000);
        } catch (err) {
            msgEl.innerHTML = '<div class="alert alert-error">' + I18n.t('settings.save_failed') + ': ' + escapeHtml(String(err)) + '</div>';
        }
    });

    // Test LLM connection
    document.getElementById('test-llm-btn').addEventListener('click', async () => {
        const settings = gatherSettings();
        if (!settings.llm.base_url || !settings.llm.api_key || !settings.llm.model_name) {
            testResultEl.innerHTML = '<div class="alert alert-error">' + I18n.t('settings.fill_llm_config') + '</div>';
            return;
        }

        const testBtn = document.getElementById('test-llm-btn');
        testBtn.disabled = true;
        testBtn.innerHTML = '<span class="spinner"></span> ' + I18n.t('settings.testing');
        testResultEl.innerHTML = '<div class="alert alert-info">' + I18n.t('settings.test_saving') + '</div>';

        try {
            await window.go.main.App.SaveSettings(settings);
            await window.go.main.App.TestLLM();
            testResultEl.innerHTML = '<div class="alert alert-success">' + I18n.t('settings.test_success') + '</div>';
            App.onLLMConfigured();
        } catch (err) {
            testResultEl.innerHTML = '<div class="alert alert-error">' + I18n.t('settings.test_failed') + ': ' + escapeHtml(String(err)) + '</div>';
        } finally {
            testBtn.disabled = false;
            testBtn.textContent = I18n.t('settings.test_connection');
        }
    });
});
