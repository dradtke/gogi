package main

import (
	"gogi"
	"fmt"
)

func main() {
	gogi.Init()
	//gogi.ArrayTest()
	infos := gogi.GetInfos("GLib")
	for _, info := range infos {
		if info.Type == gogi.Function {
			flags, _ := info.GetFunctionFlags()
			fmt.Printf("%v\n", flags)
		}
	}
}
