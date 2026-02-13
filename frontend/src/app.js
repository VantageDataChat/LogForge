// app.js — SPA router and navigation logic

function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

const App = {
    pages: {},
    currentPage: null,
    llmConfigured: false,

    registerPage(name, renderFn) {
        this.pages[name] = renderFn;
    },

    async init() {
        // Set up nav item clicks
        document.querySelectorAll('.nav-item').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const page = link.dataset.page;
                if (!this.llmConfigured && page !== 'settings') {
                    return;
                }
                window.location.hash = page;
            });
        });

        window.addEventListener('hashchange', () => this.route());

        // Check if LLM is configured
        try {
            const configured = await window.go.main.App.IsLLMConfigured();
            this.llmConfigured = configured;
        } catch (_) {
            this.llmConfigured = false;
        }

        if (!this.llmConfigured) {
            this.updateNavState();
            window.location.hash = 'settings';
            this.navigate('settings');
        } else {
            this.updateNavState();
            this.route();
            this.pollPythonEnvStatus();
            this.tryShowWizard();
        }
    },

    route() {
        const hash = window.location.hash.replace('#', '') || 'sample';
        if (!this.llmConfigured && hash !== 'settings') {
            window.location.hash = 'settings';
            return;
        }
        this.navigate(hash);
    },

    navigate(page) {
        if (!this.pages[page]) {
            page = 'sample';
        }

        document.querySelectorAll('.nav-item').forEach(link => {
            link.classList.toggle('active', link.dataset.page === page);
        });

        const container = document.getElementById('page-container');
        container.innerHTML = '';
        this.currentPage = page;
        this.pages[page](container);
    },

    onLLMConfigured() {
        this.llmConfigured = true;
        this.updateNavState();
        this.pollPythonEnvStatus();
    },

    updateNavState() {
        document.querySelectorAll('.nav-item').forEach(link => {
            if (link.dataset.page === 'settings') return;
            if (this.llmConfigured) {
                link.classList.remove('nav-disabled');
            } else {
                link.classList.add('nav-disabled');
            }
        });
    },

    updateEnvIndicator(status, label) {
        const dot = document.querySelector('#sidebar-env-status .env-dot');
        const lbl = document.querySelector('#sidebar-env-status .env-label');
        if (dot) {
            dot.className = 'env-dot';
            if (status === 'ready') dot.classList.add('ready');
            else if (status === 'error') dot.classList.add('error');
        }
        if (lbl) lbl.textContent = label;
    },

    async tryShowWizard() {
        try {
            const show = await window.go.main.App.GetShowWizard();
            if (show) this.showWizard();
        } catch (_) { /* ignore */ }
    },

    showWizard() {
        const overlay = document.createElement('div');
        overlay.id = 'wizard-overlay';
        overlay.innerHTML = `
            <div class="wizard-dialog">
                <div class="wizard-header">
                    <span class="wizard-logo">🚀</span>
                    <h2>欢迎使用 LogForge</h2>
                    <p class="text-muted">智能网络日志格式化系统</p>
                </div>
                <div class="wizard-steps">
                    <div class="wizard-step">
                        <span class="wizard-step-num">1</span>
                        <div>
                            <div class="wizard-step-title">样本分析</div>
                            <div class="wizard-step-desc">粘贴一段日志样本，AI 将自动生成 Python 解析代码</div>
                        </div>
                    </div>
                    <div class="wizard-step">
                        <span class="wizard-step-num">2</span>
                        <div>
                            <div class="wizard-step-title">代码验证</div>
                            <div class="wizard-step-desc">系统自动验证生成的代码，确保可以正确运行</div>
                        </div>
                    </div>
                    <div class="wizard-step">
                        <span class="wizard-step-num">3</span>
                        <div>
                            <div class="wizard-step-title">批量处理</div>
                            <div class="wizard-step-desc">选择输入目录，一键批量处理所有日志文件并导出 Excel</div>
                        </div>
                    </div>
                    <div class="wizard-step">
                        <span class="wizard-step-num">4</span>
                        <div>
                            <div class="wizard-step-title">项目管理</div>
                            <div class="wizard-step-desc">历史项目可随时查看、编辑代码或重新执行</div>
                        </div>
                    </div>
                </div>
                <div class="wizard-footer">
                    <label class="wizard-checkbox">
                        <input type="checkbox" id="wizard-dont-show">
                        <span>不再显示此向导</span>
                    </label>
                    <button class="btn btn-primary" id="wizard-close-btn">开始使用</button>
                </div>
            </div>
        `;
        document.body.appendChild(overlay);

        document.getElementById('wizard-close-btn').addEventListener('click', async () => {
            const dontShow = document.getElementById('wizard-dont-show').checked;
            if (dontShow) {
                try { await window.go.main.App.SetShowWizard(false); } catch (_) {}
            }
            overlay.classList.add('wizard-fade-out');
            setTimeout(() => overlay.remove(), 250);
        });

        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) {
                document.getElementById('wizard-close-btn').click();
            }
        });
    },

    async pollPythonEnvStatus() {
        this.updateEnvIndicator('pending', 'Python 环境初始化中...');

        const banner = document.createElement('div');
        banner.id = 'pyenv-banner';
        banner.className = 'alert alert-info';
        banner.innerHTML = '<span class="spinner"></span> 正在自动初始化 Python 环境...';
        document.body.appendChild(banner);

        const maxAttempts = 60;
        for (let i = 0; i < maxAttempts; i++) {
            try {
                const status = await window.go.main.App.GetPythonEnvReady();
                if (status.ready) {
                    banner.className = 'alert alert-success';
                    banner.innerHTML = '✅ Python 环境已就绪';
                    this.updateEnvIndicator('ready', 'Python 环境就绪');
                    setTimeout(() => banner.remove(), 3000);
                    return;
                }
                if (status.error) {
                    banner.className = 'alert alert-error';
                    banner.innerHTML = '❌ 环境初始化失败: ' + escapeHtml(status.error);
                    this.updateEnvIndicator('error', '环境异常');
                    setTimeout(() => banner.remove(), 8000);
                    return;
                }
            } catch (_) { /* ignore */ }
            await new Promise(r => setTimeout(r, 1000));
        }
        banner.className = 'alert alert-warning';
        banner.innerHTML = '⏱ 环境初始化超时，请在设置中手动初始化';
        this.updateEnvIndicator('error', '初始化超时');
        setTimeout(() => banner.remove(), 8000);
    }
};
