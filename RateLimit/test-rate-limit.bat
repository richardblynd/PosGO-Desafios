@echo off
REM Script para testar o rate limiter no Windows
REM Usage: test-rate-limit.bat

set BASE_URL=http://localhost:8080
set API_ENDPOINT=%BASE_URL%/api/v1/users

echo 🚀 Testando Rate Limiter
echo =========================
echo.

echo 🔍 Teste 1: Health Check
curl -s "%BASE_URL%/health"
echo.
echo.

echo 🔍 Teste 2: Limitacao por IP
echo Fazendo 5 requisicoes seguidas...
for /L %%i in (1,1,5) do (
    echo Requisicao %%i:
    curl -s -w "Status: %%{http_code}" "%API_ENDPOINT%"
    echo.
    timeout /t 1 /nobreak >nul
)
echo.

echo 🔍 Teste 3: Requisicoes com Token
echo Com token abc123:
curl -s -w "Status: %%{http_code}" -H "API_KEY: abc123" "%API_ENDPOINT%"
echo.
echo Com token premium:
curl -s -w "Status: %%{http_code}" -H "API_KEY: premium" "%API_ENDPOINT%"
echo.
echo.

echo ✅ Testes concluidos!
echo.
echo 💡 Dicas:
echo - Configure IP_RATE_LIMIT=2 no .env para ver bloqueio mais facilmente
echo - Use diferentes tokens para testar limites especificos
echo - Monitore os headers X-RateLimit-* nas respostas
pause