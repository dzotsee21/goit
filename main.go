package main

import (
	"flag"
	"goit/src/api"
)

func main() {
	initPtr := flag.Bool("init", false, "initialize Goit")
	barePtr := flag.Bool("bare", false, "initialize bare Goit")
	addPtr := flag.Bool("add", false, "add files to stage")
	// rmPtr := flag.Bool("rm", false, "remove files")

	flag.Parse()

	args := flag.Args()

	if *initPtr {
		api.Init(*barePtr)
	}

	if *addPtr {
		var path string
		if len(args) > 0 {
			path = args[0]
		}

		api.Add(path, "")
	}
}
