// Utilitário para atualizar o status no DOM com verificação de elemento
function updateStatus(message) {
    const statusElement = document.getElementById('status');
    if (!statusElement) {
        console.error('Elemento de status não encontrado no DOM.');
        return;
    }
    statusElement.innerText = message;
}

// Função para carregar o módulo WASM com tratamento de erro
async function loadWasmModule() {
    try {
        const response = await fetch('wasm/packetcrypt.wasm');
        if (!response.ok) {
            throw new Error(`Falha ao carregar WASM: ${response.status} ${response.statusText}`);
        }
        const wasmBinary = await response.arrayBuffer();
        const module = await WebAssembly.instantiate(wasmBinary);
        return module;
    } catch (error) {
        throw new Error(`Erro ao carregar o módulo WASM: ${error.message}`);
    }
}

// Função para criar uma instância do minerador
function createMiner(module) {
    return {
        start: function () {
            // Substitua por chamada real ao módulo WASM quando disponível
            // Exemplo: module.instance.exports.startMining();
            console.log("Minerador PKT iniciado!");
        },
        stop: function () {
            // Adicione lógica para parar o minerador, se aplicável
            console.log("Minerador PKT parado!");
        }
    };
}

// Função principal para iniciar a mineração
async function initializeMining(source = 'botão') {
    updateStatus(`Iniciando mineração via ${source}...`);

    try {
        const module = await loadWasmModule();
        const miner = createMiner(module);
        miner.start();
        updateStatus(`Mineração iniciada via ${source}!`);
        return miner;
    } catch (error) {
        updateStatus(`Erro ao iniciar mineração via ${source}: ${error.message}`);
        console.error(`Erro na mineração (${source}):`, error);
        throw error;
    }
}

// Gerenciamento de estado para evitar múltiplas inicializações
let minerInstance = null;

// Função para iniciar a mineração (usada por botão ou extensão)
async function startMiner(source = 'extensão') {
    if (minerInstance) {
        updateStatus(`Mineração já está em execução (iniciada via ${source}).`);
        return minerInstance;
    }

    try {
        minerInstance = await initializeMining(source);
        return minerInstance;
    } catch (error) {
        minerInstance = null; // Reseta o estado em caso de erro
        throw error;
    }
}

// Evento de clique no botão HTML
function setupButtonListener() {
    const connectButton = document.getElementById('connect-btn');
    if (!connectButton) {
        console.error('Botão "connect-btn" não encontrado no DOM.');
        return;
    }

    connectButton.addEventListener('click', async () => {
        try {
            await startMiner('botão');
        } catch (error) {
            // Erro já tratado em startMiner, mas pode adicionar mais ações aqui se necessário
        }
    });
}

// Escuta mensagens da extensão com validação de origem
function setupMessageListener() {
    window.addEventListener('message', (event) => {
        // Valida a origem da mensagem
        if (event.source !== window || !event.data?.type) return;
        if (event.data.type !== 'BIGFOOT_START_MINER') return;

        // Adicional: valide o domínio ou origem, se aplicável
        if (event.origin !== window.location.origin) {
            console.warn('Mensagem recebida de origem não confiável:', event.origin);
            return;
        }

        startMiner('extensão').catch(() => {
            // Erro já tratado em startMiner, mas pode adicionar mais ações aqui
        });
    });
}

// Inicialização do script
document.addEventListener('DOMContentLoaded', () => {
    setupButtonListener();
    setupMessageListener();
});
