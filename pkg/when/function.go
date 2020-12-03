package when

type FunctionInput struct {
	Name   string      `json:"name"`
	Params interface{} `json:"params"`
	Result bool        `json:"result"`
}
