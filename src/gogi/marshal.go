package gogi

/*
#cgo pkg-config: glib-2.0
#include <glib.h>
#include <glib-object.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <girepository.h>

GList *empty_glist = NULL;
GSList *empty_gslist = NULL;

// conversions
gint32 int_from_pointer(gpointer p) { return GPOINTER_TO_INT(p); }
gpointer pointer_from_int(gint i) { return GINT_TO_POINTER(i); }
guint32 uint_from_pointer(gpointer p) { return GPOINTER_TO_UINT(p); }
gpointer pointer_from_uint(guint i) { return GUINT_TO_POINTER(i); }
gulong size_from_pointer(gpointer p) { return GPOINTER_TO_SIZE(p); }
gpointer pointer_from_size(gulong i) { return GSIZE_TO_POINTER(i); }
char *from_gchar(gchar *str) { return (char*)str; }
gchar *to_gchar(char *str) { return (gchar*)str; }

// array access
gint array_length(GArray *array) {
	printf("C array length is %d\n", array->len);
	return array->len;
}

GValue *array_get(GArray *array, gint i) {
	return &g_array_index(array, GValue, i);
}

// bitflag support
gboolean and(gint flags, gint position) {
	return flags & position;
}

// --- Test Methods --- //

static inline GValue *new_value(gpointer data, GType type) {
	GValue *value = (GValue*)malloc(sizeof(GValue));
	memset(value, 0, sizeof(GValue));
	g_value_init(value, type);
	switch (type) {
		case G_TYPE_STRING:
			g_value_set_string(value, data);
			break;
		default:
			g_value_set_pointer(value, data);
			break;
	}
	return value;
}

// return an array of strings stuffed in a gvalue
GValue *array_test() {
	GValue *item1 = new_value(g_strdup("hello"), G_TYPE_STRING);
	GValue *item2 = new_value(g_strdup("world"), G_TYPE_STRING);

	GArray *array = g_array_sized_new(FALSE, TRUE, sizeof(gchar*), 2);
	g_array_insert_val(array, 0, item1);
	g_array_insert_val(array, 1, item2);

	GValue *value = new_value(array, G_TYPE_POINTER);

	return value;
}
*/
import "C"
import (
	"container/list"
	"fmt"
	"reflect"
	"unsafe"
)

// returns the C type and the necessary marshaling code
// ???: is anything below this really necessary?
func GoToC(typeInfo *GiInfo, arg Argument, cvar string) (ctype string, marshal string) {
	govar := arg.name

	var ref string
	dir := arg.info.GetDirection()
	if dir == Out || dir == InOut {
		ref = "*"
	}

	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		// do array stuff
		switch typeInfo.GetArrayType() {
		case C.GI_ARRAY_TYPE_C:
			arg.name = govar + "_ar"
			ar_ctype, _ := GoToC(typeInfo.GetParamType(0), arg, cvar + "_ar")
			ctype = "*" + ar_ctype
			cvar_len := cvar + "_len"
			cvar_val := cvar + "_val"
			marshal = cvar_len + " := len(" + ref + govar + ")\n\t" +
			          cvar_val + " := make([]" + ar_ctype + ", " + cvar_len + ")\n\t" +
			          "for i := 0; i < " + cvar_len + "; i++ {\n\t" +
					  "\t" + cvar_val + "[i] = (*C.gchar)(C.CString((" + ref + govar + ")[i]))\n\t" +
					  "}\n\t" +
					  cvar + " = (" + ref + ar_ctype + ")(unsafe.Pointer(&" + cvar_val + "))"
		}
	} else {
		switch tag {
		case C.GI_TYPE_TAG_INT32:
			ctype = "C.gint32"
			marshal = fmt.Sprintf("%s = (%s)(%s)", cvar, ctype, ref + govar)
		case C.GI_TYPE_TAG_UTF8:
			ctype = "*C.gchar"
			marshal = "// TODO: marshal strings"
		}
	}
	return
}

