package gogi

/*
#cgo pkg-config: glib-2.0
#include <glib.h>
#include <girepository.h>

GList *EMPTY_GLIST = NULL;

char *from_gchar(gchar *str) { return (char*)str; }
gchar *to_gchar(char *str) { return (gchar*)str; }

// bitflag support
gboolean and(gint flags, gint position) {
	return flags & position;
}
*/
import "C"
import (
	"container/list"
	"fmt"
	"reflect"
	"strings"
)

var goTypes = map[int]string {
	(int)(C.GI_TYPE_TAG_VOID):     "",
	(int)(C.GI_TYPE_TAG_BOOLEAN):  "bool",
	(int)(C.GI_TYPE_TAG_INT8):     "int8",
	(int)(C.GI_TYPE_TAG_INT16):    "int16",
	(int)(C.GI_TYPE_TAG_INT32):    "int32",
	(int)(C.GI_TYPE_TAG_INT64):    "int64",
	(int)(C.GI_TYPE_TAG_UINT8):    "uint8",
	(int)(C.GI_TYPE_TAG_UINT16):   "uint16",
	(int)(C.GI_TYPE_TAG_UINT32):   "uint32",
	(int)(C.GI_TYPE_TAG_UINT64):   "uint64",
	(int)(C.GI_TYPE_TAG_FLOAT):    "float32",
	(int)(C.GI_TYPE_TAG_DOUBLE):   "float64",
	(int)(C.GI_TYPE_TAG_GTYPE):    "int",
	(int)(C.GI_TYPE_TAG_UTF8):     "string",
	(int)(C.GI_TYPE_TAG_FILENAME): "string",
	// skip a couple
	(int)(C.GI_TYPE_TAG_GLIST):    "list.List",
	(int)(C.GI_TYPE_TAG_GSLIST):   "list.List",
	// skip a couple
	//(int)(C.GI_TYPE_TAG_UNICHAR):  "rune",
}

var cTypes = map[int]string {
	(int)(C.GI_TYPE_TAG_VOID):     "void",
	(int)(C.GI_TYPE_TAG_BOOLEAN):  "gboolean",
	(int)(C.GI_TYPE_TAG_INT8):     "gint8",
	(int)(C.GI_TYPE_TAG_INT16):    "gint16",
	(int)(C.GI_TYPE_TAG_INT32):    "gint32",
	(int)(C.GI_TYPE_TAG_INT64):    "gint64",
	(int)(C.GI_TYPE_TAG_UINT8):    "guint8",
	(int)(C.GI_TYPE_TAG_UINT16):   "guint16",
	(int)(C.GI_TYPE_TAG_UINT32):   "guint32",
	(int)(C.GI_TYPE_TAG_UINT64):   "guint64",
	(int)(C.GI_TYPE_TAG_FLOAT):    "gfloat",
	(int)(C.GI_TYPE_TAG_DOUBLE):   "gdouble",
	(int)(C.GI_TYPE_TAG_GTYPE):    "GType",
	(int)(C.GI_TYPE_TAG_UTF8):     "gchar",
	(int)(C.GI_TYPE_TAG_FILENAME): "gchar",
	// skip a couple
	(int)(C.GI_TYPE_TAG_GLIST):    "GList",
	(int)(C.GI_TYPE_TAG_GSLIST):   "GSList",
	// skip a couple
	//(int)(C.GI_TYPE_TAG_UNICHAR):  "gunichar",
}

