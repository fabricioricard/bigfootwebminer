@echo off
REM Instalar o Git se necessário
REM Instalar o Rust manualmente (o Windows tem o processo de instalação diferente)
REM Baixar e instalar o PacketCrypt
curl -s https://sh.rustup.rs | sh

REM Instalar o PacketCrypt
cargo install --git https://github.com/pkt-world/packetcrypt_rs.git --locked --features jit

REM Iniciar o minerador com a sua carteira PKT
packetcrypt ann -p pkt1q2phzyfzd7aufszned7q2h77t4u0kl3exxgyuqf http://pool.pkt.world