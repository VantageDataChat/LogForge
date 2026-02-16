// i18n.js — 多语言支持

const translations = {
    'zh-CN': {
        // 通用
        'app.title': 'LogForge',
        'app.subtitle': '智能日志格式化',
        'common.save': '保存',
        'common.cancel': '取消',
        'common.confirm': '确定',
        'common.delete': '删除',
        'common.browse': '浏览...',
        'common.loading': '加载中...',
        'common.ok': '知道了',
        'common.start': '开始使用',
        
        // 导航
        'nav.sample': '样本分析',
        'nav.batch': '批量处理',
        'nav.projects': '项目管理',
        'nav.settings': '设置',
        
        // 对话框
        'dialog.warning': '提示',
        'dialog.error': '错误',
        'dialog.info': '提示',
        'dialog.confirm': '确认',
        
        // 设置页面
        'settings.title': '设置',
        'settings.desc': '配置 LLM 连接和默认目录',
        'settings.setup_banner': '首次使用，请先配置 LLM 参数并测试连接通过后才能使用其他功能。',
        'settings.llm_config': 'LLM 配置',
        'settings.base_url': 'Base URL',
        'settings.base_url_placeholder': '例如: https://api.deepseek.com/v1',
        'settings.api_key': 'API Key',
        'settings.api_key_placeholder': '输入 API Key',
        'settings.model': 'Model Name',
        'settings.model_placeholder': '例如: deepseek-chat',
        'settings.test_connection': '测试连接',
        'settings.save_settings': '保存设置',
        'settings.default_dirs': '默认目录',
        'settings.default_input_dir': '默认输入目录',
        'settings.default_input_placeholder': '日志文件默认目录',
        'settings.default_output_dir': '默认输出目录',
        'settings.default_output_placeholder': 'Excel 输出默认目录',
        'settings.other': '其他',
        'settings.sample_lines': '采样条数（浏览日志文件时取前几行作为样本）',
        'settings.sample_lines_placeholder': '默认 5',
        'settings.show_wizard': '启动时显示使用向导',
        'settings.language': '界面语言',
        'settings.saved': '设置已保存',
        'settings.save_failed': '保存失败',
        'settings.load_failed': '加载设置失败',
        'settings.test_success': '✅ LLM 连接测试通过',
        'settings.test_failed': '❌ 测试失败',
        'settings.testing': '测试中...',
        'settings.test_saving': '正在保存设置并测试连接...',
        'settings.fill_llm_config': '请先填写完整的 LLM 配置',
        
        // 样本分析页面
        'sample.title': '样本分析',
        'sample.desc': '粘贴少量日志样本，或浏览日志文件取前几行作为样本，AI 将自动分析格式并生成 Python 处理程序',
        'sample.input_title': '样本日志输入',
        'sample.project_name': '项目名称',
        'sample.project_name_placeholder': '为本次分析命名，例如: nginx访问日志',
        'sample.input_label': '粘贴几条样本日志条目，或点击下方按钮从日志文件中提取',
        'sample.input_placeholder': '在此粘贴样本日志内容...\n\n例如:\n2024-01-15 10:23:45 INFO [nginx] 192.168.1.100 GET /api/users 200 0.032s',
        'sample.browse_log': '浏览日志文件',
        'sample.analyze': '开始分析',
        'sample.analyzing': '正在分析样本并生成代码，请稍候...',
        'sample.result_title': '生成的 Python 代码',
        'sample.validated': '已验证',
        'sample.not_validated': '未验证',
        'sample.project_id': '项目名称',
        'sample.enter_name': '请输入项目名称',
        'sample.enter_sample': '请输入样本日志内容',
        'sample.browse_failed': '浏览日志文件失败',
        
        // 批量处理页面
        'batch.title': '批量处理',
        'batch.desc': '使用已生成的 Python 程序批量处理日志文件，输出 Excel 格式',
        'batch.config_title': '处理配置',
        'batch.select_project': '选择项目',
        'batch.select_project_placeholder': '-- 请选择项目 --',
        'batch.input_dir': '输入目录',
        'batch.input_dir_placeholder': '日志文件所在目录路径',
        'batch.output_dir': '输出目录',
        'batch.output_dir_placeholder': 'Excel 输出目录路径',
        'batch.output_name': '输出文件名（不含 .xlsx 后缀）',
        'batch.output_name_placeholder': '默认使用项目名称',
        'batch.start_processing': '开始处理',
        'batch.progress_title': '处理进度',
        'batch.log_title': '运行日志',
        'batch.result_title': '处理结果',
        'batch.total_files': '总文件数',
        'batch.success': '成功',
        'batch.failed': '失败',
        'batch.completed': '批量处理完成',
        'batch.processing_failed': '批量处理失败',
        'batch.open_output': '打开输出目录',
        'batch.open_failed': '打开目录失败',
        'batch.select_project_alert': '请选择项目',
        'batch.select_input_alert': '请选择输入目录',
        'batch.select_output_alert': '请选择输出目录',
        'batch.started': '批处理已启动...',
        'batch.start_failed': '启动失败',
        'batch.progress_failed': '获取进度失败',
        'batch.status.running': '处理中',
        'batch.status.completed': '已完成',
        'batch.status.failed': '失败',
        'batch.status.fixing': '修复中',
        'batch.status.idle': '空闲',
        'batch.current_file': '当前',
        'batch.load_projects_failed': '加载项目失败',
        
        // 项目管理页面
        'projects.title': '项目管理',
        'projects.desc': '管理已生成的日志处理程序，查看、编辑或重新运行',
        'projects.empty': '暂无项目',
        'projects.empty_hint': '请先在「样本分析」页面生成代码',
        'projects.created_at': '创建时间',
        'projects.name': '项目名称',
        'projects.status': '状态',
        'projects.actions': '操作',
        'projects.view': '查看',
        'projects.detail_title': '项目详情',
        'projects.back': '← 返回列表',
        'projects.sample_data': '样本数据',
        'projects.code': 'Python 代码',
        'projects.save_code': '保存代码',
        'projects.rerun': '重新运行',
        'projects.delete_project': '删除项目',
        'projects.code_saved': '代码已保存',
        'projects.rerun_started': '批处理已启动，请前往「批量处理」页面查看进度',
        'projects.delete_confirm': '确定要删除此项目吗？',
        'projects.delete_failed': '删除失败',
        'projects.load_failed': '加载项目失败',
        'projects.status.draft': '草稿',
        'projects.status.validated': '已验证',
        'projects.status.executed': '已执行',
        'projects.status.failed': '失败',
        
        // 环境状态
        'env.checking': '环境检测中',
        'env.initializing': 'Python 环境初始化中...',
        'env.ready': 'Python 环境就绪',
        'env.error': '环境异常',
        'env.timeout': '初始化超时',
        'env.init_banner': '正在自动初始化 Python 环境...',
        'env.init_success': '✅ Python 环境已就绪',
        'env.init_failed': '❌ 环境初始化失败',
        'env.init_timeout': '⏱ 环境初始化超时，请在设置中手动初始化',
        
        // 向导
        'wizard.welcome': '欢迎使用 LogForge',
        'wizard.subtitle': '智能网络日志格式化系统',
        'wizard.step1_title': '样本分析',
        'wizard.step1_desc': '粘贴一段日志样本，AI 将自动生成 Python 解析代码',
        'wizard.step2_title': '代码验证',
        'wizard.step2_desc': '系统自动验证生成的代码，确保可以正确运行',
        'wizard.step3_title': '批量处理',
        'wizard.step3_desc': '选择输入目录，一键批量处理所有日志文件并导出 Excel',
        'wizard.step4_title': '项目管理',
        'wizard.step4_desc': '历史项目可随时查看、编辑代码或重新执行',
        'wizard.dont_show': '不再显示此向导',
        'wizard.start': '开始使用',
        
        // 目录选择
        'dir.select_input': '选择输入目录',
        'dir.select_output': '选择输出目录',
        'dir.select_default_input': '选择默认输入目录',
        'dir.select_default_output': '选择默认输出目录',
    },
    
    'en': {
        // Common
        'app.title': 'LogForge',
        'app.subtitle': 'Smart Log Formatter',
        'common.save': 'Save',
        'common.cancel': 'Cancel',
        'common.confirm': 'OK',
        'common.delete': 'Delete',
        'common.browse': 'Browse...',
        'common.loading': 'Loading...',
        'common.ok': 'Got it',
        'common.start': 'Get Started',
        
        // Navigation
        'nav.sample': 'Sample Analysis',
        'nav.batch': 'Batch Processing',
        'nav.projects': 'Projects',
        'nav.settings': 'Settings',
        
        // Dialogs
        'dialog.warning': 'Warning',
        'dialog.error': 'Error',
        'dialog.info': 'Info',
        'dialog.confirm': 'Confirm',
        
        // Settings page
        'settings.title': 'Settings',
        'settings.desc': 'Configure LLM connection and default directories',
        'settings.setup_banner': 'First time setup: Please configure LLM parameters and test the connection before using other features.',
        'settings.llm_config': 'LLM Configuration',
        'settings.base_url': 'Base URL',
        'settings.base_url_placeholder': 'e.g., https://api.deepseek.com/v1',
        'settings.api_key': 'API Key',
        'settings.api_key_placeholder': 'Enter API Key',
        'settings.model': 'Model Name',
        'settings.model_placeholder': 'e.g., deepseek-chat',
        'settings.test_connection': 'Test Connection',
        'settings.save_settings': 'Save Settings',
        'settings.default_dirs': 'Default Directories',
        'settings.default_input_dir': 'Default Input Directory',
        'settings.default_input_placeholder': 'Default directory for log files',
        'settings.default_output_dir': 'Default Output Directory',
        'settings.default_output_placeholder': 'Default directory for Excel output',
        'settings.other': 'Other',
        'settings.sample_lines': 'Sample Lines (number of lines to read when browsing log files)',
        'settings.sample_lines_placeholder': 'Default: 5',
        'settings.show_wizard': 'Show wizard on startup',
        'settings.language': 'Language',
        'settings.saved': 'Settings saved',
        'settings.save_failed': 'Save failed',
        'settings.load_failed': 'Failed to load settings',
        'settings.test_success': '✅ LLM connection test passed',
        'settings.test_failed': '❌ Test failed',
        'settings.testing': 'Testing...',
        'settings.test_saving': 'Saving settings and testing connection...',
        'settings.fill_llm_config': 'Please fill in complete LLM configuration',
        
        // Sample analysis page
        'sample.title': 'Sample Analysis',
        'sample.desc': 'Paste sample log entries or browse a log file, AI will analyze the format and generate Python code',
        'sample.input_title': 'Sample Log Input',
        'sample.project_name': 'Project Name',
        'sample.project_name_placeholder': 'Name this analysis, e.g., nginx-access-logs',
        'sample.input_label': 'Paste sample log entries, or click the button below to extract from a log file',
        'sample.input_placeholder': 'Paste sample log content here...\n\nExample:\n2024-01-15 10:23:45 INFO [nginx] 192.168.1.100 GET /api/users 200 0.032s',
        'sample.browse_log': 'Browse Log File',
        'sample.analyze': 'Analyze',
        'sample.analyzing': 'Analyzing sample and generating code, please wait...',
        'sample.result_title': 'Generated Python Code',
        'sample.validated': 'Validated',
        'sample.not_validated': 'Not Validated',
        'sample.project_id': 'Project Name',
        'sample.enter_name': 'Please enter project name',
        'sample.enter_sample': 'Please enter sample log content',
        'sample.browse_failed': 'Failed to browse log file',
        
        // Batch processing page
        'batch.title': 'Batch Processing',
        'batch.desc': 'Use generated Python code to batch process log files and output Excel format',
        'batch.config_title': 'Processing Configuration',
        'batch.select_project': 'Select Project',
        'batch.select_project_placeholder': '-- Select a project --',
        'batch.input_dir': 'Input Directory',
        'batch.input_dir_placeholder': 'Directory containing log files',
        'batch.output_dir': 'Output Directory',
        'batch.output_dir_placeholder': 'Directory for Excel output',
        'batch.output_name': 'Output Filename (without .xlsx extension)',
        'batch.output_name_placeholder': 'Defaults to project name',
        'batch.start_processing': 'Start Processing',
        'batch.progress_title': 'Processing Progress',
        'batch.log_title': 'Execution Log',
        'batch.result_title': 'Processing Result',
        'batch.total_files': 'Total Files',
        'batch.success': 'Success',
        'batch.failed': 'Failed',
        'batch.completed': 'Batch processing completed',
        'batch.processing_failed': 'Batch processing failed',
        'batch.open_output': 'Open Output Directory',
        'batch.open_failed': 'Failed to open directory',
        'batch.select_project_alert': 'Please select a project',
        'batch.select_input_alert': 'Please select input directory',
        'batch.select_output_alert': 'Please select output directory',
        'batch.started': 'Batch processing started...',
        'batch.start_failed': 'Failed to start',
        'batch.progress_failed': 'Failed to get progress',
        'batch.status.running': 'Running',
        'batch.status.completed': 'Completed',
        'batch.status.failed': 'Failed',
        'batch.status.fixing': 'Fixing',
        'batch.status.idle': 'Idle',
        'batch.current_file': 'Current',
        'batch.load_projects_failed': 'Failed to load projects',
        
        // Projects page
        'projects.title': 'Project Management',
        'projects.desc': 'Manage generated log processing programs, view, edit or re-run',
        'projects.empty': 'No projects',
        'projects.empty_hint': 'Please generate code in the "Sample Analysis" page first',
        'projects.created_at': 'Created At',
        'projects.name': 'Project Name',
        'projects.status': 'Status',
        'projects.actions': 'Actions',
        'projects.view': 'View',
        'projects.detail_title': 'Project Details',
        'projects.back': '← Back to List',
        'projects.sample_data': 'Sample Data',
        'projects.code': 'Python Code',
        'projects.save_code': 'Save Code',
        'projects.rerun': 'Re-run',
        'projects.delete_project': 'Delete Project',
        'projects.code_saved': 'Code saved',
        'projects.rerun_started': 'Batch processing started, please check progress in "Batch Processing" page',
        'projects.delete_confirm': 'Are you sure you want to delete this project?',
        'projects.delete_failed': 'Delete failed',
        'projects.load_failed': 'Failed to load project',
        'projects.status.draft': 'Draft',
        'projects.status.validated': 'Validated',
        'projects.status.executed': 'Executed',
        'projects.status.failed': 'Failed',
        
        // Environment status
        'env.checking': 'Checking environment',
        'env.initializing': 'Initializing Python environment...',
        'env.ready': 'Python environment ready',
        'env.error': 'Environment error',
        'env.timeout': 'Initialization timeout',
        'env.init_banner': 'Automatically initializing Python environment...',
        'env.init_success': '✅ Python environment ready',
        'env.init_failed': '❌ Environment initialization failed',
        'env.init_timeout': '⏱ Environment initialization timeout, please initialize manually in settings',
        
        // Wizard
        'wizard.welcome': 'Welcome to LogForge',
        'wizard.subtitle': 'Smart Network Log Formatter',
        'wizard.step1_title': 'Sample Analysis',
        'wizard.step1_desc': 'Paste a log sample, AI will automatically generate Python parsing code',
        'wizard.step2_title': 'Code Validation',
        'wizard.step2_desc': 'System automatically validates the generated code to ensure it runs correctly',
        'wizard.step3_title': 'Batch Processing',
        'wizard.step3_desc': 'Select input directory, batch process all log files and export to Excel',
        'wizard.step4_title': 'Project Management',
        'wizard.step4_desc': 'Historical projects can be viewed, edited or re-executed at any time',
        'wizard.dont_show': 'Don\'t show this wizard again',
        'wizard.start': 'Get Started',
        
        // Directory selection
        'dir.select_input': 'Select Input Directory',
        'dir.select_output': 'Select Output Directory',
        'dir.select_default_input': 'Select Default Input Directory',
        'dir.select_default_output': 'Select Default Output Directory',
    }
};

