# StressTest - Sistema CLI para Testes de Carga

Um sistema de linha de comando desenvolvido em Go para realizar testes de carga em serviços web. A aplicação permite configurar o número de requisições, nível de concorrência e gera relatórios detalhados sobre o desempenho.

## 📋 Funcionalidades

- ✅ Testes de carga com concorrência configurável
- ✅ Relatórios detalhados com métricas de desempenho
- ✅ Distribuição de códigos de status HTTP
- ✅ Execução via CLI nativa ou Docker
- ✅ Testes unitários abrangentes
- ✅ Interface simples e intuitiva

## 🚀 Instalação e Execução

### Pré-requisitos

- Go 1.21+ (para execução nativa)
- Docker (para execução containerizada)

### Execução via CLI (Nativa)

1. **Clone e compile o projeto:**
```bash
# Clone o repositório
git clone <url-do-repositorio>
cd StressTest

# Compile a aplicação (Windows)
go build -o stresstest.exe

# Execute a aplicação (Windows)
.\stresstest.exe --url=http://google.com --requests=1000 --concurrency=10

# Compile a aplicação (Linux/macOS)
go build -o stresstest

# Execute a aplicação (Linux/macOS)
./stresstest --url=http://google.com --requests=1000 --concurrency=10
```

### Execução via Docker

1. **Construir a imagem Docker:**
```bash
docker build -t stresstest .
```

2. **Executar o container:**
```bash
# Exemplo básico
docker run --rm stresstest --url=http://google.com --requests=1000 --concurrency=10

# Outro exemplo
docker run --rm stresstest --url=https://httpbin.org/status/200 --requests=100 --concurrency=5
```

## 📖 Parâmetros de Linha de Comando

| Parâmetro | Descrição | Obrigatório | Exemplo |
|-----------|-----------|-------------|---------|
| `--url` | URL do serviço a ser testado | ✅ Sim | `--url=http://example.com` |
| `--requests` | Número total de requisições | ✅ Sim | `--requests=1000` |
| `--concurrency` | Número de chamadas simultâneas | ❌ Não (padrão: 1) | `--concurrency=10` |

### Exemplos de Uso

```bash
# Teste básico com 100 requisições sequenciais
./stresstest --url=http://httpbin.org/status/200 --requests=100

# Teste com 1000 requisições e 50 workers simultâneos
./stresstest --url=https://jsonplaceholder.typicode.com/posts/1 --requests=1000 --concurrency=50

# Teste de alta concorrência
./stresstest --url=http://example.com/api/health --requests=5000 --concurrency=100
```

## 📊 Relatório de Saída

Após a execução, o sistema gera um relatório detalhado contendo:

```
=== RELATÓRIO DO TESTE DE CARGA ===
Tempo total de execução: 2.347s
Total de requests realizados: 1000
Requests com status 200 (sucesso): 987
Tempo médio por request: 23.2ms
Requests por segundo: 426.15

Distribuição de códigos de status HTTP:
  200: 987 requests
  404: 8 requests
  500: 5 requests

=== FIM DO RELATÓRIO ===
```

### Métricas Disponíveis

- **Tempo total de execução**: Duração completa do teste
- **Total de requests**: Número de requisições executadas
- **Requests com sucesso**: Requisições que retornaram status 200
- **Tempo médio por request**: Latência média das requisições
- **Requests por segundo**: Taxa de throughput
- **Distribuição de status codes**: Contagem por código de status HTTP

## 🧪 Executando os Testes

### Testes Unitários

```bash
# Executar todos os testes
go test

# Executar testes com verbosidade
go test -v

# Executar testes com coverage
go test -cover

# Gerar relatório de coverage em HTML
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Testes de Benchmark

```bash
# Executar benchmarks
go test -bench=.

# Benchmark com profiling de CPU
go test -bench=. -cpuprofile=cpu.prof

# Benchmark com profiling de memória
go test -bench=. -memprofile=mem.prof
```

### Estrutura dos Testes

- `TestParseFlags`: Validação do parsing de argumentos CLI
- `TestRunStressTest`: Teste da funcionalidade principal de stress testing
- `TestWorker`: Teste das goroutines workers
- `TestWorkerWithTimeout`: Teste de comportamento com timeouts
- `BenchmarkRunStressTest`: Benchmark de performance

## 🏗️ Arquitetura do Sistema

### Componentes Principais

1. **CLI Parser**: Processa argumentos da linha de comando
2. **Worker Pool**: Gerencia concorrência com goroutines
3. **HTTP Client**: Realiza requisições com timeout configurado
4. **Results Collector**: Coleta e processa resultados
5. **Report Generator**: Gera relatórios formatados

### Fluxo de Execução

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Parse     │───▶│  Create      │──▶│   Start     │
│   CLI Args  │    │  Worker Pool │    │   Workers   │
└─────────────┘    └──────────────┘    └─────────────┘
                            │                   │
                            ▼                   ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Generate  │◀──│   Collect    │◀───│   Execute   │
│   Report    │    │   Results    │    │   Requests  │
└─────────────┘    └──────────────┘    └─────────────┘
```

## 🐳 Docker

### Construindo a Imagem

```bash
# Build da imagem
docker build -t stresstest .

# Build com tag específica
docker build -t stresstest:v1.0.0 .

# Build sem cache
docker build --no-cache -t stresstest .
```

### Características do Container

- **Imagem base**: Alpine Linux (mínima)
- **Usuário**: Non-root para segurança
- **Certificados SSL**: Incluídos para HTTPS
- **Tamanho**: ~15MB (multi-stage build)

## 🛠️ Desenvolvimento

### Estrutura do Projeto

```
StressTest/
├── main.go              # Aplicação principal
├── main_test.go         # Testes unitários
├── go.mod               # Módulo Go
├── Dockerfile           # Container Docker
├── .dockerignore        # Exclusões Docker
└── README.md            # Documentação
```

## 🔧 Configuração Avançada

### Timeout de Requisições

O timeout padrão é de 30 segundos. Para modificar, edite a linha em `worker()`:

```go
client := &http.Client{
    Timeout: 30 * time.Second, // Modificar aqui
}
```

### Limites de Concorrência

O sistema automaticamente ajusta a concorrência se ela for maior que o número de requisições.