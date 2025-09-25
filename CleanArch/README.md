# CleanArch - Desafio

Este projeto implementa uma arquitetura limpa para um sistema de pedidos, com suporte a REST, gRPC e GraphQL.

## Pré-requisitos
- Docker e Docker Compose
- Go 1.20+

## Passos para rodar o projeto

### 1. Subir banco de dados e dependências

```sh
docker compose up
```
Isso irá preparar o banco de dados automaticamente.


### 2. Rodar a aplicação

```sh
go run .
```

## Serviços e Portas


- **REST:**
  - Endpoint: `GET /orders` e `POST /order`
  - Porta: **8000**
  - Exemplo: `http://localhost:8000/orders`

- **gRPC:**
  - Service: `OrderService`
  - Porta: **50051**

- **GraphQL:**
  - Playground: `http://localhost:8080/`
  - Endpoint: `/query`
  - Porta: **8080**

## Testando as APIs

- Utilize o arquivo `api/create_order.http` e `api/list_orders.http` para testar as requisições REST.
- Para gRPC, utilize uma ferramenta como [grpcurl](https://github.com/fullstorydev/grpcurl) ou [BloomRPC](https://github.com/bloomrpc/bloomrpc).
- Para GraphQL, acesse o playground na porta 8081 e utilize as queries/mutations:

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

## Observações
- Certifique-se de que as portas 8080 (REST), 8081 (GraphQL) e 50051 (gRPC) estejam livres.
- O projeto utiliza Clean Architecture e pode ser expandido facilmente para outros serviços.
