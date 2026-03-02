# Script para deploy no Google Cloud Run (Windows PowerShell)
# Certifique-se de ter o gcloud CLI instalado e autenticado

# Configurações
$PROJECT_ID = "avid-influence-474915-r5"
$SERVICE_NAME = "weather-api"
$REGION = "us-central1"
$WEATHER_API_KEY = "c6f326b808d94a238f7144531251210"

Write-Host "Building Docker image..." -ForegroundColor Green
docker build -t "gcr.io/$PROJECT_ID/$SERVICE_NAME" . --no-cache

Write-Host "Pushing image to GCR..." -ForegroundColor Green
docker push "gcr.io/$PROJECT_ID/$SERVICE_NAME"

Write-Host "Deploying to Cloud Run..." -ForegroundColor Green
gcloud run deploy $SERVICE_NAME `
  --image "gcr.io/$PROJECT_ID/$SERVICE_NAME" `
  --platform managed `
  --region $REGION `
  --allow-unauthenticated `
  --port 8080
  # --set-env-vars "WEATHER_API_KEY=$WEATHER_API_KEY"

Write-Host "Deployment completed!" -ForegroundColor Green