package main

import (
	"Go-Simple-Licensing-System/SimpleLicensing"
	"fmt"
)

func init() {
	Licensing.CheckLicense("http://127.0.0.1:8080/", false, false)
}

func main() {
	fmt.Println("License was varified!")
	for {
	}
}
