package main

import (
	"gogi"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: main.go <namespace> <function>")
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
		if info.GetName() == os.Args[2] {
			gofunc, cfunc := gogi.WriteFunction(info)
			println("/*\n" + cfunc + "\n*/\n")
			println(gofunc)
		}
	}
}
