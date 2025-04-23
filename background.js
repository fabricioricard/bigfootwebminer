chrome.runtime.onInstalled.addListener(() => {
    console.log('[BIGFOOT] Extensão instalada.');
});

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'START_MINING') {
        // Lógica para iniciar o minerador
        console.log('[BIGFOOT] Iniciando mineração em background...');
        sendResponse({ success: true });
    }
});