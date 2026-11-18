# Delcom Blog — Backend (Golang Gin + PostgreSQL)

```bash
cp .env.example .env
psql -U postgres -c "CREATE DATABASE pabwe_p9_blog;"
go mod tidy            # mengunduh dependensi & membuat go.sum
go run cmd/server/main.go     # http://localhost:8081 , Swagger: /swagger/index.html
go test -v -cover ./internal/...
go build -o server.exe cmd/server/main.go
```
Akun bawaan: `admin/admin` (admin) dan `member/member` (member).
Spesifikasi OpenAPI ada di `docs/` (dilayani dinamis di `/api.json`).
