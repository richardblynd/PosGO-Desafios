# Rate Limiter em Go

Um sistema de rate limiting robusto implementado em Go que suporta limitação por endereço IP e token de acesso, com backend plugável para armazenamento (Redis/Memory).

## 📋 Características

- ✅ **Limitação por IP**: Controla requisições por endereço IP
- ✅ **Limitação por Token**: Controla requisições por token de acesso via header `API_KEY`
- ✅ **Prioridade de Token**: Configurações de token sobrepõem configurações de IP
- ✅ **Backend Plugável**: Suporte para Redis e armazenamento em memória
- ✅ **Middleware HTTP**: Fácil integração com servidores web
- ✅ **Configuração Flexível**: Via variáveis de ambiente ou arquivo `.env`
- ✅ **Docker Ready**: Inclui Docker e Docker Compose
- ✅ **Testes Abrangentes**: Cobertura completa de testes

## 🚀 Início Rápido

### Pré-requisitos

- Go 1.21+
- Docker & Docker Compose (opcional, para Redis)

### Instalação Local (Armazenamento em Memória)

1. Clone o repositório:
```bash
git clone <repository-url>
cd ratelimiter
```

2. Instale as dependências:
```bash
go mod tidy
```

3. Configure o ambiente (opcional):
```bash
cp .env.example .env
# Edite .env conforme necessário
```

4. Execute o servidor:
```bash
go run cmd/server/main.go
```

O servidor estará disponível em `http://localhost:8080`

### Instalação com Docker (Redis)

1. Execute com Docker Compose:
```bash
docker-compose up --build
```

O servidor estará disponível em `http://localhost:8080` com Redis como backend.

## ⚙️ Configuração

### Variáveis de Ambiente

Todas as configurações podem ser definidas via variáveis de ambiente ou arquivo `.env`:

```bash
# Configuração do Servidor
PORT=8080

# Configuração de Armazenamento
STORAGE_TYPE=memory          # Opções: redis, memory
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Configuração de Rate Limiting
IP_RATE_LIMIT=10            # Requisições por segundo por IP
IP_BLOCK_DURATION=5m        # Tempo de bloqueio quando limite for excedido
TOKEN_RATE_LIMIT=100        # Requisições por segundo padrão para tokens
TOKEN_BLOCK_DURATION=5m     # Tempo de bloqueio padrão para tokens

# Configurações Específicas de Tokens
TOKEN_ABC123_RATE_LIMIT=50
TOKEN_ABC123_BLOCK_DURATION=10m

TOKEN_XYZ789_RATE_LIMIT=200
TOKEN_XYZ789_BLOCK_DURATION=2m

TOKEN_PREMIUM_RATE_LIMIT=1000
TOKEN_PREMIUM_BLOCK_DURATION=1m
```

### Formato de Duração

As durações podem ser especificadas em vários formatos:
- `5m` - 5 minutos
- `30s` - 30 segundos
- `1h` - 1 hora
- `300` - 300 segundos (fallback)

## 🔧 Como Usar

### Requisições Básicas

Faça requisições para qualquer endpoint:

```bash
curl http://localhost:8080/api/v1/users
curl http://localhost:8080/api/v1/orders
curl http://localhost:8080/health
```

### Requisições com Token

Use o header `API_KEY` para autenticação por token:

```bash
curl -H "API_KEY: abc123" http://localhost:8080/api/v1/users
curl -H "API_KEY: premium" http://localhost:8080/api/v1/orders
```

### 📁 Arquivos de Exemplo HTTP

O projeto inclui arquivos `.http` prontos para testar todas as funcionalidades:

```
examples/
├── requests.http           # Requisições básicas
├── rate-limit-scenarios.http  # Cenários específicos de rate limiting  
├── performance-tests.http     # Testes de performance
├── test-config.env           # Configuração otimizada para testes
└── README.md                # Guia dos arquivos de exemplo
```

**Como usar no VS Code:**
1. Instale a extensão "REST Client"
2. Abra qualquer arquivo `.http` na pasta `examples/`
3. Clique em "Send Request" para executar
4. Observe headers de resposta e status codes

### Headers de Resposta

O sistema inclui headers informativos em todas as respostas:

```
X-RateLimit-Limit: 10        # Limite de requisições por segundo
X-RateLimit-Remaining: 7     # Requisições restantes na janela atual
X-RateLimit-Reset: 1609459200 # Timestamp do próximo reset
```

### Resposta de Limite Excedido

Quando o limite é excedido, você receberá:

**Status:** `429 Too Many Requests`

```json
{
  "error": "Too Many Requests",
  "message": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "code": 429,
  "timestamp": "2023-09-24T10:30:00Z"
}
```

## 🏗️ Arquitetura

### Estrutura do Projeto

