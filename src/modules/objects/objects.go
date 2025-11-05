package objects

import (
	filesmodule "goit/src/modules/files"
	"goit/src/modules/utils"
	"path/filepath"
	"regexp"
	"strings"
)

func Write(str string) string {
	filesmodule.Write(filepath.Join(filesmodule.GoitPath(""), "objects", utils.Hash(str)), str)

	return utils.Hash(str)
}

func Exists(objectHash string) bool {
	return filesmodule.Exists(filepath.Join(filesmodule.GoitPath(""), "objects", objectHash))
}

func CommitToc(hash string) map[string]interface{} {
	return filesmodule.FlattenNestedTree(fileTree(treeHash(read(hash)), nil), nil, "")
}

func treeHash(str string) string {
	if isType(str) == "commit" {
		re := regexp.MustCompile(`\s+`)

		return re.Split(str, -1)[1]
	}

	return ""
}

func isType(str string) string {
	types := map[string]string{"commit": "commit", "tree": "tree", "blob": "tree"}

	strType, ok := types[strings.Split(str, " ")[0]]

	if !ok {
		return "blob"
	} else {
		return strType
	}
}

func fileTree(treeHash string, tree map[string]interface{}) map[string]interface{} {
	if tree != nil {
		return fileTree(treeHash, map[string]interface{}{})
	}

	objectLines := utils.Lines(read(treeHash))

	for _, line := range objectLines {
		lineTokens := strings.Split(line, " ")
		if lineTokens[0] == "tree" {
			tree[lineTokens[2]] = fileTree(lineTokens[1], map[string]interface{}{})
		} else {
			tree[lineTokens[2]] = lineTokens[1]
		}
	}

	return tree
}

func read(objectHash string) string {
	if objectHash != "" {
		objectPath := filepath.Join(filesmodule.GoitPath(""), "objects", objectHash)
		if filesmodule.Exists(objectPath) {
			return filesmodule.Read(objectPath)
		}
	}

	return ""
}
