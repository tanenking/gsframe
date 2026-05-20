package gsframe

import (
	"github.com/tanenking/gsframe/internal/constants"
)

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) bool {
	return constants.PathExists(path)
}

// 判断所给路径是否为文件夹
func IsDir(path string) (is bool, exists bool) {
	return constants.IsDir(path)
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return constants.IsFile(path)
}

func CheckFileIsExist(filename string) bool {
	return constants.CheckFileIsExist(filename)
}
