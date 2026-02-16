// sample.js — 样本分析页面

App.registerPage('sample', function(container) {
    container.innerHTML = `
        <h2 class="page-header">样本分析</h2>
        <p class="page-desc">粘贴少量日志样本，或浏览日志文件取前几行作为样本，AI 将自动分析格式并生成 Python 处理程序</p>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                样本日志输入
            </div>
            <div class="form-group">
                <label for="project-name">项目名称</label>
                <input type="text" id="project-name" placeholder="为本次分析命名，例如: nginx访问日志">
            </div>
            <div class="form-group">
                <label for="sample-input">粘贴几条样本日志条目，或点击下方按钮从日志文件中提取</label>
                <textarea id="sample-input" rows="10" placeholder="在此粘贴样本日志内容...&#10;&#10;例如:&#10;2024-01-15 10:23:45 INFO [nginx] 192.168.1.100 GET /api/users 200 0.032s"></textarea>
            </div>
            <div class="btn-group">
                <button id="browse-log-btn" class="btn btn-default">
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    浏览日志文件
                </button>
                <button id="analyze-btn" class="btn btn-primary">
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 10V3L4 14h7v7l9-11h-7z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    开始分析
                </button>
            </div>
        </div>
        <div id="sample-loading" style="display:none;">
            <div class="card">
                <div class="flex-center gap-12">
                    <span class="spinner"></span>
                    <span class="text-secondary">正在分析样本并生成代码，请稍候...</span>
                </div>
            </div>
        </div>
        <div id="sample-result" style="display:none;">
            <div class="card">
                <div class="flex-between mb-8">
                    <div class="card-title" style="margin-bottom:0;">
                        <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" stroke-linecap="round" stroke-linejoin="round"/></svg>
                        生成的 Python 代码
                    </div>
                    <div id="validation-status"></div>
                </div>
                <div id="sample-errors"></div>
                <pre class="code-block"><code id="generated-code"></code></pre>
                <div class="mt-12 text-xs text-muted" id="project-id-display"></div>
            </div>
        </div>
    `;

    const analyzeBtn = document.getElementById('analyze-btn');
    const browseLogBtn = document.getElementById('browse-log-btn');
    const projectNameInput = document.getElementById('project-name');
    const sampleInput = document.getElementById('sample-input');
    const resultDiv = document.getElementById('sample-result');
    const loadingDiv = document.getElementById('sample-loading');
    const codeEl = document.getElementById('generated-code');
    const statusEl = document.getElementById('validation-status');
    const errorsEl = document.getElementById('sample-errors');
    const projectIdEl = document.getElementById('project-id-display');

    // Browse log file — read first N lines as sample, auto-fill project name
    browseLogBtn.addEventListener('click', async () => {
        try {
            const result = await window.go.main.App.BrowseLogFile();
            if (!result) return; // user cancelled
            sampleInput.value = result.sample_text;
            if (!projectNameInput.value.trim()) {
                projectNameInput.value = result.project_name;
            }
        } catch (err) {
            showError('浏览日志文件失败: ' + String(err));
        }
    });

    analyzeBtn.addEventListener('click', async () => {
        const name = projectNameInput.value.trim();
        if (!name) {
            showAlert('请输入项目名称');
            return;
        }

        const text = sampleInput.value.trim();
        if (!text) {
            showAlert('请输入样本日志内容');
            return;
        }

        analyzeBtn.disabled = true;
        resultDiv.style.display = 'none';
        loadingDiv.style.display = 'block';
        errorsEl.innerHTML = '';

        try {
            const result = await window.go.main.App.AnalyzeSample(name, text);
            loadingDiv.style.display = 'none';
            resultDiv.style.display = 'block';

            codeEl.textContent = result.code;

            if (result.valid) {
                statusEl.innerHTML = '<span class="badge badge-success">已验证</span>';
            } else {
                statusEl.innerHTML = '<span class="badge badge-warning">未验证</span>';
            }

            if (result.errors && result.errors.length > 0) {
                errorsEl.innerHTML = '<div class="alert alert-error mb-8">' +
                    result.errors.map(e => escapeHtml(e)).join('<br>') + '</div>';
            }

            projectIdEl.textContent = '项目名称: ' + name;
        } catch (err) {
            loadingDiv.style.display = 'none';
            resultDiv.style.display = 'block';
            codeEl.textContent = '';
            statusEl.innerHTML = '';
            errorsEl.innerHTML = '<div class="alert alert-error">' + escapeHtml(String(err)) + '</div>';
            projectIdEl.textContent = '';
        } finally {
            analyzeBtn.disabled = false;
        }
    });
});
