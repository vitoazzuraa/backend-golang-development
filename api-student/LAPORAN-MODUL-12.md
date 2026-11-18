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

---

## 3. Daftar kesalahan yang sengaja ditanam

Bagian ini diisi setiap kali menemukan kesalahan. Setiap baris memuat lokasi,
gejala, akar masalah, dan cara memperbaikinya.

| No | Lokasi | Gejala | Akar masalah | Perbaikan | Ketahuan dari |
|---|---|---|---|---|---|
| (belum ada) | | | | | |

---

## 4. Perbedaan dari modul sebelumnya

Bagian ini mencatat perbedaan antara kode kita dengan modul, untuk bahan
belajar. Bukan kesalahan yang harus dibereskan.

| No | Bagian | Modul | Kode kita | Alasan |
|---|---|---|---|---|
| (belum ada) | | | | |

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
