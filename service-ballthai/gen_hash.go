package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var password string
	if len(os.Args) > 1 {
		password = os.Args[1]
	} else {
		password = "admin123"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", string(hash))
}
