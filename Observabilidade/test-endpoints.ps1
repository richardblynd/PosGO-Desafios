# Script PowerShell para testar os endpoints do sistema
# Certifique-se de que o sistema está rodando: docker-compose up

$BaseURL = "http://localhost:8080"

Write-Host "🧪 Testando Sistema de Temperatura por CEP" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

Write-Host "`n1️⃣ Teste com CEP válido (01310-100 - São Paulo)" -ForegroundColor Yellow
$body1 = @{ cep = "01310100" } | ConvertTo-Json
try {
    $response1 = Invoke-RestMethod -Uri "$BaseURL/cep" -Method POST -Body $body1 -ContentType "application/json"
    $response1 | ConvertTo-Json -Depth 3
    Write-Host "Status: 200" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
}

Write-Host "`n2️⃣ Teste com CEP válido (20040-020 - Rio de Janeiro)" -ForegroundColor Yellow
$body2 = @{ cep = "20040020" } | ConvertTo-Json
try {
    $response2 = Invoke-RestMethod -Uri "$BaseURL/cep" -Method POST -Body $body2 -ContentType "application/json"
    $response2 | ConvertTo-Json -Depth 3
    Write-Host "Status: 200" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
}

Write-Host "`n3️⃣ Teste com CEP inválido (poucos dígitos)" -ForegroundColor Yellow
$body3 = @{ cep = "123" } | ConvertTo-Json
try {
    $response3 = Invoke-RestMethod -Uri "$BaseURL/cep" -Method POST -Body $body3 -ContentType "application/json"
    $response3 | ConvertTo-Json -Depth 3
} catch {
    $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
    $errorResponse | ConvertTo-Json -Depth 3
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Orange
}

Write-Host "`n4️⃣ Teste com CEP inexistente" -ForegroundColor Yellow
$body4 = @{ cep = "99999999" } | ConvertTo-Json
try {
    $response4 = Invoke-RestMethod -Uri "$BaseURL/cep" -Method POST -Body $body4 -ContentType "application/json"
    $response4 | ConvertTo-Json -Depth 3
} catch {
    $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
    $errorResponse | ConvertTo-Json -Depth 3
    Write-Host "Status: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Orange
}

Write-Host "`n5️⃣ Teste health check Serviço A" -ForegroundColor Yellow
try {
    $health1 = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    $health1 | ConvertTo-Json -Depth 3
    Write-Host "Status: 200" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n6️⃣ Teste health check Serviço B" -ForegroundColor Yellow
try {
    $health2 = Invoke-RestMethod -Uri "http://localhost:8081/health" -Method GET
    $health2 | ConvertTo-Json -Depth 3
    Write-Host "Status: 200" -ForegroundColor Green
} catch {
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n🔍 Para visualizar os traces:" -ForegroundColor Cyan
Write-Host "   Acesse: http://localhost:9411" -ForegroundColor White
Write-Host "   Clique em 'Run Query' para ver os traces das requisições acima" -ForegroundColor White
Write-Host "`n✅ Testes concluídos!" -ForegroundColor Green