#!/bin/bash

# Script para converter PKT para BigChain (blockchain) mantendo BigCrypt (minerador)
echo "Convertendo PKT para BigChain..."

# Função para substituir referências PKT para BigChain
replace_pkt_references() {
    find . -name "*.go" -type f -exec sed -i "s/PKT Cash/BigChain/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/PKT blockchain/BigChain blockchain/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/PKT Network/BigChain Network/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pkt-cash/bigchain/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pktd/bigchaind/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pktwallet/bigchainwallet/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pktctl/bigchainctl/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pktconfig/bigchainconfig/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pktlog/bigchainlog/g" {} \;
    
    # Arquivos de documentação
    find . -name "*.md" -type f -exec sed -i "s/PKT Cash/BigChain/g" {} \;
    find . -name "*.md" -type f -exec sed -i "s/PKT blockchain/BigChain blockchain/g" {} \;
    find . -name "*.md" -type f -exec sed -i "s/pktd/bigchaind/g" {} \;
    find . -name "*.md" -type f -exec sed -i "s/pktwallet/bigchainwallet/g" {} \;
    find . -name "*.md" -type f -exec sed -i "s/pktctl/bigchainctl/g" {} \;
}

# Executar substituições
replace_pkt_references

echo "Conversão para BigChain concluída!"
