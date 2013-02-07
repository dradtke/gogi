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
	fmt.Printf("%s\n", info.GetName())
	for i := 0; i < info.GetNArgs(); i++ {
		arg := info.GetArg(i)
		argType := arg.GetType()
		if argType.GetTag() == gogi.ArrayTag {
			switch argType.GetArrayType() {
				case gogi.CArray:
					println("It's a C array.")
				case gogi.GArray:
					println("It's a GArray.")
				case gogi.PtrArray:
					println("It's a ptr array.")
				case gogi.ByteArray:
					println("It's a byte array.")
			}
		}
	}
}
