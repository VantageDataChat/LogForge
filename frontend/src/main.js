// Main entry point — bootstrap the SPA
document.addEventListener('DOMContentLoaded', async () => {
    await I18n.init();
    App.init();
});
