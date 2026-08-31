# ---------- 构建阶段 ----------
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 先复制依赖清单，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /sso-server ./cmd/server

# ---------- 运行阶段 ----------
FROM alpine:3.20

RUN apk add --no-cache tzdata ca-certificates \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /sso-server ./sso-server
# 兜底配置：生产请通过环境变量（JWT_SECRET、SSO_INTERNAL_TOKEN）注入敏感项，
# 或挂载自定义 config/config.yaml 覆盖
COPY config/config.example.yaml ./config/config.yaml

ENV TZ=Asia/Shanghai

EXPOSE 8080

ENTRYPOINT ["./sso-server"]
