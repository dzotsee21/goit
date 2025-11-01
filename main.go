package main

import (
	"flag"
	"fmt"
	commands "goit/src"
)

func main() {

	initPtr := flag.Bool("init", true, "initialize Giot dir")

	flag.Parse()

	if (*initPtr) {
		commands.Init()
	}

	args := flag.Args()
	if len(args) > 0 {
		fmt.Printf("args: %v\n", args)
	}

}