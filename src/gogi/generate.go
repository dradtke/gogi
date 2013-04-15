package gogi

import (
	"fmt"
	"strings"
)

type Argument struct {
	info *GiInfo
	typ *GiInfo
	name string
	cname string
	marshal string
}

// return a marshaled Go function and any necessary C wrapper
func WriteFunction(info *GiInfo, owner *GiInfo) (g string, c string) {
	symbol := info.GetSymbol()
	if blacklist[symbol] || cExports[symbol] {
		return
	}
	cExports[symbol] = true
	prefix := GetPrefix(info)

	flags := info.GetFunctionFlags()
	argc := info.GetNArgs()
	retc := 0

	for i := 0; i < argc; i++ {
		dir := info.GetArg(i).GetDirection()
		switch dir {
			case In: // default, do nothing
			case Out: argc-- ; retc++
			case InOut: return "", "" // quit early
		}
	}

	var ownerName string
	if owner != nil {
		ownerName = prefix + owner.GetName()
		castFunc(prefix, owner.GetName(), &c)
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
			g += ", " ; c += ", "
		} else if retc > 0 {
			c += ", "
		}
	}

	args := make([]Argument, 0)
	rets := make([]Argument, 0)
	argsAndRets := make([]Argument, 0)
	for i := 0; i < argc + retc; i++ {
		arg := info.GetArg(i)
		dir := arg.GetDirection()
		gotype, gp := GoType(arg.GetType())
		ctype, cp := CType(arg.GetType())
		if gotype == "" || ctype == "" || blacklist[gotype] {
			// argument failed to marshal
			g = "" ; c = ""
			return
		}

		if i > 0 {
			c += ", "
			if dir == In {
				g += ", "
			}
		}

		name := arg.GetName()
		newArg := Argument{arg,arg.GetType(),name,"c_"+name,""}
		argsAndRets = append(argsAndRets, newArg)
		if dir == In {
			args = append(args, newArg)
			g += fmt.Sprintf("%s %s", noKeywords(args[i].name), gp + gotype)
			c += fmt.Sprintf("%s %s", ctype, cp + args[i].name)
		} else if dir == Out {
			rets = append(rets, newArg)
			c += fmt.Sprintf("%s *%s", ctype, cp + arg.GetName())
		}
	}
	g += ") "
	c += ") "

	var returns bool
	if returnType.GetTag() != VoidTag || returnType.IsPointer() {
		retc++
		rets = append(rets, Argument{nil,returnType,"retval","c_retval",""})
		returns = true
	}

	var retLine string
	for i, ret := range rets {
		retType, retMarshal := MarshalToGo(ret.typ, ret.name, ret.cname)
		if retType == "" {
			g = "" ; c = "" ; return
		}
		if blacklist[strings.Trim(retType, "*")] {
			g = "" ; c = "" ; return
		}
		if i > 0 {
			retLine += ", "
		}
		retLine += retType
		rets[i].marshal = retMarshal
	}
	if retc > 1 {
		retLine = "(" + retLine + ")"
	}
	g += retLine + " "

	g += "{\n"
	c += "{\n"

	// marshal
	for i, arg := range args {
		ctype, marshal := MarshalToC(args[i].typ, arg, arg.cname)
		// TODO: remove the check for "C.", it shouldn't be needed
		if ctype == "" || ctype == "C." {
			g = ""; c = ""
			return
		}
		g += fmt.Sprintf("\tvar %s %s\n", arg.cname, ctype)
		g += fmt.Sprintf("\t%s\n", marshal)
	}

	/*
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
	*/

	g += "\t"
	for i, ret := range rets {
		if i > 0 {
			g += ", "
		}
		g += ret.cname
	}
	if retc > 0 {
		g += " := "
	}

	g += "C.gogi_" + symbol + "("
	if owner != nil && flags.IsMethod {
		switch owner.Type {
			case Object:
				g += "self.As" + owner.GetName() + "()"
			default:
				g += "self.ptr"
		}
	}

	for i, arg := range args {
		if i > 0 || (owner != nil && flags.IsMethod) {
			g += ", "
		}
		g += arg.cname
	}
	g += ")\n"

	for _, ret := range rets {
		g += "\t" + ret.marshal + "\n"
	}
	if retc > 0 {
		g += "\treturn "
		for i, ret := range rets {
			if i > 0 {
				g += ", "
			}
			g += ret.name
		}
		g += "\n"
	}

	// TODO: marshal return values back

	/*
	if hasReturnValue {
		if owner != nil && flags.IsConstructor {
			// wrap the return value in a Go struct
			structName := owner.GetName()
			if owner.Type == Object {
				structName = GetImplName(structName)
			}
			g += fmt.Sprintf("\treturn &%s{C.%s((C.gpointer)(c_retval))}\n", structName, cast)
		} else {
			g += "\t" + returnValueMarshal + "\n\treturn retval\n"
		}
	}
	*/

	// TODO: catch errno
	c += "\t"
	if returns {
		c += "return "
	}
	c += info.GetSymbol() + "("
	if owner != nil && flags.IsMethod {
		c += "self"
	}

	//c += strings.Join(c_argnames, ", ")
	for i, arg := range argsAndRets {
		if i > 0 || (owner != nil && flags.IsMethod) {
			c += ", "
		}
		c += arg.name
	}

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

	prefix := GetPrefix(info)

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

	prefix := GetPrefix(info)

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
		if !blacklist[prefix + name] {
			cast := castFunc(prefix, name, &c)
			g += fmt.Sprintf("func (ob %s) As%s() *C.%s {\n", implName, name, prefix + name)
			g += fmt.Sprintf("\treturn C.%s((C.gpointer)(ob.ptr))\n", cast)
			g += "}\n"
		}
		// ???: better way to tell when to stop?
		if name == "Object" || name == "ParamSpec" {
			break
		}
		// workaround for this sometimes being written out twice
		oldName := name
		for name == oldName {
			iter = iter.GetParent() ; defer iter.Free()
			name = iter.GetName()
		}
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
	prefix := GetPrefix(info)
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

// Some argument names overlap with Go keywords; use this method to rename them
func noKeywords(name string) string {
	switch name {
		case "type": return "typ"
		case "func": return "fun"
		case "len": return "length"
		case "string": return "str"
	}
	return name
}

// Gets the name for an enum value. Used to avoid naming conflicts
func enumValueName(enum, value string) string {
	return enum + value
}

// Gets the C function for casting to a specific type and writes it if it hasn't been yet
func castFunc(prefix, n string, c *string) string {
	name := "as_" + strings.ToLower(n)
	if !cExports[name] {
		cExports[name] = true
		(*c) += fmt.Sprintf("%s *%s(gpointer ob) {\n", prefix + n, name)
		(*c) += fmt.Sprintf("\treturn (%s*)ob;\n", prefix + n)
		(*c) += "}\n"
	}
	return name
}
