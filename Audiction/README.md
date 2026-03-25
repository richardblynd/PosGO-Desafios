# Auction Go - Leilão Automático

Sistema de leilões em Go com fechamento automático baseado em tempo configurável.

## Funcionalidades

- Criação de leilões com fechamento automático
- Lances (bids) em leilões ativos
- Consulta de leilões, lances e usuários
- Fechamento automático via goroutine após expiração do tempo configurado

## Variáveis de Ambiente

As variáveis de ambiente são configuradas no arquivo `cmd/auction/.env`:

| Variável | Descrição | Exemplo |
|--|--|--|
| `AUCTION_INTERVAL` | Duração do leilão antes do fechamento automático | `20s`, `5m`, `1h` |
| `BATCH_INSERT_INTERVAL` | Intervalo para inserção em lote de bids | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de bids | `4` |
| `MONGODB_URL` | URL de conexão com MongoDB | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |
| `MONGO_INITDB_ROOT_USERNAME` | Usuário root do MongoDB | `admin` |
| `MONGO_INITDB_ROOT_PASSWORD` | Senha root do MongoDB | `admin` |

### Configurando o tempo do leilão

Para alterar a duração dos leilões, edite a variável `AUCTION_INTERVAL` no arquivo `cmd/auction/.env`:

```env
# Exemplos de valores válidos:
AUCTION_INTERVAL=30s    # 30 segundos
AUCTION_INTERVAL=5m     # 5 minutos
AUCTION_INTERVAL=1h     # 1 hora
```

Se a variável não estiver definida ou for inválida, o padrão é **5 minutos**.

## Rodando com Docker Compose

```bash
docker-compose up --build
```

Isso irá:
1. Subir um container MongoDB
2. Compilar e rodar a aplicação Go
3. Expor a API na porta `8080`

Para parar:

```bash
docker-compose down
```

## Rodando Localmente (sem Docker)

### Pré-requisitos

- Go 1.20+
- MongoDB rodando localmente ou acessível via rede

### Passos

1. Ajuste a `MONGODB_URL` no `cmd/auction/.env` para apontar para o seu MongoDB local:
   ```env
   MONGODB_URL=mongodb://admin:admin@localhost:27017/auctions?authSource=admin
   ```

2. Execute a aplicação:
   ```bash
   go run cmd/auction/main.go
   ```

## Endpoints da API

| Método | Rota | Descrição |
|--|--|--|
| `POST` | `/auction` | Criar leilão |
| `GET` | `/auction` | Listar leilões (query params: `status`, `category`, `productName`) |
| `GET` | `/auction/:auctionId` | Buscar leilão por ID |
| `GET` | `/auction/winner/:auctionId` | Buscar lance vencedor |
| `POST` | `/bid` | Criar lance |
| `GET` | `/bid/:auctionId` | Listar lances de um leilão |
| `POST` | `/user` | Criar usuário |
| `GET` | `/user/:userId` | Buscar usuário por ID |

## Testes

### Teste de fechamento automático

O teste verifica que um leilão criado é automaticamente fechado após o tempo configurado em `AUCTION_INTERVAL`.

**Pré-requisito:** MongoDB rodando (localmente ou via Docker).

```bash
# Subir apenas o MongoDB
docker-compose up mongodb -d

# Rodar o teste
go test ./internal/infra/database/auction/ -v -run TestAuctionAutoClose -timeout 30s
```

O teste:
1. Cria um leilão com `AUCTION_INTERVAL=3s`
2. Verifica que o status inicial é `Active` (0)
3. Aguarda 4 segundos
4. Verifica que o status mudou automaticamente para `Completed` (1)
