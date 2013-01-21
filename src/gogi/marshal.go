package gogi

/*
#cgo pkg-config: glib-2.0
#include <glib.h>
#include <string.h>

GList *empty_glist = NULL;
GSList *empty_gslist = NULL;

// conversions
gint int_from_pointer(gpointer p) { return GPOINTER_TO_INT(p); }
gpointer pointer_from_int(gint i) { return GINT_TO_POINTER(i); }
guint uint_from_pointer(gpointer p) { return GPOINTER_TO_UINT(p); }
gpointer pointer_from_uint(guint i) { return GUINT_TO_POINTER(i); }
gulong size_from_pointer(gpointer p) { return GPOINTER_TO_SIZE(p); }
gpointer pointer_from_size(gulong i) { return GSIZE_TO_POINTER(i); }
char *from_gchar(gchar *str) { return (char*)str; }
gchar *to_gchar(char *str) { return (gchar*)str; }
GVariant *to_variant(gpointer ptr) { return (GVariant*)ptr; }

// general method for reading a variant into a gpointer
gpointer read_variant(GVariant *variant, const gchar *type) {
	if (!strcmp(type, "b")) {
		// boolean
		gboolean value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "y")) {
		// byte
		guchar value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "n")) {
		// int16
		gint16 value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "q")) {
		// uint16
		guint16 value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "i") || !strcmp(type, "h")) {
		// int32, handle
		gint32 value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "u")) {
		// uint32
		guint32 value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "x")) {
		// int64
		gint64 value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "t")) {
		// uint64
		guint64 value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else if (!strcmp(type, "d")) {
		// double
		gdouble value;
		g_variant_get(variant, type, &value);
		return GINT_TO_POINTER(value);
	} else {
		return NULL;
	}
}

gdouble read_double_variant(GVariant *variant) {
	gdouble value;
	g_variant_get(variant, "d", &value);
	return value;
}

gchar *read_string_variant(GVariant *variant, const gchar *type) {
	gchar *value;
	g_variant_get(variant, type, &value);
	return value;
}

GVariant *get_array() {
	return g_variant_new_strv(g_get_system_data_dirs(), -1);
}
*/
import "C"
import (
	"container/list"
	"reflect"
	//"unsafe"
)

/* -- Pointers -- */

// Try to read the pointer as a variant, which provides type information
func ToGo(ptr C.gpointer) (interface{}, reflect.Kind) {
	variant := C.to_variant(ptr)
	typ := C.g_variant_get_type(variant)
	typ_str := C.g_variant_type_dup_string(typ)
	defer C.g_free(C.gpointer(typ_str))

	if (GoBool(C.g_variant_type_is_basic(typ))) {
		// basic types
		switch GoString(typ_str) {
		case "b": // boolean
			value := C.gboolean(C.int_from_pointer(C.read_variant(variant, typ_str)))
			return GoBool(value), reflect.Bool
		case "y": // byte
			value := C.guint8(C.uint_from_pointer(C.read_variant(variant, typ_str)))
			return GoUInt8(value), reflect.Uint8
		case "n": // int16
			value := C.gint16(C.int_from_pointer(C.read_variant(variant, typ_str)))
			return GoInt16(value), reflect.Int16
		case "q": // uint16
			value := C.guint16(C.uint_from_pointer(C.read_variant(variant, typ_str)))
			return GoUInt16(value), reflect.Uint16
		case "i", "h": // int32, handle
			value := C.gint32(C.int_from_pointer(C.read_variant(variant, typ_str)))
			return GoInt32(value), reflect.Int32
		case "u": // uint32
			value := C.guint32(C.uint_from_pointer(C.read_variant(variant, typ_str)))
			return GoUInt32(value), reflect.Uint32
		case "x": // int64 (NOTE: this may not work correctly since gsize = unsigned long)
			value := C.gint64(C.size_from_pointer(C.read_variant(variant, typ_str)))
			return GoInt64(value), reflect.Int64
		case "t": // uint64
			value := C.guint64(C.size_from_pointer(C.read_variant(variant, typ_str)))
			return GoUInt64(value), reflect.Uint64
		case "d": // double
			value := C.read_double_variant(variant)
			return GoDouble(value), reflect.Float64
		case "s", "o", "g": // string, object path, or signature
			value := C.read_string_variant(variant, typ_str)
			return GoString(value), reflect.String
		}
	} else if (GoBool(C.g_variant_type_is_array(typ))) {
		// arrays
		// TODO: extract all values from the array
	}
	return nil, reflect.Invalid
}

// Convert an arbitrary Go value into its C representation
func ToGlib(data interface{}) C.gpointer {
	value := reflect.ValueOf(data)
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

func IntFromPointer(p C.gpointer) int {
	return GoInt(C.int_from_pointer(p))
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
		value, kind := ToGo(glist.data)
		if kind != reflect.Invalid {
			result.PushBack(value)
		}
		glist = glist.next
	}
	return result
}

func GSListToGo(gslist *C.GSList) *list.List {
	result := list.New()
	for gslist != C.empty_gslist {
		value, kind := ToGo(gslist.data)
		if kind != reflect.Invalid {
			result.PushBack(value)
		}
		gslist = gslist.next
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

func Test() {
}
