package gogi

import "fmt"

type Argument struct {
	info *GiInfo
	name string
	typ *GiInfo
}

func WriteFunction(info *GiInfo) string {
	var text string = "func "
	// TODO: check if this is a method on an object
	text += fmt.Sprintf("%s(", info.GetName())
	argc := info.GetNArgs()
	args := make([]Argument, argc)
	for i := 0; i < argc; i++ {
		arg := info.GetArg(i)
		args[i] = Argument{arg,arg.GetName(),arg.GetType()}
		text += fmt.Sprintf("%s %s", args[i].name, GoType(args[i].typ))
		if i < argc-1 {
			text += ", ";
		}
	}
	text += ") "
	// TODO: check for a return value
	text += "{\n";
	// marshal
	for i := 0; i < argc; i++ {
		cname := "c_" + args[i].name
		ctype, marshal := GoToC(args[i].typ, args[i].name, cname)
		text += fmt.Sprintf("\tvar %s %s\n", cname, ctype)
		text += fmt.Sprintf("\t%s\n", marshal)
	}
	text += "}"
	return text
}
