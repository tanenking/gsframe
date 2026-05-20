package configx

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ParseConfig(config_file string, config_value interface{}) (err error) {
	err = readConfig(config_file, config_value)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	return
}

func readConfig(config_file string, config_value interface{}) (err error) {
	var filePtr *os.File
	fullPath := config_file
	if filepath.IsAbs(fullPath) {
		filePtr, err = os.Open(fullPath)
		if err != nil {
			fmt.Printf("Open file failed [Err:%s]\n", err.Error())
			return
		}
	} else {
		dir, _ := os.Getwd()
		dir = strings.ReplaceAll(dir, "\\", "/")
		fullPath = fmt.Sprintf("%s/%s", dir, config_file)
		filePtr, err = os.Open(fullPath)
		if err != nil {
			paths := strings.Split(dir, "/")
			paths = paths[0 : len(paths)-1]
			dir = strings.Join(paths, "/")
			fmt.Printf("not found config %s, try found in gopath\n", fullPath)
			// dir = os.Getenv(constants.Env_Gopath)
			fullPath = fmt.Sprintf("%s/%s", dir, config_file)
			filePtr, err = os.Open(fullPath)
			if err != nil {
				fmt.Printf("Open file failed [Err:%s]\n", err.Error())
				return
			}
		}
	}

	defer filePtr.Close()

	allbytes, err := io.ReadAll(filePtr)
	if err != nil {
		fmt.Printf("Read file failed [Err:%s]\n", err.Error())
		return
	}

	err = json.Unmarshal(allbytes, config_value)
	if err != nil {
		fmt.Printf("Parse file failed [Err:%s]\n", err.Error())
		return
	}

	fmt.Printf("Parse file [%s] success \n", config_file)
	return
}
