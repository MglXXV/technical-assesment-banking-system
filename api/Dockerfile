FROM golang:1.25-bookworm

# Instalamos gcc, librerías base y 'file' para diagnóstico
RUN apt-get update && apt-get install -y gcc libc6-dev file

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilamos forzando Intel 64-bits
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bank-api ./src/main.go
RUN chmod +x bank-api

EXPOSE 8080

# Este CMD imprimirá la información antes de ejecutar para ver dónde está el choque
CMD echo "--- INFO CONTENEDOR ---" && uname -m && \
    echo "--- INFO BINARIO ---" && file ./bank-api && \
    echo "--- INICIANDO API ---" && ./bank-api