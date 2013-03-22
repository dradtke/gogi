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

// return a marshaled Go function and any necessary C wrapper
func WriteFunction(info *GiInfo, owner *GiInfo) (g string, c string) {
	flags := info.GetFunctionFlags()
	symbol := info.GetSymbol()
	argc := info.GetNArgs()

	g += "func "
	if owner != nil && !flags.IsConstructor {
		g += "(self *" + owner.GetName() + ") "
	}

	returnType := info.GetReturnType() ; defer returnType.Free()
	{
		ctype, cp := CType(returnType, In)
		c += ctype + " " + cp
	}

	g += CamelCase(info.GetName())
	c += "gogi_" + symbol + "("
	if owner != nil {
		if flags.IsConstructor {
			g += owner.GetName()
		} else {
			c += owner.GetObjectTypeName() + " *self"
			if argc > 0 {
				c += ", "
			}
		}
	}
	g += "("

	args := make([]Argument, argc)
	for i := 0; i < argc; i++ {
		arg := info.GetArg(i)
		dir := arg.GetDirection()
		args[i] = Argument{arg,arg.GetName(),"",arg.GetType()}
		gotype, gp := GoType(args[i].typ, dir)
		ctype, cp := CType(args[i].typ, dir)
		g += fmt.Sprintf("%s %s", noKeywords(args[i].name), gp + gotype)
		c += fmt.Sprintf("%s %s", ctype, cp + args[i].name)
		if i < argc-1 {
			g += ", "
			c += ", "
		}
	}
	g += ") "
	c += ") "

	hasReturnValue := (returnType.GetTag() != VoidTag)
	returnValueType, returnValueMarshal := MarshalToGo(returnType, "retval", "c_retval")
	if hasReturnValue {
		g += returnValueType + " "
	}

	g += "{\n"
	c += "{\n"
	// marshal
	for i := 0; i < argc; i++ {
		args[i].cname = "c_" + args[i].name
		ctype, marshal := MarshalToC(args[i].typ, args[i], args[i].cname)
		g += fmt.Sprintf("\tvar %s %s\n", args[i].cname, ctype)
		g += fmt.Sprintf("\t%s\n", marshal)
		g += fmt.Sprint("\n")
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

	g += "\t"
	if hasReturnValue {
		g += "c_retval, _ := "
	}
	g += "C.gogi_" + symbol + "("
	if owner != nil && !flags.IsConstructor {
		g += "self.ptr"
		if argc > 0 {
			g += ", "
		}
	}
	g +=  strings.Join(go_argnames, ", ") + ")\n"
	if hasReturnValue {
		if owner != nil && flags.IsConstructor {
			// wrap the return value in a Go struct
			implName := GetImplName(owner.GetName())
			// return &implName{(c_return_type)(retval)}
			g += fmt.Sprintf("\treturn &%s{(%s)(retval)}\n", implName, "*C." + owner.GetObjectTypeName());
		} else {
			g += "\t" + returnValueMarshal + "\n\treturn retval\n"
		}
	}

	// TODO: catch errno
	c += "\t"
	if hasReturnValue {
		c += "return "
	}
	c += info.GetSymbol() + "("
	if owner != nil && !flags.IsConstructor {
		c += "self"
		if argc > 0 {
			c += ", "
		}
	}
	c += strings.Join(c_argnames, ", ") + ");\n"

	g += "}\n"
	c += "}\n"

	return
}

func WriteObject(info *GiInfo) (g string, c string) {
	iter := info
	name := iter.GetName() ; typeName := iter.GetObjectTypeName()
	
	// interface
	g += fmt.Sprintf("type %s interface {\n", name)
	g += fmt.Sprintf("\tAs%s() *C.%s\n", name, typeName)
	g += "}\n"

	// implementation
	if !info.IsAbstract() {
		implName := GetImplName(name)
		g += fmt.Sprintf("type %s struct {\n", implName)
		g += fmt.Sprintf("\tptr *C.%s\n", typeName)
		g += "}\n"

		// ???: do this for abstract types?
		for {
			g += fmt.Sprintf("func (ob *%s) As%s() *C.%s {\n", implName, name, typeName)
			g += fmt.Sprintf("\treturn (*C.%s)(ob.ptr)\n", typeName)
			g += "}\n"
			// ???: better way to tell when to stop?
			if name == "Object" {
				break
			}
			iter = iter.GetParent() ; defer iter.Free()
			name = iter.GetName() ; typeName = iter.GetObjectTypeName()
		}
	}

	// do its methods
	method_count := info.GetNObjectMethods()
	for i := 0; i < method_count; i++ {
		method := info.GetObjectMethod(i)
		g_, c_ := WriteFunction(method, info)
		g += g_ + "\n"
		c += c_ + "\n"
	}

	g += "\n"
	if c != "" {
		c += "\n"
	}

	return
}

func WriteEnum(info *GiInfo) (g string, c string) {
	g += fmt.Sprintf("type %s C.%s\n", info.GetName(), info.GetRegisteredTypeName())
	g += "const (\n"

	value_count := info.GetNEnumValues()
	for i := 0; i < value_count; i++ {
		value := info.GetEnumValue(i) ; defer value.Free()
		// ???: how to avoid name clashes?
		g += fmt.Sprintf("\t%s = %d\n", CamelCase(value.GetName()), value.GetValue())
	}
	g += ")\n"

	return
}

// some argument names overlap with Go keywords; use this method to rename them
func noKeywords(name string) string {
	switch name {
		case "type": return "typ"
	}
	return name
}
