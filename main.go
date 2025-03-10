package main

import (
	"github.com/CosmicBDry/helm-push/cmd"
)

//declare -x GOOS=linux  #编译系统类型为linux
//declare -x GOARCH=amd64 #编译架构为amd64
//go build -ldflags="-w -s" -o helm-push main.go  #编译成二进制程序

func main() {

	cmd.Execute()

}
