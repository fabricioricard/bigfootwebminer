package bigchainwallet

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
)

var walletDBPath = filepath.Join(os.Getenv("HOME"), ".bigchainwallet", "wallet.db")

// StartWallet inicializa a wallet
func StartWallet() error {
    // Cria diretório da wallet, se não existir
    walletDir := filepath.Dir(walletDBPath)
    if _, err := os.Stat(walletDir); os.IsNotExist(err) {
        err := os.MkdirAll(walletDir, 0700)
        if err != nil {
            return fmt.Errorf("erro criando diretório da wallet: %v", err)
        }
    }

    // Cria arquivo do banco, se não existir
    if _, err := os.Stat(walletDBPath); os.IsNotExist(err) {
        f, err := os.Create(walletDBPath)
        if err != nil {
            return fmt.Errorf("erro criando banco da wallet: %v", err)
        }
        f.Close()
        fmt.Println("Wallet criada em:", walletDBPath)
        return nil
    }

    // Aqui você pode carregar configurações ou o banco, por enquanto só retorna ok
    fmt.Println("Wallet existente encontrada em:", walletDBPath)
    return nil
}

// Função de exemplo para simular operação da wallet
func GetBalance() (float64, error) {
    // Exemplo: sem conexão real, retorna 0
    return 0, errors.New("função GetBalance ainda não implementada")
}
