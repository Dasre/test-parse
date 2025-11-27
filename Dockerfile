FROM golang:1.21-alpine AS builder

WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.* ./

# 下載依賴
RUN go mod download

# 複製源碼
COPY . .

# 編譯
RUN CGO_ENABLED=0 go build -o validator ./cmd/validator

# 最終映像
FROM alpine:3.19

# 複製執行檔
COPY --from=builder /app/validator /validator

# 複製產品配置
COPY products.yaml /products.yaml

# 複製規則
COPY rules /rules

WORKDIR /workspace

ENTRYPOINT ["/validator"]
