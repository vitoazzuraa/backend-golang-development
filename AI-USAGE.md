# AI-USAGE

Dokumentasi penggunaan bantuan AI (opencode) dalam pengembangan project ini.
Catatan disusun per minggu dan diurutkan sesuai urutan pengerjaan, dari Week 1
hingga pekerjaan terbaru.

## Prinsip Penggunaan

Pada Week 1, AI tidak pernah diminta untuk menulis kode langsung ke file. AI
hanya diminta memberikan:

1. **Petunjuk / kisi-kisi** cara mengimplementasikan suatu fungsi
2. **Review kode** untuk menemukan kesalahan
3. **Koreksi konsep** jika ada pemahaman yang keliru

Pada Week 2 sampai Week 5, peran AI bergeser menjadi pengarah, reviewer, dan
pendamping untuk materi yang masih baru: memetakan struktur project, menjelaskan
pola implementasi dari contoh modul, dan memeriksa hasil pekerjaan. Kode tetap
ditulis sendiri: AI tidak menulis langsung ke file.

Pada Week 6, AI dipakai untuk memeriksa kondisi repository sebelum masuk materi
authorization: menemukan sisa masalah, memandu perbaikan langkah demi langkah,
lalu memverifikasi hasilnya. AI hanya menulis file dokumentasi (README dan
catatan ini) ketika diminta secara eksplisit.

## Week 1 — Dasar Go, Pointer, dan Struct (13–18 Agustus 2026)

### 1. Folder `pointer` — `updateSlice` (menambah item ke slice)

- Meminta petunjuk cara mengimplementasikan fungsi untuk menambahkan item baru
  ke slice melalui pointer.
- AI menjelaskan konsep: dereference pointer (`*s`) lalu `append`, dan wajib
  assign balik (`*s = append(*s, newItem)`) karena `append` dapat membuat array
  baru.

### 2. Folder `pointer` — `swap` (menukar nilai)

- Bertanya apa yang salah dengan function `swap` yang masih kosong.
- AI menjelaskan kesalahan umum: menukar pointer-nya saja (`a, b = b, a`) tidak
  mengubah nilai asli; harus dereference (`*a, *b = *b, *a`).

### 3. Folder `pointer` — review kode lengkap

- Menyerahkan kode utuh untuk diperiksa.
- Ditemukan bug pada `swap`: setelah `*a = *b`, nilai asli `*a` hilang sehingga
  kedua variabel menjadi nilai `b`.
- AI menyarankan variabel sementara atau tuple assignment, serta tips
  menambahkan `Println` untuk memverifikasi hasil.

### 4. Folder `struct` — kisi-kisi function dan `main`

- Meminta kisi-kisi (outline) saja untuk method pada struct `Student` dan fungsi
  `main`.
- AI menunjukkan kesalahan syntax: method Go hanya boleh memiliki satu receiver
  (`func (s *Student, Grade float64) UpdateGrade()` tidak valid), dan field
  `Grade` bertipe `string` sehingga parameter seharusnya `string`.
- Diberikan outline isi `GetInfo` dan alur `main` (scan input, panggil method).

### 5. Folder `struct` — review kesalahan dan penggunaan pointer receiver

- Meminta identifikasi kesalahan pada kode dan receiver mana saja yang perlu
  pointer.
- Temuan AI:
  - `GetInfo` dideklarasikan return `string` tetapi body mengembalikan hasil
    `fmt.Printf` (bertipe `(int, error)`) — tidak akan compile.
  - Variabel `pilihan` tidak pernah dibaca input (`fmt.Scan` hilang di dalam
    loop), sehingga `switch` selalu masuk `case 0`.
  - `UpdateGrade`, `Activate`, `Deactive` perlu pointer receiver karena mengubah
    field; `GetInfo` cukup value receiver karena hanya membaca.

### 6. Folder `pointer` — perbandingan pass by value vs pass by pointer

- Meminta petunjuk di mana menambahkan versi pass by value sebagai pembanding
  dari versi pointer yang sudah ada.
- AI menyarankan lokasi penambahan: fungsi `swapValue(a, b int)` dan
  `updateSliceValue(s []string, newItem string)` di dekat fungsi pointer yang
  sudah ada, lalu menampilkan perbedaan hasil di `main`.

## Week 2 — REST API dan HTTP Deep Dive (26 Agustus 2026)

### 7. API Student — validasi, pagination, dan pengujian endpoint

- Membandingkan implementasi API dengan persyaratan tugas mandiri Week 2.
- AI membantu mengubah API `users` menjadi API `students` dengan field NIM,
  Name, Grade, dan IsActive.