func CToGo(typeInfo *GiInfo, govar string, cvar string) (gotype string, marshal string) {
	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		// TODO: implement
	} else {
		switch tag {
		case C.GI_TYPE_TAG_BOOLEAN:
			gotype = "bool"
			marshal = fmt.Sprintf("var %s %s\n", govar, gotype)
			marshal += fmt.Sprintf("\tif %s == 0 {", cvar) + "\n\t" +
					   fmt.Sprintf("\t%s = false", govar) + "\n\t" +
					   "} else {\n\t" +
					   fmt.Sprintf("\t%s = true", govar) + "\n\t" +
			           "}"
		}
	}
	return
}

func GoType(typeInfo *GiInfo, dir Direction) string {
	var result string

	// TODO: refactor this and the equivalent code in CType to its own method
	var ptr string
	if typeInfo.IsPointer() {
		ptr = "*"
	}
	var out string
	if dir == Out || dir == InOut {
		out = "*"
	}

	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		result += out + "[]" + GoType(typeInfo.GetParamType(0), In)
	} else {
		switch tag {
			// void?
		case C.GI_TYPE_TAG_VOID:
			return ""
		case C.GI_TYPE_TAG_BOOLEAN:
			return ptr + out + "bool"
		case C.GI_TYPE_TAG_INT32:
			result = ptr + out + "int32"
		case C.GI_TYPE_TAG_UTF8:
			result = out + "string"
		default:
			println("Unrecognized tag:", TypeTagToString(tag))
		}
	}
	return result
}

func CType(typeInfo *GiInfo, dir Direction) string {
	var result string

	var ptr string
	if dir == Out || dir == InOut {
		ptr += "*"
	}
	if typeInfo.IsPointer() {
		ptr += "*"
	}

	tag := typeInfo.GetTag()
	if tag == ArrayTag {
		result = CType(typeInfo.GetParamType(0), In) + ptr
	} else {
		switch tag {
		case C.GI_TYPE_TAG_VOID:
			result = "void"
		case C.GI_TYPE_TAG_BOOLEAN:
			result = "gboolean" + ptr
		case C.GI_TYPE_TAG_INT32:
			result = "gint32" + ptr
		case C.GI_TYPE_TAG_UTF8:
			result = "gchar" + ptr
		default:
			println("Unrecognized tag:", TypeTagToString(tag))
			result = ""
		}
	}
	return result
}

/* -- Pointers -- */

func ToGo(ptr *C.GValue) (interface{}, reflect.Kind) {
	typ := GoString(C.g_type_name(ptr.g_type))
	println("Type:", typ)
	value := C.g_value_peek_pointer(ptr)
	switch typ {
		case "gchar", "gint8":
			return GoInt8((C.gint8)(C.int_from_pointer(value))), reflect.Int8
		case "guchar", "guint8":
			return GoUInt8((C.guint8)(C.int_from_pointer(value))), reflect.Uint8
		case "gboolean":
			return GoBool((C.gboolean)(C.int_from_pointer(value))), reflect.Bool
		case "gint", "gint32":
			return GoInt32(C.int_from_pointer(value)), reflect.Int32
		case "guint", "guint32":
			return GoUInt32(C.uint_from_pointer(value)), reflect.Uint32
		case "glong", "gint64": // ???: better way to do this than using gsize?
			return GoLong((C.glong)(C.size_from_pointer(value))), reflect.Int64
		case "gulong", "guint64":
			return GoULong(C.size_from_pointer(value)), reflect.Uint64
		case "gshort", "gint16":
			return GoInt16((C.gint16)(C.int_from_pointer(value))), reflect.Int16
		case "gushort", "guint16":
			return GoUInt16((C.guint16)(C.int_from_pointer(value))), reflect.Uint16
		case "gpointer":
			return (unsafe.Pointer)(value), reflect.Ptr
		case "gchararray":
			return GoString((*C.gchar)(value)), reflect.String
	}
	return (unsafe.Pointer)(value), reflect.Invalid
}

