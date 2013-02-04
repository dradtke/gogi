package gogi

/*
#include <glib.h>
#include <girepository.h>
*/
import "C"
import (
	"fmt"
)

type GiType C.GIInfoType
const (
	Function = C.GI_INFO_TYPE_FUNCTION
	Callback = C.GI_INFO_TYPE_CALLBACK
	Struct = C.GI_INFO_TYPE_STRUCT
	Boxed = C.GI_INFO_TYPE_BOXED
	Enum = C.GI_INFO_TYPE_ENUM
	Flags = C.GI_INFO_TYPE_FLAGS
	Object = C.GI_INFO_TYPE_OBJECT
	Interface = C.GI_INFO_TYPE_INTERFACE
	Constant = C.GI_INFO_TYPE_CONSTANT
	//ErrorDomain = C.GI_INFO_TYPE_ERRORDOMAIN
	Union = C.GI_INFO_TYPE_UNION
	Value = C.GI_INFO_TYPE_VALUE
	Signal = C.GI_INFO_TYPE_SIGNAL
	VFunc = C.GI_INFO_TYPE_VFUNC
	Property = C.GI_INFO_TYPE_PROPERTY
	Field = C.GI_INFO_TYPE_FIELD
	Arg = C.GI_INFO_TYPE_ARG
	Type = C.GI_INFO_TYPE_TYPE
	Unresolved = C.GI_INFO_TYPE_UNRESOLVED
)

type GiInfo struct {
	ptr *C.GIBaseInfo
	Type GiType
}

func NewGiInfo(ptr *C.GIBaseInfo) *GiInfo {
	return &GiInfo{ptr, (GiType)(C.g_base_info_get_type(ptr))}
}

/* -- Base Info -- */

func (info *GiInfo) GetName() string {
	return GoString(C.g_base_info_get_name(info.ptr))
}

func (info *GiInfo) GetNamespace() string {
	return GoString(C.g_base_info_get_namespace(info.ptr))
}

func (info *GiInfo) IsDeprecated() bool {
	return GoBool(C.g_base_info_is_deprecated(info.ptr))
}

func (info *GiInfo) GetAttribute(attr string) string {
	return GoString(C.g_base_info_get_attribute(info.ptr, GlibString(attr)))
}

/* -- Function Info -- */

type FunctionFlags struct {
	IsMethod bool
	IsConstructor bool
	IsGetter bool
	IsSetter bool
	WrapsVFunc bool
	Throws bool
}

func NewFunctionFlags(bits C.GIFunctionInfoFlags) *FunctionFlags {
	var flags FunctionFlags
	PopulateFlags(&flags, (C.gint)(bits), []C.gint{
		C.GI_FUNCTION_IS_METHOD,
		C.GI_FUNCTION_IS_CONSTRUCTOR,
		C.GI_FUNCTION_IS_GETTER,
		C.GI_FUNCTION_IS_SETTER,
		C.GI_FUNCTION_WRAPS_VFUNC,
		C.GI_FUNCTION_THROWS,
	})
	return &flags
}

func (info *GiInfo) GetFunctionSymbol() (string, error) {
	if info.Type != Function {
		return "", fmt.Errorf("gogi: expected function info, received %v", info.Type)
	}
	return GoString(C.g_function_info_get_symbol((*C.GIFunctionInfo)(info.ptr))), nil
}

func (info *GiInfo) GetFunctionFlags() (*FunctionFlags, error) {
	if info.Type != Function {
		return nil, fmt.Errorf("gogi: expected function info, received %v", info.Type)
	}
	return NewFunctionFlags(C.g_function_info_get_flags((*C.GIFunctionInfo)(info.ptr))), nil
}
