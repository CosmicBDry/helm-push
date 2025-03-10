package tools

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// os.ReadFile读取Chart.yaml文件返回完整fileBytes字节流数据
func OsReadFile(filepath string) ([]byte, error) {
	fileBytes, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

// *os.File.Read(bindFileByte)读取Chart.yaml并写入bindFileByte字节流，但bindFileByte需要make初始化并指定长度才能被用；
// 导致bindFileByte中长度多出的部分会用0填充，破坏yaml原格式，最终yaml.Unmarshal无法解析数据到map实例中;
// 因此，需要对字节流bindFileByte中有效数据进行截取，如bindFileByte[:n]，截取后才能被yaml.Unmarshal解析。
func ReadFile(filepath string, bindFileByte []byte, chmod os.FileMode) ([]byte, error) {

	osFile, err := os.OpenFile(filepath, os.O_RDWR, chmod)
	if err != nil {
		return nil, err
	}

	defer osFile.Close()

	n, err := osFile.Read(bindFileByte)

	if err != nil {
		return nil, err
	}

	return bindFileByte[:n], nil
}

func WriteFile(filepath string, fileByte []byte, chmod os.FileMode) error {

	err := os.WriteFile(filepath, fileByte, chmod)
	return err
}

// 将字符串写入文件中，当文件不存在时创建文件，存在则清空文件内容重新写入
func WriteFileCreate(filepath string, content string, chmod os.FileMode) error {
	osFile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, chmod)
	if err != nil {
		return fmt.Errorf("os.OpenFile: %s", err.Error())
	}
	_, err = osFile.WriteString(content)
	if err != nil {
		return fmt.Errorf("osFile.WriteString: %s", err.Error())
	}
	return nil
}

// 自动换行缩进，读取data字符串中的每一行，每行的行首缩进为intent
func LineIndentBuidler(data, intent string) string {
	var builder strings.Builder
	lines := strings.Split(data, "\n")

	for index, line := range lines {
		builder.WriteString(intent)
		builder.WriteString("[" + strconv.Itoa(index+1) + "]")
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	return builder.String()

}
