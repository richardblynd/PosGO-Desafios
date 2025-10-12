# Weather API - Consulta de Clima por CEP

Este projeto implementa uma REST API em Go que recebe um CEP brasileiro, identifica a cidade correspondente e retorna a temperatura atual em Celsius, Fahrenheit e Kelvin.

Deploy realizado no Google Cloud Run na URL:  
https://weather-api-1066335039627.us-central1.run.app

## 🚀 Funcionalidades

- ✅ Validação de CEP (8 dígitos)
- ✅ Integração com API viaCEP para buscar localização
- ✅ Integração com WeatherAPI para obter dados meteorológicos
- ✅ Conversão automática de temperaturas (°C, °F, K)
- ✅ Respostas padronizadas com códigos HTTP apropriados
- ✅ Testes automatizados
- ✅ Containerização com Docker
- ✅ Pronto para Google Cloud Run

## 📋 Requisitos

### Para executar localmente:
- Go 1.21+
- Docker e Docker Compose (opcional)

### Para deploy na nuvem:
- Conta Google Cloud Platform
- gcloud CLI configurado
- Chave de API do WeatherAPI

## 🔧 Configuração

### 1. Obter chave da WeatherAPI

1. Acesse [WeatherAPI](https://www.weatherapi.com/)
2. Crie uma conta gratuita
3. Obtenha sua API Key no dashboard

### 2. Configurar variáveis de ambiente

Copie o arquivo de exemplo:
```bash
cp .env.example .env
```

Edite o arquivo `.env` e configure sua chave:
```env
WEATHER_API_KEY=sua_chave_aqui
PORT=8080
```

## 🏃‍♂️ Como Executar

### Opção 1: Execução Local (Go)

```bash
# Instalar dependências
go mod tidy

# Executar aplicação
go run main.go
```

### Opção 2: Docker Compose (Recomendado)

```bash
# Build e execução
docker-compose up --build

# Execução em background
docker-compose up -d
```

### Opção 3: Docker Manual

```bash
# Build da imagem
docker build -t weather-api .

# Executar container
docker run -p 8080:8080 -e WEATHER_API_KEY=sua_chave weather-api
```

## 📡 Endpoints da API

### GET /weather/{zipcode}

Retorna informações de temperatura para o CEP informado.

**Parâmetros:**
- `zipcode`: CEP brasileiro de 8 dígitos (com ou sem hífen)

**Exemplos de requisição:**
```bash
# CEP válido
curl http://localhost:8080/weather/01310100

# CEP com hífen (será removido automaticamente)
curl http://localhost:8080/weather/01310-100
```

**Respostas:**

#### ✅ Sucesso (200 OK)
```json
{
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.5
}
```

#### ❌ CEP Inválido (422 Unprocessable Entity)
```json
{
  "message": "invalid zipcode"
}
```

#### ❌ CEP Não Encontrado (404 Not Found)
```json
{
  "message": "can not find zipcode"
}
```

## 🧪 Executar Testes

### Testes Unitários

```bash
# Todos os testes
go test ./...

# Testes com detalhes
go test -v ./...

# Testes com coverage
go test -cover ./...
```

### Testes da API (Arquivos .http)

O projeto inclui arquivos `.http` para testar a API diretamente no VS Code:

- **`api-tests.http`**: Testes locais completos (casos de sucesso e erro)
- **`production-tests.http`**: Testes para ambiente de produção
- **`performance-tests.http`**: Testes de performance e edge cases

#### Como usar:

1. **Instalar extensão REST Client no VS Code**
2. **Executar a aplicação localmente:**
   ```bash
   docker-compose up
   ```
3. **Abrir qualquer arquivo `.http`**
4. **Clicar em "Send Request" acima de cada requisição**

#### Exemplos de teste:
```http
### CEP válido - São Paulo/SP
GET http://localhost:8080/weather/01310100

### CEP inválido - muito curto
GET http://localhost:8080/weather/0131010

### CEP não encontrado
GET http://localhost:8080/weather/99999999
```

## 🚀 Deploy no Google Cloud Run

### Pré-requisitos

1. **Instalar Google Cloud CLI:**
   - [Instruções oficiais](https://cloud.google.com/sdk/docs/install)

2. **Autenticar:**
   ```bash
   gcloud auth login
   gcloud config set project SEU_PROJECT_ID
   ```

3. **Habilitar APIs necessárias:**
   ```bash
   gcloud services enable run.googleapis.com
   gcloud services enable containerregistry.googleapis.com
   ```

### Deploy Manual

#### Linux/Mac:
```bash
chmod +x deploy.sh
./deploy.sh
```

#### Windows (PowerShell):
```powershell
.\deploy.ps1
```

Antes de executar, edite os scripts com suas configurações:
- `PROJECT_ID`: ID do seu projeto GCP
- `WEATHER_API_KEY`: Sua chave da WeatherAPI

### Deploy Automático (GitHub Actions)

1. **Configure secrets no GitHub:**
   - `GCP_PROJECT_ID`: ID do projeto
   - `GCP_SA_KEY`: JSON da service account
   - `WEATHER_API_KEY`: Chave da WeatherAPI

2. **Push para branch main:**
   ```bash
   git push origin main
   ```

O deploy será executado automaticamente via GitHub Actions.

## 🏗️ Estrutura do Projeto

```
.
├── main.go                          # Ponto de entrada da aplicação
├── go.mod                           # Dependências Go
├── go.sum                           # Checksums das dependências
├── Dockerfile                       # Configuração Docker
├── docker-compose.yml               # Orquestração local
├── deploy.sh                        # Script deploy Linux/Mac
├── deploy.ps1                       # Script deploy Windows
├── .env.example                     # Exemplo de variáveis
├── api-tests.http                   # Testes API locais
├── production-tests.http            # Testes API produção
├── performance-tests.http           # Testes de performance
├── .github/
│   └── workflows/
│       └── deploy.yml               # CI/CD GitHub Actions
├── internal/
│   ├── handlers/
│   │   ├── weather.go               # Handlers HTTP
│   │   └── weather_test.go          # Testes dos handlers
│   ├── services/
│   │   ├── temperature.go           # Conversão de temperaturas
│   │   ├── temperature_test.go      # Testes de conversão
│   │   ├── validation.go            # Validação de CEP
│   │   ├── validation_test.go       # Testes de validação
│   │   ├── viacep.go                # Integração viaCEP
│   │   └── weather.go               # Integração WeatherAPI
│   └── models/
│       └── models.go                # Estruturas de dados
└── README.md                        # Este arquivo
```

## 🔧 APIs Utilizadas

### viaCEP
- **URL:** https://viacep.com.br/
- **Propósito:** Buscar informações de localização por CEP
- **Gratuita:** Sim
- **Limites:** Não documentados

### WeatherAPI
- **URL:** https://www.weatherapi.com/
- **Propósito:** Dados meteorológicos atuais
- **Gratuita:** Sim (até 1 milhão de calls/mês)
- **Requer:** Chave de API

## 🧮 Fórmulas de Conversão

- **Celsius → Fahrenheit:** `F = C × 1.8 + 32`
- **Celsius → Kelvin:** `K = C + 273`