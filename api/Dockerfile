FROM golang:1.25-bookworm

# Install gcc, base libraries, and 'file' for diagnostics
RUN apt-get update && apt-get install -y gcc libc6-dev file

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compile forcing Intel 64-bits
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bank-api ./src/main.go
RUN chmod +x bank-api

EXPOSE 8080

# This CMD will print the information before executing to see where the crash is
CMD echo "--- CONTAINER INFO ---" && uname -m && \
    echo "--- BINARY INFO ---" && file ./bank-api && \
    echo "--- STARTING API ---" && ./bank-api