```
├── cmd/server/          # Ponto de entrada da aplicação
│   └── main.go
├── internal/
│   ├── config/         # Sistema de configuração
│   │   ├── config.go
│   │   └── config_test.go
│   ├── middleware/     # Middleware HTTP
│   │   ├── ratelimiter.go
│   │   └── ratelimiter_test.go
│   ├── ratelimiter/   # Lógica core do rate limiter
│   │   ├── ratelimiter.go
│   │   └── ratelimiter_test.go
│   └── storage/       # Abstrações de armazenamento
│       ├── interface.go
│       ├── memory.go
│       ├── memory_test.go
│       └── redis.go
├── examples/          # Arquivos de teste HTTP
├── docker-compose.yml # Configuração Docker
├── Dockerfile        # Imagem Docker
└── .env.example     # Exemplo de configuração
```

### Componentes Principais

1. **Storage Interface**: Abstração para diferentes backends de armazenamento
2. **RateLimiter**: Lógica central de limitação de taxa
3. **Middleware**: Integração HTTP com extração de IP e token
4. **Config**: Sistema flexível de configuração

### Estratégia de Armazenamento

O sistema suporta múltiplos backends via interface:

```go
type Storage interface {
    CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration, blockDuration time.Duration) (*RateLimitResult, error)
    IsBlocked(ctx context.Context, key string) (bool, time.Duration, error)
    Block(ctx context.Context, key string, duration time.Duration) error
    Close() error
    Health(ctx context.Context) error
}
```

**Implementações Disponíveis:**
- **Redis**: Para ambiente de produção distribuído
- **Memory**: Para desenvolvimento e testes

## 🧪 Testes

Execute os testes completos:

```bash
# Todos os testes unitários e de integração
go test ./... -v

# Testes com cobertura de código
go test ./... -cover

# Testes com benchmark de performance
go test ./... -bench=. -benchmem

# Testes apenas dos pacotes internos (sem cmd/server)
go test ./internal/... -cover -v
```

### Cobertura de Código

O projeto possui alta cobertura de testes:

```
✅ Config Package:      89.7% coverage  
✅ Middleware Package:  78.6% coverage  
✅ RateLimiter Package: 44.1% coverage  
✅ Storage Package:     35.7% coverage  
```

### Executando Testes Específicos

```bash
# Testar apenas configuração
go test ./internal/config -v

# Testar apenas rate limiter
go test ./internal/ratelimiter -v  

# Testar apenas middleware com cobertura
go test ./internal/middleware -cover -v

# Testar apenas storage com benchmark
go test ./internal/storage -bench=. -benchmem

# Testar cenário específico
go test ./internal/ratelimiter -run TestRateLimiterPriority

# Gerar relatório de cobertura HTML
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Testes com Arquivos HTTP

Para testes interativos, use os arquivos na pasta `examples/`:

```bash
# Configure limites baixos para ver o bloqueio rapidamente
cp examples/test-config.env .env

# Inicie o servidor
go run cmd/server/main.go

# Use os arquivos .http no VS Code com REST Client
# ou execute os scripts:
test-rate-limit.bat  # Windows
./test-rate-limit.sh # Linux/Mac
```

### Tipos de Teste Disponíveis

O projeto inclui testes abrangentes organizados por pacote:

**📋 Testes de Configuração** (`internal/config/config_test.go`)
- Carregamento de variáveis de ambiente
- Parsing de durações (5m, 30s, 1h, etc.)
- Configurações específicas de tokens
- Valores padrão e fallbacks

**🔒 Testes de Rate Limiter** (`internal/ratelimiter/ratelimiter_test.go`)
- Limitação por IP com validação de endereços
- Limitação por token com diferentes configurações
- Prioridade de token sobre IP
- Health checks e error handling

**🌐 Testes de Middleware** (`internal/middleware/ratelimiter_test.go`)
- Integração HTTP completa
- Extração de IP de headers proxy (X-Forwarded-For, X-Real-IP)
- Headers de rate limiting (X-RateLimit-*)
- Respostas de erro 429 formatadas
- Processamento de tokens API_KEY

**💾 Testes de Storage** (`internal/storage/memory_test.go`)
- Algoritmo sliding window
- Funcionalidade de bloqueio temporal
- Performance benchmarks
- Cleanup automático de dados antigos

**⚡ Benchmarks de Performance**
```bash
# Exemplo de resultado:
BenchmarkMemoryStorage-8    944710   5439 ns/op   11451 B/op   10 allocs/op
```

## 🔍 Monitoramento

### Health Check

Endpoint para verificar saúde do serviço:

```bash
curl http://localhost:8080/health
```

### Logs

O sistema produz logs informativos:
```
2023/09/24 10:30:00 Server starting on port 8080...
2023/09/24 10:30:05 Rate limiter error: invalid IP address
```

## 🐳 Docker

### Build Local

```bash
docker build -t ratelimiter .
docker run -p 8080:8080 ratelimiter
```

### Docker Compose

```bash
# Iniciar serviços
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar serviços
docker-compose down
```

## 📊 Exemplos Práticos

### Cenário 1: Limitação por IP

```bash
# Configure IP_RATE_LIMIT=5 no .env
# Faça 6 requisições rapidamente:
for i in {1..6}; do
  curl -w "Status: %{http_code}\n" http://localhost:8080/api/v1/users
