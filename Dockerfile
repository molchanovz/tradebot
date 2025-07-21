FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .. .

ENV CGO_ENABLED=0

RUN go build ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache \
    curl\
    mupdf-dev\
    mupdf-tools\
    libx11 \
    glib \
    fontconfig \
    freetype \
    ca-certificates \
    wget \
    tar \
    xz

RUN wget https://github.com/pdfcpu/pdfcpu/releases/download/v0.11.0/pdfcpu_0.11.0_Linux_x86_64.tar.xz && \
    mkdir -p /tmp/pdfcpu && \
    tar -xJf pdfcpu_0.11.0_Linux_x86_64.tar.xz -C /tmp/pdfcpu --strip-components=1 && \
    mv /tmp/pdfcpu/pdfcpu /usr/local/bin/ && \
    rm -rf /tmp/pdfcpu pdfcpu_0.11.0_Linux_x86_64.tar.xz && \
    chmod +x /usr/local/bin/pdfcpu


COPY --from=builder /app/main .
COPY pkg/client/googlesheet/credentials.json .
COPY pkg/client/googlesheet/token.json .
COPY assets/font.ttf .
COPY assets/barcodes pkg/barcodes

CMD ["./main"]