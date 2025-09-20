// Função para registrar logs com prefixo consistente
const log = (message, level = 'log') => {
    console[level](`[BIGFOOT] ${message}`);
};

// Função para verificar se o script está no contexto correto
const isValidContext = () => {
    // Verifica se window e window.postMessage existem
    if (typeof window === 'undefined' || !window.postMessage) {
        return false;
    }

    // Verifica se está na página esperada (alinhado com o manifest.json)
    const allowedHost = 'https://bigfootwebminer.vercel.app';
    return window.location.href.startsWith(allowedHost);
};

// Função para iniciar o minerador via postMessage com confirmação de evento
const startMiningWithEvent = async (timeoutMs = 5000) => {
    if (!isValidContext()) {
        throw new Error('Ambiente não compatível: este script deve rodar em uma página do BIGFOOT WebMiner.');
    }

    return new Promise((resolve, reject) => {
        // Configura um listener para confirmação do app.js
        const confirmationHandler = (event) => {
            if (event.data?.type === 'BIGFOOT_MINER_STARTED') {
                log('Minerador iniciado com sucesso via postMessage!');
                resolve(true);
            }
        };

        // Adiciona o listener para o evento de confirmação
        window.addEventListener('message', confirmationHandler);

        // Configura um timeout para falha
        const timeout = setTimeout(() => {
            window.removeEventListener('message', confirmationHandler);
            reject(new Error('Falha ao iniciar o minerador: tempo esgotado.'));
        }, timeoutMs);

        try {
            // Envia mensagem para o app.js
            log('Enviando mensagem para iniciar o minerador...');
            window.postMessage({ type: 'BIGFOOT_START_MINER' }, window.location.origin);
        } catch (error) {
            clearTimeout(timeout);
            window.removeEventListener('message', confirmationHandler);
            reject(new Error(`Erro ao enviar mensagem: ${error.message}`));
        }
    });
};

// Função principal para executar o script
(async () => {
    try {
        // Verifica o contexto antes de executar qualquer lógica
        if (!isValidContext()) {
            log('Script injetado em um contexto inválido. Este script deve rodar em https://bigfootwebminer.vercel.app.', 'warn');
            return;
        }

        log('Conectando com o WebMiner...');
        await startMiningWithEvent();
    } catch (error) {
        log(`Erro ao conectar: ${error.message}`, 'error');
        // Opcional: enviar erro para um serviço de monitoramento (ex.: Sentry)
        // if (window.Sentry) Sentry.captureException(error);
    }
})();
