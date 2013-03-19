package gogi

/*
#include <glib.h>
#include <glib-object.h>
#include <girepository.h>
*/
import "C"
import (
	//"fmt"
	"strings"
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

func InfoTypeToString(typ GiType) string {
	return GoString(C.g_info_type_to_string((C.GIInfoType)(typ)))
}

type GiInfo struct {
	ptr *C.GIBaseInfo
	Type GiType
}

func NewGiInfo(ptr *C.GIBaseInfo) *GiInfo {
	typ := (GiType)(C.g_base_info_get_type(ptr))
	return &GiInfo{ptr, typ}
}

func (info *GiInfo) Free() {
	C.g_base_info_unref(info.ptr)
}

type TypeTag C.GITypeTag
const (
	VoidTag = C.GI_TYPE_TAG_VOID
	BooleanTag = C.GI_TYPE_TAG_BOOLEAN
	Int8Tag = C.GI_TYPE_TAG_INT8
	Uint8Tag = C.GI_TYPE_TAG_UINT8
	Int16Tag = C.GI_TYPE_TAG_INT16
	Uint16Tag = C.GI_TYPE_TAG_UINT16
	Int32Tag = C.GI_TYPE_TAG_INT32
	Uint32Tag = C.GI_TYPE_TAG_UINT32
	Int64Tag = C.GI_TYPE_TAG_INT64
	Uint64Tag = C.GI_TYPE_TAG_UINT64
	FloatTag = C.GI_TYPE_TAG_FLOAT
	DoubleTag = C.GI_TYPE_TAG_DOUBLE
	GTypeTag = C.GI_TYPE_TAG_GTYPE
	Utf8Tag = C.GI_TYPE_TAG_UTF8
	FilenameTag = C.GI_TYPE_TAG_FILENAME
	// non-basic types
	ArrayTag = C.GI_TYPE_TAG_ARRAY
	InterfaceTag = C.GI_TYPE_TAG_INTERFACE
	GListTag = C.GI_TYPE_TAG_GLIST
	GSListTag = C.GI_TYPE_TAG_GSLIST
	GHashTag = C.GI_TYPE_TAG_GHASH
	ErrorTag = C.GI_TYPE_TAG_ERROR
	// another basic type
	UnicharTag = C.GI_TYPE_TAG_UNICHAR
)

/* -- Base Info -- */

func (info *GiInfo) GetName() string {
	return GoString(C.g_base_info_get_name(info.ptr))
}

func (info *GiInfo) GetFullName() string {
	return strings.ToLower(GoString(C.g_base_info_get_namespace(info.ptr))) + "_" + info.GetName()
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

func (info *GiInfo) GetReturnType() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_callable_info_get_return_type((*C.GICallableInfo)(info.ptr))))
}

func (info *GiInfo) GetCallerOwns() Transfer {
	return (Transfer)(C.g_callable_info_get_caller_owns((*C.GICallableInfo)(info.ptr)))
}

func (info *GiInfo) MayReturnNull() bool {
	return GoBool(C.g_callable_info_may_return_null((*C.GICallableInfo)(info.ptr)))
}

func (info *GiInfo) GetReturnAttribute(name string) string {
	_name := GlibString(name) ; defer C.g_free((C.gpointer)(_name))
	return GoString(C.g_callable_info_get_return_attribute((*C.GICallableInfo)(info.ptr), _name))
}

// iterate return attributes?

func (info *GiInfo) GetNArgs() int {
	return GoInt(C.g_callable_info_get_n_args((*C.GICallableInfo)(info.ptr)))
}

func (info *GiInfo) GetArg(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_callable_info_get_arg((*C.GICallableInfo)(info.ptr), GlibInt(n))))
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

func (info *GiInfo) GetSymbol() string {
	return GoString(C.g_function_info_get_symbol((*C.GIFunctionInfo)(info.ptr)))
}

func (info *GiInfo) GetFunctionFlags() *FunctionFlags {
	return NewFunctionFlags(C.g_function_info_get_flags((*C.GIFunctionInfo)(info.ptr)))
}

