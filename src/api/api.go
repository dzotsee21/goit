package api

import (
	"fmt"
	"goit/src/modules/config"
	"goit/src/modules/diff"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/index"
	"goit/src/modules/utils"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func Init(isBare bool) {
	var currentFiles []string

	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		currentFiles = append(currentFiles, file.Name())
	}

	if filesmodule.InRepo() {
		return
	}

	goitStructure := map[string]interface{}{
		"HEAD":    "ref: refs/heads/master\n",
		"config":  config.ObjectToStr(map[string]interface{}{"core": map[string]interface{}{"": map[string]interface{}{"bare": isBare == true}}}),
		"objects": map[string]interface{}{},
		"refs": map[string]interface{}{
			"heads": map[string]interface{}{},
		},
	}

	base := ".goit"

	createStructure(base, goitStructure)
}

func createStructure(basePath string, structure map[string]interface{}) error {
	err := os.Mkdir(basePath, os.FileMode(0755))
	if err != nil {
		log.Fatal(err)
	}

	for name, value := range structure {
		filePath := filepath.Join(basePath, name)

		switch v := value.(type) {
		case string:
			if err := os.WriteFile(filePath, []byte(v), 0644); err != nil {
				return err
			}
		case map[string]interface{}:
			if err := createStructure(filePath, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func Add(path, _ string) {
	filesmodule.AssertInRepo()

	addedFiles := filesmodule.LsRecursive(path)

	if len(addedFiles) == 0 {
		fmt.Println(filesmodule.PathFromRepoRoot(path) + " didn't match any files")
		return
	} else {
		for _, file := range addedFiles {
			index.UpdateIndex(file, []string{"add"})
		}
	}
}

func Rm(path string, cmds []string) {
	filesmodule.AssertInRepo()

	filesToRm := index.MatchingFiles(path)

	if slices.Contains(filesToRm, "f") {
		fmt.Println("unsupported")
	}

	if len(filesToRm) == 0 {
		fmt.Println(filesmodule.PathFromRepoRoot(path) + " did not match any files")
	}

	if filesmodule.Exists(path) && utils.IsDir(path) && slices.Contains(filesToRm, "r") {
		fmt.Println("not removing " + path + " recursively without -r")
	} else {
		changesToRm := utils.Intersection(diff.AddedOrModifiedFiles(), filesToRm)
		if len(changesToRm) > 0 {
			fmt.Println("these files have changes:\n" + strings.Join(changesToRm, "\n") + "\n")
		} else {
			for _, file := range filesToRm {
				fileCopyPath := filesmodule.WorkingCopyPath(file)

				if filesmodule.Exists(fileCopyPath) {
					err := os.Remove(fileCopyPath)
					if err != nil {
						log.Fatal(err)
					}
				}
			}

			for _, file := range filesToRm {
				index.UpdateIndex(file, []string{"remove"})
			}
		}
	}
}
