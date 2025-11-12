package merge

import "goit/src/modules/objects"

func IsAForceFetch(receiverHash, giverHash string) bool {
	return !objects.IsAncestor(giverHash, receiverHash)
}