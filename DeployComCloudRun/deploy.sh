#!/bin/bash

# Script para deploy no Google Cloud Run
# Certifique-se de ter o gcloud CLI instalado e autenticado

# Configurações
PROJECT_ID="your-project-id"
SERVICE_NAME="weather-api"
REGION="us-central1"
WEATHER_API_KEY="your-weather-api-key"

# Build da imagem Docker
echo "Building Docker image..."
docker build -t gcr.io/$PROJECT_ID/$SERVICE_NAME .

# Push da imagem para Google Container Registry
echo "Pushing image to GCR..."
docker push gcr.io/$PROJECT_ID/$SERVICE_NAME

# Deploy no Cloud Run
echo "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=$WEATHER_API_KEY \
  --port 8080

echo "Deployment completed!"