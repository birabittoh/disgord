package main

import "github.com/birabittoh/disgord/src/ui"

func main() {
	if err := ui.Main(); err != nil {
		panic(err)
	}
}
