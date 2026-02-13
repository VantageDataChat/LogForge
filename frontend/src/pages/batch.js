// batch.js — 批量处理页面

App.registerPage('batch', function(container) {
    container.innerHTML = `
        <h2 class="page-header">批量处理</h2>
        <p class="page-desc">使用已生成的 Python 程序批量处理日志文件，输出 Excel 格式</p>
        <div class="card">
            <div class="card-title">
                <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" stroke-linecap="round" stroke-linejoin="round"/></svg>
                处理配置
            </div>
            <div class="form-group">
                <label for="batch-project">选择项目</label>
                <select id="batch-project" class="form-select">
                    <option value="">加载中...</option>
                </select>
            </div>
            <div class="form-group">
                <label for="batch-input-dir">输入目录</label>
                <div class="input-with-btn">
                    <input type="text" id="batch-input-dir" placeholder="日志文件所在目录路径" readonly>
                    <button class="btn btn-default btn-sm" id="browse-input-btn">浏览...</button>
                </div>
            </div>
            <div class="form-group">
                <label for="batch-output-dir">输出目录</label>
                <div class="input-with-btn">
                    <input type="text" id="batch-output-dir" placeholder="Excel 输出目录路径" readonly>
                    <button class="btn btn-default btn-sm" id="browse-output-btn">浏览...</button>
                </div>
            </div>
            <button id="batch-start-btn" class="btn btn-primary">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                开始处理
            </button>
        </div>
        <div id="batch-progress-section" style="display:none;">
            <div class="card">
                <div class="flex-between mb-8">
                    <div class="card-title" style="margin-bottom:0;">
                        <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                        处理进度
                    </div>
                    <span id="batch-status-badge"></span>
                </div>
                <div class="progress-bar-container">
                    <div class="progress-bar-fill" id="batch-progress-bar" style="width:0%"></div>
                </div>
                <div class="flex-between text-xs text-muted mt-8">
                    <span id="batch-progress-text">0%</span>
                    <span id="batch-file-info"></span>
                </div>
                <div class="section-divider"></div>
                <div class="text-xs text-muted mb-8" style="font-weight:600;text-transform:uppercase;letter-spacing:0.5px;">运行日志</div>
                <div class="log-area" id="batch-log"></div>
            </div>
        </div>
        <div id="batch-result-section" style="display:none;">
            <div class="card">
                <div class="card-title">
                    <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    处理结果
                </div>
                <div id="batch-result-content"></div>
            </div>
        </div>
    `;

    const startBtn = document.getElementById('batch-start-btn');
    const projectSelect = document.getElementById('batch-project');
    const inputDirInput = document.getElementById('batch-input-dir');
    const outputDirInput = document.getElementById('batch-output-dir');
    const progressSection = document.getElementById('batch-progress-section');
    const resultSection = document.getElementById('batch-result-section');
    const progressBar = document.getElementById('batch-progress-bar');
    const progressText = document.getElementById('batch-progress-text');
    const fileInfo = document.getElementById('batch-file-info');
    const statusBadge = document.getElementById('batch-status-badge');
    const logArea = document.getElementById('batch-log');
    const resultContent = document.getElementById('batch-result-content');

    // Load projects into dropdown
    (async () => {
        try {
            const projects = await window.go.main.App.ListProjects();
            projectSelect.innerHTML = '<option value="">-- 请选择项目 --</option>';
            if (projects && projects.length > 0) {
                projects.forEach(p => {
                    const label = p.name || p.id.substring(0, 8);
                    projectSelect.innerHTML += '<option value="' + p.id + '">' + escapeHtml(label) + '</option>';
                });
            }
        } catch (_) {
            projectSelect.innerHTML = '<option value="">加载项目失败</option>';
        }
    })();

    // Load default dirs
    (async () => {
        try {
            const settings = await window.go.main.App.GetSettings();
            if (settings.default_input_dir) inputDirInput.value = settings.default_input_dir;
            if (settings.default_output_dir) outputDirInput.value = settings.default_output_dir;
        } catch (_) {}
    })();

    // Directory browse buttons
    document.getElementById('browse-input-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory('选择输入目录');
            if (dir) inputDirInput.value = dir;
        } catch (_) {}
    });

    document.getElementById('browse-output-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory('选择输出目录');
            if (dir) outputDirInput.value = dir;
        } catch (_) {}
    });

    let pollTimer = null;
    let lastLogMessage = '';

    // Clean up polling timer when navigating away from this page
    const originalNavigate = App.navigate.bind(App);
    const cleanupPolling = () => {
        if (pollTimer) {
            clearInterval(pollTimer);
            pollTimer = null;
        }
    };
    // Override navigate temporarily to clean up on page change
    App.navigate = function(page) {
        cleanupPolling();
        App.navigate = originalNavigate;
        originalNavigate(page);
    };

    startBtn.addEventListener('click', async () => {
        const projectId = projectSelect.value;
        const inputDir = inputDirInput.value.trim();
        const outputDir = outputDirInput.value.trim();

        if (!projectId) { alert('请选择项目'); return; }
        if (!inputDir) { alert('请选择输入目录'); return; }
        if (!outputDir) { alert('请选择输出目录'); return; }

        startBtn.disabled = true;
        progressSection.style.display = 'block';
        resultSection.style.display = 'none';
        logArea.textContent = '';
        progressBar.style.width = '0%';
        progressText.textContent = '0%';

        try {
            await window.go.main.App.RunBatch(projectId, inputDir, outputDir);
            appendLog('批处理已启动...');
            startPolling();
        } catch (err) {
            appendLog('启动失败: ' + err);
            startBtn.disabled = false;
        }
    });

    function startPolling() {
        if (pollTimer) clearInterval(pollTimer);
        pollTimer = setInterval(async () => {
            try {
                const p = await window.go.main.App.GetBatchProgress();
                updateProgress(p);
                if (p.status === 'completed' || p.status === 'failed') {
                    clearInterval(pollTimer);
                    pollTimer = null;
                    startBtn.disabled = false;
                    showResult(p);
                }
            } catch (err) {
                appendLog('获取进度失败: ' + err);
            }
        }, 1000);
    }

    function updateProgress(p) {
        const pct = Math.round((p.progress || 0) * 100);
        progressBar.style.width = pct + '%';
        progressText.textContent = pct + '%';

        if (p.current_file) {
            fileInfo.textContent = '当前: ' + p.current_file;
        }

        const statusMap = {
            'running': ['处理中', 'badge badge-info'],
            'completed': ['已完成', 'badge badge-success'],
            'failed': ['失败', 'badge badge-error'],
            'fixing': ['修复中', 'badge badge-warning'],
            'idle': ['空闲', 'badge badge-info']
        };
        const [label, cls] = statusMap[p.status] || [p.status, 'badge badge-info'];
        statusBadge.innerHTML = '<span class="' + cls + '">' + label + '</span>';

        if (p.message && p.message !== lastLogMessage) {
            appendLog(p.message);
            lastLogMessage = p.message;
        }
    }

    function showResult(p) {
        resultSection.style.display = 'block';
        const total = p.total_files || 0;
        const processed = p.processed || 0;
        const failed = p.failed || 0;
        const succeeded = processed - failed;

        let html = '<div class="stat-row">';
        html += '<div class="stat-card"><div class="stat-value">' + total + '</div><div class="stat-label">总文件数</div></div>';
        html += '<div class="stat-card"><div class="stat-value success">' + succeeded + '</div><div class="stat-label">成功</div></div>';
        html += '<div class="stat-card"><div class="stat-value danger">' + failed + '</div><div class="stat-label">失败</div></div>';
        html += '</div>';

        if (p.status === 'completed') {
            html += '<div class="alert alert-success">批量处理完成</div>';
        } else if (p.status === 'failed') {
            html += '<div class="alert alert-error">批量处理失败' + (p.message ? ': ' + escapeHtml(p.message) : '') + '</div>';
        }

        resultContent.innerHTML = html;
    }

    function appendLog(msg) {
        const time = new Date().toLocaleTimeString();
        logArea.textContent += '[' + time + '] ' + msg + '\n';
        logArea.scrollTop = logArea.scrollHeight;
    }
});
