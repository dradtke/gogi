package main

import (
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

var KnownPackages map[string]Deps

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
	if len(os.Args) < 3 {
		fmt.Println("usage: main.go <namespace> <function>")
		return
	}

	// TODO: save this in a config file
	KnownPackages = map[string]Deps {
		"Gtk" : Deps{
			[]string{"gtk+-3.0"},
			[]string{"gtk/gtk.h"},
		},
	}

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
		if info.GetName() == os.Args[2] {
			gofunc, cfunc := gogi.WriteFunction(info)
			c_code += cfunc + "\n"
			go_code += gofunc + "\n"
		}
	}

	pkg := strings.ToLower(namespace)
	deps, deps_exist := KnownPackages[namespace]
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
