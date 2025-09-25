#!/bin/bash

# Script para testar o rate limiter
# Usage: ./test-rate-limit.sh

BASE_URL="http://localhost:8080"
API_ENDPOINT="$BASE_URL/api/v1/users"

echo "🚀 Testando Rate Limiter"
echo "========================="

# Função para fazer requisição e mostrar resultado
make_request() {
    local url=$1
    local headers=$2
    local description=$3
    
    echo "📋 $description"
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "Status: %{http_code} | Time: %{time_total}s" -H "$headers" "$url")
    else
        response=$(curl -s -w "Status: %{http_code} | Time: %{time_total}s" "$url")
    fi
    
    echo "   $response"
    echo
}

# Teste 1: Health Check
echo "🔍 Teste 1: Health Check"
make_request "$BASE_URL/health" "" "Verificando saúde do serviço"

# Teste 2: Requisições normais (IP)
echo "🔍 Teste 2: Limitação por IP"
echo "Fazendo 5 requisições seguidas (limite padrão: 10)..."
for i in {1..5}; do
    make_request "$API_ENDPOINT" "" "Requisição $i por IP"
    sleep 0.1
done

# Teste 3: Requisições com token
echo "🔍 Teste 3: Requisições com Token"
make_request "$API_ENDPOINT" "API_KEY: abc123" "Requisição com token abc123"
make_request "$API_ENDPOINT" "API_KEY: premium" "Requisição com token premium"

# Teste 4: Exceder limite (se configurado baixo)
echo "🔍 Teste 4: Testando limite (configure IP_RATE_LIMIT=2 para ver bloqueio)"
echo "Fazendo 4 requisições rápidas..."
for i in {1..4}; do
    make_request "$API_ENDPOINT" "" "Requisição rápida $i"
done

echo "✅ Testes concluídos!"
echo
echo "💡 Dicas:"
echo "- Configure IP_RATE_LIMIT=2 no .env para ver bloqueio mais facilmente"
echo "- Use diferentes tokens para testar limites específicos"
echo "- Monitore os headers X-RateLimit-* nas respostas"