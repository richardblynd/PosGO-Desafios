
FROM golang:1.24.2-alpine AS base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM base AS build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app .

FROM gcr.io/distroless/static-debian12

COPY --from=build /bin/app /app

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["./app"]


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