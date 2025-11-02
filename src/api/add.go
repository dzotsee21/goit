package api

import (
	"fmt"
	"goit/src/utils"
)

func Add(path, _ string) {
	utils.AssertInRepo()

	addedFiles := utils.LsRecursive(path)

	if len(addedFiles) == 0 {
		fmt.Println("didn't match any files")
		return
	} else {
		fmt.Println(addedFiles)
	}
}