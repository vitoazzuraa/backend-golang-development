package main

import (
	"fmt"
)

func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func main() {
	var a int
	var b int

	fmt.Print("Input angka a: ")
	fmt.Scan(&a)

	fmt.Print("Input angka b: ")
	fmt.Scan(&b)

	swap(&a, &b)
	fmt.Println("Nilai a dan b setelah ditukar, a:", a, "& nilai b:", b)

	var name string

	names := []string{}

	fmt.Print("Input kata untuk di masukkan ke slice: ")
	fmt.Scan(&name)

	updateSlice(&names, name)

	fmt.Println("Value slice sekarang:", names)
}
