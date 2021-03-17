package consolelogger

import "fmt"

var nesting = 0

func EmptyLine() {
	fmt.Println(nestPadding())
}

func Info(msg string) {
	fmt.Println(nestPadding() + msg)
}

func Infof(msg string, attr ...interface{}) {
	fmt.Printf(nestPadding()+msg, attr...)
}

func IncrementNesting() {
	nesting++
}

func DecreaseNesting() {
	nesting--
}

func nestPadding() string {
	if nesting == 0 {
		return ""
	}

	res := ""

	for i := 0; i < nesting; i++ {
		res += "   "
	}

	res += "| "

	return res
}
