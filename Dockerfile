# -----------------------------
# СТАДИЯ СБОРКИ (Go + MuPDF)
# -----------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Установим зависимости для сборки Go и MuPDF
RUN apk add --no-cache \
    build-base \
    git \
    wget \
    tar \
    xz \
    cmake

# Скачиваем и собираем MuPDF
ENV MUPDF_VERSION=1.25.6

RUN wget https://mupdf.com/downloads/archive/mupdf-${MUPDF_VERSION}-source.tar.gz && \
    tar -xzf mupdf-${MUPDF_VERSION}-source.tar.gz && \
    cd mupdf-${MUPDF_VERSION}-source && \
    make build=release && \
    mkdir -p /usr/local/lib /usr/local/include && \
    cp build/release/libmupdf.a /usr/local/lib/ && \
    cp -r include/mupdf /usr/local/include/ && \
    cd .. && rm -rf mupdf-${MUPDF_VERSION}-source*

# Установка переменных окружения для сборки Go с CGO
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-I/usr/local/include"
ENV CGO_LDFLAGS="-L/usr/local/lib -lmupdf"

# Копируем и устанавливаем зависимости Go
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY .. .

# Сборка приложения
RUN go build -o main ./cmd/main.go

# -----------------------------
# ФИНАЛЬНЫЙ МИНИМАЛЬНЫЙ ОБРАЗ
# -----------------------------
FROM alpine:latest

WORKDIR /app

# Установка необходимых зависимостей для запуска
RUN apk add --no-cache \
    libx11 \
    glib \
    fontconfig \
    freetype \
    ca-certificates \
    wget \
    tar \
    xz \
    libjpeg-turbo \
    openjpeg \
    jbig2dec \
    harfbuzz \
    zlib

# Установка pdfcpu (для работы с PDF)
RUN wget https://github.com/pdfcpu/pdfcpu/releases/download/v0.11.0/pdfcpu_0.11.0_Linux_x86_64.tar.xz && \
    mkdir -p /tmp/pdfcpu && \
    tar -xJf pdfcpu_0.11.0_Linux_x86_64.tar.xz -C /tmp/pdfcpu --strip-components=1 && \
    mv /tmp/pdfcpu/pdfcpu /usr/local/bin/ && \
    rm -rf /tmp/pdfcpu pdfcpu_0.11.0_Linux_x86_64.tar.xz && \
    chmod +x /usr/local/bin/pdfcpu

COPY --from=builder /app/main .
COPY pkg/google/utils/credentials.json .
COPY pkg/google/utils/token.json .
COPY assets/font.ttf .
COPY assets/barcodes pkg/barcodes

CMD ["./main"]
