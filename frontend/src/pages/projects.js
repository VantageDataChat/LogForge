// projects.js — 项目管理页面

App.registerPage('projects', function(container) {
    container.innerHTML = `
        <h2 class="page-header">项目管理</h2>
        <p class="page-desc">管理已生成的日志处理程序，查看、编辑或重新运行</p>
        <div id="projects-list-section">
            <div class="card" id="projects-card">
                <div class="flex-center gap-8 text-muted">
                    <span class="spinner"></span>
                    <span>加载中...</span>
                </div>
            </div>
        </div>
        <div id="project-detail-section" style="display:none;">
            <div class="card">
                <div class="flex-between mb-16">
                    <div class="card-title" style="margin-bottom:0;">项目详情</div>
                    <button class="btn btn-default btn-sm" id="back-to-list-btn">← 返回列表</button>
                </div>
                <div class="stat-row mb-16">
                    <div class="stat-card" style="text-align:left;padding:12px 16px;">
                        <div class="text-xs text-muted" style="margin-bottom:4px;">项目名称</div>
                        <div class="text-sm" id="detail-name" style="font-weight:600;"></div>
                    </div>
                    <div class="stat-card" style="padding:12px 16px;">
                        <div class="text-xs text-muted" style="margin-bottom:4px;">状态</div>
                        <div id="detail-status"></div>
                    </div>
                    <div class="stat-card" style="text-align:right;padding:12px 16px;">
                        <div class="text-xs text-muted" style="margin-bottom:4px;">创建时间</div>
                        <div class="text-sm" id="detail-created"></div>
                    </div>
                </div>
                <div class="form-group">
                    <label>样本数据</label>
                    <textarea id="detail-sample" rows="4" readonly style="opacity:0.7;"></textarea>
                </div>
                <div class="form-group">
                    <label>Python 代码</label>
                    <textarea id="detail-code" rows="14"></textarea>
                </div>
                <div class="btn-group">
                    <button class="btn btn-primary btn-sm" id="save-code-btn">保存代码</button>
                    <button class="btn btn-default btn-sm" id="rerun-btn">重新运行</button>
                    <button class="btn btn-danger btn-sm" id="delete-btn">删除项目</button>
                </div>
                <div id="detail-message" class="mt-12"></div>
            </div>
            <div id="rerun-section" style="display:none;">
                <div class="card">
                    <div class="card-title">重新运行</div>
                    <div class="form-group">
                        <label for="rerun-input-dir">输入目录</label>
                        <div class="input-with-btn">
                            <input type="text" id="rerun-input-dir" placeholder="日志文件所在目录路径" readonly>
                            <button class="btn btn-default btn-sm" id="browse-rerun-input-btn">浏览...</button>
                        </div>
                    </div>
                    <div class="form-group">
                        <label for="rerun-output-dir">输出目录</label>
                        <div class="input-with-btn">
                            <input type="text" id="rerun-output-dir" placeholder="Excel 输出目录路径" readonly>
                            <button class="btn btn-default btn-sm" id="browse-rerun-output-btn">浏览...</button>
                        </div>
                    </div>
                    <button class="btn btn-primary btn-sm" id="confirm-rerun-btn">确认运行</button>
                </div>
            </div>
        </div>
    `;

    const listSection = document.getElementById('projects-list-section');
    const detailSection = document.getElementById('project-detail-section');
    const projectsCard = document.getElementById('projects-card');
    let currentProjectId = null;

    loadProjects();

    async function loadProjects() {
        try {
            const projects = await window.go.main.App.ListProjects();
            renderProjectList(projects);
        } catch (err) {
            projectsCard.innerHTML = '<div class="alert alert-error">' + escapeHtml(String(err)) + '</div>';
        }
    }

    function renderProjectList(projects) {
        if (!projects || projects.length === 0) {
            projectsCard.innerHTML = `
                <div class="empty-state">
                    <svg class="empty-state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1"><path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <p>暂无项目</p>
                    <p class="text-xs text-muted mt-8">请先在「样本分析」页面生成代码</p>
                </div>`;
            return;
        }

        let html = '<table class="table"><thead><tr>';
        html += '<th>创建时间</th><th>项目名称</th><th>状态</th><th>操作</th>';
        html += '</tr></thead><tbody>';

        projects.forEach(p => {
            const time = new Date(p.created_at).toLocaleString();
            const statusBadge = getStatusBadge(p.status);
            const name = p.name || p.id.substring(0, 8) + '...';
            html += '<tr>';
            html += '<td class="text-sm">' + time + '</td>';
            html += '<td class="text-sm">' + escapeHtml(name) + '</td>';
            html += '<td>' + statusBadge + '</td>';
            html += '<td><button class="btn btn-default btn-sm view-btn" data-id="' + escapeHtml(p.id) + '">查看</button></td>';
            html += '</tr>';
        });

        html += '</tbody></table>';
        projectsCard.innerHTML = html;

        projectsCard.querySelectorAll('.view-btn').forEach(btn => {
            btn.addEventListener('click', () => showDetail(btn.dataset.id));
        });
    }

    async function showDetail(id) {
        try {
            const p = await window.go.main.App.GetProject(id);
            currentProjectId = id;

            document.getElementById('detail-name').textContent = p.name || p.id;
            document.getElementById('detail-status').innerHTML = getStatusBadge(p.status);
            document.getElementById('detail-created').textContent = new Date(p.created_at).toLocaleString();
            document.getElementById('detail-sample').value = p.sample_data || '';
            document.getElementById('detail-code').value = p.code || '';
            document.getElementById('detail-message').innerHTML = '';
            document.getElementById('rerun-section').style.display = 'none';

            listSection.style.display = 'none';
            detailSection.style.display = 'block';
        } catch (err) {
            alert('加载项目失败: ' + err);
        }
    }

    document.getElementById('back-to-list-btn').addEventListener('click', () => {
        detailSection.style.display = 'none';
        listSection.style.display = 'block';
        loadProjects();
    });

    document.getElementById('save-code-btn').addEventListener('click', async () => {
        if (!currentProjectId) return;
        const code = document.getElementById('detail-code').value;
        const msgEl = document.getElementById('detail-message');
        try {
            await window.go.main.App.UpdateProjectCode(currentProjectId, code);
            msgEl.innerHTML = '<div class="alert alert-success">代码已保存</div>';
            setTimeout(() => { msgEl.innerHTML = ''; }, 3000);
        } catch (err) {
            msgEl.innerHTML = '<div class="alert alert-error">' + escapeHtml(String(err)) + '</div>';
        }
    });

    document.getElementById('rerun-btn').addEventListener('click', () => {
        const rerunSection = document.getElementById('rerun-section');
        rerunSection.style.display = rerunSection.style.display === 'none' ? 'block' : 'none';
    });

    // Directory browse for rerun
    document.getElementById('browse-rerun-input-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory('选择输入目录');
            if (dir) document.getElementById('rerun-input-dir').value = dir;
        } catch (_) {}
    });

    document.getElementById('browse-rerun-output-btn').addEventListener('click', async () => {
        try {
            const dir = await window.go.main.App.SelectDirectory('选择输出目录');
            if (dir) document.getElementById('rerun-output-dir').value = dir;
        } catch (_) {}
    });

    document.getElementById('confirm-rerun-btn').addEventListener('click', async () => {
        if (!currentProjectId) return;
        const inputDir = document.getElementById('rerun-input-dir').value.trim();
        const outputDir = document.getElementById('rerun-output-dir').value.trim();
        if (!inputDir || !outputDir) { alert('请选择输入和输出目录'); return; }

        const msgEl = document.getElementById('detail-message');
        try {
            await window.go.main.App.RerunProject(currentProjectId, inputDir, outputDir);
            msgEl.innerHTML = '<div class="alert alert-success">批处理已启动，请前往「批量处理」页面查看进度</div>';
        } catch (err) {
            msgEl.innerHTML = '<div class="alert alert-error">' + escapeHtml(String(err)) + '</div>';
        }
    });

    document.getElementById('delete-btn').addEventListener('click', async () => {
        if (!currentProjectId) return;
        if (!confirm('确定要删除此项目吗？')) return;

        try {
            await window.go.main.App.DeleteProject(currentProjectId);
            currentProjectId = null;
            detailSection.style.display = 'none';
            listSection.style.display = 'block';
            loadProjects();
        } catch (err) {
            alert('删除失败: ' + err);
        }
    });

    function getStatusBadge(status) {
        const map = {
            'draft': ['草稿', 'badge badge-info'],
            'validated': ['已验证', 'badge badge-success'],
            'executed': ['已执行', 'badge badge-success'],
            'failed': ['失败', 'badge badge-error']
        };
        const [label, cls] = map[status] || [status, 'badge badge-info'];
        return '<span class="' + cls + '">' + label + '</span>';
    }
});