// returns the C type and the necessary marshaling code
func MarshalToC(arg Argument) (ctype string, marshal string) {
	typeInfo := arg.typ
	cvar := arg.cname
	govar := noKeywords(arg.name)
	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		switch typeInfo.GetArrayType() {
			case C.GI_ARRAY_TYPE_C:
				arg.name = govar + "_ar"
				ar_ctype, _ := MarshalToC(Argument{typ:typeInfo.GetParamType(0), cname:cvar + "_ar", name:arg.name, dir:arg.dir})
				ctype = "*" + ar_ctype
				cvar_len := cvar + "_len"
				cvar_val := cvar + "_val"
				marshal = cvar_len + " := len(" + govar + ")\n\t" +
				          cvar_val + " := make([]" + ar_ctype + ", " + cvar_len + ")\n\t" +
				          "for i := 0; i < " + cvar_len + "; i++ {\n\t" +
						  "\t" + cvar_val + "[i] = (*C.gchar)(C.CString((" + govar + ")[i]))\n\t" +
						  "}\n\t" +
						  cvar + " = (" + ar_ctype + ")(unsafe.Pointer(&" + cvar_val + "))"
		}
	} else {
		var p string
		ctype, p = CType(typeInfo)
		if ctype == "" {
			return "", ""
		}
		ctype = p + "C." + ctype
		switch tag {
			case C.GI_TYPE_TAG_VOID:
				if ctype == "C.gpointer" {
					marshal = fmt.Sprintf("%s = (C.gpointer)(reflect.ValueOf(%s).Pointer())", cvar, govar)
				}
			case C.GI_TYPE_TAG_BOOLEAN:
				marshal = "if " + govar + " {\n\t" +
					  "\t" + cvar + " = 1\n\t" +
					  "} else {\n\t" +
					  "\t" + cvar + " = 0\n\t" +
					  "}"
			case C.GI_TYPE_TAG_INT8,
			     C.GI_TYPE_TAG_INT16,
			     C.GI_TYPE_TAG_INT32,
			     C.GI_TYPE_TAG_INT64,
			     C.GI_TYPE_TAG_UINT8,
			     C.GI_TYPE_TAG_UINT16,
			     C.GI_TYPE_TAG_UINT32,
			     C.GI_TYPE_TAG_UINT64,
			     C.GI_TYPE_TAG_FLOAT,
			     C.GI_TYPE_TAG_DOUBLE,
			     C.GI_TYPE_TAG_GTYPE,
			     C.GI_TYPE_TAG_UNICHAR:
				marshal = fmt.Sprintf("%s = (%s)(%s)", cvar, ctype, govar)
			case C.GI_TYPE_TAG_UTF8, C.GI_TYPE_TAG_FILENAME:
				marshal = fmt.Sprintf("%s = (%s)(C.CString(%s))", cvar, ctype, govar)
			case C.GI_TYPE_TAG_INTERFACE:
				interfaceInfo := typeInfo.GetTypeInterface()
				switch interfaceInfo.Type {
					case Enum, Flags:
						ctype = "C." + GetPrefix(interfaceInfo) + interfaceInfo.GetName()
						marshal = fmt.Sprintf("%s = (%s)(%s)", cvar, ctype, govar)
					case Object:
						marshal = fmt.Sprintf("%s = (%s).As%s()", cvar, govar, interfaceInfo.GetName())
					case Struct:
						marshal = fmt.Sprintf("%s = (%s).ptr", cvar, govar)
				}
			case C.GI_TYPE_TAG_GLIST:
				ctype = "C.GList"
				marshal = "// TODO: marshal glist"
			case C.GI_TYPE_TAG_GSLIST:
				ctype = "C.GSList"
				marshal = "// TODO: marshal gslist"
			default:
				ctype = "<CAN'T MARSHAL TO C: " + TypeTagToString(tag) + ">"
				//ctype = "gint"
		}
	}
	return
}

