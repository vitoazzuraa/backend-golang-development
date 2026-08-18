package main

import (
	"fmt"
)

type Student struct {
	ID       int
	Name     string
	Grade    float64
	isActive bool
}

func (s Student) GetInfo() string {
	return fmt.Sprintf("%v, %v, %f, %t", s.ID, s.Name, s.Grade, s.isActive)
}

func (s *Student) UpdateGrade(newGrade float64) {
	s.Grade = newGrade
}

func (s *Student) Activate() {
	s.isActive = true
}

func (s *Student) Deactive() {
	s.isActive = false
}

func main() {
	var s Student

	fmt.Print("Input nama: ")
	fmt.Scan(&s.Name)

	fmt.Print("Input id: ")
	fmt.Scan(&s.ID)

	fmt.Print("Input grade: ")
	fmt.Scan(&s.Grade)

	fmt.Println("Informasi data")
	fmt.Println(s.GetInfo())

	var newGrade float64

	fmt.Print("Masukkan nilai grade baru: ")
	fmt.Scan(&newGrade)

	s.UpdateGrade(newGrade)

	fmt.Println("Informasi data terbaru")
	fmt.Println(s.GetInfo())

	var pilihan int

	for {
		fmt.Println("1. Ubah status aktif")
		fmt.Println("2. Ubah status deactive")
		fmt.Println("0. Keluar")

		fmt.Print("Masukkan pilihan: ")
		fmt.Scan(&pilihan)

		switch pilihan {
		case 1:
			s.Activate()
			fmt.Println("Status", s.isActive)
		case 2:
			s.Deactive()
			fmt.Println("Status", s.isActive)
		case 0:
			return
		default:
			fmt.Print("Kamu salah membuat pilihan")
		}
	}
}
