# bizmopol

## Manual usage

Start backend
```bash
cd backend
go mod tidy
go run cmd/api/main.go
```

Start frontend
```bash
cd frontend
npm install
npm run dev
```

## Automatic usage (recommended)
Just run 
```bash
docker compose -f docker-comopose.yml up
```