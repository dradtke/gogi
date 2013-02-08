package main

import (
	"gogi"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: main.go <namespace>")
		return
	}
	gogi.Init()
	//gogi.ArrayTest()
	infos, err := gogi.GetInfos(os.Args[1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	for _, info := range infos {
		if info.GetName() == "init" {
			println(gogi.WriteFunction(info))
		}
	}
}
