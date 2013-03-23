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
	symbol := info.GetSymbol()
	if blacklist[symbol] || cExports[symbol] {
		return
	}
	cExports[symbol] = true
	prefix := getPrefix(info)

	flags := info.GetFunctionFlags()
	argc := info.GetNArgs()

	var ownerName string
	if owner != nil {
		ownerName = prefix + owner.GetName()
	}

	g += "func "

	returnType := info.GetReturnType() ; defer returnType.Free()
	{
		ctype, cp := CType(returnType)
		if ctype == "" {
			g = ""; c = ""
			return
		} else if (ctype == "gchar" && cp != "") {
			ctype = "const " + ctype
		}

		c += ctype + " " + cp
	}

	if owner != nil {
		g += owner.GetName()
	}
	g += CamelCase(info.GetName()) + "("
	c += "gogi_" + symbol + "("
	if owner != nil && flags.IsMethod {
		c += ownerName + " *self"
		g += "self " + owner.GetName()
		if argc > 0 {
			g += ", "
			c += ", "
		}
	}

	args := make([]Argument, argc)
	for i := 0; i < argc; i++ {
		arg := info.GetArg(i)
		args[i] = Argument{arg,arg.GetName(),"",arg.GetType()}
		gotype, gp := GoType(args[i].typ)
		ctype, cp := CType(args[i].typ)
		if gotype == "" || ctype == "" || blacklist[gotype] {
			// argument failed to marshal
			g = ""; c = ""
			return
		}
		dir := arg.GetDirection()
		if dir == Out || dir == InOut {
			cp += "*"
			gp += "*"
		}
		g += fmt.Sprintf("%s %s", noKeywords(args[i].name), gp + gotype)
		c += fmt.Sprintf("%s %s", ctype, cp + args[i].name)
		if i < argc-1 {
			g += ", "
			c += ", "
		}
	}
	g += ") "
	c += ") "

	hasReturnValue := (returnType.GetTag() != VoidTag || returnType.IsPointer())
	var returnValueType, returnValueMarshal string
	if hasReturnValue {
		returnValueType, returnValueMarshal = MarshalToGo(returnType, "retval", "c_retval")
		if returnValueType == "" {
			g = ""; c = ""
			return
		}
		plainReturnType := strings.Trim(returnValueType, "*")
		if blacklist[plainReturnType] {
			g = ""; c = ""
			return
		}
		g += returnValueType + " "
	}

	g += "{\n"
	c += "{\n"
	// marshal
	for i := 0; i < argc; i++ {
		args[i].cname = "c_" + args[i].name
		ctype, marshal := MarshalToC(args[i].typ, args[i], args[i].cname)
		// TODO: remove the check for "C.", it shouldn't be needed
		if ctype == "" || ctype == "C." {
			g = ""; c = ""
			return
		}
		g += fmt.Sprintf("\tvar %s %s\n", args[i].cname, ctype)
		g += fmt.Sprintf("\t%s\n", marshal)
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
	if owner != nil && flags.IsMethod {
		switch owner.Type {
			case Object:
				g += "self.As" + owner.GetName() + "()"
			default:
				g += "self.ptr"
		}
		if argc > 0 {
			g += ", "
		}
	}
	g +=  strings.Join(go_argnames, ", ") + ")\n"
	if hasReturnValue {
		if owner != nil && flags.IsConstructor {
			// wrap the return value in a Go struct
			structName := owner.GetName()
			if owner.Type == Object {
				structName = GetImplName(structName)
			}
			g += fmt.Sprintf("\treturn &%s{(%s)(c_retval)}\n", structName, "*C." + ownerName)
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
	if owner != nil && flags.IsMethod {
		c += "self"
		if argc > 0 {
			c += ", "
		}
	}
	c += strings.Join(c_argnames, ", ")
	if flags.Throws {
		if argc > 0 || flags.IsMethod {
			c += ", "
		}
		c += "NULL"
	}
	c += ");\n"

	g += "}\n"
	c += "}\n"

	return
}

func WriteStruct(info *GiInfo) (g string, c string) {
	// for now, skip gtype and foreign structs
	if info.IsGTypeStruct() || info.IsForeign() {
		return
	}

	name := info.GetName()

	if blacklist[name] {
		return
	}

	prefix := getPrefix(info)

	g += fmt.Sprintf("type %s struct {\n", name)
	g += fmt.Sprintf("\tptr *C.%s\n", prefix + name)
	g += "}\n"

	// do its methods
	method_count := info.GetNStructMethods()
	for i := 0; i < method_count; i++ {
		method := info.GetStructMethod(i)
		if method.IsDeprecated() {
			continue
		}
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

func WriteObject(info *GiInfo) (g string, c string) {
	iter := info
	name := iter.GetName()

	if blacklist[name] {
		return
	}

	prefix := getPrefix(info)
	
	// interface
	g += fmt.Sprintf("type %s interface {\n", name)
	g += fmt.Sprintf("\tAs%s() *C.%s\n", name, prefix + name)
	g += "}\n"

	// implementation
	// ???: does it matter if it's abstract?
	implName := GetImplName(name)
	g += fmt.Sprintf("type %s struct {\n", implName)
	g += fmt.Sprintf("\tptr *C.%s\n", prefix + name)
	g += "}\n"

	// ???: do this for abstract types?
	for {
		castFunc := "as_" + strings.ToLower(name)
		if !cExports[castFunc] {
			cExports[castFunc] = true
			c += fmt.Sprintf("%s *%s(gpointer ob) {\n", prefix + name, castFunc)
			c += fmt.Sprintf("\treturn (%s*)ob;\n", prefix + name)
			c += "}\n"
		}
		g += fmt.Sprintf("func (ob %s) As%s() *C.%s {\n", implName, name, prefix + name)
		g += fmt.Sprintf("\treturn C.%s((C.gpointer)(ob.ptr))\n", castFunc)
		g += "}\n"
		// ???: better way to tell when to stop?
		if name == "Object" || name == "ParamSpec" {
			break
		}
		iter = iter.GetParent() ; defer iter.Free()
		name = iter.GetName()
	}

	// do its methods
	method_count := info.GetNObjectMethods()
	for i := 0; i < method_count; i++ {
		method := info.GetObjectMethod(i)
		if method.IsDeprecated() {
			continue
		}
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
	name := info.GetName()
	prefix := getPrefix(info)
	symbol := prefix + info.GetName()
	g += fmt.Sprintf("type %s C.%s\n", name, symbol)
	g += "const (\n"

	value_count := info.GetNEnumValues()
	for i := 0; i < value_count; i++ {
		value := info.GetEnumValue(i) ; defer value.Free()
		// ???: how to avoid name clashes?
		g += fmt.Sprintf("\t%s = %d\n", enumValueName(name, CamelCase(value.GetName())), value.GetValue())
	}
	g += ")\n"

	return
}

// some argument names overlap with Go keywords; use this method to rename them
func noKeywords(name string) string {
	switch name {
		case "type": return "typ"
		case "func": return "fun"
	}
	return name
}

func enumValueName(enum, value string) string {
	return enum + value
}
