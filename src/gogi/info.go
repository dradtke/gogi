package gogi

/*
#include <glib.h>
#include <glib-object.h>
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

func (info *GiInfo) Free() {
	C.g_base_info_unref(info.ptr)
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
	_attr := GlibString(attr) ; defer C.g_free((C.gpointer)(_attr))
	return GoString(C.g_base_info_get_attribute(info.ptr, _attr))
}

/* -- Callables -- */

type Transfer C.GITransfer
const (
	Nothing = C.GI_TRANSFER_NOTHING
	Container = C.GI_TRANSFER_CONTAINER
	Everything = C.GI_TRANSFER_EVERYTHING
)

func (info *GiInfo) IsCallable() bool {
	switch info.Type {
	case Function, Signal, VFunc:
		return true
	}
	return false
}

func (info *GiInfo) GetReturnType() (*GiInfo, error) {
	if !info.IsCallable() {
		return nil, fmt.Errorf("gogi: expected callable info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_callable_info_get_return_type((*C.GICallableInfo)(info.ptr)))), nil
}

func (info *GiInfo) GetCallerOwns() (Transfer, error) {
	if !info.IsCallable() {
		return (Transfer)(C.G_MAXINT), fmt.Errorf("gogi: expected callable info, received %v", info.Type)
	}
	return (Transfer)(C.g_callable_info_get_caller_owns((*C.GICallableInfo)(info.ptr))), nil
}

func (info *GiInfo) MayReturnNull() (bool, error) {
	if !info.IsCallable() {
		return false, fmt.Errorf("gogi: expected callable info, received %v", info.Type)
	}
	return GoBool(C.g_callable_info_may_return_null((*C.GICallableInfo)(info.ptr))), nil
}

func (info *GiInfo) GetReturnAttribute(name string) (string, error) {
	if !info.IsCallable() {
		return "", fmt.Errorf("gogi: expected callable info, received %v", info.Type)
	}
	_name := GlibString(name) ; defer C.g_free((C.gpointer)(_name))
	return GoString(C.g_callable_info_get_return_attribute((*C.GICallableInfo)(info.ptr), _name)), nil
}

// iterate return attributes?

func (info *GiInfo) GetNArgs() (int, error) {
	if !info.IsCallable() {
		return 0, fmt.Errorf("gogi: expected callable info, received %v", info.Type)
	}
	return GoInt(C.g_callable_info_get_n_args((*C.GICallableInfo)(info.ptr))), nil
}

func (info *GiInfo) GetArg(n int) (*GiInfo, error) {
	if !info.IsCallable() {
		return nil, fmt.Errorf("gogi: expected callable info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_callable_info_get_arg((*C.GICallableInfo)(info.ptr), GlibInt(n)))), nil
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

func (info *GiInfo) GetFunctionProperty() (*GiInfo, error) {
	if info.Type != Function {
		return nil, fmt.Errorf("gogi: expected function info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_function_info_get_property((*C.GIFunctionInfo)(info.ptr)))), nil
}

func (info *GiInfo) GetFunctionVFunc() (*GiInfo, error) {
	if info.Type != Function {
		return nil, fmt.Errorf("gogi: expected function info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_function_info_get_vfunc((*C.GIFunctionInfo)(info.ptr)))), nil
}

// invoke?

/* -- Signal Info -- */

type SignalFlags struct {
	RunFirst bool
	RunLast bool
	RunCleanup bool
	NoRecurse bool
	Detailed bool
	Action bool
	NoHooks bool
	MustCollect bool
	Deprecated bool
}

func NewSignalFlags(bits C.GSignalFlags) *SignalFlags {
	var flags SignalFlags
	PopulateFlags(&flags, (C.gint)(bits), []C.gint{
		C.G_SIGNAL_RUN_FIRST,
		C.G_SIGNAL_RUN_LAST,
		C.G_SIGNAL_RUN_CLEANUP,
		C.G_SIGNAL_NO_RECURSE,
		C.G_SIGNAL_DETAILED,
		C.G_SIGNAL_ACTION,
		C.G_SIGNAL_NO_HOOKS,
		C.G_SIGNAL_MUST_COLLECT,
		C.G_SIGNAL_DEPRECATED,
	})
	return &flags
}

func (info *GiInfo) GetSignalFlags() (*SignalFlags, error) {
	if info.Type != Signal {
		return nil, fmt.Errorf("gogi: expected signal info, received %v", info.Type)
	}
	return NewSignalFlags(C.g_signal_info_get_flags((*C.GISignalInfo)(info.ptr))), nil
}

func (info *GiInfo) GetClassClosure() (*GiInfo, error) {
	if info.Type != Signal {
		return nil, fmt.Errorf("gogi: expected signal info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_signal_info_get_class_closure((*C.GISignalInfo)(info.ptr)))), nil
}

func (info *GiInfo) TrueStopsEmit() (bool, error) {
	if info.Type != Signal {
		return false, fmt.Errorf("gogi: expected signal info, received %v", info.Type)
	}
	return GoBool(C.g_signal_info_true_stops_emit((*C.GISignalInfo)(info.ptr))), nil
}