func MarshalToGo(arg Argument) (gotype string, marshal string) {
	typeInfo := arg.typ
	govar := arg.name
	cvar := arg.cname
	tag := typeInfo.GetTag()
	eq := ":="
	if arg.dir == InOut {
		eq = "="
	}
	if tag == ArrayTag {
		var ptr string
		arrayType := typeInfo.GetParamType(0)
		gotype, ptr = GoType(arrayType)
		gotype = "[]" + ptr + gotype
		marshal = "// TODO: marshal"
		switch typeInfo.GetArrayType() {
			case C.GI_ARRAY_TYPE_C:
			default:
				// TODO: implement other array types
				return "", ""
		}
	} else {
		var ptr string
		if typeInfo.IsPointer() {
			ptr = "*"
		}
		gotype = goTypes[(int)(tag)]
		if gotype != "" {
			gotype = ptr + gotype
		}
		switch tag {
			case C.GI_TYPE_TAG_VOID:
				if ptr != "" {
					marshal = fmt.Sprintf("%s = reflect.ValueOf(%s).Interface()", govar, cvar)
				}
			case C.GI_TYPE_TAG_BOOLEAN:
				marshal = fmt.Sprintf("%s %s %s != 0", govar, eq, cvar)
			case C.GI_TYPE_TAG_INT8,
			     C.GI_TYPE_TAG_INT16,
			     C.GI_TYPE_TAG_INT32,
			     C.GI_TYPE_TAG_INT64,
			     C.GI_TYPE_TAG_UINT8,
			     C.GI_TYPE_TAG_UINT16,
			     C.GI_TYPE_TAG_UINT32,
			     C.GI_TYPE_TAG_UINT64,
			     C.GI_TYPE_TAG_FLOAT,
			     C.GI_TYPE_TAG_DOUBLE,
			     C.GI_TYPE_TAG_GTYPE,
			     C.GI_TYPE_TAG_UNICHAR:
				marshal = fmt.Sprintf("%s %s (%s)(%s)", govar, eq, gotype, cvar)
			case C.GI_TYPE_TAG_UTF8, C.GI_TYPE_TAG_FILENAME:
				gotype = "string"
				marshal = fmt.Sprintf("%s %s C.GoString((*C.char)(%s))", govar, eq, cvar)
			case C.GI_TYPE_TAG_INTERFACE:
				interfaceInfo := typeInfo.GetTypeInterface()
				name := interfaceInfo.GetName()
				switch interfaceInfo.Type {
					case Object:
						//gotype = ptr + name
						gotype = name
						marshal = fmt.Sprintf("%s %s &%s{C.as_%s((C.gpointer)(%s))}", govar, eq, GetImplName(name), strings.ToLower(gotype), cvar)
					case Struct:
						//gotype = ptr + name
						gotype = "*" + name
						var addr string
						if ptr == "" {
							addr = "&"
						}
						marshal = fmt.Sprintf("%s %s &%s{%s}", govar, eq, name, addr + cvar)
					default:
						marshal = fmt.Sprintf("// TODO: marshal %d", interfaceInfo.Type)
				}
			case C.GI_TYPE_TAG_GLIST, C.GI_TYPE_TAG_GSLIST:
				gotype = "*list.List"
				marshal = fmt.Sprintf("%s %s list.New()\n", govar, eq) +
				          fmt.Sprintf("\tfor %s != nil {\n", cvar) +
					  fmt.Sprintf("\t\t%s.PushBack(%s.data)\n", govar, cvar) +
					  fmt.Sprintf("\t\t%s = %s.next\n", cvar, cvar) +
					  fmt.Sprintf("\t}\n")
			default:
				gotype = "<CAN'T MARSHAL TO GO: " + TypeTagToString(tag) + ">"
		}
	}
	return
}

func GoType(typeInfo *GiInfo) (string, string) {
	var ptr string
	if typeInfo.IsPointer() {
		ptr = "*"
	}
	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		gotype, p := GoType(typeInfo.GetParamType(0))
		return gotype, "[]" + p
		//return (refOut(dir) + "[]" + GoType(typeInfo.GetParamType(0), In))
	} else {
		val, ok := goTypes[(int)(tag)]
		if ok {
			if val == "" && ptr != "" {
				return "interface{}", ptr[1:]
			} else if val == "string" && ptr != "" {
				return val, ptr[1:]
			} else {
				return val, ptr
			}
		}

		// check non-primitive tags
		// TODO: find callbacks
		switch tag {
			case C.GI_TYPE_TAG_INTERFACE:
				interfaceType := typeInfo.GetTypeInterface()
				// for now, ignore types not in this namespace
				if interfaceType.GetNamespace() != cNamespace {
					return "", ""
				}

				if interfaceType.Type == Callback {
					// TODO: enable callbacks
					return "", ""
				} else if interfaceType.Type == Object {
					// objects are interfaces, so don't include pointers
					return interfaceType.GetName(), ""
				} else if interfaceType.Type == Struct {
					// always pass structs around as pointers
					return interfaceType.GetName(), "*"
				} else {
					return interfaceType.GetName(), ptr
				}
		}
	}

	//println("go unrecognized:", TypeTagToString(tag))
	return "", ptr
}

func CType(typeInfo *GiInfo) (string, string) {
	var ptr string
	if typeInfo.IsPointer() {
		ptr = "*"
	}
	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		// TODO: re-enable useful array functions
		ctype, p := CType(typeInfo.GetParamType(0))
		return ctype, "*" + p
	} else {
		val, ok := cTypes[(int)(tag)]
		if ok {
			if tag == C.GI_TYPE_TAG_VOID && ptr != "" {
				return "gpointer", ptr[1:]
			} else {
				return val, ptr
			}
		}

		switch tag {
			case C.GI_TYPE_TAG_INTERFACE:
				interfaceType := typeInfo.GetTypeInterface()

				if interfaceType.Type == Callback {
					// TODO: enable callbacks
					return "", ""
				} else {
					return GetPrefix(interfaceType) + interfaceType.GetName(), ptr
				}

				// TODO: print this out to stderr
				//fmt.Printf("unrecognized interface type: %d [%s]\n", interfaceType.Type, interfaceType.GetName())
		}
	}

	//println(" c unrecognized:", TypeTagToString(tag))
	return "", ptr
}


