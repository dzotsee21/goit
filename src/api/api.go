package api

import (
	"fmt"
	"goit/src/modules/config"
	"goit/src/modules/diff"
	filesmodule "goit/src/modules/files"
	"goit/src/modules/index"
	"goit/src/modules/merge"
	"goit/src/modules/objects"
	"goit/src/modules/refs"
	"goit/src/modules/utils"
	workingcopy "goit/src/modules/working_copy"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
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

func Commit(cmds map[string]string) string {
	filesmodule.AssertInRepo()

	treeHash := writeTree()

	var headDesc string
	if refs.IsHeadDetached() {
		headDesc = "detached HEAD"
	} else {
		headDesc = refs.HeadBranchName()
	}
	if refs.Hash("HEAD").(string) != "" && treeHash == objects.TreeHash(objects.Read(refs.Hash("HEAD").(string))) {
		fmt.Println("# on" + headDesc + "\nnothing to commit, working dir clean")
	} else {

		conflictedPaths := index.ConflictedPaths()
		if refs.IsMergeInProgress() != "" && len(conflictedPaths) > 0 {
			fmt.Printf("%#v\n", conflictedPaths)
		} else {
			var m string
			if refs.IsMergeInProgress() != "" {
				m = filesmodule.Read(filesmodule.GoitPath("MERGE_MSG"))
			} else {
				m = cmds["m"]
			}

			commitHash := objects.WriteCommit(treeHash, m, refs.CommitParentHashes())

			updateRef("HEAD", commitHash)

			if refs.IsMergeInProgress() != "" {
				err := os.Remove(filesmodule.GoitPath("MERGE_MSG"))
				if err != nil {
					fmt.Println("Couldn't find the MERGE_MSG")
				}
				return "merge made by the three-way strategy"
			} else {
				return "[" + headDesc + " " + commitHash + "] " + m
			}
		}
	}

	return ""
}

func writeTree() string {
	filesmodule.AssertInRepo()

	return objects.WriteTree(filesmodule.NestFlatTree(index.Toc()))
}

func updateRef(refToUpdate, refToUpdateTo string) {
	filesmodule.AssertInRepo()

	hash := refs.Hash(refToUpdateTo).(string)

	if !objects.Exists(hash) {
		fmt.Println(refToUpdateTo + " not a valid SHA1")
	}
	if !refs.IsRef(refToUpdate) {
		fmt.Println("cannot lock the ref " + refToUpdate)
	}
	if objects.IsType(objects.Read(hash)) != "commit" {
		branch := refs.TerminalRef(refToUpdate)
		fmt.Println(branch + " cannot refer to non-commit object " + hash + "\n")
	} else {
		refs.Write(refs.TerminalRef(refToUpdate), hash)
	}
}

func Branch(name interface{}, opts []string) string {
	filesmodule.AssertInRepo()

	if name == nil {
		branchHeads := utils.MapKeys(refs.LocalHeads())
		localBranches := ""
		for _, branch := range branchHeads {
			if branch == refs.HeadBranchName() {
				localBranches += "* " + branch
			} else {
				localBranches += "  " + branch
			}
		}

		return localBranches
	}

	if refs.Hash("HEAD") == "" {
		fmt.Println(refs.HeadBranchName() + " not a valid object name")
	}
	if refs.Exists(refs.ToLocalRef(name.(string))) {
		fmt.Println("A branch named " + name.(string) + " already exists")
	} else {
		updateRef(refs.ToLocalRef(name.(string)), refs.Hash("HEAD").(string))
	}

	return ""
}

func Checkout(ref string) string {
	toHash := refs.Hash(ref).(string)

	if !objects.Exists(toHash) {
		fmt.Println(ref + " did not match any file(s) known to Goit")
	}
	if objects.IsType(objects.Read(toHash)) != "commit" {
		fmt.Println("reference is not a tree: " + ref)
	}
	if ref == refs.HeadBranchName() || ref == filesmodule.Read(filesmodule.GoitPath("HEAD")) {
		return "already on " + ref
	} else {
		paths := diff.ChangedFilesCommitWouldOverwrite(toHash)
		if len(paths) > 0 {
			fmt.Println("local changes would be lost\n" + strings.Join(paths, "\n") + "\n")
		} else {
			err := os.Chdir(filesmodule.WorkingCopyPath(""))
			if err != nil {
				log.Fatal(err)
			}

			isDetachingHead := objects.Exists(ref)

			workingcopy.Write(diff.Diff(refs.Hash("HEAD"), toHash))

			if isDetachingHead {
				refs.Write("HEAD", toHash)
			} else {
				refs.Write("HEAD", "ref: "+refs.ToLocalRef(ref))
			}

			index.Write(index.TocToIndex(objects.CommitToc(toHash)))

			if isDetachingHead {
				return "note: checking out " + toHash + "\nYou are in detached HEAD state."
			} else {
				return "switched to branch " + ref
			}
		}
	}

	return ""
}

func Diff(ref1, ref2 interface{}, cmds []string) string {
	filesmodule.AssertInRepo()

	if ref1 != nil && refs.Hash(ref1.(string)) == "" {
		fmt.Println("ambiguous argument " + ref1.(string) + ": unknown revision")
	}
	if ref2 != nil && refs.Hash(ref2.(string)) == "" {
		fmt.Println("ambiguous argument " + ref2.(string) + ": unknown revision")
	} else {
		nameToStatus := diff.NameStatus(diff.Diff(refs.Hash(ref1.(string)), refs.Hash(ref2.(string))))

		statusKeys := utils.MapKeys(nameToStatus)

		changedFiles := ""
		for _, key := range statusKeys {
			changedFiles += nameToStatus[key].(string) + " " + key
		}

		return changedFiles
	}

	return ""
}

func Remote(command, name, path string) string {
	filesmodule.AssertInRepo()

	if command != "add" {
		fmt.Println("unsupported")
	}

	_, exists := config.Read()["remote"].(map[string]interface{})[name]
	if exists {
		fmt.Println("remote " + name + " already exists")
	} else {
		config.Write(utils.SetIn(config.Read(), []interface{}{"remote", name, "url", path}))
		return "\n"
	}

	return ""
}

func Fetch(remote, branch interface{}) []string {
	filesmodule.AssertInRepo()

	if remote == nil || branch == nil {
		fmt.Println("unsupported")
	}
	_, exists := config.Read()["remote"].(map[string]interface{})[remote.(string)]
	if !exists {
		fmt.Println(remote.(string) + " does not appear to be a git repository")
	} else {
		remoteUrl := config.Read()["remote"].(map[string]interface{})[remote.(string)].(map[string]interface{})["url"]
		remoteRef := refs.ToRemoteRef(remote.(string), branch.(string))

		newHash := utils.OnRemote(remoteUrl.(string))(func(ref interface{}) interface{} {
			return refs.Hash(ref.(string))
		}, branch.(string))

		oldHash := refs.Hash(remoteRef).(string)

		remoteObjects := utils.OnRemote(remoteUrl.(string))(func(_ interface{}) interface{} {
			return objects.AllObjects()
		}).([]string)

		for _, obj := range remoteObjects {
			objects.Write(obj)
		}

		updateRef(remoteRef, newHash.(string))

		refs.Write("FETCH_HEAD", newHash.(string)+" branch "+branch.(string)+" of "+remoteUrl.(string))

		var addit string
		if merge.IsAForceFetch(oldHash, newHash.(string)) {
			addit = " (forced)"
		} else {
			addit = ""
		}

		return []string{"From " + remoteUrl.(string),
			"Count " + strconv.Itoa(len(remoteObjects)),
			branch.(string) + " -> " + remote.(string) + "/" + branch.(string) + addit + "\n",
		}
	}

	return []string{}
}

func Merge(ref string) string {
	filesmodule.AssertInRepo()

	receiverHash := refs.Hash("HEAD").(string)

	giverHash := refs.Hash(ref).(string)

	if refs.IsHeadDetached() {
		log.Fatal("unsupported")
	}
	if giverHash == "" || objects.IsType(objects.Read(giverHash)) != "commit" {
		log.Fatal(ref + ": expected commit type")
	}
	if objects.IsUpToDate(receiverHash, giverHash) {
		return "Already up-to-date"
	} else {
		paths := diff.ChangedFilesCommitWouldOverwrite(giverHash)
		if len(paths) > 0 {
			log.Fatal("local changes would be lost\n" + strings.Join(paths, "\n"))
		}
		if merge.CanFastForward(receiverHash, giverHash) {
			merge.WriteFastForwardMerge(receiverHash, giverHash)

			return "fast-forward"
		} else {
			merge.WriteNonFastForwardMerge(receiverHash, giverHash, ref)

			if merge.HasConflicts(receiverHash, giverHash) {
				return "auto merge failed. fix conflicts and commit the result."
			} else {
				return Commit(map[string]string{})
			}
		}
	}
}

func Pull(remote, branch string) string {
	filesmodule.AssertInRepo()

	Fetch(remote, branch)

	return Merge("FETCH_HEAD")
}

func Push(remote, branch interface{}, cmds map[string]string) string {
	filesmodule.AssertInRepo()

	if remote == nil || branch == nil {
		log.Fatal("unsupported")
	} else {
		remotePath := config.Read()["remote"].(map[string]map[string]string)[remote.(string)]["url"]
		remoteCall := utils.OnRemote(remotePath)

		if remoteCall(func(interface{}) interface{} {
			return refs.IsCheckedOut
		}, branch.(string)).(bool) {
			log.Fatal("refusing to update checked out branch " + branch.(string))
		} else {
			receiverHash := remoteCall(func(interface{}) interface{} {
				return refs.Hash
			}, branch.(string)).(string)

			giverHash := refs.Hash(branch.(string)).(string)

			if objects.IsUpToDate(receiverHash, giverHash) {
				return "already up-to-date"
			}
			_, exists := cmds["f"]
			if !exists && !merge.CanFastForward(receiverHash, giverHash) {
				log.Fatal("failed to push some refs to " + remotePath)
			} else {
				for _, obj := range objects.AllObjects() {
					remoteCall(func(interface{}) interface{} {
						return objects.Write
					}, obj)
				}

				remoteCall(func(interface{}) interface{} {
					return updateRef
				}, refs.ToLocalRef(branch.(string)), giverHash)

				updateRef(refs.ToRemoteRef(remote.(string), branch.(string)), giverHash)
				
				objLen := strconv.Itoa(len(objects.AllObjects()))
				return "[To " + remotePath + "\nCount " + objLen + "\n" + branch.(string) + " -> " + branch.(string)
			}
		}
	}

	return ""
}