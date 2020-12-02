package logs

import (
	"encoding/json"
	"os"
	"reflect"

	gabs "github.com/Jeffail/gabs/v2"
)

var loggerInstance *os.File
var currentPipelineFilePath string

func Open(path string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	loggerInstance = f
}

func SetCurrentPipelineFilePath(path string) {
	currentPipelineFilePath = path
}

func Log(e interface{}) {
	msg := toJSON(e)

	_, err := loggerInstance.WriteString(msg + "\n")
	if err != nil {
		panic(err)
	}
}

func toJSON(e interface{}) string {
	msg, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	jsonEvent, err := gabs.ParseJSON(msg)
	if err != nil {
		panic(err)
	}

	jsonEvent.Set(reflect.TypeOf(e).Name(), "type")
	jsonEvent.Set(currentPipelineFilePath, "location", "file")

	return jsonEvent.String()
}
