package main

import (
    "log"

    "bigchain/bigchainwallet" // apenas funções públicas da wallet
)

func main() {
    err := bigchainwallet.StartWallet() // StartWallet deve ser uma função pública
    if err != nil {
        log.Fatal(err)
    }
}
