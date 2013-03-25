package gogi

import (
	"fmt"
	"strings"
)

type Argument struct {
	info *GiInfo
	typ *GiInfo
	dir Direction
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
	argAndRetc := argc
	retc := 0

	for i := 0; i < argAndRetc; i++ {
		dir := info.GetArg(i).GetDirection()
		switch dir {
			case In: // default, do nothing
			case Out: argc-- ; retc++
			case InOut: retc++
		}
	}

	var ownerName string
	if owner != nil {
		ownerName = owner.GetName()
		castFunc(prefix, ownerName, &c)
	}

	g += "func "

	returnType := info.GetReturnType() ; defer returnType.Free()
	{
		ctype, cp := CType(returnType)
		if ctype == "" {
			return "", ""
		} else if (ctype == "gchar" && cp != "" && returnType.GetTag() != ArrayTag) {
			// ???: add this for arrays or not?
			ctype = "const " + ctype
		}

		c += ctype + " " + cp
	}

	if owner != nil {
		g += ownerName
	}
	g += CamelCase(info.GetName()) + "("
	c += "gogi_" + symbol + "("

	cParamLine := make([]string, 0)
	gParamLine := make([]string, 0)

	if owner != nil && flags.IsMethod {
		cParamLine = append(cParamLine, prefix + ownerName + " *self")
		gArg := "self "
		if owner.Type == Struct {
			gArg += "*"
		}
		gArg += ownerName
		gParamLine = append(gParamLine, gArg)
	}

	args := make([]Argument, 0)
	rets := make([]Argument, 0)
	argsAndRets := make([]Argument, 0)
	for i := 0; i < argAndRetc; i++ {
		arg := info.GetArg(i)
		dir := arg.GetDirection()
		typ := arg.GetType()
		gotype, gp := GoType(typ)
		ctype, cp := CType(typ)
		if gotype == "" || ctype == "" || blacklist[gotype] {
			// argument failed to marshal
			return "", ""
		}

		name := arg.GetName()
		if symbol == "g_base64_decode_inplace" && name == "text" {
			fmt.Printf("%s (%s):\n", name, TypeTagToString(typ.GetTag()))
			fmt.Printf("direction: %d\n", dir)
			fmt.Printf("caller allocates: %t\n", arg.IsCallerAllocates())
			fmt.Printf("is return value: %t\n", arg.IsReturnValue())
			fmt.Printf("is optional: %t\n", arg.IsOptional())
			fmt.Printf("may be null: %t\n", arg.MayBeNull())
			fmt.Printf("ownership transfer: %d\n", arg.GetOwnershipTransfer())
			fmt.Printf("is pointer: %t\n", arg.GetType().IsPointer())
			fmt.Println()
		}
		newArg := Argument{arg,arg.GetType(),dir,name,"c_"+name,""}
		argsAndRets = append(argsAndRets, newArg)
		if dir == In {
			args = append(args, newArg)
			if ctype == "gchar" && cp != "" && typ.GetTag() != ArrayTag {
				ctype = "const " + ctype
			}
			gParamLine = append(gParamLine, fmt.Sprintf("%s %s", noKeywords(name), gp + gotype))
			cParamLine = append(cParamLine, fmt.Sprintf("%s %s", ctype, cp + name))
		} else if dir == Out {
			rets = append(rets, newArg)
			cp += "*"
			cParamLine = append(cParamLine, fmt.Sprintf("%s %s", ctype, cp + name))
		} else if dir == InOut {
			args = append(args, newArg)
			rets = append(rets, newArg)
			cp += "*"
			gParamLine = append(gParamLine, fmt.Sprintf("%s %s", noKeywords(name), gp + gotype))
			cParamLine = append(cParamLine, fmt.Sprintf("%s %s", ctype, cp + name))
		}
	}
	if flags.Throws {
		cParamLine = append(cParamLine, "GError **error")
	}
	g += strings.Join(gParamLine, ", ") + ") "
	c += strings.Join(cParamLine, ", ") + ") "

	var returns bool
	if returnType.GetTag() != VoidTag || returnType.IsPointer() {
		retc++
		rets = append(rets, Argument{nil,returnType,In,"retval","c_retval",""})
		returns = true
	}

	gParamLine = make([]string, 0)
	for i, ret := range rets {
		retType, retMarshal := MarshalToGo(ret)
		if retType == "" {
			return "", ""
		}
		if blacklist[strings.Trim(retType, "*")] {
			return "", ""
		}
		gParamLine = append(gParamLine, retType)
		rets[i].marshal = retMarshal
	}
	if flags.Throws {
		gParamLine = append(gParamLine, "error")
	}
	if len(gParamLine) > 0 {
		g += "(" + strings.Join(gParamLine, ", ") + ") "
	}

	g += "{\n"
	c += "{\n"

	// marshal
	for _, arg := range args {
		ctype, marshal := MarshalToC(arg)
		// TODO: remove the check for "C.", it shouldn't be needed
		if ctype == "" || ctype == "C." {
			return "", ""
		}
		g += fmt.Sprintf("\tvar %s %s\n", arg.cname, ctype)
		g += fmt.Sprintf("\t%s\n", marshal)
	}

	for i, ret := range rets {
		if i == len(rets)-1 && returns {
			break
		}
		if ret.dir == Out {
			ctype, cp := CType(ret.typ)
			/*
			if ret.info.IsCallerAllocates() && cp != "" {
				cp = cp[1:]
			}
			*/
			g += fmt.Sprintf("\tvar %s %sC.%s\n", ret.cname, cp, ctype)
		}
	}
	if flags.Throws {
		g += "\tvar c_error *C.GError\n"
	}
	g += "\t"
	if returns {
		// TODO: catch and use errno here
		g += "c_retval, _ := "
	}

	gParamLine = make([]string, 0)
	if owner != nil && flags.IsMethod {
		switch owner.Type {
			case Object:
				gParamLine = append(gParamLine, fmt.Sprintf("self.As%s()", ownerName))
			case Struct:
				gParamLine = append(gParamLine, "self.ptr")
		}
	}
	for _, arg := range argsAndRets {
		name := arg.cname
		if arg.dir == Out {
			name = "&" + name
		}
		gParamLine = append(gParamLine, name)
	}
	if flags.Throws {
		gParamLine = append(gParamLine, "&c_error")
	}
	g += fmt.Sprintf("C.gogi_%s(%s)\n", symbol, strings.Join(gParamLine, ", "))

	for _, ret := range rets {
		g += "\t" + ret.marshal + "\n"
	}
	if retc > 0 || flags.Throws {
		gParamLine = make([]string, 0)
		for _, ret := range rets {
			gParamLine = append(gParamLine, ret.name)
		}
		if flags.Throws {
			e := "GError{Code:(int)((*c_error).code), Message:C.GoString((*C.char)((*c_error).message))}"
			g += "\tif c_error != nil {\n"
			g += "\t\treturn " + strings.Join(append(gParamLine, e), ", ") + "\n"
			g += "\t}\n"
			g += "\treturn " + strings.Join(append(gParamLine, "nil"), ", ") + "\n"
		} else {
			g += "\treturn " + strings.Join(gParamLine, ", ") + "\n"
		}
	}

	c += "\t"
	if returns {
		c += "return "
	}
	c += info.GetSymbol()

	cParamLine = make([]string, 0)
	if owner != nil && flags.IsMethod {
		cParamLine = append(cParamLine, "self")
	}

	for _, arg := range argsAndRets {
		cParamLine = append(cParamLine, arg.name)
	}

	if flags.Throws {
		cParamLine = append(cParamLine, "error")
	}
	c += "(" + strings.Join(cParamLine, ", ") + ");\n"

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
