package main

import (
    "fmt"
    "bigchain/bigchainwallet"
)

func main() {
    fmt.Println("BigWallet iniciado")

    // Exemplo de uso mínimo (ajuste conforme suas funções reais)
    cfg := bigchainwallet.LoadConfig()
    bigchainwallet.StartWallet(cfg)
}
