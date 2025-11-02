package main

import (
	"flag"
	"fmt"
	"goit/src/api"
)

func main() {

	initPtr := flag.Bool("init", false, "initialize Goit")
	barePtr := flag.Bool("bare", false, "initialize bare Goit")
	addPtr := flag.Bool("add", false, "add files in stage")

	flag.Parse()

	if (*initPtr) {
		api.Init(*barePtr)
	}

	if (*addPtr) {
		api.Add(".", "")
	}

	args := flag.Args()
	if len(args) > 0 {
		fmt.Printf("args: %v\n", args)
	}

}