# AI-USAGE

Dokumentasi penggunaan bantuan AI (opencode) dalam pengembangan project ini.

## Prinsip Penggunaan

Pada pekerjaan Week 1, AI **tidak pernah diminta untuk menulis kode langsung**
ke file. AI hanya diminta memberikan:

1. **Petunjuk / kisi-kisi** cara mengimplementasikan suatu fungsi
2. **Review kode** untuk menemukan kesalahan
3. **Koreksi konsep** jika ada pemahaman yang keliru

Pada pengembangan Week 2, AI berperan sebagai pengarah, reviewer, dan
pendamping untuk memahami materi yang masih baru. AI membantu memetakan
struktur project, menjelaskan pola implementasi dari contoh modul, dan
memeriksa hasil pengujian API.

## Daftar Bantuan yang Diberikan

### 1. Folder `pointer` — `updateSlice` (menambah item ke slice)

- Meminta petunjuk cara mengimplementasikan fungsi untuk menambahkan item baru ke slice melalui pointer.
- AI menjelaskan konsep: dereference pointer (`*s`) lalu `append`, dan wajib assign balik (`*s = append(*s, newItem)`) karena `append` dapat membuat array baru.

### 2. Folder `pointer` — `swap` (menukar nilai)

- Bertanya apa yang salah dengan function `swap` yang masih kosong.
- AI menjelaskan kesalahan umum: menukar pointer-nya saja (`a, b = b, a`) tidak mengubah nilai asli; harus dereference (`*a, *b = *b, *a`).

### 3. Folder `pointer` — review kode lengkap

- Menyerahkan kode utuh untuk diperiksa.
- Ditemukan bug pada `swap`: setelah `*a = *b`, nilai asli `*a` hilang sehingga kedua variabel menjadi nilai `b`.
- AI menyarankan variabel sementara atau tuple assignment, serta tips menambahkan `Println` untuk memverifikasi hasil.

### 4. Folder `struct` — kisi-kisi function dan `main`

- Meminta kisi-kisi (outline) saja untuk method pada struct `Student` dan fungsi `main`.
- AI menunjukkan kesalahan syntax: method Go hanya boleh memiliki satu receiver (`func (s *Student, Grade float64) UpdateGrade()` tidak valid), dan field `Grade` bertipe `string` sehingga parameter seharusnya `string`.
- Diberikan outline isi `GetInfo` dan alur `main` (scan input, panggil method).

### 5. Folder `struct` — review kesalahan dan penggunaan pointer receiver

- Meminta identifikasi kesalahan pada kode dan receiver mana saja yang perlu pointer.
- Temuan AI:
  - `GetInfo` dideklarasikan return `string` tetapi body mengembalikan hasil `fmt.Printf` (bertipe `(int, error)`) — tidak akan compile.
  - Variabel `pilihan` tidak pernah dibaca input (`fmt.Scan` hilang di dalam loop), sehingga `switch` selalu masuk `case 0`.
  - `UpdateGrade`, `Activate`, `Deactive` perlu pointer receiver karena mengubah field; `GetInfo` cukup value receiver karena hanya membaca.

### 6. Folder `pointer` — perbandingan pass by value vs pass by pointer

- Meminta petunjuk di mana menambahkan versi pass by value sebagai pembanding dari versi pointer yang sudah ada.
- AI menyarankan lokasi penambahan: fungsi `swapValue(a, b int)` dan `updateSliceValue(s []string, newItem string)` di dekat fungsi pointer yang sudah ada, lalu menampilkan perbedaan hasil di `main`.

### 7. Week 2 - API Student

- Membandingkan implementasi API dengan persyaratan tugas mandiri Week 2.
- Membantu mengubah API `users` menjadi API `students` dengan field NIM, Name,
  Grade, dan IsActive.
- Membantu menerapkan validasi NIM unik dengan status `409 Conflict`, validasi
  field dengan status `422`, pagination, pencarian, filter, dan pengurutan.
- Menemukan dan memperbaiki bug response validasi serta mutasi sebagian pada
  PATCH yang gagal.
- Membantu menulis kontrak endpoint pada `README.md`.
- Membantu pengujian API menggunakan `curl.exe`, termasuk memahami opsi `-i`,
  pengiriman JSON melalui PowerShell, header `Location` dan `X-Request-Id`,
  serta membaca status HTTP dari response.

### 8. Week 4 - Clean Architecture sebagai mentor

- Peran AI sebagai mentor bertahap, bukan worker. Setiap langkah dirujuk ke
  bagian modul agar mudah dicek ulang.
- Membagi Modul 4 menjadi langkah kecil: kerangka folder, `helper`,
  `rules` murni, test, `service`, `middleware`, `route`, `logger`,
  `app`, `main` ramping, hapus file lama.
- Review tiap file sebelum lanjut. Temuan utama:
  - `helper/request.go` kurang import `time` dan belum definisikan
    `maxPageLimit` serta `allowedSort` student.
  - `student_service.go` jalur import `helper` salah, import belum terpakai,
    panggil `fail` lama, hitung halaman manual, `translateError` belum ada,
    `Patch` belum pakai `ApplyPatch` dan `IsEmptyPatch`.
  - `route/route.go` masih tempelan tanpa fungsi `Register` dan nama lama.
  - `config/app.go` belum pakai `middleware.Register`.
  - `config/logger.go` nama file `app.Log` harus `app.log`.
- Koreksi konsep: beda method dan fungsi, beda adapter dan use case,
  kenapa dua whitelist tidak digabung, kenapa `ORDER BY` butuh whitelist,
  kenapa `TrimSpace` di rules tidak terbawa, kenapa `CountTotalPages`
  harus jaga limit agar tidak panic.
- Bantuan efektivitas: urutan CRUD, receiver `s` untuk service, `gofmt`,
  beda `go build` dan `go vet` dan `gofmt -l` serta `gofmt -w`,
  cara baca error `&` di PowerShell dengan kutip URL.
- Pengujian: `go vet`, `go build`, `go test` tiga PASS,
  `curl` health dan list, cek `api-student/logs/app.log` ada baris
  `http_request` dengan `request_id`, `method`, `path`, `status`, `duration`.
- Lokasi log disesuaikan ke `api-student/logs/app.log` agar tunggal
  untuk struktur monorepo ini.

## Ringkasan Peran AI

| Tipe Bantuan | Contoh |
|---|---|
| Petunjuk / kisi-kisi | Cara mengisi `updateSlice`, outline method struct, pecah Modul 4 per langkah |
| Review & koreksi kode | Bug `swap`, return type `GetInfo`, `pilihan` tidak discan, import dan return di service dan route |
| Koreksi konsep | Kapan receiver harus pointer, perbedaan by value vs by pointer, batas layer dan whitelist |
| Pengujian API | Menjalankan `curl.exe`, opsi `-i`, header, membaca status HTTP, cek `app.log` |
