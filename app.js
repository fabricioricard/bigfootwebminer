document.getElementById('connect-btn').addEventListener('click', async function() {
    document.getElementById('status').innerText = "Conectando e iniciando mineração...";

    try {
        const miner = await startMining();
        document.getElementById('status').innerText = "Mineração Iniciada!";

        // Aqui você pode mandar os ganhos para sua carteira (implementação do servidor)
        miner.start();

    } catch (error) {
        document.getElementById('status').innerText = "Erro ao iniciar mineração.";
        console.error("Erro ao iniciar o minerador:", error);
    }
});

async function startMining() {
    // Carregar o arquivo WASM do minerador
    const response = await fetch('wasm/packetcrypt.wasm');
    const wasmBinary = await response.arrayBuffer();
    
    const minerModule = await WebAssembly.instantiate(wasmBinary);
    
    const miner = {
        start: function() {
            // Aqui é onde o código de mineração será executado
            // Exemplo: minerModule.exports.startMining();
            console.log("Minerador PKT iniciado!");
        }
    };

    return miner;
}