func (info *GiInfo) GetFunctionProperty() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_function_info_get_property((*C.GIFunctionInfo)(info.ptr))))
}

func (info *GiInfo) GetFunctionVFunc() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_function_info_get_vfunc((*C.GIFunctionInfo)(info.ptr))))
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

func (info *GiInfo) GetSignalFlags() *SignalFlags {
	return NewSignalFlags(C.g_signal_info_get_flags((*C.GISignalInfo)(info.ptr)))
}

func (info *GiInfo) GetClassClosure() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_signal_info_get_class_closure((*C.GISignalInfo)(info.ptr))))
}

func (info *GiInfo) TrueStopsEmit() bool {
	return GoBool(C.g_signal_info_true_stops_emit((*C.GISignalInfo)(info.ptr)))
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

func (info *GiInfo) GetVFuncFlags() *VFuncFlags {
	return NewVFuncFlags(C.g_vfunc_info_get_flags((*C.GIVFuncInfo)(info.ptr)))
}

func (info *GiInfo) GetOffset() int {
	// TODO: check for a value of 0xFFFF, which means it's unknown
	return GoInt(C.g_vfunc_info_get_offset((*C.GIVFuncInfo)(info.ptr)))
}

func (info *GiInfo) GetVFuncSignal() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_vfunc_info_get_signal((*C.GIVFuncInfo)(info.ptr))))
}

func (info *GiInfo) GetInvoker() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_vfunc_info_get_invoker((*C.GIVFuncInfo)(info.ptr))))
}

/* -- RegisteredType Info -- */

func (info *GiInfo) IsRegisteredType() bool {
	switch info.Type {
	case Enum, Interface, Object, Struct, Union:
		return true
	}
	return false
}

func (info *GiInfo) GetRegisteredTypeName() string {
	return GoString(C.g_registered_type_info_get_type_name((*C.GIRegisteredTypeInfo)(info.ptr)))
}

func (info *GiInfo) GetRegisteredTypeInit() string {
	return GoString(C.g_registered_type_info_get_type_init((*C.GIRegisteredTypeInfo)(info.ptr)))
}

// TODO: get gtype?

/* -- Enum Info -- */

func (info *GiInfo) GetNEnumValues() int {
	return GoInt(C.g_enum_info_get_n_values((*C.GIEnumInfo)(info.ptr)))
}

