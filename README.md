# Pemrograman Backend Lanjut dengan Go

Repository ini berisi latihan Week 1 dan API Student berbasis PostgreSQL,
lengkap dengan authentication berbasis token.

## Struktur Project

```text
latihan-fiber/
  main.go
  variabel/main.go
  pointer/main.go
  struct/main.go

api-student/
  main.go
  app/
    model/        student.go, user.go, auth.go
    repository/   student_repository.go, user_repository.go, token_repository.go
    service/      student_service.go, student_rules.go, auth_service.go, auth_rules.go
  config/         app.go, env.go, logger.go
  database/       postgres.go
  helper/         context.go, jwt.go, request.go, response.go, security.go
  middleware/     middleware.go, auth.go
  migrations/     001_create_students.sql, 002_auth.sql
  route/          route.go
```

## Menjalankan Week 1

```powershell
go run ./latihan-fiber
go run ./latihan-fiber/variabel
go run ./latihan-fiber/pointer
go run ./latihan-fiber/struct
```

## Menjalankan API Student

```powershell
go run ./api-student
```

Setelah server berjalan, `GET /api/v1/health` memeriksa server sekaligus koneksi
database: `200` bila keduanya sehat, `503` bila database tidak dapat dihubungi.

## Menyiapkan Database

Buat database dan jalankan kedua migration dari root repository:

```powershell
psql -h localhost -U postgres -c "CREATE DATABASE go_backend;"
psql -h localhost -U postgres -d go_backend -f .\api-student\migrations\001_create_students.sql
psql -h localhost -U postgres -d go_backend -f .\api-student\migrations\002_auth.sql
```

Migration pertama membuat tabel `students`, constraint unik pada `nim`, dan index
untuk pencarian nama. Migration kedua membuat tabel `users` dengan index unik
case-insensitive pada `username` dan `email`, serta tabel `refresh_tokens` untuk
menyimpan refresh token dalam bentuk hash.

## Environment Variable

Buat `.env` berdasarkan `.env.example`. File `.env` tidak boleh di-commit.
`JWT_SECRET` wajib diisi minimal 32 karakter — aplikasi berhenti saat dijalankan
bila nilainya kosong atau terlalu pendek.

| Variabel | Kegunaan | Contoh aman |
|---|---|---|
| `APP_PORT` | Port aplikasi | `3000` |
| `APP_NAME` | Nama aplikasi pada log | `Praktikum Backend Lanjut` |
| `LOG_LEVEL` | Level log (`debug`, `info`, `warn`, `error`) | `info` |
| `DB_HOST` | Host PostgreSQL | `localhost` |
| `DB_PORT` | Port PostgreSQL | `5432` |
| `DB_USER` | User database | `postgres` |
| `DB_PASSWORD` | Password database | Diisi lokal, jangan di-upload |
| `DB_NAME` | Nama database | `go_backend` |
| `DB_SSLMODE` | Mode SSL koneksi | `disable` untuk lokal |
| `DB_MAX_CONNS` | Maksimum koneksi pool | `10` |
| `JWT_SECRET` | Kunci penandatangan access token | Diisi lokal, minimal 32 karakter |
| `JWT_ISSUER` | Issuer klaim `iss` pada token | `praktikum-backend` |
| `JWT_ACCESS_TTL_MINUTES` | Umur access token dalam menit | `15` |
| `JWT_REFRESH_TTL_DAYS` | Umur refresh token dalam hari | `7` |
| `ALLOWED_ORIGINS` | Origin CORS yang diizinkan, dipisah koma | `http://localhost:5173` |

## Authentication

`POST /api/v1/auth/register` dan `POST /api/v1/auth/login` menukar kredensial
dengan sepasang token: access token JWT (HS256, bawaan 15 menit) dan refresh
token acak (bawaan 7 hari). Refresh token disimpan di database sebagai hash
SHA-256 dan dirotasi setiap kali dipakai — token lama langsung dicabut.