func GoBool(b C.gboolean) bool {
	if b == C.gboolean(0) {
		return false
	}
	return true
}

func GlibBool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func GoChar(c C.gchar) int8 {
	return int8(c)
}

func GlibChar(i int8) C.gchar {
	return C.gchar(i)
}

func GoUChar(c C.guchar) uint {
	return uint(c)
}

func GlibUChar(i uint) C.guchar {
	return C.guchar(i)
}

func GoInt(i C.gint) int {
	return int(i)
}

func GlibInt(i int) C.gint {
	return C.gint(i)
}

func GoUInt(i C.guint) uint {
	return uint(i)
}

func GlibUInt(i uint) C.guint {
	return C.guint(i)
}

func GoInt8(i C.gint8) int8 {
	return int8(i)
}

func GlibInt8(i int8) C.gint8 {
	return C.gint8(i)
}

func GoUInt8(i C.guint8) uint8 {
	return uint8(i)
}

func GlibUInt8(i uint8) C.guint8 {
	return C.guint8(i)
}

func GoInt16(i C.gint16) int16 {
	return int16(i)
}

func GlibInt16(i int16) C.gint16 {
	return C.gint16(i)
}

func GoUInt16(i C.guint16) uint16 {
	return uint16(i)
}

func GlibUInt16(i uint16) C.guint16 {
	return C.guint16(i)
}

func GoInt32(i C.gint32) int32 {
	return int32(i)
}

func GlibInt32(i int32) C.gint32 {
	return C.gint32(i)
}

func GoUInt32(i C.guint32) uint32 {
	return uint32(i)
}

func GlibUInt32(i uint32) C.guint32 {
	return C.guint32(i)
}

func GoInt64(i C.gint64) int64 {
	return int64(i)
}

func GlibInt64(i int64) C.gint64 {
	return C.gint64(i)
}

func GoUInt64(i C.guint64) uint64 {
	return uint64(i)
}

func GlibUInt64(i uint64) C.guint64 {
	return C.guint64(i)
}

func GoShort(s C.gshort) int16 {
	return int16(s)
}

func GlibShort(s int16) C.gshort {
	return C.gshort(s)
}

func GoUShort(s C.gushort) uint16 {
	return uint16(s)
}

func GlibUShort(s uint16) C.gushort {
	return C.gushort(s)
}

func GoLong(l C.glong) int64 {
	return int64(l)
}

func GlibLong(l int64) C.glong {
	return C.glong(l)
}

func GoULong(l C.gulong) uint64 {
	return uint64(l)
}

func GlibULong(l uint64) C.gulong {
	return C.gulong(l)
}

// TODO: gint8, gint16, etc.

func GoFloat(f C.gfloat) float32 {
	return float32(f)
}

func GlibFloat(f float32) C.gfloat {
	return C.gfloat(f)
}

func GoDouble(d C.gdouble) float64 {
	return float64(d)
}

func GlibDouble(d float64) C.gdouble {
	return C.gdouble(d)
}

func GoString(str *C.gchar) string {
	return C.GoString(C.from_gchar(str))
}

func GlibString(str string) *C.gchar {
	return C.to_gchar(C.CString(str))
}

func GListToGo(glist *C.GList) *list.List {
	result := list.New()
	for glist != nil {
		result.PushBack(glist.data)
		glist = glist.next
	}
	return result
}

func PopulateFlags(data interface{}, bits C.gint, flags []C.gint) {
	value := reflect.ValueOf(data).Elem()
	for i := range flags {
		value.Field(i).SetBool(GoBool(C.and(bits, flags[i])))
	}
}

func refOut(dir Direction) string {
	if dir == Out || dir == InOut {
		return "*"
	}
	return ""
}

func refPointer(typeInfo *GiInfo, dir Direction) string {
	var ptr string
	//ptr := refOut(dir)
	if typeInfo.IsPointer() {
		ptr += "*"
	}
	return ptr
}
