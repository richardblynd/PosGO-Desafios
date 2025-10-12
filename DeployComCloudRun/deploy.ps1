# Script para deploy no Google Cloud Run (Windows PowerShell)
# Certifique-se de ter o gcloud CLI instalado e autenticado

# Configurações
$PROJECT_ID = "your-project-id"
$SERVICE_NAME = "weather-api"
$REGION = "us-central1"
$WEATHER_API_KEY = "your-weather-api-key"

Write-Host "Building Docker image..." -ForegroundColor Green
docker build -t "gcr.io/$PROJECT_ID/$SERVICE_NAME" .

Write-Host "Pushing image to GCR..." -ForegroundColor Green
docker push "gcr.io/$PROJECT_ID/$SERVICE_NAME"

Write-Host "Deploying to Cloud Run..." -ForegroundColor Green
gcloud run deploy $SERVICE_NAME `
  --image "gcr.io/$PROJECT_ID/$SERVICE_NAME" `
  --platform managed `
  --region $REGION `
  --allow-unauthenticated `
  --set-env-vars "WEATHER_API_KEY=$WEATHER_API_KEY" `
  --port 8080

Write-Host "Deployment completed!" -ForegroundColor Green