package environment

import "os"

const GitRefTypeTag = "tag"

func GitRefType() string {
	return os.Getenv("SEMAPHORE_GIT_REF_TYPE")
}
