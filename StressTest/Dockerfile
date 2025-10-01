# Multi-stage build para reduzir o tamanho da imagem final
FROM golang:1.21-alpine AS builder

# Instalar certificados SSL para HTTPS
RUN apk --no-cache add ca-certificates

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivo go mod
COPY go.mod ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar a aplicação com otimizações
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o stresstest .

# Imagem final mínima
FROM alpine:latest

# Instalar certificados SSL
RUN apk --no-cache add ca-certificates

# Criar usuário não-root
RUN adduser -D -s /bin/sh stresstest

# Definir diretório de trabalho
WORKDIR /app

# Copiar o binário da aplicação do stage de build
COPY --from=builder /app/stresstest .

# Definir o usuário para execução
USER stresstest

# Definir o comando padrão
ENTRYPOINT ["./stresstest"]

# Exemplo de uso:
# docker run --rm stresstest --url=http://google.com --requests=1000 --concurrency=10