# BIGchain Mining Guide

## Requirements
- Linux/Ubuntu system
- Go 1.18+
- Internet connection

## Setup Instructions
1. Clone repository: `git clone [seu_repositório]`
2. Compile: `cd bigchain && go install -v .`
3. Generate address: `go run generate_bigchain_address.go`
4. Start mining: `~/go/bin/bigchain --generate --miningaddr=[SEU_ENDEREÇO] --connect=150.136.245.118:8433`

## Network Settings
- P2P Port: 8433
- RPC Port: 8334 (optional)
- Seed Node: 150.136.245.118:8433
