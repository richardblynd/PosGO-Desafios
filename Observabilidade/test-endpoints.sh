#!/bin/bash

# Script para testar os endpoints do sistema
# Certifique-se de que o sistema está rodando: docker-compose up

BASE_URL="http://localhost:8080"

echo "🧪 Testando Sistema de Temperatura por CEP"
echo "=========================================="

echo -e "\n1️⃣ Teste com CEP válido (01310-100 - São Paulo)"
curl -X POST $BASE_URL/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}' \
  -w "\nStatus: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || echo "Resposta recebida (instale jq para formatação JSON)"

echo -e "\n2️⃣ Teste com CEP válido (20040-020 - Rio de Janeiro)" 
curl -X POST $BASE_URL/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "20040020"}' \
  -w "\nStatus: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || echo "Resposta recebida"

echo -e "\n3️⃣ Teste com CEP inválido (poucos dígitos)"
curl -X POST $BASE_URL/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}' \
  -w "\nStatus: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || echo "Resposta recebida"

echo -e "\n4️⃣ Teste com CEP inexistente"
curl -X POST $BASE_URL/cep \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}' \
  -w "\nStatus: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || echo "Resposta recebida"

echo -e "\n5️⃣ Teste health check Serviço A"
curl -X GET http://localhost:8080/health \
  -w "\nStatus: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || echo "Resposta recebida"

echo -e "\n6️⃣ Teste health check Serviço B"
curl -X GET http://localhost:8081/health \
  -w "\nStatus: %{http_code}\n" \
  -s | jq '.' 2>/dev/null || echo "Resposta recebida"

echo -e "\n🔍 Para visualizar os traces:"
echo "   Acesse: http://localhost:9411"
echo "   Clique em 'Run Query' para ver os traces das requisições acima"
echo -e "\n✅ Testes concluídos!"