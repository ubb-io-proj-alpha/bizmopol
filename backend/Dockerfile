FROM debian:trixie-slim

RUN apt-get update && apt-get install -y \
    golang-go \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

RUN go install github.com/air-verse/air@latest

RUN go mod download
RUN go mod tidy

EXPOSE 8080

CMD ["/root/go/bin/air", "-c", ".air.toml"]
#CMD ["go", "run", "cmd/api/main.go"]
