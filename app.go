package main

import (
	"fmt"
	"os"

	"github.com/jplein/tmux-term-drawer/drawer"
)

func handleError(err error) {
	os.Stderr.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	os.Exit(1)
}

func main() {
	err := drawer.Toggle()
	if err != nil {
		handleError(err)
	}
}
