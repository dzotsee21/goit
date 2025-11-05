package refs

import (
	filesmodule "goit/src/modules/files"
	"goit/src/modules/objects"
	"goit/src/modules/utils"
	"regexp"
)

func Hash(refOrHash string) interface{} {
	if objects.Exists(refOrHash) {
		return refOrHash
	} else {
		terminalRef := TerminalRef(refOrHash)
		if terminalRef == "FETCH_HEAD" {
			return fetchHeadBranchToMerge(headBranchName())
		}
		if exists(terminalRef) {
			return filesmodule.Read(filesmodule.GoitPath(terminalRef))
		}
	}

	return ""
}

func TerminalRef(ref string) string {
	if ref == "HEAD" && !isHeadDetached() {
		headPath := filesmodule.GoitPath("HEAD")
		content := filesmodule.Read(headPath)

		re := regexp.MustCompile(`ref: (refs/heads/.+)`)
		matches := re.FindStringSubmatch(content)

		if len(matches) > 1 {
			return matches[1]
		}

		return ""
	}
	if isRef(ref) {
		return ref
	} else {
		return toLocalRef(ref)
	}
}

func isHeadDetached() bool {
	headPath := filesmodule.GoitPath("HEAD")
	content := filesmodule.Read(headPath)

	re := regexp.MustCompile(`ref: (refs/heads/.+)`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 0 {
		return true
	} else {
		return false
	}
}

func isRef(ref string) bool {
	re1 := regexp.MustCompile(`^refs/heads/[A-Za-z-]+$`)
	matches1 := re1.FindStringSubmatch(ref)
	re2 := regexp.MustCompile(`^refs/remotes/[A-Za-z-]+/[A-Za-z-]+$`)
	matches2 := re2.FindStringSubmatch(ref)

	specialRefs := []string{"HEAD", "FETCH_HEAD", "MERGE_HEAD"}
	for _, s := range specialRefs {
		if ref == s {
			return true
		}
	}

	return len(matches1) > 0 || len(matches2) > 0
}

func toLocalRef(name string) string {
	return "refs/heads/" + name
}

func fetchHeadBranchToMerge(branchName string) []string {
	lines := utils.Lines(filesmodule.Read(filesmodule.GoitPath("FETCH_HEAD")))

	re := regexp.MustCompile(`^.+ branch ` + branchName + ` of`)
	var filteredLines []string
	for _, line := range lines {
		if re.MatchString(line) {
			filteredLines = append(filteredLines, line)
		}
	}

	re1 := regexp.MustCompile(`^([^ ]+) `)
	var result []string
	for _, fLine := range filteredLines {
		matches := re1.FindStringSubmatch(fLine)
		if len(matches) > 1 {
			result = append(result, matches[1])
		}
	}

	return result
}

func headBranchName() string {
	if !isHeadDetached() {
		content := filesmodule.Read(filesmodule.GoitPath("HEAD"))
		re1 := regexp.MustCompile(`refs/heads/(.+)`)
		matches := re1.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func exists(ref string) bool {
	return isRef(ref) && filesmodule.Exists(filesmodule.GoitPath(ref))
}
