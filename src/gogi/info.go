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

type TypeTag C.GITypeTag
const (
	Void = C.GI_TYPE_TAG_VOID
	Boolean = C.GI_TYPE_TAG_BOOLEAN
	Int8 = C.GI_TYPE_TAG_INT8
	Uint8 = C.GI_TYPE_TAG_UINT8
	Int16 = C.GI_TYPE_TAG_INT16
	Uint16 = C.GI_TYPE_TAG_UINT16
	Int32 = C.GI_TYPE_TAG_INT32
	Uint32 = C.GI_TYPE_TAG_UINT32
	Int64 = C.GI_TYPE_TAG_INT64
	Uint64 = C.GI_TYPE_TAG_UINT64
	Float = C.GI_TYPE_TAG_FLOAT
	Double = C.GI_TYPE_TAG_DOUBLE
	GType = C.GI_TYPE_TAG_GTYPE
	Utf8 = C.GI_TYPE_TAG_UTF8
	Filename = C.GI_TYPE_TAG_FILENAME
	// non-basic types
	GArray = C.GI_TYPE_TAG_ARRAY
	GInterface = C.GI_TYPE_TAG_INTERFACE
	GList = C.GI_TYPE_TAG_GLIST
	GSList = C.GI_TYPE_TAG_GSLIST
	GHash = C.GI_TYPE_TAG_GHASH
	GError = C.GI_TYPE_TAG_ERROR
	// another basic type
	Unichar = C.GI_TYPE_TAG_UNICHAR
)

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

