// app.js — SPA router and navigation logic

function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

// ---- Custom Dialog System ----

function _showDialog(type, message, options = {}) {
    return new Promise((resolve) => {
        const icons = {
            warning: '⚠️',
            error: '❌',
            info: 'ℹ️',
            confirm: '❓',
        };
        const titleKeys = {
            warning: 'dialog.warning',
            error: 'dialog.error',
            info: 'dialog.info',
            confirm: 'dialog.confirm',
        };
        const title = options.title || I18n.t(titleKeys[type] || 'dialog.info');
        const icon = icons[type] || 'ℹ️';
        const isConfirm = type === 'confirm';

        const overlay = document.createElement('div');
        overlay.className = 'dialog-overlay';
        overlay.innerHTML = `
            <div class="dialog-box">
                <div class="dialog-body">
                    <div class="dialog-icon dialog-icon-${type}">${icon}</div>
                    <div class="dialog-content">
                        <div class="dialog-title">${escapeHtml(title)}</div>
                        <div class="dialog-message">${escapeHtml(message)}</div>
                    </div>
                </div>
                <div class="dialog-footer">
                    ${isConfirm ? '<button class="btn btn-default btn-sm dialog-cancel-btn">' + I18n.t('common.cancel') + '</button>' : ''}
                    <button class="btn btn-primary btn-sm dialog-ok-btn">${isConfirm ? I18n.t('common.confirm') : I18n.t('common.ok')}</button>
                </div>
            </div>
        `;
        document.body.appendChild(overlay);

        function close(result) {
            overlay.classList.add('dialog-fade-out');
            setTimeout(() => overlay.remove(), 200);
            resolve(result);
        }

        overlay.querySelector('.dialog-ok-btn').addEventListener('click', () => close(true));
        const cancelBtn = overlay.querySelector('.dialog-cancel-btn');
        if (cancelBtn) cancelBtn.addEventListener('click', () => close(false));

        // Click backdrop to dismiss (cancel for confirm, ok for alert)
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) close(isConfirm ? false : true);
        });

        // Focus the OK button for keyboard accessibility
        overlay.querySelector('.dialog-ok-btn').focus();
    });
}

function showAlert(message, options) {
    return _showDialog('warning', message, options);
}

function showError(message, options) {
    return _showDialog('error', message, options);
}

function showConfirm(message, options) {
    return _showDialog('confirm', message, options);
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

    navigate(page, params) {
        if (!this.pages[page]) {
            page = 'sample';
        }

        // Only update pageParams if params is explicitly provided,
        // so hashchange-triggered navigations don't clear pending params.
        if (params !== undefined) {
            this.pageParams = params;
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
                    <h2 data-i18n="wizard.welcome">${I18n.t('wizard.welcome')}</h2>
                    <p class="text-muted" data-i18n="wizard.subtitle">${I18n.t('wizard.subtitle')}</p>
                </div>
                <div class="wizard-steps">
                    <div class="wizard-step">
                        <span class="wizard-step-num">1</span>
                        <div>
                            <div class="wizard-step-title" data-i18n="wizard.step1_title">${I18n.t('wizard.step1_title')}</div>
                            <div class="wizard-step-desc" data-i18n="wizard.step1_desc">${I18n.t('wizard.step1_desc')}</div>
                        </div>
                    </div>
                    <div class="wizard-step">
                        <span class="wizard-step-num">2</span>
                        <div>
                            <div class="wizard-step-title" data-i18n="wizard.step2_title">${I18n.t('wizard.step2_title')}</div>
                            <div class="wizard-step-desc" data-i18n="wizard.step2_desc">${I18n.t('wizard.step2_desc')}</div>
                        </div>
                    </div>
                    <div class="wizard-step">
                        <span class="wizard-step-num">3</span>
                        <div>
                            <div class="wizard-step-title" data-i18n="wizard.step3_title">${I18n.t('wizard.step3_title')}</div>
                            <div class="wizard-step-desc" data-i18n="wizard.step3_desc">${I18n.t('wizard.step3_desc')}</div>
                        </div>
                    </div>
                    <div class="wizard-step">
                        <span class="wizard-step-num">4</span>
                        <div>
                            <div class="wizard-step-title" data-i18n="wizard.step4_title">${I18n.t('wizard.step4_title')}</div>
                            <div class="wizard-step-desc" data-i18n="wizard.step4_desc">${I18n.t('wizard.step4_desc')}</div>
                        </div>
                    </div>
                </div>
                <div class="wizard-footer">
                    <label class="wizard-checkbox">
                        <input type="checkbox" id="wizard-dont-show">
                        <span data-i18n="wizard.dont_show">${I18n.t('wizard.dont_show')}</span>
                    </label>
                    <button class="btn btn-primary" id="wizard-close-btn" data-i18n="wizard.start">${I18n.t('wizard.start')}</button>
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
        this.updateEnvIndicator('pending', I18n.t('env.initializing'));

        const banner = document.createElement('div');
        banner.id = 'pyenv-banner';
        banner.className = 'alert alert-info';
        banner.innerHTML = '<span class="spinner"></span> ' + I18n.t('env.init_banner');
        document.body.appendChild(banner);

        const maxAttempts = 60;
        for (let i = 0; i < maxAttempts; i++) {
            try {
                const status = await window.go.main.App.GetPythonEnvReady();
                if (status.ready) {
                    banner.className = 'alert alert-success';
                    banner.innerHTML = I18n.t('env.init_success');
                    this.updateEnvIndicator('ready', I18n.t('env.ready'));
                    setTimeout(() => banner.remove(), 3000);
                    return;
                }
                if (status.error) {
                    banner.className = 'alert alert-error';
                    banner.innerHTML = I18n.t('env.init_failed') + ': ' + escapeHtml(status.error);
                    this.updateEnvIndicator('error', I18n.t('env.error'));
                    setTimeout(() => banner.remove(), 8000);
                    return;
                }
            } catch (_) { /* ignore */ }
            await new Promise(r => setTimeout(r, 1000));
        }
        banner.className = 'alert alert-warning';
        banner.innerHTML = I18n.t('env.init_timeout');
        this.updateEnvIndicator('error', I18n.t('env.timeout'));
        setTimeout(() => banner.remove(), 8000);
    }
};
