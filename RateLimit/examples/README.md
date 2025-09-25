# Exemplos de Requisições HTTP

Esta pasta contém arquivos de exemplo para testar o rate limiter usando ferramentas como REST Client do VS Code.

## 📁 Arquivos Disponíveis

### `requests.http`
Requisições básicas para testar todas as funcionalidades principais:
- Health check
- Endpoints básicos
- Diferentes tokens (abc123, xyz789, premium, etc.)
- Formato Authorization Bearer
- Requisições múltiplas para teste de limite

### `rate-limit-scenarios.http`  
Cenários específicos de rate limiting:
- Teste de limite por IP
- Token sobrepondo IP
- Bloqueio temporal
- Diferentes limites por token
- Headers de rate limiting
- Simulação de proxy (X-Forwarded-For)
- Teste de carga básico
- Validação de resposta de erro

### `performance-tests.http`
Testes de performance e stress:
- Requisições sequenciais
- Simulação de concorrência
- Diferentes endpoints
- Headers e extração de IP
- Tokens inválidos/especiais
- Monitoramento de headers

### `test-config.env`
Configuração otimizada para testes:
- Limites baixos para demonstrar bloqueio
- Tempos de bloqueio curtos
- Configurações para diferentes cenários

## 🚀 Como Usar

### No VS Code com REST Client

1. Instale a extensão "REST Client" no VS Code
2. Abra qualquer arquivo `.http`
3. Clique em "Send Request" acima de cada requisição
4. Observe os headers de resposta e status codes

### Com curl (manual)

```bash
# Exemplo básico
curl -v http://localhost:8080/api/v1/users

# Com token
curl -v -H "API_KEY: premium" http://localhost:8080/api/v1/users

# Para ver headers de rate limiting
curl -v -H "API_KEY: abc123" http://localhost:8080/api/v1/users
```

### Com scripts automatizados

Use os scripts na raiz do projeto:
```bash
# Windows
test-rate-limit.bat

# Linux/Mac
./test-rate-limit.sh
```

## 📊 Configuração Recomendada para Testes

Para ver o rate limiting em ação rapidamente, use a configuração em `test-config.env`:

```bash
# Copie a configuração de teste
cp examples/test-config.env .env

# Reinicie o servidor
go run cmd/server/main.go
```

Com esta configuração:
- **IP**: 3 req/s, bloqueio 30s
- **Token padrão**: 5 req/s, bloqueio 1min  
- **Token abc123**: 8 req/s, bloqueio 45s
- **Token premium**: 50 req/s, bloqueio 10s

## 🧪 Cenários de Teste

### 1. Rate Limiting Básico
Execute `rate-limit-scenarios.http` > "Cenário 1" rapidamente para ver bloqueio.

### 2. Prioridade de Token
1. Exceda o limite por IP 
2. Use requisição com token - deve funcionar

### 3. Diferentes Tokens
Compare limites entre `abc123`, `xyz789` e `premium`.

### 4. Headers de Monitoramento
Observe headers `X-RateLimit-*` nas respostas:
- `X-RateLimit-Limit`: Limite por segundo
- `X-RateLimit-Remaining`: Requisições restantes
- `X-RateLimit-Reset`: Próximo reset

### 5. Simulação de Proxy
Use headers `X-Forwarded-For` para simular diferentes IPs.

## 💡 Dicas

- **Para desenvolvimento**: Use `STORAGE_TYPE=memory`
- **Para produção**: Use `STORAGE_TYPE=redis` 
- **Para testes rápidos**: Configure limites baixos (3-5 req/s)
- **Para stress test**: Configure limites altos (100+ req/s)

## 🔍 Troubleshooting

### Rate limiting não funciona
- Verifique se o servidor está rodando na porta 8080
- Confirme a configuração no `.env`
- Execute requisições rapidamente (< 1 segundo entre elas)

### Sempre retorna 200
- Limites muito altos, reduza `IP_RATE_LIMIT` para 2-3
- Use requisições sequenciais rápidas

### Headers não aparecem
- Alguns clientes HTTP não mostram todos os headers
- Use `curl -v` para ver headers completos