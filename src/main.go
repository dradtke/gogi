package main

import (
	"encoding/json"
	"fmt"
	"gogi"
	"path/filepath"
	"os"
	"os/exec"
	"strings"
)

type Deps struct {
	Pkgs []string
	Headers []string
}

func CreatePackageRoot(pkg string) string {
	root := filepath.Join(os.TempDir(), "gogi-" + pkg)
	os.Remove(root) ; os.Mkdir(root, os.ModePerm)
	return root
}

func OpenSourceFile(root string, pkg string) *os.File {
	src := filepath.Join(root, "src/" + pkg)
	os.Remove(src) ; os.MkdirAll(src, os.ModePerm)
	f, _ := os.Create(filepath.Join(src, pkg + ".go"))
	return f
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: main.go <namespace>")
		return
	}

	knownPackages := make(map[string]Deps)
	deps_config, _ := os.Open("deps.json")
	deps_decoder := json.NewDecoder(deps_config)
	deps_decoder.Decode(&knownPackages)
	deps_config.Close()

	namespace := os.Args[1]
	gogi.Init()

	infos, err := gogi.GetInfos(namespace)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	println("Generating...")
	var c_code string
	var go_code string
	for _, info := range infos {
		/*
		if info.Type == gogi.Function {
			go, c := gogi.WriteFunction(info, nil)
			c_code += c + "\n"
			go_code += go + "\n"
		} else if info.Type == gogi.Object {
			decl := gogi.WriteObject(info)
			go_code += decl + "\n"
		}
		*/
		if info.Type == gogi.Object && info.GetName() == "Window" {
			g, c := gogi.WriteObject(info)
			go_code += g + "\n"
			c_code += c + "\n"
		}
	}

	pkg := strings.ToLower(namespace)
	deps, deps_exist := knownPackages[namespace]
	pkg_root := CreatePackageRoot(pkg)

	f := OpenSourceFile(pkg_root, pkg)
	f.WriteString("package " + pkg + "\n\n")
	f.WriteString("/*\n")
	if deps_exist {
		f.WriteString(fmt.Sprintf("#cgo pkg-config: %s\n", strings.Join(deps.Pkgs, " ")))
		for _, header := range deps.Headers {
			f.WriteString(fmt.Sprintf("#include <%s>\n", header))
		}
	}
	f.WriteString(c_code + "\n")
	f.WriteString("*/\nimport \"C\"\n")
	// TODO: find a way to keep track of additional imports
	f.WriteString("import \"unsafe\"\n\n")
	f.WriteString(go_code)
	f.Close()

	// now build it
	println("Compiling...")
	gopath := filepath.SplitList(os.Getenv("GOPATH"))
	gopath = append(gopath, pkg_root)
	os.Setenv("GOPATH", strings.Join(gopath, string(filepath.ListSeparator)))
	cmd := exec.Command("go", "install", pkg)
	cmd.Run()

	println("done.")
}
