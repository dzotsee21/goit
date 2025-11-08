package workingcopy

import (
	"goit/src/modules/diff"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/objects"
	"goit/src/modules/utils"
	"log"
	"os"
)

func Write(dif map[string]interface{}) {

	composeConflict := func(receiverFileHash, giverFileHash string) string {
		return "<<<<<\n" + objects.Read(receiverFileHash) + "\n======\n" + objects.Read(giverFileHash) + "\n>>>>>>\n"
	}

	difKeys := utils.MapKeys(dif)

	for _, key := range difKeys {
		info := dif[key].(map[string]interface{}) 
		if info["status"] == diff.FILE_STATUS["ADD"] {
			if info["receiver"] != nil {
				filesmodule.Write(filesmodule.WorkingCopyPath(key), objects.Read(info["receiver"].(string)))
			}
			if info["giver"] != nil {
				filesmodule.Write(filesmodule.WorkingCopyPath(key), objects.Read(info["giver"].(string)))
			}
		}
		if info["status"] == diff.FILE_STATUS["CONFLICT"] {
			filesmodule.Write(filesmodule.WorkingCopyPath(key), composeConflict(info["receiver"].(string), info["giver"].(string)))
		}
		if info["status"] == diff.FILE_STATUS["MODIFY"] {
			filesmodule.Write(filesmodule.WorkingCopyPath(key), objects.Read(info["giver"].(string)))
		}
		if info["status"] == diff.FILE_STATUS["DELETE"] {
			err := os.Remove(filesmodule.WorkingCopyPath(key))
			if err != nil {
				log.Fatal(err)
			}
		}

		entries, _ := os.ReadDir(filesmodule.WorkingCopyPath(""))

		for _, entry := range entries {
			if entry.Name() != ".goit" {
				filesmodule.RmEmptyDirs(entry.Name())
			}
		}
	}
}