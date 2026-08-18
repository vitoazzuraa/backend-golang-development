package main

import (
	"fmt"
)

func main() {
	var nama string
	var nim int
	var ipk float64
	var isActive bool
	var mataKuliah []string

	mahasiswa := map[string]string{}

	var pilihan int

	for {
		fmt.Println("1. Tambah nama")
		fmt.Println("2. Cari nama")
		fmt.Println("3. Hapus nama")
		fmt.Println("4. Daftar nama")
		fmt.Println("5. Input lainnya")
		fmt.Println("0. Keluar")

		fmt.Print("Masukkan pilihan: ")
		fmt.Scan(&pilihan)

		switch pilihan {
		case 1:
			fmt.Print("Masukkan nama: ")
			fmt.Scan(&nama)

			index, ditemukan := mahasiswa[nama]

			if ditemukan {
				fmt.Println("Nama sudah ada di map:", index)
				continue
			} else {
				mahasiswa[nama] = nama
			}
		case 2:
			fmt.Print("Masukkan nama: ")
			fmt.Scan(&nama)

			index, ditemukan := mahasiswa[nama]

			if ditemukan {
				fmt.Println("Nama ada di map", index)
			} else {
				fmt.Println("Nama tidak ditemukan")
			}
		case 3:
			fmt.Print("Hapus nama: ")
			fmt.Scan(&nama)

			index, ditemukan := mahasiswa[nama]

			if ditemukan {
				fmt.Println("Nama ditemukan:", index)
				delete(mahasiswa, nama)
			} else {
				fmt.Println("Nama tidak ditemukan")
			}
		case 4:
			for _, list := range mahasiswa {
				fmt.Println(list)
			}
		case 5:
			var matkul string

			fmt.Print("Input nama: ")
			fmt.Scan(&nama)

			index, ditemukan := mahasiswa[nama]

			if ditemukan {
				isActive = true

				fmt.Print("Input nim: ")
				fmt.Scan(&nim)

				fmt.Print("Input ipk: ")
				fmt.Scan(&ipk)

				fmt.Print("Input mata kuliah favorit: ")
				fmt.Scan(&matkul)

				mataKuliah = append(mataKuliah, matkul)

			fmt.Println("Nama:", index)
			fmt.Println("Nim:", nim)
			fmt.Println("Ipk:", ipk)

			if isActive {
				fmt.Println("Status: aktif")
			} else {
				fmt.Println("Status: nonaktif")
			}

			fmt.Println("Mata Kuliah Favorit:", mataKuliah)
			} else {
				fmt.Println("Mahasiswa tidak ditemukan")
			}
		case 0:
			return
		default:
			fmt.Println("Pilihan anda salah!")
		}
	}
}