done

# Resultado:
# Requisições 1-5: Status 200
# Requisição 6: Status 429
```

### Cenário 2: Token Sobrepõe IP

```bash
# Mesmo com IP bloqueado, token funciona:
curl -H "API_KEY: premium" http://localhost:8080/api/v1/users
# Status: 200 (mesmo com IP bloqueado)
```

### Cenário 3: Diferentes Limites por Token

```bash
# Token regular (limit padrão: 100 req/s)
curl -H "API_KEY: regular-token" http://localhost:8080/api/v1/users

# Token premium (limit configurado: 1000 req/s)
curl -H "API_KEY: premium" http://localhost:8080/api/v1/users
```

## 🔧 Desenvolvimento

### Executar em Modo Desenvolvimento

```bash
# Com hot reload (se tiver air instalado)
air

# Ou modo normal
go run cmd/server/main.go
```

### Adicionar Novo Backend de Armazenamento

1. Implemente a interface `storage.Storage`
2. Adicione a inicialização em `cmd/server/main.go`
3. Crie testes em `internal/storage/[new_backend]_test.go`

### Estrutura de Testes

Os testes estão organizados junto ao código que testam:

```
internal/
├── config/config_test.go      # Testa sistema de configuração
├── middleware/ratelimiter_test.go  # Testa integração HTTP
├── ratelimiter/ratelimiter_test.go # Testa lógica core
└── storage/memory_test.go     # Testa implementação de storage
```

**Vantagens desta estrutura:**
- ✅ Cobertura de código precisa (`go test ./... -cover`)
- ✅ Testes podem acessar funções não-exportadas
- ✅ Organização clara por funcionalidade
- ✅ Fácil manutenção e localização

### Configurar Tokens Personalizados

Adicione no `.env`:
```bash
TOKEN_MEUTOKEN_RATE_LIMIT=500
TOKEN_MEUTOKEN_BLOCK_DURATION=30s
```

## 🚨 Solução de Problemas

### Redis Não Conecta

```
Failed to connect to Redis: dial tcp localhost:6379: connect: connection refused
```

**Solução**: Use `STORAGE_TYPE=memory` ou inicie o Redis:
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

### Rate Limit Não Funciona

1. Verifique se os headers estão sendo enviados corretamente
2. Confirme a configuração no `.env`
3. Verifique os logs do servidor

### Performance Issues

- Para alta carga, use Redis em cluster
- Ajuste timeouts de conexão
- Considere usar `STORAGE_TYPE=memory` para desenvolvimento

## 📝 API Reference

### Endpoints Disponíveis

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/health` | GET | Health check do serviço |
| `/api/v1/users` | GET | Exemplo de endpoint protegido |
| `/api/v1/orders` | GET | Exemplo de endpoint protegido |

### Headers Suportados

| Header | Descrição | Exemplo |
|--------|-----------|---------|
| `API_KEY` | Token de autenticação | `API_KEY: abc123` |
| `Authorization` | Bearer token (alternativo) | `Authorization: Bearer abc123` |
| `X-Forwarded-For` | IP real (para proxies) | `X-Forwarded-For: 192.168.1.1` |
| `X-Real-IP` | IP real (alternativo) | `X-Real-IP: 192.168.1.1` |

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Adicione testes para novas funcionalidades
4. Execute os testes: `go test ./... -cover -v`
5. Faça commit das mudanças
6. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a MIT License.

---

## 🎯 Atendimento aos Requisitos

✅ **Rate limiter como middleware injetado ao servidor web**  
✅ **Configuração do número máximo de requisições por segundo**  
✅ **Opção de tempo de bloqueio configurável**  
✅ **Configurações via variáveis de ambiente ou arquivo .env**  
✅ **Limitação por IP e por token de acesso**  
✅ **Token no formato API_KEY: <TOKEN>**  
✅ **Resposta HTTP 429 com mensagem específica quando limite excedido**  
✅ **Armazenamento Redis com strategy pattern para fácil troca**  
✅ **Lógica do limiter separada do middleware**  
✅ **Testes automatizados completos**  
✅ **Docker e docker-compose funcionais**  
✅ **Servidor na porta 8080**  
✅ **Configuração de token sobrepõe configuração de IP**  

*Projeto desenvolvido seguindo as melhores práticas de Go e arquitetura limpa.*