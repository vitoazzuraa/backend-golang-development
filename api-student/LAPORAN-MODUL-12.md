# Laporan Modul 12: Testing, Swagger, dan Docker

Nama: Christianus Primavito
NIM: 434241058
Mata Kuliah: Praktikum Pemrograman Backend Lanjut
Repository: https://github.com/vitoazzuraa/backend-golang-development.git

Berkas ini disusun bertahap, mengikuti setiap langkah pengerjaan dan setiap
commit yang dibuat. Bagian yang belum dikerjakan diberi tanda.

---

## 1. Apa yang dikerjakan di modul ini

Modul 12 isinya tiga hal sekaligus: menulis test untuk kode yang sudah ada
sejak modul sebelumnya, mendokumentasikan API memakai Swagger, lalu membungkus
seluruh sistem ke dalam Docker.

Karena Modul 9 sampai 11 adalah edisi puzzle, Modul 12 juga memuat kesalahan
yang sengaja ditanam, kali ini di dalam berkas testnya. Tugas saya adalah
menemukannya lalu melaporkannya.

---

## 2. Urutan pengerjaan

### 2.1 Test aturan batas sks dan pindah kelas

Berkas: `app/service/course_rules_test.go`

Berkas ini memuat dua test yang menguji fungsi yang sudah kita punya sejak
Modul 9. Keduanya memakai pola table-driven, yaitu kasus uji ditulis sebagai
tabel lalu diperulang.

Hasil pengujian: 7 subtest, semuanya lulus.

Tidak ada kesalahan yang ditanam di bagian ini. Fungsi `WithinCreditLimit` dan
`CreditsAfterSwap` sudah benar sejak awal.

Commit: `tambah test aturan batas sks dan pindah kelas`

### 2.2 Test normalisasi tag

Berkas: `app/service/review_rules_test.go`

Berkas ini menguji fungsi `NormalizeTags` yang kita buat untuk menutupi
kekurangan sejak modul ulasan. Fungsinya membersihkan tag sebelum disimpan:
membuang spasi di tepi, mengubah ke huruf kecil, menghapus yang kosong, dan
menghapus yang kembar sambil mempertahankan urutan kemunculan pertama.

Waktu berkas ini pertama kali disalin dari modul apa adanya, kodenya tidak bisa
diperiksa. Pesan yang muncul:

    wrong signature for TestNormalizeTags, must be: func TestNormalizeTags(t *testing.T)

Ternyata tanda tangan fungsinya memakai `t testing.T`, yaitu salinan struct,
padahal yang diperlukan adalah `t *testing.T`, yaitu alamat struct. Ini
kesalahan pertama yang sengaja ditanam di modul ini.

Setelah diperbaiki, enam subtest semuanya lulus.

Commit: `tambah test normalisasi tag`

### 2.3 Test penyusun pesan notifikasi

Berkas: `app/service/notification_rules_test.go`

Berkas ini menguji `BuildMessage`, yaitu fungsi yang menyusun pesan yang akan
dikirim kepada mahasiswa. Tiga kejadian diuji: berhasil terdaftar, keluar dari
kelas, dan pindah kelas. Ditambah satu pengujian untuk kejadian yang tidak
dikenal, yang harus mengembalikan error khusus.

Setiap pesan diperiksa tiga hal: subjeknya persis, alamat tujuannya benar, dan
isinya menyebut nama penggunanya.

Hasil pengujian: 4 subtest, semuanya lulus.

Tidak ada kesalahan yang ditanam di bagian ini. Bentuknya sudah sesuai dengan
modul.

Commit: `tambah test penyusun pesan notifikasi`

### 2.4 Tiruan repository untuk unit test

Berkas: `app/service/fakes_test.go`

Berkas ini berisi tiruan repository, yaitu pengganti MongoDB dan PostgreSQL yang
menyimpan datanya di memori. Tujuannya supaya test bisa berjalan cepat tanpa
perlu database sungguhan.

Ada dua tiruan. Yang pertama untuk ulasan, yang menyimpan data di dalam potong
memori dan memakai pengunci karena Fiber melayani permintaan pada goroutine yang
berbeda. Yang kedua untuk kelas, yang hanya mengisi dua metode yang dipakai.

Ada satu baris yang menarik di awal berkas:

    var _ repository.ReviewRepository = (*fakeReviewRepo)(nil)

Baris ini memeriksa saat dikompilasi apakah tiruannya benar-benar memenuhi
interface yang ditiru. Kalau suatu hari bentuknya berubah, baris inilah yang
pertama kali gagal, bukan test yang memakainya.

Satu perbedaan dari modul yang perlu dicatat: modul memanggil `FindByCourse`
dengan `courseID int`, sedangkan kode kita memakai `model.ReviewQuery`. Ini
ternyata ketidakkonsistenan dalam modulnya sendiri, karena Modul 10 memang
memakai `model.ReviewQuery`. Jadi tiruannya yang mengikuti bentuk Modul 10.

Commit: `tambah tiruan repository untuk unit test`

---

## 3. Daftar kesalahan yang sengaja ditanam

Bagian ini diisi setiap kali menemukan kesalahan. Setiap baris memuat lokasi,
gejala, akar masalah, dan cara memperbaikinya.

| No | Lokasi | Gejala | Akar masalah | Perbaikan | Ketahuan dari |
|---|---|---|---|---|---|
| 1 | `app/service/review_rules_test.go`, fungsi `TestNormalizeTags` | kode tidak bisa diperiksa, muncul pesan `wrong signature for TestNormalizeTags` | tanda tangan fungsi memakai `t testing.T` berupa nilai, sedangkan kerangka test Go menghendaki `t *testing.T` berupa alamat | tambahkan tanda bintang menjadi `func TestNormalizeTags(t *testing.T)` | compiler |

---

## 4. Perbedaan dari modul sebelumnya

Bagian ini mencatat perbedaan antara kode kita dengan modul, untuk bahan
belajar. Bukan kesalahan yang harus dibereskan.

| No | Bagian | Modul | Kode kita | Alasan |
|---|---|---|---|---|
| 1 | `FindByCourse` pada `fakeReviewRepo` | `FindByCourse(ctx, courseID int)` | `FindByCourse(ctx, q model.ReviewQuery)` | Modul 10 memakai `model.ReviewQuery`, jadi kode kita sudah sesuai Modul 10. Modul 12 yang tidak konsisten dengan modul sebelumnya, jadi tiruannya yang mengikuti kita |

---

## 5. Bagian Swagger

(belum dikerjakan)

---

## 6. Bagian Docker

(belum dikerjakan)

---

## 7. Tugas mandiri

(belum dikerjakan)

---

## 8. Yang saya bawa pulang

(belum diisi)