Seluruh endpoint `students` dan `GET /api/v1/auth/me` membutuhkan header
`Authorization: Bearer <access_token>`. Password disimpan sebagai hash bcrypt
dan tidak pernah ikut di dalam response.

## Kontrak API Auth

| Metode | Endpoint | Contoh body permintaan | Status yang mungkin dikembalikan |
|---|---|---|---|
| POST | `/api/v1/auth/register` | `{"username":"vito","email":"vito@example.com","password":"rahasia123"}` | `201`, `400`, `409`, `415`, `422` |
| POST | `/api/v1/auth/login` | `{"username":"vito","password":"rahasia123"}` | `200`, `400`, `401`, `403`, `415`, `422`, `429` |
| POST | `/api/v1/auth/refresh` | `{"refresh_token":"..."}` | `200`, `400`, `401`, `415` |
| POST | `/api/v1/auth/logout` | `{"refresh_token":"..."}` | `200`, `400`, `415` |
| GET | `/api/v1/auth/me` | Tidak ada | `200`, `401` |

Response `POST /api/v1/auth/login` dan `/auth/refresh`:

```json
{"success":true,"message":"login berhasil","data":{"access_token":"...","refresh_token":"...","token_type":"Bearer","expires_in":900}}
```

`401` dikembalikan saat kredensial salah atau token tidak sah, `403` saat akun
dinonaktifkan, dan `429` saat login dari satu IP gagal lebih dari 5 kali dalam
satu menit. Username duplikat menghasilkan `409`. Response register menyertakan
header `Location` yang menunjuk ke `/api/v1/users/{id}`; endpoint detail user
tersebut belum tersedia.

## Error Database

Repository menerjemahkan error PostgreSQL menjadi error aplikasi. Data tidak
ditemukan menghasilkan `404`, NIM duplikat menghasilkan `409`, dan database
yang tidak dapat dihubungi pada health check menghasilkan `503`.

## Kontrak API Student

Data disimpan di PostgreSQL dan seluruh endpoint membutuhkan header
`Authorization: Bearer <access_token>`. `page` memiliki nilai bawaan 1, `limit`
memiliki nilai bawaan 10 dan dibatasi maksimum 100. Field yang dapat digunakan
untuk sorting adalah `id`, `nim`, `name`, `grade`, dan `created_at`.

| Metode | Endpoint | Parameter | Contoh body permintaan | Status yang mungkin dikembalikan | Contoh respons |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | Query: `page`, `limit`, `search`, `sort`, `order`, `is_active` | Tidak ada | `200`, `401` | `{"success":true,"message":"daftar student berhasil diambil","data":[...],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| GET | `/api/v1/students/:id` | Path: `id` | Tidak ada | `200`, `400`, `401`, `404` | `{"success":true,"message":"student ditemukan","data":{"id":1,"nim":24001,"name":"Sari","grade":90,"is_active":true,"created_at":"2026-09-23T10:00:00+07:00"}}` |
| POST | `/api/v1/students` | Tidak ada | `{"nim":24001,"name":"Sari","grade":90}` | `201`, `400`, `401`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil dibuat","data":{...}}` |
| PUT | `/api/v1/students/:id` | Path: `id`; body wajib: `nim`, `name`, `grade`, `is_active` | `{"nim":24001,"name":"Sari Baru","grade":95,"is_active":false}` | `200`, `400`, `401`, `404`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil diganti seluruhnya","data":{...}}` |
| PATCH | `/api/v1/students/:id` | Path: `id`; body hanya field yang diubah | `{"is_active":true}` | `200`, `400`, `401`, `404`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil diperbarui sebagian","data":{...}}` |
| DELETE | `/api/v1/students/:id` | Path: `id` | Tidak ada | `204`, `400`, `401`, `404` | Tidak ada body |

Request dengan body wajib menggunakan `Content-Type: application/json`. Response
POST berhasil menyertakan header `Location`, dan setiap response menyertakan
header `X-Request-Id`.

## Pemeriksaan Kode

```powershell
gofmt -l .
go vet ./...
go build ./...
go test ./...
```