const I18n = {
    currentLang: 'zh-CN',
    
    init() {
        // 从设置中加载语言，如果没有则使用操作系统语言
        this.loadLanguage();
    },
    
    async loadLanguage() {
        try {
            const settings = await window.go.main.App.GetSettings();
            if (settings.language) {
                this.currentLang = settings.language;
            } else {
                // 使用操作系统语言
                const sysLang = navigator.language || navigator.userLanguage || 'zh-CN';
                this.currentLang = sysLang.startsWith('zh') ? 'zh-CN' : 'en';
            }
        } catch (_) {
            // 默认使用操作系统语言
            const sysLang = navigator.language || navigator.userLanguage || 'zh-CN';
            this.currentLang = sysLang.startsWith('zh') ? 'zh-CN' : 'en';
        }
        this.updateUI();
    },
    
    t(key) {
        const lang = translations[this.currentLang] || translations['zh-CN'];
        return lang[key] || key;
    },
    
    setLanguage(lang) {
        this.currentLang = lang;
        this.updateUI();
    },
    
    updateUI() {
        // 更新所有带 data-i18n 属性的元素
        document.querySelectorAll('[data-i18n]').forEach(el => {
            const key = el.getAttribute('data-i18n');
            el.textContent = this.t(key);
        });
        
        // 更新所有带 data-i18n-placeholder 属性的元素
        document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
            const key = el.getAttribute('data-i18n-placeholder');
            el.placeholder = this.t(key);
        });
        
        // 更新 HTML lang 属性
        document.documentElement.lang = this.currentLang;
        
        // 更新页面标题
        document.title = this.t('app.title') + ' - ' + this.t('app.subtitle');
    }
};