func (info *GiInfo) GetSymbol() (string, error) {
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

/* -- VFunc Info -- */

type VFuncFlags struct {
	MustChainUp bool
	MustOverride bool
	MustNotOverride bool
	Throws bool
}

func NewVFuncFlags(bits C.GIVFuncInfoFlags) *VFuncFlags {
	var flags VFuncFlags
	PopulateFlags(&flags, (C.gint)(bits), []C.gint{
		C.GI_VFUNC_MUST_CHAIN_UP,
		C.GI_VFUNC_MUST_OVERRIDE,
		C.GI_VFUNC_MUST_NOT_OVERRIDE,
		C.GI_VFUNC_THROWS,
	})
	return &flags
}

func (info *GiInfo) GetVFuncFlags() (*VFuncFlags, error) {
	if info.Type != VFunc {
		return nil, fmt.Errorf("gogi: expected vfunc info, received %v", info.Type)
	}
	return NewVFuncFlags(C.g_vfunc_info_get_flags((*C.GIVFuncInfo)(info.ptr))), nil
}

func (info *GiInfo) GetOffset() (int, error) {
	if info.Type != VFunc {
		return 0, fmt.Errorf("gogi: expected vfunc info, received %v", info.Type)
	}
	// TODO: check for a value of 0xFFFF, which means it's unknown
	return GoInt(C.g_vfunc_info_get_offset((*C.GIVFuncInfo)(info.ptr))), nil
}

func (info *GiInfo) GetVFuncSignal() (*GiInfo, error) {
	if info.Type != VFunc {
		return nil, fmt.Errorf("gogi: expected vfunc info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_vfunc_info_get_signal((*C.GIVFuncInfo)(info.ptr)))), nil
}

func (info *GiInfo) GetInvoker() (*GiInfo, error) {
	if info.Type != VFunc {
		return nil, fmt.Errorf("gogi: expected vfunc info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_vfunc_info_get_invoker((*C.GIVFuncInfo)(info.ptr)))), nil
}

/* -- RegisteredType Info -- */

func (info *GiInfo) IsRegisteredType() bool {
	switch info.Type {
	case Enum, Interface, Object, Struct, Union:
		return true
	}
	return false
}

func (info *GiInfo) GetRegisteredTypeName() (string, error) {
	if !info.IsRegisteredType() {
		return "", fmt.Errorf("gogi: expected registered type info, received %v", info.Type)
	}
	return GoString(C.g_registered_type_info_get_type_name((*C.GIRegisteredTypeInfo)(info.ptr))), nil
}

func (info *GiInfo) GetRegisteredTypeInit() (string, error) {
	if !info.IsRegisteredType() {
		return "", fmt.Errorf("gogi: expected registered type info, received %v", info.Type)
	}
	return GoString(C.g_registered_type_info_get_type_init((*C.GIRegisteredTypeInfo)(info.ptr))), nil
}

// TODO: get gtype?

/* -- Enum Info -- */

func (info *GiInfo) GetNValues() (int, error) {
	if info.Type != Enum {
		return 0, fmt.Errorf("gogi: expected enum info, received %v", info.Type)
	}
	return GoInt(C.g_enum_info_get_n_values((*C.GIEnumInfo)(info.ptr))), nil
}

func (info *GiInfo) GetValue(n int) (*GiInfo, error) {
	if info.Type != Enum {
		return nil, fmt.Errorf("gogi: expected enum info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_enum_info_get_value((*C.GIEnumInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNEnumMethods() (int, error) {
	if info.Type != Enum {
		return 0, fmt.Errorf("gogi: expected enum info, received %v", info.Type)
	}
	return GoInt(C.g_enum_info_get_n_methods((*C.GIEnumInfo)(info.ptr))), nil
}

func (info *GiInfo) GetEnumMethod(n int) (*GiInfo, error) {
	if info.Type != Enum {
		return nil, fmt.Errorf("gogi: expected enum info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_enum_info_get_method((*C.GIEnumInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetStorageType() (TypeTag, error) {
	if info.Type != Enum {
		return 0, fmt.Errorf("gogi: expected enum info, received %v", info.Type)
	}
	return (TypeTag)(C.g_enum_info_get_storage_type((*C.GIEnumInfo)(info.ptr))), nil
}

/* -- Object Info -- */

func (info *GiInfo) GetObjectTypeName() (string, error) {
	if info.Type != Object {
		return "", fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoString(C.g_object_info_get_type_name((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetObjectTypeInit() (string, error) {
	if info.Type != Object {
		return "", fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoString(C.g_object_info_get_type_init((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) IsAbstract() (bool, error) {
	if info.Type != Object {
		return false, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoBool(C.g_object_info_get_abstract((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) IsFundamental() (bool, error) {
	if info.Type != Object {
		return false, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoBool(C.g_object_info_get_fundamental((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetParent() (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_parent((*C.GIObjectInfo)(info.ptr)))), nil
}

func (info *GiInfo) GetNInterfaces() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_interfaces((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetInterface(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_interface((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNFields() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_fields((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetField(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_field((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNObjectProperties() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_properties((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetObjectProperty(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_property((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNObjectMethods() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_methods((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetObjectMethod(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_method((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNSignals() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_signals((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetObjectSignal(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_signal((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNVFuncs() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_vfuncs((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetVFunc(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_vfunc((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

func (info *GiInfo) GetNConstants() (int, error) {
	if info.Type != Object {
		return 0, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return GoInt(C.g_object_info_get_n_constants((*C.GIObjectInfo)(info.ptr))), nil
}

func (info *GiInfo) GetConstant(n int) (*GiInfo, error) {
	if info.Type != Object {
		return nil, fmt.Errorf("gogi: expected object info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_constant((*C.GIObjectInfo)(info.ptr), GlibInt(n)))), nil
}

/* -- Arg Info -- */

type Direction C.GIDirection
const (
	In = C.GI_DIRECTION_IN
	Out = C.GI_DIRECTION_OUT
	InOut = C.GI_DIRECTION_INOUT
)

type ScopeType C.GIScopeType
const (
	Invalid = C.GI_SCOPE_TYPE_INVALID
	Call = C.GI_SCOPE_TYPE_CALL
	Async = C.GI_SCOPE_TYPE_ASYNC
	Notified = C.GI_SCOPE_TYPE_NOTIFIED
)

func (info *GiInfo) GetDirection() (Direction, error) {
	if info.Type != Arg {
		return 0, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return (Direction)(C.g_arg_info_get_direction((*C.GIArgInfo)(info.ptr))), nil
}

func (info *GiInfo) IsCallerAllocates() (bool, error) {
	if info.Type != Arg {
		return false, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return GoBool(C.g_arg_info_is_caller_allocates((*C.GIArgInfo)(info.ptr))), nil
}

func (info *GiInfo) IsReturnValue() (bool, error) {
	if info.Type != Arg {
		return false, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return GoBool(C.g_arg_info_is_return_value((*C.GIArgInfo)(info.ptr))), nil
}

func (info *GiInfo) IsOptional() (bool, error) {
	if info.Type != Arg {
		return false, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return GoBool(C.g_arg_info_is_optional((*C.GIArgInfo)(info.ptr))), nil
}

func (info *GiInfo) MayBeNull() (bool, error) {
	if info.Type != Arg {
		return false, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return GoBool(C.g_arg_info_may_be_null((*C.GIArgInfo)(info.ptr))), nil
}

func (info *GiInfo) GetOwnershipTransfer() (Transfer, error) {
	if info.Type != Arg {
		return 0, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return (Transfer)(C.g_arg_info_get_ownership_transfer((*C.GIArgInfo)(info.ptr))), nil
}

func (info *GiInfo) GetScope() (ScopeType, error) {
	if info.Type != Arg {
		return 0, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return (ScopeType)(C.g_arg_info_get_scope((*C.GIArgInfo)(info.ptr))), nil
}

// TODO: get closure/destroy?

func (info *GiInfo) GetType() (*GiInfo, error) {
	if info.Type != Arg {
		return nil, fmt.Errorf("gogi: expected arg info, received %v", info.Type)
	}
	return NewGiInfo((*C.GIBaseInfo)(C.g_arg_info_get_type((*C.GIArgInfo)(info.ptr)))), nil
}
