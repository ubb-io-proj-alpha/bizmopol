FROM debian:trixie-slim

RUN apt-get update && apt-get install -y \
    golang-go \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN mkdir -p /app/data

COPY . .

EXPOSE 8080

CMD ["go", "run", "cmd/api/main.go"]
