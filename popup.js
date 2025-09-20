import init, { start_mining } from './pkg/packetcrypt_rs.js';

// Configurações
const WALLET_ADDRESS = 'pkt1q2phzyfzd7aufszned7q2h77t4u0kl3exxgyuqf'; // Mover para um arquivo de configuração no futuro

// Utilitários DOM
const getElement = (id) => document.getElementById(id) || null;
const updateStatus = (message) => {
    const status = getElement('status');
    if (status) status.textContent = message;
};
const toggleLoading = (isLoading) => {
    const startBtn = getElement('start-btn');
    const stopBtn = getElement('stop-btn');
    const spinner = getElement('loading-spinner');
    if (startBtn && stopBtn && spinner) {
        startBtn.setAttribute('aria-busy', isLoading.toString());
        startBtn.disabled = isLoading;
        stopBtn.disabled = !isLoading;
        spinner.classList.toggle('hidden', !isLoading);
    }
};

// Estado da mineração
let isMining = false;

// Função para iniciar a mineração
const startMining = async () => {
    try {
        toggleLoading(true);
        updateStatus('Iniciando mineração...');

        await init();
        start_mining(WALLET_ADDRESS);
        
        isMining = true;
        updateStatus(`Minerando para: ${WALLET_ADDRESS}`);
        
        // Notificar o background.js (se existir)
        chrome.runtime.sendMessage({ type: 'START_MINING', wallet: WALLET_ADDRESS });
    } catch (error) {
        isMining = false;
        updateStatus('Erro ao iniciar mineração.');
        console.error('[BIGFOOT] Erro ao iniciar mineração:', error);
    } finally {
        toggleLoading(false);
    }
};

// Função para parar a mineração
const stopMining = () => {
    try {
        toggleLoading(true);
        updateStatus('Parando mineração...');

        // Supondo que o módulo WASM tenha uma função stop_mining
        // Se não houver, você precisará implementar uma lógica para parar
        // Exemplo: stop_mining();
        
        isMining = false;
        updateStatus('Mineração parada.');
        
        // Notificar o background.js (se existir)
        chrome.runtime.sendMessage({ type: 'STOP_MINING' });
    } catch (error) {
        updateStatus('Erro ao parar mineração.');
        console.error('[BIGFOOT] Erro ao parar mineração:', error);
    } finally {
        toggleLoading(false);
    }
};

// Inicialização
window.addEventListener('DOMContentLoaded', () => {
    const startBtn = getElement('start-btn');
    const stopBtn = getElement('stop-btn');

    if (!startBtn || !stopBtn) {
        console.error('[BIGFOOT] Botões não encontrados no DOM.');
        return;
    }

    startBtn.addEventListener('click', startMining);
    stopBtn.addEventListener('click', stopMining);
});
