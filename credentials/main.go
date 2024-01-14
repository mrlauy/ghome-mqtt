package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	file, err := os.OpenFile("../.credentials", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("failed to open file", err)
	}
	defer file.Close()

	var username string
	var password string
	flag.StringVar(&username, "u", "", "username")
	flag.StringVar(&password, "p", "", "password")
	flag.Parse()

	if len(username) < 1 || len(password) < 1 {
		log.Fatal("no username or password provided")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)

	user := strings.ToLower(username)
	if _, err := file.WriteString(fmt.Sprintf("%s:%s\n", user, string(hashedPassword))); err != nil {
		log.Fatal("failed to write credentials to file", err)
	}

	log.Printf("successful stored credentials for user: %s", username)
}
