package gogi

/*
#cgo pkg-config: glib-2.0
#include <glib.h>
#include <string.h>

GList *empty_glist = NULL;
GSList *empty_gslist = NULL;

gint int_from_pointer(gpointer p) { return GPOINTER_TO_INT(p); }
gpointer pointer_from_int(gint i) { return GINT_TO_POINTER(i); }
char *from_gchar(gchar *str) { return (char*)str; }
gchar *to_gchar(char *str) { return (gchar*)str; }

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
	} else {
		return NULL;
	}
}
*/
import "C"
import (
	"container/list"
	"reflect"
	"unsafe"
)

/* -- Booleans -- */
func GoBool(b C.gboolean) bool {
	if b == C.gboolean(0) {
		return false
	}
	return true
}

func GBool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

/* -- Pointers -- */

// Since C has no idea what's contained in the pointer, we can't
// really infer anything from it.
func GoPointer(gptr C.gpointer) unsafe.Pointer {
	ptr := unsafe.Pointer(gptr)
	return ptr
}

// Try to read the pointer as a variant, which provides type information
// ???: return the type as well?
func GoPointerFromVariant(variant *C.GVariant) interface{} {
	typ := C.g_variant_get_type_string(variant)

	switch GoString(typ) {
		case "b": // boolean
			value := C.gboolean(C.int_from_pointer(C.read_variant(variant, typ)))
			return GoBool(value)
		case "y": // byte
			value := C.guchar(C.int_from_pointer(C.read_variant(variant, typ)))
			return GoUChar(value)
		default:
			return nil
	}
	return nil
}

// Use Go reflection to determine the correct pointer value
func GPointer(data interface{}) C.gpointer {
	value := reflect.ValueOf(data)
	switch value.Kind() {
	case reflect.Int:
		i := GInt(data.(int))
		return C.pointer_from_int(i)
	case reflect.String:
		gstr := GString(data.(string))
		return C.gpointer(gstr)
	default:
		ptr := value.Pointer()
		return C.gpointer(ptr)
	}
	return nil
}

func IntFromPointer(p C.gpointer) int {
	return GoInt(C.int_from_pointer(p))
}

/* -- Chars -- */

func GoChar(c C.gchar) int8 {
	return int8(c)
}

func GChar(i int8) C.gchar {
	return C.gchar(i)
}

func GoUChar(c C.guchar) uint {
	return uint(c)
}

func GuChar(i uint) C.guchar {
	return C.guchar(i)
}

/* -- Ints -- */

func GoInt(i C.gint) int {
	return int(i)
}

func GInt(i int) C.gint {
	return C.gint(i)
}

func GoUInt(i C.guint) uint {
	return uint(i)
}

func GuInt(i uint) C.guint {
	return C.guint(i)
}

/* -- Shorts -- */

func GoShort(s C.gshort) int16 {
	return int16(s)
}

func GShort(s int16) C.gshort {
	return C.gshort(s)
}

func GoUShort(s C.gushort) uint16 {
	return uint16(s)
}

func GuShort(s uint16) C.gushort {
	return C.gushort(s)
}

/* -- Longs -- */

func GoLong(l C.glong) int64 {
	return int64(l)
}

func GLong(l int64) C.glong {
	return C.glong(l)
}

func GoULong(l C.gulong) uint64 {
	return uint64(l)
}

func GuLong(l uint64) C.gulong {
	return C.gulong(l)
}

// TODO: gint8, gint16, etc.

/* -- Floats -- */

func GoFloat(f C.gfloat) float32 {
	return float32(f)
}

func GFloat(f float32) C.gfloat {
	return C.gfloat(f)
}

/* -- Doubles -- */

func GoDouble(d C.gdouble) float64 {
	return float64(d)
}

func GDouble(d float64) C.gdouble {
	return C.gdouble(d)
}

/* -- Strings -- */

func GoString(str *C.gchar) string {
	return C.GoString(C.from_gchar(str))
}

func GString(str string) *C.gchar {
	return C.to_gchar(C.CString(str))
}

// TODO: gsize, gssize, goffset, gintptr, guintptr

/* -- Lists -- */

// This method just fills the list with gpointers
func GoList(glist *C.GList) *list.List {
	result := list.New()
	for glist != C.empty_glist {
		result.PushBack(glist.data)
		glist = glist.next
	}
	return result
}

func GoSList(gslist *C.GSList) *list.List {
	result := list.New()
	for gslist != C.empty_gslist {
		result.PushBack(gslist.data)
		gslist = gslist.next
	}
	return result
}

func GList(golist *list.List) *C.GList {
	glist := C.empty_glist
	for e := golist.Front(); e != nil; e = e.Next() {
		data := GPointer(e.Value)
		glist = C.g_list_prepend(glist, data)
	}
	glist = C.g_list_reverse(glist)
	return glist
}

func GSList(golist *list.List) *C.GSList {
	gslist := C.empty_gslist
	for e := golist.Front(); e != nil; e = e.Next() {
		data := GPointer(e.Value)
		gslist = C.g_slist_prepend(gslist, data)
	}
	gslist = C.g_slist_reverse(gslist)
	return gslist
}
