package whencli

import gabs "github.com/Jeffail/gabs/v2"

func fromJSON(s string) *gabs.Container {
	res, err := gabs.ParseJSON([]byte(s))
	if err != nil {
		panic(err)
	}

	return res
}
