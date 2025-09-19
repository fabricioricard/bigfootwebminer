# BigChain

BigChain é uma implementação completa de blockchain com mineração proof-of-work de largura de rede ociosa, baseada no algoritmo BigCrypt.

## O que é BigChain

BigChain é uma blockchain de camada 1 que utiliza o algoritmo BigCrypt para proof-of-work baseado em largura de banda. O projeto é um fork do PKT Cash, adaptado para implementar melhorias no algoritmo de mineração de largura de rede ociosa.

## Componentes

- **bigchaind** - Nó completo da blockchain BigChain
- **bigchainwallet** - Carteira para a blockchain BigChain  
- **bigchainctl** - Ferramenta de linha de comando para interagir com a blockchain

## Algoritmo BigCrypt

O BigCrypt é um algoritmo de proof-of-work que:
- Utiliza largura de banda de rede ociosa
- Incentiva a infraestrutura de rede descentralizada
- Recompensa mineradores pela transferência de dados
- Implementa validação baseada em anúncios (announcements)

## Instalação

### Requisitos
- Go 1.14 ou superior
- Git

### Compilação

```bash
git clone https://github.com/bigchain/bigchaind
cd bigchaind
./do
```

## Mineração

Para minerar na rede BigChain, você precisará do minerador BigCrypt separado (implementado em Rust).

## Licença

BigChain é licenciado sob a licença ISC Copyfree.
