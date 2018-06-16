package main

import (
	"fmt"
	"gogi"
	"os"
	"strings"
)

func Display(info *gogi.GiInfo) {
	name := info.GetName()
	fmt.Println(name)
	fmt.Println(strings.Repeat("=", len(name)))
	fmt.Printf("type: %s\n", gogi.InfoTypeToString(info.Type))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run introspector.go <namespace> [symbol...]")
		return
	}

	namespace := os.Args[1]
	gogi.Init()

	loaded := gogi.LoadNamespace(namespace)
	if !loaded {
		fmt.Printf("Failed to load namespace '%s'\n", namespace)
		return
	}

	if len(os.Args) == 2 {
		infos := gogi.GetInfos(namespace)
		for _, info := range infos {
			fmt.Println(info.GetName())
		}
	} else {
		for i := 2; i < len(os.Args); i++ {
			symbol := os.Args[i]
			info := gogi.GetInfoByName(namespace, symbol)
			if info != nil {
				Display(info)
			} else {
				fmt.Printf("Symbol '%s' not found\n", symbol)
			}
		}
	}
}
