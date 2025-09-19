#!/bin/bash

# Script para converter PacketCrypt para BigCrypt
echo "Convertendo PacketCrypt para BigCrypt..."

# Função para substituir em arquivos
replace_in_files() {
    find . -name "*.go" -type f -exec sed -i "s/PacketCrypt/BigCrypt/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/packetcrypt/bigcrypt/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/PacketCryptAnn/BigCryptAnn/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/PcAnn/BcAnn/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/ValidatePcAnn/ValidateBcAnn/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/ValidatePcBlock/ValidateBcBlock/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/PcCoinbaseCommit/BcCoinbaseCommit/g" {} \;
    find . -name "*.go" -type f -exec sed -i "s/pcCoinbasePrefix/bcCoinbasePrefix/g" {} \;
}

# Executar substituições
replace_in_files

echo "Conversão concluída!"
