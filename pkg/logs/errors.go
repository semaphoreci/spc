package logs

// revive:disable:max-public-structs

type Location struct {
	File string   `json:"file"`
	Path []string `json:"path"`
}

type ErrorChangeInMissingBranch struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func (e *ErrorChangeInMissingBranch) Error() string {
	return e.Message
}

type ErrorInvalidWhenExpression struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func (e *ErrorInvalidWhenExpression) Error() string {
	return e.Message
}

type ErrorChangeInGitFailure struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func (e *ErrorChangeInGitFailure) Error() string {
	return e.Message
}
