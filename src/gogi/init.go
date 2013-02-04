package gogi

/*
#cgo pkg-config: glib-2.0
#include <glib-object.h>
*/
import "C"

func Init() {
	C.g_type_init()
}
