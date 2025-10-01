package main

import (
	"fmt"
	"os"

	"github.com/rizesky/mckmt/internal/auth"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run tools/hash-password.go <password>")
		fmt.Println("Example: go run tools/hash-password.go admin123")
		os.Exit(1)
	}

	password := os.Args[1]

	// Use our auth package's password manager with proper config
	passwordManager := auth.NewPasswordManager(auth.DefaultPasswordConfig())

	// Generate hash using our mechanism
	hash, err := passwordManager.HashPassword(password)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Hash: %s\n", hash)

	// Verify the hash works
	valid, err := passwordManager.VerifyPassword(password, hash)
	if err != nil {
		fmt.Printf("Error verifying hash: %v\n", err)
		os.Exit(1)
	}

	if !valid {
		fmt.Printf("Error: Generated hash doesn't match password!\n")
		os.Exit(1)
	}

	fmt.Println("✓ Hash verification successful")
}
