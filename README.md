# Delcom Blog API (Golang Gin & PostgreSQL)

Backend REST API untuk aplikasi **Delcom Blog — Platform Publikasi Artikel Multi-Tier & AI Generator dengan Integrasi Search Engine** menggunakan Golang Gin, PostgreSQL 17, GORM, Clean Architecture (Feature-First), dan TDD.

## Menjalankan Server

1. Pastikan PostgreSQL berjalan dan database `pabwe_p9_blog` telah dibuat.
2. Salin file lingkungan:
   ```bash
   cp .env.example .env
   ```
3. Jalankan server:
   ```bash
   go run main.go
   ```
   Server aktif pada: `http://localhost:8081`
   Swagger UI: `http://localhost:8081/swagger/index.html`

## Menjalankan Pengujian (TDD & 100% Coverage)

```bash
go test -v -cover ./internal/...
```
