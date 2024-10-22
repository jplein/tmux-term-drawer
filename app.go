package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jplein/tmux-term-drawer/drawer"
)

func handleError(err error) {
	os.Stderr.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	os.Exit(1)
}

func main() {
	var socketName string
	flag.StringVar(&socketName, "socket", "", "The name of the tmux socket to use")

	flag.Parse()

	err := drawer.Toggle(socketName)
	if err != nil {
		handleError(err)
	}
}
