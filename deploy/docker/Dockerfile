FROM node:20-alpine AS web
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/out ./web/out
RUN go build -o /out/testpay ./cmd/testpay

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/testpay .
COPY deploy/config/testpay.render.yaml /etc/testpay/testpay.yaml
EXPOSE 7700
CMD ["./testpay", "start", "--config", "/etc/testpay/testpay.yaml", "--no-dashboard"]
