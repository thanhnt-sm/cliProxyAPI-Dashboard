package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$D3AOObVxJq4T3VjWp7R2U.CeoTjpgLEA.vpHQ/L8clQjAXqFk.eIO"
	key := "34ba56f38983bb7f1d32bc6a0c6d54a0"

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	if err != nil {
		fmt.Printf("Mismatch: %v\n", err)
	} else {
		fmt.Println("Match!")
	}
}