func (info *GiInfo) GetEnumValue(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_enum_info_get_value((*C.GIEnumInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNEnumMethods() int {
	return GoInt(C.g_enum_info_get_n_methods((*C.GIEnumInfo)(info.ptr)))
}

func (info *GiInfo) GetEnumMethod(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_enum_info_get_method((*C.GIEnumInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetStorageType() TypeTag {
	return (TypeTag)(C.g_enum_info_get_storage_type((*C.GIEnumInfo)(info.ptr)))
}

// this acts on GIValueInfo
func (info *GiInfo) GetValue() int64 {
	return (int64)(C.g_value_info_get_value((*C.GIValueInfo)(info.ptr)))
}

/* -- Object Info -- */

func (info *GiInfo) GetObjectTypeName() string {
	return GoString(C.g_object_info_get_type_name((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetObjectTypeInit() string {
	return GoString(C.g_object_info_get_type_init((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) IsAbstract() bool {
	return GoBool(C.g_object_info_get_abstract((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) IsFundamental() bool {
	return GoBool(C.g_object_info_get_fundamental((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetParent() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_parent((*C.GIObjectInfo)(info.ptr))))
}

func (info *GiInfo) GetNObjectInterfaces() int {
	return GoInt(C.g_object_info_get_n_interfaces((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetObjectInterface(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_interface((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNFields() int {
	return GoInt(C.g_object_info_get_n_fields((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetField(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_field((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNObjectProperties() int {
	return GoInt(C.g_object_info_get_n_properties((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetObjectProperty(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_property((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNObjectMethods() int {
	return GoInt(C.g_object_info_get_n_methods((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetObjectMethod(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_method((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNSignals() int {
	return GoInt(C.g_object_info_get_n_signals((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetObjectSignal(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_signal((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNVFuncs() int {
	return GoInt(C.g_object_info_get_n_vfuncs((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetVFunc(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_vfunc((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetNConstants() int {
	return GoInt(C.g_object_info_get_n_constants((*C.GIObjectInfo)(info.ptr)))
}

func (info *GiInfo) GetConstant(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_object_info_get_constant((*C.GIObjectInfo)(info.ptr), GlibInt(n))))
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

func (info *GiInfo) GetDirection() Direction {
	return (Direction)(C.g_arg_info_get_direction((*C.GIArgInfo)(info.ptr)))
}

func (info *GiInfo) IsCallerAllocates() bool {
	return GoBool(C.g_arg_info_is_caller_allocates((*C.GIArgInfo)(info.ptr)))
}

func (info *GiInfo) IsReturnValue() bool {
	return GoBool(C.g_arg_info_is_return_value((*C.GIArgInfo)(info.ptr)))
}

func (info *GiInfo) IsOptional() bool {
	return GoBool(C.g_arg_info_is_optional((*C.GIArgInfo)(info.ptr)))
}

func (info *GiInfo) MayBeNull() bool {
	return GoBool(C.g_arg_info_may_be_null((*C.GIArgInfo)(info.ptr)))
}

func (info *GiInfo) GetOwnershipTransfer() Transfer {
	return (Transfer)(C.g_arg_info_get_ownership_transfer((*C.GIArgInfo)(info.ptr)))
}

func (info *GiInfo) GetScope() ScopeType {
	return (ScopeType)(C.g_arg_info_get_scope((*C.GIArgInfo)(info.ptr)))
}

// TODO: get closure/destroy?

func (info *GiInfo) GetType() *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_arg_info_get_type((*C.GIArgInfo)(info.ptr))))
}

/* -- Type Info -- */

type ArrayType C.GIArrayType
const (
	CArray = C.GI_ARRAY_TYPE_C
	GArray = C.GI_ARRAY_TYPE_ARRAY
	PtrArray = C.GI_ARRAY_TYPE_PTR_ARRAY
	ByteArray = C.GI_ARRAY_TYPE_BYTE_ARRAY
)

func TypeTagToString(tag TypeTag) string {
	return GoString(C.g_type_tag_to_string((C.GITypeTag)(tag)))
}

func (info *GiInfo) IsPointer() bool {
	return GoBool(C.g_type_info_is_pointer((*C.GITypeInfo)(info.ptr)))
}

func (info *GiInfo) GetTag() TypeTag {
	return (TypeTag)(C.g_type_info_get_tag((*C.GITypeInfo)(info.ptr)))
}

func (info *GiInfo) GetParamType(n int) *GiInfo {
	return NewGiInfo((*C.GIBaseInfo)(C.g_type_info_get_param_type((*C.GITypeInfo)(info.ptr), GlibInt(n))))
}

func (info *GiInfo) GetTypeInterface() *GiInfo {
	return NewGiInfo(C.g_type_info_get_interface((*C.GITypeInfo)(info.ptr)))
}

func (info *GiInfo) GetArrayLength() int {
	return GoInt(C.g_type_info_get_array_length((*C.GITypeInfo)(info.ptr)))
}

func (info *GiInfo) GetArrayFixedSize() int {
	return GoInt(C.g_type_info_get_array_fixed_size((*C.GITypeInfo)(info.ptr)))
}

func (info *GiInfo) IsZeroTerminated() bool {
	return GoBool(C.g_type_info_is_zero_terminated((*C.GITypeInfo)(info.ptr)))
}

func (info *GiInfo) GetArrayType() ArrayType {
	return (ArrayType)(C.g_type_info_get_array_type((*C.GITypeInfo)(info.ptr)))
}
