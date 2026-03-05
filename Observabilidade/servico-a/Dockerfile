FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar todos os arquivos
COPY . .

# Gerar go.sum, baixar dependências e compilar
RUN go mod tidy && go mod download && CGO_ENABLED=0 GOOS=linux go build -o main .

# Estágio final
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copiar binário
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]