package helper

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func ToJson(v interface{}) string {
	b, err := jsoniter.ConfigFastest.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func FromJson(j string, out interface{}) error {
	if len(j) <= 0 {
		return fmt.Errorf("FromJson, data empty")
	}
	err := jsoniter.ConfigFastest.Unmarshal([]byte(j), out)
	return err
}
