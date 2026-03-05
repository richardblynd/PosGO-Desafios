# Sistema de Temperatura por CEP com Observabilidade

Este projeto implementa um sistema distribuído em Go que recebe um CEP, identifica a cidade e retorna o clima atual (temperatura em graus Celsius, Fahrenheit e Kelvin) juntamente com o nome da cidade. O sistema implementa observabilidade completa usando OpenTelemetry e Zipkin para tracing distribuído.

## 📋 Arquitetura

O sistema é composto por dois serviços principais:

### Serviço A (Input Handler) - Porta 8080
- **Responsabilidade**: Receber e validar entrada de CEP
- **Endpoint**: `POST /cep`
- **Validação**: CEP deve conter exatamente 8 dígitos numéricos
- **Ação**: Encaminha CEPs válidos para o Serviço B

### Serviço B (Weather Orchestrator) - Porta 8081
- **Responsabilidade**: Orquestrar busca de localização e temperatura
- **Endpoint**: `POST /weather`
- **Integrações**: ViaCEP (localização) + WeatherAPI (temperatura)
- **Conversões**: Celsius → Fahrenheit e Kelvin

### Componentes de Observabilidade
- **Zipkin**: Interface web para visualização de traces (Porta 9411)
- **OpenTelemetry Collector**: Coleta e processa telemetria (Portas 4317/4318)

## 🚀 Como Executar

### Pré-requisitos
- Docker
- Docker Compose
- (Opcional) Chave da WeatherAPI para dados reais de temperatura

### 1. Clone e Configure

```bash
git clone <seu-repositorio>
cd Observabilidade
```

### 2. Configure a API Key (Opcional)

Para dados reais de temperatura, obtenha uma chave gratuita em [WeatherAPI](https://www.weatherapi.com/) e configure:

```bash
# Windows PowerShell
$env:WEATHER_API_KEY="sua_chave_aqui"

# Linux/Mac
export WEATHER_API_KEY="sua_chave_aqui"
```

**Nota**: Sem a API key, o sistema usa temperaturas simuladas baseadas no nome da cidade.

### 3. Execute o Sistema

```bash
docker-compose up --build
```

### 4. Aguarde a Inicialização

O sistema estará pronto quando você ver as mensagens:
```
servico-a    | Serviço A iniciado na porta 8080
servico-b    | Serviço B iniciado na porta 8081
zipkin       | Started @xxxms
```

## 📡 Testando o Sistema

### 1. Teste com CEP Válido

```bash
# PowerShell
$body = @{
    cep = "01310100"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/cep" -Method POST -Body $body -ContentType "application/json"

# Curl (Linux/Mac/Git Bash)
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}'
```

**Resposta esperada:**
```json
{
  "city": "São Paulo",
  "temp_C": 25.3,
  "temp_F": 77.54,
  "temp_K": 298.45
}
```

### 2. Teste com CEP Inválido

```bash
# PowerShell
$body = @{
    cep = "123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/cep" -Method POST -Body $body -ContentType "application/json"

# Curl
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

**Resposta esperada:**
```json
{
  "message": "invalid zipcode"
}
```

### 3. Teste com CEP Inexistente

```bash
# PowerShell
$body = @{
    cep = "99999999"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/cep" -Method POST -Body $body -ContentType "application/json"

# Curl
curl -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}'
```

**Resposta esperada:**
```json
{
  "message": "can not find zipcode"
}
```

## 📊 Visualizando Traces

### Acessar Zipkin UI
Abra seu navegador em: http://localhost:9411

### Como Visualizar
1. **Execute algumas requisições** nos endpoints acima
2. **Acesse o Zipkin** em http://localhost:9411
3. **Clique em "Run Query"** para ver os traces recentes
4. **Clique em um trace** para ver detalhes da requisição distribuída

### O que Você Verá
- **Trace completo**: Da requisição no Serviço A até a resposta
- **Spans individuais**: 
  - `handle-cep`: Processamento no Serviço A
  - `forward-to-service-b`: Comunicação entre serviços
  - `handle-weather`: Processamento no Serviço B
  - `get-location-from-cep`: Consulta ViaCEP
  - `get-temperature`: Consulta WeatherAPI
- **Timing detalhado**: Tempo gasto em cada operação
- **Propagação de contexto**: Como o trace flui entre serviços

## 🔧 Endpoints Disponíveis

### Serviço A
- `POST /cep` - Recebe CEP e encaminha para processamento
- `GET /health` - Health check

### Serviço B  
- `POST /weather` - Processa CEP e retorna dados meteorológicos
- `GET /health` - Health check

### Zipkin
- `GET /` - Interface web (http://localhost:9411)

## 🐛 Troubleshooting

### Problema: Containers não iniciam
```bash
# Verificar logs
docker-compose logs

# Rebuild forçado
docker-compose down
docker-compose up --build --force-recreate
```

### Problema: Sem traces no Zipkin
1. Verifique se o Zipkin está acessível: http://localhost:9411
2. Execute requisições nos serviços
3. Aguarde alguns segundos e clique "Run Query" no Zipkin

### Problema: Erro de conexão entre serviços
- Verifique se todos os containers estão rodando: `docker-compose ps`
- Verifique logs específicos: `docker-compose logs servico-a`

### Problema: CEPs não encontrados
- CEPs de teste que funcionam: `01310100`, `20040020`, `30112000`
- Para outros CEPs, verifique no [ViaCEP](https://viacep.com.br/)

## 📁 Estrutura do Projeto

```
Observabilidade/
├── servico-a/              # Serviço de Input
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── servico-b/              # Serviço de Orquestração
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── otel-config/            # Configuração OpenTelemetry
│   └── otel-collector-config.yaml
├── docker-compose.yml      # Orquestração dos serviços
├── README.md              # Esta documentação
└── requisitos.txt         # Especificações do projeto
```

## 🔍 Códigos de Resposta HTTP

| Código | Cenário | Mensagem |
|--------|---------|----------|
| 200 | Sucesso | Dados de temperatura |
| 404 | CEP não encontrado | "can not find zipcode" |
| 422 | CEP inválido | "invalid zipcode" |
| 500 | Erro interno | "internal server error" |

## 🧪 Validações Implementadas

### Serviço A
- ✅ CEP deve ser string
- ✅ CEP deve conter exatamente 8 dígitos
- ✅ Encaminhamento via HTTP para Serviço B
- ✅ Tracing completo da requisição

### Serviço B
- ✅ Validação de formato do CEP
- ✅ Consulta na API ViaCEP
- ✅ Tratamento de CEP não encontrado
- ✅ Consulta de temperatura (WeatherAPI + fallback simulado)
- ✅ Conversões de temperatura (C → F, C → K)
- ✅ Spans para medição de tempo de APIs externas

## 🌡️ Fórmulas de Conversão

- **Celsius para Fahrenheit**: `F = C × 1.8 + 32`
- **Celsius para Kelvin**: `K = C + 273.15`

## 🔗 APIs Utilizadas

- **ViaCEP**: https://viacep.com.br/ (Consulta de localização por CEP)
- **WeatherAPI**: https://www.weatherapi.com/ (Dados meteorológicos)