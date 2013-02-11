package gogi

import (
	"fmt"
	"strings"
)

type Argument struct {
	info *GiInfo
	name string
	cname string
	typ *GiInfo
}

func CamelCase(str string) (name string) {
	words := strings.Split(str, "_")
	for _, word := range words {
		name += strings.Title(word)
	}
	return
}

// return a marshaled Go function and any necessary C wrapper
func WriteFunction(info *GiInfo) (string, string) {
	var text string = "func "
	var wrapper string
	c_func := info.GetFullName()

	// TODO: check if this is a method on an object
	ret := info.GetReturnType() ; defer ret.Free()
	ret_ctype := CType(ret, In)
	wrapper += ret_ctype + " "

	text += CamelCase(info.GetName()) + "("
	wrapper += "gogi_" + c_func + "("

	argc := info.GetNArgs()
	args := make([]Argument, argc)
	for i := 0; i < argc; i++ {
		arg := info.GetArg(i)
		args[i] = Argument{arg,arg.GetName(),"",arg.GetType()}
		text += fmt.Sprintf("%s %s", args[i].name, GoType(args[i].typ, arg.GetDirection()))
		wrapper += fmt.Sprintf("%s %s", CType(args[i].typ, args[i].info.GetDirection()), args[i].name)
		if i < argc-1 {
			text += ", "
			wrapper += ", "
		}
	}
	text += ") "
	wrapper += ") "

	// TODO: check for a return value
	ret_gotype, ret_marshal := CToGo(ret, "retval", "c_retval")
	if ret_gotype != "" {
		text += ret_gotype + " "
	}

	text += "{\n" // Go function open
	wrapper += "{\n"
	// marshal
	for i := 0; i < argc; i++ {
		args[i].cname = "c_" + args[i].name
		ctype, marshal := GoToC(args[i].typ, args[i], args[i].cname)
		text += fmt.Sprintf("\tvar %s %s\n", args[i].cname, ctype)
		text += fmt.Sprintf("\t%s\n", marshal)
		text += fmt.Sprint("\n")
	}
	go_argnames := make([]string, len(args))
	c_argnames := make([]string, len(args))
	for i, arg := range args {
		dir := arg.info.GetDirection()
		if dir == Out || dir == InOut {
			go_argnames[i] += "&"
			//c_argnames[i] += "&"
		}
		go_argnames[i] += arg.cname
		c_argnames[i] += arg.name
	}
	text += "\tc_retval, _ := C.gogi_" + c_func + "(" + strings.Join(go_argnames, ", ") + ")\n"
	if ret_gotype != "" {
		text += "\t" + ret_marshal + "\n\treturn retval\n"
	}

	// TODO: catch errno
	wrapper += "\t"
	if ret_ctype != "void" {
		wrapper += "return "
	}
	wrapper += c_func + "(" + strings.Join(c_argnames, ", ") + ");\n"

	text += "}"
	wrapper += "}"

	return text, wrapper
}