// Convert an arbitrary Go value into its C representation
// TODO: create a method that takes a variable name and type tag
// and turns it into the appropriate code to marshal it
func ToGlib(data interface{}) C.gpointer {
	value := reflect.ValueOf(data)
	// TODO: fill this in
	switch value.Kind() {
		case reflect.Int:
			i := GlibInt(data.(int))
			return C.pointer_from_int(i)
		case reflect.String:
			gstr := GlibString(data.(string))
			return C.gpointer(gstr)
	}
	ptr := value.Pointer()
	return C.gpointer(ptr)
}

func ToSlice(array *C.GArray) []interface{} {
	length := GoInt(C.array_length(array))
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = C.array_get(array, (C.gint)(i))
	}
	return result
}

/* -- Booleans -- */

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

/* -- Chars -- */

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

/* -- Ints -- */

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

/* -- Shorts -- */

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

/* -- Longs -- */

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

/* -- Floats -- */

func GoFloat(f C.gfloat) float32 {
	return float32(f)
}

func GlibFloat(f float32) C.gfloat {
	return C.gfloat(f)
}

/* -- Doubles -- */

func GoDouble(d C.gdouble) float64 {
	return float64(d)
}

func GlibDouble(d float64) C.gdouble {
	return C.gdouble(d)
}

/* -- Strings -- */

func GoString(str *C.gchar) string {
	return C.GoString(C.from_gchar(str))
}

func GlibString(str string) *C.gchar {
	return C.to_gchar(C.CString(str))
}

// TODO: gsize, gssize, goffset, gintptr, guintptr

/* -- Lists -- */

func GListToGo(glist *C.GList) *list.List {
	result := list.New()
	for glist != C.empty_glist {
		result.PushBack(glist.data)
		glist = glist.next
		/*
		value, kind := ToGo((C.GValue)(glist.data))
		if kind != reflect.Invalid {
			result.PushBack(value)
		}
		glist = glist.next
		*/
	}
	return result
}

func GSListToGo(gslist *C.GSList) *list.List {
	result := list.New()
	for gslist != C.empty_gslist {
		/*
		value, kind := ToGo(gslist.data)
		if kind != reflect.Invalid {
			result.PushBack(value)
		}
		gslist = gslist.next
		*/
	}
	return result
}

func GoToGList(golist *list.List) *C.GList {
	glist := C.empty_glist
	for e := golist.Front(); e != nil; e = e.Next() {
		data := ToGlib(e.Value)
		glist = C.g_list_prepend(glist, data)
	}
	glist = C.g_list_reverse(glist)
	return glist
}

func GoToGSList(golist *list.List) *C.GSList {
	gslist := C.empty_gslist
	for e := golist.Front(); e != nil; e = e.Next() {
		data := ToGlib(e.Value)
		gslist = C.g_slist_prepend(gslist, data)
	}
	gslist = C.g_slist_reverse(gslist)
	return gslist
}

func PopulateFlags(data interface{}, bits C.gint, flags []C.gint) {
	value := reflect.ValueOf(data).Elem()
	for i := range flags {
		value.Field(i).SetBool(GoBool(C.and(bits, flags[i])))
	}
}

/* --- Test Methods --- */

func ArrayTest() {
	result, typ := ToGo(C.array_test())
	if typ != reflect.Ptr {
		println("Found non-pointer type in ArrayTest!")
		return
	}

	// ???: get gvalue to give us more information?
	// we know it's an array, so convert it to one
	dirs := ToSlice((*C.GArray)(result.(unsafe.Pointer)))
	if dirs == nil {
		println("Failed to convert to slice")
	}

	for _, dir := range dirs {
		value, ok := dir.(*C.GValue)
		if !ok {
			continue
		}

		govalue, _ := ToGo(value)
		_, ok = govalue.(string)
		if !ok {
			fmt.Println("Not a string")
		}
	}
	/*
	dirs, ok := result.([]interface{})
	if !ok {
		println("Type assertion to []interface{} failed.")
		return
	}

	fmt.Printf("%v\n", dirs)
	*/
}
