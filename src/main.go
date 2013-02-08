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
			DispFunction(info)
		}
	}
}

func DispFunction(info *gogi.GiInfo) {
	fmt.Printf("func %s(", info.GetName())
	arg_count := info.GetNArgs()
	for i := 0; i < arg_count; i++ {
		arg := info.GetArg(i)
		fmt.Printf("%s %s", arg.GetName(), gogi.GoType(arg.GetType()))
		if i != arg_count-1 {
			fmt.Print(", ")
		}
	}
	fmt.Print(")")
	// return type
}
