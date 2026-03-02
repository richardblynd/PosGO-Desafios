
# Estágio 1: Base
# Sugestão: use uma versão estável e conhecida do Go. 
# Se você realmente precisa da 1.24.x, tente apenas 'golang:1.24-alpine'
FROM golang:1.22-alpine AS base

WORKDIR /app

# O asterisco no go.sum* evita que o build quebre caso o arquivo não exista
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Estágio 2: Build
FROM base AS build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app .

# Estágio 3: Imagem final (Distroless)
FROM gcr.io/distroless/static-debian12

# Copia o binário para a raiz da imagem final
COPY --from=build /bin/app /app

ENV PORT=8080
EXPOSE 8080

# Usa o caminho absoluto para executar o binário
ENTRYPOINT ["/app"]


# Build stage
# FROM golang:1.24.2-alpine AS builder
# WORKDIR /app
# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# # Final stage
# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
# WORKDIR /root/
# COPY --from=builder /app/main .
# EXPOSE 8080
# CMD ["./main"]