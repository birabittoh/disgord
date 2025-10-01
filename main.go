package main

import "github.com/birabittoh/disgord/src"

func main() {
	if err := src.Main(); err != nil {
		panic(err)
	}
}
