package logs

type Location struct {
	File string   `json:"file"`
	Path []string `json:"path"`
}

type ErrorChangeInMissingBranch struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}
