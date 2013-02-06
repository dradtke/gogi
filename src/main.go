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
	fmt.Println("[Objects]")
	for _, info := range infos {
		if info.Type == gogi.Object {
			fmt.Printf("%v\n", info.GetName())
			n_methods, _ := info.GetNObjectMethods()
			for i := 0; i<n_methods; i++ {
				method_info, _ := info.GetObjectMethod(i)
				method_symbol, _ := method_info.GetSymbol()
				fmt.Printf("\t%s\n", method_symbol)
			}
		}
	}
}
