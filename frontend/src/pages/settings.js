// settings.js — 设置页面

App.registerPage('settings', function(container) {
    const isSetupMode = !App.llmConfigured;

    container.innerHTML = `
        ${isSetupMode ? `
        <div class="setup-banner">
            <span class="setup-banner-icon">🔧</span>
            <div class="setup-banner-text">首次使用，请先配置 LLM 参数并测试连接通过后才能使用其他功能。</div>
        </div>` : ''}
        <h2 class="page-header">设置</h2>
        <p class="page-desc">配置 LLM 连接和默认目录</p>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                LLM 配置
            </div>
            <div class="form-group">
                <label for="llm-base-url">Base URL</label>
                <input type="text" id="llm-base-url" placeholder="例如: https://api.deepseek.com/v1">
            </div>
            <div class="form-group">
                <label for="llm-api-key">API Key</label>
                <input type="password" id="llm-api-key" placeholder="输入 API Key">
            </div>
            <div class="form-group">
                <label for="llm-model">Model Name</label>
                <input type="text" id="llm-model" placeholder="例如: deepseek-chat">
            </div>
            <div class="btn-group">
                <button class="btn btn-primary" id="test-llm-btn">测试连接</button>
                <button class="btn btn-default" id="save-settings-btn">保存设置</button>
            </div>
            <div id="llm-test-result" class="mt-12"></div>
        </div>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                默认目录
            </div>
            <div class="form-group">
                <label for="default-input-dir">默认输入目录</label>
                <div class="input-with-btn">
                    <input type="text" id="default-input-dir" placeholder="日志文件默认目录" readonly>
                    <button class="btn btn-default btn-sm" id="browse-default-input-btn">浏览...</button>
                </div>
            </div>
            <div class="form-group">
                <label for="default-output-dir">默认输出目录</label>
                <div class="input-with-btn">
                    <input type="text" id="default-output-dir" placeholder="Excel 输出默认目录" readonly>
                    <button class="btn btn-default btn-sm" id="browse-default-output-btn">浏览...</button>
                </div>
            </div>
        </div>
        <div id="settings-message" class="mt-12"></div>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                其他
            </div>
            <div class="form-group">
                <label for="sample-lines">采样条数（浏览日志文件时取前几行作为样本）</label>
                <input type="number" id="sample-lines" min="1" max="1000" placeholder="默认 5">
            </div>
            <label class="wizard-checkbox" style="margin-bottom:0">
                <input type="checkbox" id="show-wizard-toggle">
                <span>启动时显示使用向导</span>
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
        } catch (err) {
            msgEl.innerHTML = '<div class="alert alert-error">加载设置失败: ' + escapeHtml(String(err)) + '</div>';
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
        };
    }

    // Directory browse buttons
    document.getElementById('browse-default-input-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory('选择默认输入目录');
            if (dir) fields.inputDir.value = dir;
        } catch (_) {}
    });

    document.getElementById('browse-default-output-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory('选择默认输出目录');
            if (dir) fields.outputDir.value = dir;
        } catch (_) {}
    });

    // Save settings
    document.getElementById('save-settings-btn').addEventListener('click', async () => {
        try {
            await window.go.main.App.SaveSettings(gatherSettings());
            msgEl.innerHTML = '<div class="alert alert-success">设置已保存</div>';
            setTimeout(() => { msgEl.innerHTML = ''; }, 3000);
        } catch (err) {
            msgEl.innerHTML = '<div class="alert alert-error">保存失败: ' + escapeHtml(String(err)) + '</div>';
        }
    });

    // Test LLM connection
    document.getElementById('test-llm-btn').addEventListener('click', async () => {
        const settings = gatherSettings();
        if (!settings.llm.base_url || !settings.llm.api_key || !settings.llm.model_name) {
            testResultEl.innerHTML = '<div class="alert alert-error">请先填写完整的 LLM 配置</div>';
            return;
        }

        const testBtn = document.getElementById('test-llm-btn');
        testBtn.disabled = true;
        testBtn.innerHTML = '<span class="spinner"></span> 测试中...';
        testResultEl.innerHTML = '<div class="alert alert-info">正在保存设置并测试连接...</div>';

        try {
            await window.go.main.App.SaveSettings(settings);
            await window.go.main.App.TestLLM();
            testResultEl.innerHTML = '<div class="alert alert-success">✅ LLM 连接测试通过</div>';
            App.onLLMConfigured();
        } catch (err) {
            testResultEl.innerHTML = '<div class="alert alert-error">❌ 测试失败: ' + escapeHtml(String(err)) + '</div>';
        } finally {
            testBtn.disabled = false;
            testBtn.textContent = '测试连接';
        }
    });
});
