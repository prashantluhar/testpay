FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/testpay ./cmd/testpay

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/testpay .
COPY deploy/config/testpay.local.yaml /etc/testpay/testpay.yaml
EXPOSE 7700
CMD ["./testpay", "start", "--config", "/etc/testpay/testpay.yaml"]