- AI membantu menerapkan validasi NIM unik dengan status `409 Conflict`,
  validasi field dengan status `422`, pagination, pencarian, filter, dan
  pengurutan.
- Menemukan dan memperbaiki bug response validasi serta mutasi sebagian pada
  PATCH yang gagal.
- AI membantu menulis kontrak endpoint pada `README.md`.
- Pengujian: memakai `curl.exe`, termasuk opsi `-i`, pengiriman JSON melalui
  PowerShell, header `Location` dan `X-Request-Id`, serta membaca status HTTP
  dari response.

## Week 3 — Database dan Repository Pattern (2 September 2026)

### 8. Memindahkan penyimpanan data dari memori ke PostgreSQL

- Titik awalnya data student disimpan di variabel global
  (`var students []model.Student` dan `var nextID = 1`), sehingga seluruh data
  hilang setiap server dimatikan dan hampir semua pekerjaan dikerjakan manual di
  handler: mencari index dengan `findStudentIndex`, memeriksa NIM dengan
  `nimExists`, menyaring dengan `cocokPencarian`, dan mengurutkan dengan
  `lessStudent`.
- AI menjelaskan perubahan konsepnya, bukan sekadar tempat menyimpan: data
  dipindahkan ke tabel `students` agar bertahan setelah server mati, pencarian
  dan pengurutan menjadi tugas database lewat `WHERE` dan `ORDER BY`, dan
  keunikan NIM dijaga constraint, bukan perulangan di kode.
- Koreksi konsep: jumlah baris kode bisa berkurang, tetapi tanggung jawabnya
  berpindah. Logika yang dulu ditulis manual di handler kini menjadi tanggung
  jawab query dan constraint di database, dan kolom seperti `created_at` baru
  muncul karena pengelolaan waktu diserahkan ke database.

### 9. Konfigurasi environment, koneksi database, dan migration

- Meminta arahan urutan langkah sebelum menulis query: memisahkan konfigurasi,
  koneksi, entity, dan akses data ke folder masing-masing.
- AI membantu memetakan struktur: `config/env.go` untuk membaca `.env`,
  `database/postgres.go` untuk koneksi, `app/model` untuk entity, dan
  `app/repository` untuk akses data. File `model.go` dan `helper.go` di root
  `api-student` dihapus karena isinya pindah ke folder baru.
- AI menjelaskan kenapa koneksi dibuat sekali dalam satu pool (`pgxpool`) lalu
  dipakai ulang oleh semua request, bukan dibuka ulang setiap request.
- AI memandu penulisan migration `001_create_students.sql`: tabel `students`,
  constraint unik pada `nim`, `CHECK (grade >= 0 AND grade <= 100)`, check nama
  tidak kosong, dan index `LOWER(name)` untuk pencarian.

### 10. Repository pattern dan penulisan query

- Meminta arahan cara memindahkan operasi CRUD ke
  `app/repository/student_repository.go`.
- AI menjelaskan peran repository sebagai satu-satunya tempat yang mengenal SQL:
  pemanggil cukup memakai method `FindAll`, `FindByID`, `Create`, `Update`, dan
  `Delete` tanpa perlu tahu bentuk query-nya.
- Koreksi konsep: nilai selalu dikirim lewat placeholder `$1`, `$2` agar tidak
  dirangkai ke dalam string SQL, dan error PostgreSQL diterjemahkan menjadi
  error aplikasi — `pgx.ErrNoRows` menjadi `ErrNotFound` (`404`) dan pelanggaran
  constraint unik dengan kode `23505` menjadi `ErrDuplicate` (`409`).
- Pengujian: menjalankan migration, lalu mengulang pengujian endpoint dengan
  `curl.exe` untuk memastikan data tidak lagi hilang ketika server dimatikan.

## Week 4 — Clean Architecture (9–10 September 2026)

### 11. Membagi Modul 4 menjadi langkah kecil

- Meminta AI berperan sebagai mentor bertahap, bukan worker. Setiap langkah
  dirujuk ke bagian modul agar mudah dicek ulang.
- Urutan yang disepakati: kerangka folder, `helper`, `rules` murni, test,
  `service`, `middleware`, `route`, `logger`, `app`, `main` ramping, lalu hapus
  file lama.

### 12. Review per file, koreksi konsep, dan pengujian

