package constants

import "os"

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) (is bool, exists bool) {
	is = false
	exists = false

	s, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		exists = true
		return
	}
	exists = true
	is = s.IsDir()
	return
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	is, _ := IsDir(path)
	return !is
}

func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
