# Pemrograman Backend Lanjut dengan Go

Repository ini berisi latihan Week 1 dan API Student Week 2.

## Struktur Project

```text
latihan-fiber/
  main.go
  variabel/main.go
  pointer/main.go
  struct/main.go

api-student/
  main.go
  model.go
  helper.go
  handler.go
```

## Menjalankan Week 1

```powershell
go run ./latihan-fiber
go run ./latihan-fiber/variabel
go run ./latihan-fiber/pointer
go run ./latihan-fiber/struct
```

## Menjalankan Week 2

```powershell
go run ./api-student
```

## Kontrak API Student

Data disimpan di memori. `page` memiliki nilai bawaan 1, `limit` memiliki nilai
bawaan 10 dan dibatasi maksimum 100. Field yang dapat digunakan untuk sorting
adalah `id`, `nim`, `name`, dan `grade`.

| Metode | Endpoint | Parameter | Contoh body permintaan | Status yang mungkin dikembalikan | Contoh respons |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | Query: `page`, `limit`, `search`, `sort`, `order`, `is_active` | Tidak ada | `200` | `{"success":true,"message":"daftar student berhasil diambil","data":[...],"meta":{"page":1,"limit":10,"total":1,"total_pages":1}}` |
| GET | `/api/v1/students/:id` | Path: `id` | Tidak ada | `200`, `400`, `404` | `{"success":true,"message":"student ditemukan","data":{"id":1,"nim":24001,"name":"Sari","grade":90,"is_active":true}}` |
| POST | `/api/v1/students` | Tidak ada | `{"nim":24001,"name":"Sari","grade":90}` | `201`, `400`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil dibuat","data":{...}}` |
| PUT | `/api/v1/students/:id` | Path: `id`; body wajib: `nim`, `name`, `grade`, `is_active` | `{"nim":24001,"name":"Sari Baru","grade":95,"is_active":false}` | `200`, `400`, `404`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil diganti seluruhnya","data":{...}}` |
| PATCH | `/api/v1/students/:id` | Path: `id`; body hanya field yang diubah | `{"is_active":true}` | `200`, `400`, `404`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil diperbarui sebagian","data":{...}}` |
| DELETE | `/api/v1/students/:id` | Path: `id` | Tidak ada | `204`, `400`, `404` | Tidak ada body |

Request dengan body wajib menggunakan `Content-Type: application/json`. Response
POST berhasil menyertakan header `Location`, dan setiap response menyertakan
header `X-Request-Id`.
