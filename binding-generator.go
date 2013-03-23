package main

import (
	"encoding/json"
	"fmt"
	"gogi"
	"io/ioutil"
	"path/filepath"
	"os"
	//"os/exec"
	"strings"
)

type Deps struct {
	Pkgs []string
	Headers []string
	Typedefs map[string]string
	Imports []string
}

var knownPackages map[string] Deps
var blacklist map[string] bool

func CreatePackageRoot(pkg string) string {
	root := filepath.Join("src/gi", pkg)
	os.Remove(root) ; os.MkdirAll(root, os.ModePerm)
	return root
}

func OpenSourceFile(root, pkg string) *os.File {
	f, _ := os.Create(filepath.Join(root, pkg + ".go"))
	return f
}

func Process(namespace string) {
	infos := gogi.GetInfos(namespace)

	fmt.Printf("Generating bindings for %s...\n", namespace)

	var c_code string
	var go_code string
	for _, info := range infos {
		if info.IsDeprecated() {
			continue
		}
		var g, c string
		switch info.Type {
			case gogi.Object:
				g, c = gogi.WriteObject(info)
			case gogi.Struct:
				g, c = gogi.WriteStruct(info)
			case gogi.Enum:
				g, c = gogi.WriteEnum(info)
			case gogi.Function:
				g, c = gogi.WriteFunction(info, nil)
			default:
				//fmt.Printf("unknown info '%s' of type %s\n", info.GetName(), gogi.InfoTypeToString(info.Type))
		}

		/*
		if info.Type == gogi.Object {
			switch info.GetName() {
				case "Window", "Bin", "Container", "Widget", "InitiallyUnowned", "Object":
					g, c := gogi.WriteObject(info)
					go_code += g + "\n"
					if c != "" {
						c_code += c + "\n"
					}
			}
		} else if info.Type == gogi.Enum {
			g, c := gogi.WriteEnum(info)
			go_code += g + "\n"
			if c != "" {
				c_code += c + "\n"
			}
		}
		*/

		if g != "" { go_code += g + "\n" }
		if c != "" { c_code += c + "\n" }
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
		for key, value := range deps.Typedefs {
			f.WriteString(fmt.Sprintf("#define %s %s\n", key, value))
		}
		f.WriteString("\n")
		f.WriteString("GList *EMPTY_GLIST = NULL;\n")
	}
	f.WriteString(c_code + "\n")
	f.WriteString("*/\nimport \"C\"\n")
	for _, imp := range deps.Imports {
		f.WriteString(fmt.Sprintf("import \"%s\"\n", imp))
	}
	// TODO: find a way to keep track of additional imports
	//f.WriteString("import \"unsafe\"\n\n")
	f.WriteString(go_code)
	f.Close()

	// now build it
	/*
	println("Compiling...")
	cmd := exec.Command("go", "install", pkg)
	err = cmd.Run()
	if err != nil {
		println(err.Error())
	}
	*/
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run binding-generator.go <namespace>")
		return
	}

	{
		knownPackages = make(map[string]Deps)
		deps_config, _ := os.Open("deps.json")
		deps_decoder := json.NewDecoder(deps_config)
		err := deps_decoder.Decode(&knownPackages)
		if err != nil {
			println(err.Error())
		}
		deps_config.Close()
	}

	{
		blacklist = make(map[string]bool)
		content, err := ioutil.ReadFile("blacklist")
		if err != nil {
			println("error reading blacklist:", err.Error())
			return
		}

		lines := strings.Split(string(content), "\n")
		var namespace string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 || strings.HasPrefix(line, "#") {
				continue
			}

			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				namespace = strings.TrimRight(strings.TrimLeft(line, "["), "]")
			} else {
				blacklist[namespace + "." + line] = true
			}
		}
	}

	namespace := os.Args[1]
	gogi.Init()

	loaded := gogi.LoadNamespace(namespace)
	if !loaded {
		fmt.Printf("Failed to load namespace '%s'\n", namespace)
		return
	}

	gogi.SetBlacklist(blacklist)

	dependencies := gogi.GetDependencies(namespace)
	for _, dep := range dependencies {
		nameAndVersion := strings.Split(dep, "-")
		name := nameAndVersion[0]
		//version := nameAndVersion[1]
		Process(name)
	}

	Process(namespace)
	fmt.Println("done.")
}
