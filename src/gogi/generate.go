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

var functionBlacklist []string = []string {
	"g_ascii_strtod",
	"g_atomic_pointer_compare_and_exchange",
	"g_atomic_pointer_set",
	"g_filename_from_uri",
	"g_get_charset",
	"g_get_filename_charsets",
	"g_once_init_enter",
	"g_strfreev",
	"g_strjoinv",
	"g_strtod",
	"g_strv_get_type",
	"g_strv_length",
	"g_unix_error_quark",
	"g_variant_get_gtype",
	"g_variant_parse",
	"g_variant_type_string_scan",
	"g_variant_type_checked_",
}

// return a marshaled Go function and any necessary C wrapper
func WriteFunction(info *GiInfo, owner *GiInfo) (g string, c string) {
	symbol := info.GetSymbol()
	if contains(symbol, functionBlacklist) {
		return
	}

	flags := info.GetFunctionFlags()
	argc := info.GetNArgs()

	var ownerName string
	if owner != nil {
		ownerName = cPrefix + owner.GetName()
	}

	g += "func "
	/*
	if owner != nil && flags.IsMethod {
		g += "(self *" + owner.GetName() + ") "
	}
	*/

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

	g += CamelCase(info.GetName()) + "("
	c += "gogi_" + symbol + "("
	if owner != nil {
		if flags.IsConstructor {
			g += owner.GetName()
		} else if flags.IsMethod {
			c += ownerName + " *self"
			g += "self *" + owner.GetName()
			if argc > 0 {
				c += ", "
				g += ", "
			}
		}
	}

	args := make([]Argument, argc)
	for i := 0; i < argc; i++ {
		arg := info.GetArg(i)
		dir := arg.GetDirection()
		args[i] = Argument{arg,arg.GetName(),"",arg.GetType()}
		gotype, gp := GoType(args[i].typ)
		ctype, cp := CType(args[i].typ)
		if dir == Out || dir == InOut {
			cp += "*"
			gp += "*"
		}
		if gotype == "" || ctype == "" {
			g = ""; c = ""
			return
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
	if owner != nil && flags.IsMethod {
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
			g += fmt.Sprintf("\treturn &%s{(%s)(retval)}\n", implName, "*C." + ownerName)
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
		// TODO: catch the error, don't just pass in null
		if argc > 0 {
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
	name := info.GetName()

	g += fmt.Sprintf("type %s struct {\n", name)
	g += fmt.Sprintf("\tptr *C.%s\n", cPrefix + name)
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
	
	// interface
	g += fmt.Sprintf("type %s interface {\n", name)
	g += fmt.Sprintf("\tAs%s() *C.%s\n", name, cPrefix + name)
	g += "}\n"

	// implementation
	if !info.IsAbstract() {
		implName := GetImplName(name)
		g += fmt.Sprintf("type %s struct {\n", implName)
		g += fmt.Sprintf("\tptr *C.%s\n", cPrefix + name)
		g += "}\n"

		// ???: do this for abstract types?
		for {
			g += fmt.Sprintf("func (ob *%s) As%s() *C.%s {\n", implName, name, cPrefix + name)
			g += fmt.Sprintf("\treturn (*C.%s)(ob.ptr)\n", cPrefix + name)
			g += "}\n"
			// ???: better way to tell when to stop?
			if name == "Object" || name == "ParamSpec" {
				break
			}
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
	symbol := info.GetRegisteredTypeName()
	if symbol == "" {
		// ???: why the hell does this happen with GLib?
		symbol = "G" + name
	}
	g += fmt.Sprintf("type %s C.%s\n", name, symbol)
	g += "const (\n"

	value_count := info.GetNEnumValues()
	for i := 0; i < value_count; i++ {
		value := info.GetEnumValue(i) ; defer value.Free()
		// ???: how to avoid name clashes?
		g += fmt.Sprintf("\t%s%s = %d\n", name, CamelCase(value.GetName()), value.GetValue())
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

func contains(elem string, blacklist []string) bool {
	for _, x := range blacklist {
		if x == elem {
			return true
		}
	}
	return false
}