- Meminta review tiap file sebelum lanjut ke langkah berikutnya. Temuan utama:
  - `helper/request.go` kurang import `time` dan belum mendefinisikan
    `maxPageLimit` serta `allowedSort` student.
  - `student_service.go` jalur import `helper` salah, import belum terpakai,
    memanggil `fail` lama, menghitung halaman manual, `translateError` belum
    ada, dan `Patch` belum memakai `ApplyPatch` serta `IsEmptyPatch`.
  - `route/route.go` masih tempelan tanpa fungsi `Register` dan masih memakai
    nama lama.
  - `config/app.go` belum memakai `middleware.Register`.
  - `config/logger.go` menulis ke `app.Log`, seharusnya `app.log`.
- Koreksi konsep: beda method dan fungsi, beda adapter dan use case, kenapa dua
  whitelist tidak digabung, kenapa `ORDER BY` butuh whitelist, kenapa
  `TrimSpace` di rules tidak terbawa, dan kenapa `CountTotalPages` harus menjaga
  `limit` agar tidak panic.
- Bantuan efektivitas: urutan CRUD, receiver `s` untuk service, penggunaan
  `gofmt`, beda `go build`, `go vet`, `gofmt -l`, dan `gofmt -w`, serta cara
  membaca error `&` di PowerShell dengan mengutip URL.
- Pengujian: `go vet`, `go build`, dan `go test` tiga PASS, lalu `curl` ke
  endpoint health dan list, serta memeriksa `api-student/logs/app.log` berisi
  baris `http_request` dengan `request_id`, `method`, `path`, `status`, dan
  `duration`. Lokasi log disesuaikan ke `api-student/logs/app.log` agar tunggal
  untuk struktur monorepo ini.

## Week 5 — Authentication dan Security (13–16 September 2026)

### 13. Skema database auth dan dependency baru

- Meminta arahan skema penyimpanan user dan token pada migration
  `002_auth.sql`.
- AI menjelaskan pilihan di dalam skema: `username` dan `email` memakai index
  unik pada `LOWER(...)` supaya `Vito` dan `vito` dianggap sama, dan
  `refresh_tokens` memakai `ON DELETE CASCADE` ke `users` agar token ikut
  terhapus bersama pemiliknya.
- Koreksi konsep: refresh token tidak disimpan apa adanya, melainkan sebagai
  hash SHA-256. Bila isi database bocor, token yang tersimpan tidak dapat
  langsung dipakai untuk login.
- Menambahkan dependency `golang-jwt/jwt/v5` dan `golang.org/x/crypto` (bcrypt),
  serta blok JWT (`JWT_SECRET`, `JWT_ISSUER`, TTL access dan refresh) pada
  `.env.example`.

### 14. Helper keamanan dan JWT

- Meminta petunjuk isi `helper/security.go` dan `helper/jwt.go`.
- AI menjelaskan pilihan pada helper keamanan: bcrypt dengan cost 12,
  `VerifyDummyPassword` untuk menyamakan waktu proses ketika username tidak
  ditemukan (mencegah penyerang menebak username dari perbedaan waktu respons),
  `RandomToken` untuk membuat refresh token, dan `SHA256Hex` untuk menyimpannya.
- AI menjelaskan isi access token: `helper/jwt.go` menandatangani klaim
  `user_id`, `username`, dan `role` dengan HS256, memakai issuer dan masa
  berlaku dari environment, dan memisahkan error token kedaluwarsa dari token
  tidak sah.

### 15. Service dan middleware authentication

- Meminta arahan urutan pengerjaan, dari aturan validasi sampai route.
- AI membantu menyusun `app/service/auth_service.go`: `Register`, `Login`,
  `Refresh`, `Logout`, dan `Me`, dengan pembagian tugas yang jelas — aturan
  validasi murni di `auth_rules.go`, akses data di repository, dan penerbitan
  token lewat satu fungsi `issueTokenPair`.
- AI menjelaskan mekanisme rotasi refresh token: token lama dicabut tepat
  sebelum pasangan token baru diterbitkan, sehingga satu refresh token hanya
  dapat dipakai sekali.
- AI menjelaskan pembagian status pada login: `401` untuk kredensial salah,
  `403` untuk akun yang dinonaktifkan, dan kenapa keduanya tidak boleh
  disamakan.
- Memandu pembuatan `middleware/auth.go`: `RequireAuth` membaca header
  `Authorization: Bearer`, menaruh identitas pada `Locals`, dan mengirim
  `WWW-Authenticate` pada `401`; `LoginRateLimiter` membatasi 5 percobaan per
  menit per IP dengan `429` dan header `Retry-After`.

### 16. Pengujian authentication

- Meminta review aturan validasi beserta pengujiannya.
- Test yang ditulis bersama: register dengan password lemah ditolak, register
  dengan data valid lolos, login kosong menghasilkan dua error, password pendek
  ditolak, dan username yang memuat karakter seperti `!` ditolak.
