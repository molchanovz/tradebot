FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -o /bin/main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache \
    curl\
    mupdf-dev \
    mupdf-tools \
    libx11 \
    glib \
    fontconfig \
    freetype \
    ca-certificates \
    wget \
    tar \
    xz

RUN wget https://github.com/pdfcpu/pdfcpu/releases/download/v0.10.2/pdfcpu_0.10.2_Linux_x86_64.tar.xz && \
    mkdir -p /tmp/pdfcpu && \
    tar -xJf pdfcpu_0.10.2_Linux_x86_64.tar.xz -C /tmp/pdfcpu --strip-components=1 && \
    mv /tmp/pdfcpu/pdfcpu /usr/local/bin/ && \
    rm -rf /tmp/pdfcpu pdfcpu_0.10.2_Linux_x86_64.tar.xz && \
    chmod +x /usr/local/bin/pdfcpu


COPY --from=builder /app/main .
COPY credentials.json .
COPY token.json .
COPY font.ttf .
COPY pkg/barcodes pkg/barcodes

CMD ["./main"]