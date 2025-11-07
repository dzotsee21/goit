package objects

import (
	"fmt"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/utils"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func Write(str string) string {
	filesmodule.Write(filepath.Join(filesmodule.GoitPath(""), "objects", utils.Hash(str)), str)

	return utils.Hash(str)
}

func Exists(objectHash string) bool {
	return filesmodule.Exists(filepath.Join(filesmodule.GoitPath(""), "objects", objectHash))
}

func CommitToc(hash string) map[string]interface{} {
	return filesmodule.FlattenNestedTree(fileTree(TreeHash(Read(hash)), nil), nil, "")
}

func fileTree(TreeHash string, tree map[string]interface{}) map[string]interface{} {
	if tree != nil {
		return fileTree(TreeHash, map[string]interface{}{})
	}

	objectLines := utils.Lines(Read(TreeHash))

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

func WriteTree(tree map[string]interface{}) string {
	treeKeys := utils.MapKeys(tree)
	
	treeObject := ""
	for _, key := range treeKeys {
		switch kType := tree[key].(type) {
		case string:
			fmt.Println(kType)
			treeObject += "blob" + tree[key].(string) + " " + key + "\n"
		default:
			treeObject += "tree " + WriteTree(tree[key].(map[string]interface{})) + " " + key + "\n"
		}
	}

	return Write(treeObject)
}

func TreeHash(str string) string {
	if IsType(str) == "commit" {
		re := regexp.MustCompile(`\s+`)

		return re.Split(str, -1)[1]
	}

	return ""
}

func IsType(str string) string {
	types := map[string]string{"commit": "commit", "tree": "tree", "blob": "tree"}

	strType, ok := types[strings.Split(str, " ")[0]]

	if !ok {
		return "blob"
	} else {
		return strType
	}
}

func Read(objectHash string) string {
	if objectHash != "" {
		objectPath := filepath.Join(filesmodule.GoitPath(""), "objects", objectHash)
		if filesmodule.Exists(objectPath) {
			return filesmodule.Read(objectPath)
		}
	}

	return ""
}

func WriteCommit(treeHash, msg string, parentHashes []string) string {
	metaData := ""
	for _, h := range parentHashes {
		metaData += "parent " + h + "\n" + "Date:  " + time.Now().String() + "\n" + "\n" + "    " + msg + "\n"
	}
	return Write("commit " + treeHash + "\n" + metaData)
}
