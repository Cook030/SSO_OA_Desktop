# audit-consumer 构建镜像
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/audit-consumer ./cmd/consumer

FROM alpine:3.20

ENV TZ=Asia/Shanghai
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/audit-consumer /usr/local/bin/audit-consumer
COPY config/config.example.yaml ./config/config.yaml

EXPOSE 8080
ENTRYPOINT ["audit-consumer", "-c", "config/config.yaml"]
