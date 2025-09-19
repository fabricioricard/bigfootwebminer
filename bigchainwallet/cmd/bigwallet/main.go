package main

import (
    "fmt"
    "bigchain/bigchainwallet"
)

func main() {
    fmt.Println("BigWallet iniciado")

    // Exemplo de uso mínimo
    cfg := bigchainwallet.LoadConfig()
    bigchainwallet.StartWallet(cfg)
}
