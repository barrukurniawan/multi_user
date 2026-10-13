# Task Management API (Multi-User)

REST API untuk mengelola task - Created by Barru Kurniawan

[![Swagger](https://img.shields.io/badge/Swagger-API_Docs-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)](https://barrukurniawan.web.id/swagger)
[![Postman](https://img.shields.io/badge/Postman-Collection-FF6C37?style=for-the-badge&logo=postman&logoColor=white)](docs/Tasks.postman_collection.json)

## Tech Stack
- Language : Go 1.26
- Framework : Gin
- Database : PostgreSQL
- ORM : GORM
- Cache : Redis (untuk Idempotency-Key)
- Auth : JWT

## Cara Menjalankan

### 1. Siapkan `.env`
```
cp .env.example .env
```
Sesuaikan isi `.env` dengan setting PostgreSQL & Redis di komputer kamu. `JWT_SECRET` wajib diisi.

### 2. Buat database
Via terminal:
```
createdb user_multi
```
Atau via psql:
```
psql -d postgres -c "CREATE DATABASE user_multi;"
```

### 3. Jalankan Redis
```
redis-server
```

### 4. Jalankan aplikasi
```
go mod tidy
go run .
```
Migration (`migrations/schema.sql`) otomatis dijalankan saat aplikasi start. Kalau mau jalankan manual:
```
psql -d user_multi -f migrations/schema.sql
```

Server jalan di `http://localhost:8080`.

### 5. Jalankan test
Test tidak butuh database / Redis (pakai mock).
```
go test -race ./tests/...
```
Jika ingin kondisi perulangan, cth : 30x
```
go test -race -count=30 ./tests/
```
Jika ingin test semua file & tampil detailnya
```
go test -v ./tests/
```

## Daftar Endpoint

| Method | Endpoint | Auth | Keterangan |
|--------|----------|------|------------|
| GET | `/health` | - | Cek server hidup |
| POST | `/api/v1/auth/register` | - | Register user baru |
| POST | `/api/v1/auth/login` | - | Login, dapat JWT token |
| POST | `/api/v1/tasks` | Bearer | Buat task (wajib header `Idempotency-Key`) |
| GET | `/api/v1/tasks` | Bearer | List task (`?status=&search=&page=&limit=`) |
| GET | `/api/v1/tasks/:id` | Bearer | Detail task |
| PUT | `/api/v1/tasks/:id` | Bearer | Update task |
| DELETE | `/api/v1/tasks/:id` | Bearer | Hapus task |
| POST | `/api/v1/tasks/:id/assign` | Bearer | Assign task ke user lain dalam tim yang sama |

Status task: `todo`, `in_progress`, `done`.

### Contoh request

Register:
```
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Barru","email":"barru@example.com","password":"rahasia123","team_name":"backend"}'
```

Login:
```
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"barru@example.com","password":"rahasia123"}'
```

Create task:
```
curl -X POST http://localhost:8080/api/v1/tasks \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: <uuid-baru>' \
  -d '{"title":"Belajar Go","description":"Ketik ulang project","status":"todo"}'
```

### Format error
Semua error dibalas dengan format JSON yang sama:
```json
{
  "status": "error",
  "code": "TASK_NOT_FOUND",
  "message": "task tidak ditemukan",
  "timestamp": "2026-10-13T07:21:25Z"
}
```

## Handling Idempotency

Ketika klik tombol "Buat Task" berkali-kali , internet lemot, jadi aplikasi mengirim ulang request yang sama. Tanpa idempotency, task-nya bisa terbuat 2x bahkan lebih. Handling Idempotency mencegahnya.

Caranya, setiap request `POST /tasks` wajib membawa header `Idempotency-Key` berisi UUID, semacam "nomor antrean" untuk request tersebut. Disimpan di Redis hingga **24 jam**. Setiap user punya key Redis-nya sendiri agar tidak bentrok dengan user lain.


## Arsitektur

Alur request:
```
Router -> Middleware -> Controller -> Service -> Repository -> PostgreSQL / Redis
```

Struktur folder:
```
main.go              -> load config, koneksi DB & Redis, rakit semua dependency
migrations/          -> schema.sql
core/
  config/            -> baca .env
  routers/           -> daftar endpoint
  controllers/       -> terima request, validasi input, kirim response
  services/          -> business logic (auth, task, idempotency, assign, notifikasi)
  repositories/      -> query ke PostgreSQL & Redis
  models/            -> struct tabel
  dto/               -> struct request & response
  helpers/           -> JWT, middleware, format error
tests/               -> unit test idempotency, auth, error handling, logging (pakai mock)
```

Catatan beberapa fitur:
- **Logging** : Setiap request menghasilkan log JSON berisi `request_id`, `method`, `path`, `status_code`, `latency_ms`. Level INFO untuk 2xx/3xx, WARN untuk 4xx, ERROR untuk 5xx.
- **Error handling** : Panic dicover oleh `RecoveryMiddleware` dan dibalas 500 dalam format JSON, tanpa stack trace di response (stack trace hanya ditulis ke log).
- **Team** : Team dibuat otomatis saat register berdasarkan `team_name`. User dengan `team_name` yang sama dianggap satu tim.
