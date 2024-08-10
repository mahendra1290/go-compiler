package main

import (
	"monkey/src/repl"
	"os"
)

func main() {
	repl.Start(os.Stdin, os.Stdout)
}
