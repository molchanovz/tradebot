FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

# Настройка сети и установка пакетов
RUN echo "network.ipv6.conf.all.disable_ipv6=1" >> /etc/sysctl.conf && \
    echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "https://mirror.yandex.ru/mirrors/alpine/v3.21/main" > /etc/apk/repositories && \
    echo "https://mirror.yandex.ru/mirrors/alpine/v3.21/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
    wget \
    tar \
    xz \
    mupdf-dev \
    mupdf-tools \
    libx11 \
    glib \
    fontconfig \
    freetype \
    ca-certificates

# Установка pdfcpu с прогресс-баром
RUN wget -q --show-progress https://github.com/pdfcpu/pdfcpu/releases/download/v0.10.2/pdfcpu_0.10.2_Linux_x86_64.tar.xz && \
    mkdir -p /tmp/pdfcpu && \
    tar -xJf pdfcpu_0.10.2_Linux_x86_64.tar.xz -C /tmp/pdfcpu --strip-components=1 && \
    mv /tmp/pdfcpu/pdfcpu /usr/local/bin/ && \
    rm -rf /tmp/pdfcpu pdfcpu_0.10.2_Linux_x86_64.tar.xz && \
    chmod +x /usr/local/bin/pdfcpu

# Копирование файлов
COPY --from=builder /app/main .
COPY credentials.json token.json font.ttf .
COPY pkg/barcodes pkg/barcodes

HEALTHCHECK --interval=30s --timeout=3s \
  CMD pgrep main || exit 1

CMD ["./main"]