- Koreksi konsep pada aturan: panjang minimum password 8 karakter, wajib memuat
  huruf dan angka, dan password umum seperti `password1` ditolak.
- Pengujian endpoint dengan `curl.exe`: register, login, memanggil endpoint
  student memakai access token, dan memastikan permintaan tanpa token
  menghasilkan `401`.

## Week 6 — Perapian Repository sebelum Authorization (23 September 2026)

### 17. Perbaikan syntax error yang sudah ter-commit

- Meminta pemeriksaan menyeluruh sebelum masuk materi authorization, karena
  repository sudah cukup lama tidak dijalankan setelah perubahan terakhir.
- Temuan AI: `api-student/app/repository/user_repository.go` memuat tiga
  pemanggilan `Scan(...)` yang kehilangan koma penutup, dan error yang sama
  ternyata sudah ikut ter-commit sehingga `go build`, `go vet`, dan `go test`
  gagal sejak commit terakhir.
- AI menjelaskan aturan Go yang menjadi penyebabnya: titik-koma disisipkan
  otomatis di akhir baris, sehingga argumen terakhir wajib diakhiri koma selama
  tanda `)` penutup berada di baris sendiri.
- AI memandu perbaikan tiga baris tersebut, bukan menulis ulang file, lalu
  memverifikasi hasilnya dengan `gofmt`, `go vet`, `go build`, dan `go test`.

### 18. Konfigurasi `.env`, `.gitignore`, dan kebersihan repository

- Meminta pemeriksaan konfigurasi dan status git.
- Temuan AI: `.env` lokal tidak memuat blok `JWT_SECRET` padahal `main.go`
  menghentikan aplikasi bila secret kurang dari 32 karakter; `go.mod` tampil
  termodifikasi walaupun isinya byte-identik dengan HEAD; folder `logs/` di root
  adalah sisa logger lama; dan enam file modul serta folder tooling belum
  diabaikan.
- AI memandu penambahan blok JWT pada `.env`, penambahan `.commandcode/` dan
  `*.docx` ke `.gitignore`, serta penghapusan folder `logs/` lama karena logger
  aktif kini hanya menulis ke `api-student/logs/app.log`.
- AI menjelaskan kenapa `git status` bisa melaporkan file sebagai berubah
  padahal `git diff`-nya kosong: isi file identik dengan yang tersimpan di Git,
  yang berbeda hanya line ending akibat `core.autocrlf=true`.

### 19. Memahami `gofmt` dan perbedaan line ending

- Bertanya kenapa `gofmt -w` mengubah banyak file tetapi tidak ada perubahan
  yang terlihat pada `git diff`.
- AI menjelaskan urutan pemakaian yang benar: `gofmt -l` untuk melihat daftar
  file yang akan berubah, `gofmt -d` untuk melihat diff-nya, dan `gofmt -w`
  untuk menuliskan perubahannya.
- AI menjelaskan hasil pemeriksaan: enam file benar-benar berubah formatnya
  (perataan field, spasi di akhir baris, baris kosong berisi tab, dan newline di
  akhir file), sedangkan belasan file lain hanya berubah line ending dari CRLF
  ke LF. Cara membedakannya: `git diff --raw` untuk perubahan isi dan
  `git ls-files --eol` untuk line ending.

### 20. Penulisan ulang README dan verifikasi akhir

- Meminta README diperbarui karena struktur project dan endpoint sudah berubah.
- AI menulis ulang README: struktur folder terbaru, section authentication,
  kontrak API auth, status `401` pada seluruh endpoint student yang kini
  membutuhkan token, tabel environment termasuk variabel JWT, migration
  `002_auth.sql`, dan catatan health check.
- Verifikasi akhir sebelum lanjut ke materi berikutnya: `gofmt -l` bersih,
  `go vet`, `go build`, dan `go test` lolos seluruhnya.

## Ringkasan Peran AI

| Tipe Bantuan | Contoh |
|---|---|
| Petunjuk / kisi-kisi | Cara mengisi `updateSlice`, outline method struct, urutan langkah Modul 4, urutan perapian repository |
| Review & koreksi kode | Bug `swap`, return type `GetInfo`, `pilihan` tidak discan, import service dan route, koma hilang pada `Scan(...)` |
| Koreksi konsep | Receiver pointer, pass by value vs pointer, batas layer dan whitelist, perpindahan data memori ke database, rotasi refresh token, aturan koma di Go |
| Pengujian | `curl.exe` dan pembacaan status HTTP, pemeriksaan `app.log`, `gofmt`, `go vet`, `go build`, `go test` |
