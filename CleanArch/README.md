# CleanArch - Sistema de Pedidos

Este projeto implementa uma arquitetura limpa (Clean Architecture) para um sistema de pedidos, com suporte a REST, gRPC e GraphQL.

## Pré-requisitos
- Docker e Docker Compose

## Como executar

### Executar com Docker (Recomendado)

O projeto está completamente dockerizado e pode ser executado com um único comando:

```bash
docker compose up
```

Este comando irá:
- ✅ Inicializar o MySQL com o banco `orders`
- ✅ Inicializar o RabbitMQ para mensageria
- ✅ Executar as migrações automaticamente
- ✅ Iniciar a aplicação com todos os serviços

### Executar em modo desenvolvimento

Se preferir executar localmente (necessário Go 1.24+):

1. Subir apenas as dependências:
```bash
docker compose up mysql rabbitmq -d
```

2. Executar as migrações:
```bash
docker compose up migrate
```

3. Rodar a aplicação:
```bash
go run cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go
```

## Serviços Disponíveis

Após executar `docker compose up`, os seguintes serviços estarão disponíveis:

- **🌐 REST API:**
  - Endpoints: `GET /orders` e `POST /order`
  - Porta: **8000**
  - URL: `http://localhost:8000`

- **⚡ gRPC:**
  - Service: `OrderService`
  - Porta: **50051**
  - Endpoint: `localhost:50051`

- **📊 GraphQL:**
  - Playground: `http://localhost:8080/`
  - Endpoint: `/query`
  - Porta: **8080**

- **🗄️ MySQL:**
  - Host: `localhost:3306`
  - Database: `orders`
  - User: `root` / Password: `root`

- **🐰 RabbitMQ Management:**
  - URL: `http://localhost:15672`
  - User: `guest` / Password: `guest`

## Testando as APIs

### REST API

**Criar pedido:**
```bash
# PowerShell
Invoke-RestMethod -Uri "http://localhost:8000/order" -Method POST -ContentType "application/json" -Body '{"id":"order-123", "price": 100.5, "tax": 0.5}'

# cURL (Linux/Mac)
curl -X POST -H "Content-Type: application/json" -d '{"id":"order-123", "price": 100.5, "tax": 0.5}' http://localhost:8000/order
```

**Listar pedidos:**
```bash
# PowerShell
Invoke-RestMethod -Uri "http://localhost:8000/orders" -Method GET

# cURL (Linux/Mac)
curl http://localhost:8000/orders
```

### Arquivos de teste HTTP

- Utilize os arquivos `api/create_order.http` e `api/list_orders.http` no VS Code com a extensão REST Client.

### gRPC

- Para gRPC, utilize ferramentas como [grpcurl](https://github.com/fullstorydev/grpcurl) ou [BloomRPC](https://github.com/bloomrpc/bloomrpc).

### GraphQL

Acesse o playground em `http://localhost:8080/` e utilize as queries/mutations:

### Mutation para criar ordem
```graphql
mutation {
  createOrder(input: { id: "123", Price: 100.0, Tax: 10.0 }) {
    id
    Price
    Tax
    FinalPrice
  }
}
```

### Query para listar ordens
```graphql
query {
  listOrders {
    id
    Price
    Tax
    FinalPrice
  }
}
```

## Arquitetura

O projeto segue os princípios da **Clean Architecture** com as seguintes camadas:

```
├── cmd/ordersystem/     # Main application e dependency injection (Wire)
├── internal/
│   ├── entity/          # Entidades de domínio
│   ├── usecase/         # Casos de uso (business logic)
│   ├── infra/
│   │   ├── database/    # Repository implementations
│   │   ├── web/         # REST handlers
│   │   ├── grpc/        # gRPC service
│   │   └── graph/       # GraphQL resolvers
│   └── event/           # Event handlers para RabbitMQ
├── pkg/events/          # Event dispatcher
└── migrations/          # Database migrations
```

## Tecnologias Utilizadas

- **Go 1.24** - Linguagem principal
- **MySQL 8.0** - Banco de dados
- **RabbitMQ** - Message broker
- **Docker & Docker Compose** - Containerização
- **Wire** - Dependency Injection
- **gqlgen** - GraphQL server
- **gRPC** - API de alta performance
- **Viper** - Configuração

## Observações

- ✅ **Inicialização automática**: Todas as dependências são iniciadas automaticamente
- ✅ **Migrações automáticas**: Banco de dados é configurado automaticamente
- ✅ **Healthchecks**: Serviços aguardam dependências estarem prontas
- ✅ **Portas configuradas**: 8000 (REST), 8080 (GraphQL), 50051 (gRPC)
- ✅ **Clean Architecture**: Código organizado e testável

## Troubleshooting

**Problema com portas ocupadas:**
```bash
docker compose down
# Aguarde alguns segundos
docker compose up
```

**Limpar dados do banco:**
```bash
docker compose down -v  # Remove volumes
docker compose up
```
