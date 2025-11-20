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
	"goit/src/modules/status"
	"goit/src/modules/utils"
	workingcopy "goit/src/modules/working_copy"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func Init(bare interface{}) {
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

	isBare := bare.(string) == "true"
	goitStructure := map[string]interface{}{
		"HEAD":    "ref: refs/heads/master",
		"config":  config.ObjectToStr(map[string]interface{}{"core": map[string]interface{}{"": map[string]interface{}{"bare": isBare}}}),
		"objects": map[string]interface{}{},
		"refs": map[string]interface{}{
			"heads": map[string]interface{}{},
		},
	}

	cwd, _ := os.Getwd()
	if isBare {
		filesmodule.WriteFilesFromTree(goitStructure, cwd)
	} else {
		filesmodule.WriteFilesFromTree(map[string]interface{}{".goit": goitStructure}, cwd)
	}
}

func Add(path string) {
	filesmodule.AssertInRepo()

	var filesToIgnore []string
	if filesmodule.Exists(".goitignore") {
		content, err := os.ReadFile(".goitignore")
		if err != nil {
			log.Fatal(err)
		}

		fileNames := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
		filesToIgnore = fileNames
	}

	addedFiles := filesmodule.LsRecursive(path, filesToIgnore)

	if len(addedFiles) == 0 {
		fmt.Println(filesmodule.PathFromRepoRoot(path) + " didn't match any files")
		return
	} else {
		for _, file := range addedFiles {
			index.UpdateIndex(file, []string{"add"})
		}
	}
}

func Rm(path string) {
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

	indexToc := index.Toc()

	return objects.WriteTree(filesmodule.NestFlatTree(indexToc))
}

func updateRef(refToUpdate, refToUpdateTo string) error {
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

	return nil
}

func Branch(name interface{}) string {
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

	if refs.Hash("HEAD").(string) == "" {
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

func Diff(ref1, ref2 interface{}) string {
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

	remoteVal, exists := config.Read()["remote"]
	if exists {
		_, exists = remoteVal.(map[string]interface{})[name]
		if exists {
			fmt.Println("remote " + name + " already exists")
		}
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
	_, exists := config.Read()["[remote " + "\"" + remote.(string) + "\"]"]
	if !exists {
		fmt.Println(remote.(string) + " does not appear to be a git repository")
	} else {
		remotePath := config.Read()["[remote " + "\"" + remote.(string) + "\"]"].(map[string]interface{})["url"].(string)

		remoteRef := refs.ToRemoteRef(remote.(string), branch.(string))

		newHash := utils.OnRemote(remotePath)(func(interface{}) interface{} {
			return refs.Hash(branch.(string))
		}, branch.(string))

		oldHash := refs.Hash(remoteRef).(string)

		remoteObjects := utils.OnRemote(remotePath)(func(interface{}) interface{} {
			return objects.AllObjects()
		}).([]string)

		for _, obj := range remoteObjects {
			objects.Write(obj)
		}

		updateRef(remoteRef, newHash.(string))

		refs.Write("FETCH_HEAD", newHash.(string)+" branch "+branch.(string)+" of "+ remotePath)

		var addit string
		if merge.IsAForceFetch(oldHash, newHash.(string)) {
			addit = " (forced)"
		} else {
			addit = ""
		}

		return []string{"From " + remotePath,
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

func Push(remote, branch interface{}, cmd string) string {
	filesmodule.AssertInRepo()

	if remote == nil || branch == nil {
		log.Fatal("unsupported")
	} else {

		remotePath := config.Read()["[remote " + "\"" + remote.(string) + "\"]"].(map[string]interface{})["url"].(string)

		if utils.OnRemote(remotePath)(func(interface{}) interface{} {
			return refs.IsCheckedOut(branch.(string))
		}, branch.(string)).(bool) {
			log.Fatal("refusing to update checked out branch " + branch.(string))
		} else {
			receiverHash := utils.OnRemote(remotePath)(func(interface{}) interface{} {
				return refs.Hash(branch.(string))
			}, branch.(string)).(string)

			giverHash := refs.Hash(branch.(string)).(string)

			if objects.IsUpToDate(receiverHash, giverHash) {
				return "already up-to-date"
			}

			exists := cmd == "f"
			if !exists && !merge.CanFastForward(receiverHash, giverHash) {
				log.Fatal("failed to push some refs to " + remotePath)
			} else {
				for _, obj := range objects.AllObjects() {
					utils.OnRemote(remotePath)(func(interface{}) interface{} {
						return objects.Write(obj)
					}, obj)
				}

				utils.OnRemote(remotePath)(func(interface{}) interface{} {
					return updateRef(refs.ToLocalRef(branch.(string)), giverHash)
				}, refs.ToLocalRef(branch.(string)), giverHash)

				updateRef(refs.ToRemoteRef(remote.(string), branch.(string)), giverHash)

				objLen := strconv.Itoa(len(objects.AllObjects()))
				return "[To " + remotePath + "\nCount " + objLen + "\n" + branch.(string) + " -> " + branch.(string)
			}
		}
	}

	return ""
}

func Status() string {
	filesmodule.AssertInRepo()

	return status.ToString()
}

func Clone(remotePath, targetPath string, isBare interface{}) {

	if !filesmodule.Exists(remotePath) || !utils.OnRemote(remotePath)(func(interface{}) interface{} {
		return filesmodule.InRepo()
	}).(bool) {
		log.Fatal("repository " + remotePath + " doesn't exist")
	}

	files, _ := os.ReadDir(targetPath)
	if filesmodule.Exists(targetPath) && len(files) > 0 {
		log.Fatal(targetPath + " already exists and is not empty")
	} else {
		// wd, _ := os.Getwd()
		// remotePath = filepath.Clean(filepath.Join(wd, remotePath))

		if !filesmodule.Exists(targetPath) {
			err := os.Mkdir(targetPath, 0755)
			if err != nil {
				log.Fatal(err)
			}
		}

		utils.OnRemote(targetPath)(func(interface{}) interface{} {
			Init(isBare)

			cwd, _ := os.Getwd()
			rel, _ := filepath.Rel(cwd, remotePath)
			Remote("add", "origin", rel)

			remoteHeadHash := utils.OnRemote(remotePath)(func(interface{}) interface{} {
				return refs.Hash("master")
			}, "master")

			if remoteHeadHash != nil {
				Fetch("origin", "master")
				merge.WriteFastForwardMerge(nil, remoteHeadHash)
			}

			return "cloning into " + targetPath
		})
	}
}
