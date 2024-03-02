package main

import (
	"fmt"
	"project/slave/cabfile"
)

func main() {

	fmt.Println("Reading from file...")
	data := cabfile.Read()
	fmt.Println("Data read:", data)

	fmt.Println("Setting floor 2...")
	cabfile.Set(2)

	fmt.Println("Reading from file...")
	data = cabfile.Read()
	fmt.Println("Data read:", data)

	fmt.Println("Clearing floor 3...")
	cabfile.Clear(3)

	fmt.Println("Reading from file...")
	data = cabfile.Read()
	fmt.Println("Data read:", data)

	
	fmt.Println("Clearing floor2...")
	cabfile.Clear(2)

	fmt.Println("Reading from file...")
	data = cabfile.Read()
	fmt.Println("Data read:", data)
}