package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/birabittoh/disgord/src/deezer"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	email := os.Getenv("DEEZER_EMAIL")
	password := os.Getenv("DEEZER_PASSWORD")

	if len(os.Args) > 2 {
		email = os.Args[1]
		password = os.Args[2]
	}

	if email == "" || password == "" {
		fmt.Fprintln(os.Stderr, "Usage: deezer-auth <email> <password>")
		fmt.Fprintln(os.Stderr, "   or: set DEEZER_EMAIL and DEEZER_PASSWORD in .env")
		os.Exit(1)
	}

	logger := slog.Default()
	arl, err := deezer.Login(logger, email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\narl: %s\n", arl)
}
