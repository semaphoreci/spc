package consolelogger

import (
	"fmt"
	"strings"
)

var nesting = 0

func EmptyLine() {
	fmt.Println(nestPadding())
}

func Info(msg string) {
	for _, line := range strings.Split(msg, "\n") {
		fmt.Println(nestPadding() + line)
	}
}

func InfoNumberListLn(num int, msg string) {
	fmt.Printf("%03d | %s\n", num, msg)
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

	res += " | "

	return res
}
