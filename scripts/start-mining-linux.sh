#!/bin/bash

# Instalar dependências necessárias
sudo apt update
sudo apt install -y gcc git make curl libc6-dev

# Instalar o Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Instalar o minerador PacketCrypt
~/.cargo/bin/cargo install --git https://github.com/pkt-world/packetcrypt_rs.git --locked --features jit

# Iniciar o minerador com sua carteira PKT
~/.cargo/bin/packetcrypt ann -p pkt1q2phzyfzd7aufszned7q2h77t4u0kl3exxgyuqf http://pool.pkt.world