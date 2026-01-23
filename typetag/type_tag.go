package typetag

import (
	"github.com/open-move/sui-go-sdk/types"
)

type StructTag struct {
	Address    types.Address
	Module     string
	Name       string
	TypeParams []TypeTag
}

type TypeTag struct {
	Bool    *struct{}
	U8      *struct{}
	U64     *struct{}
	U128    *struct{}
	Address *struct{}
	Signer  *struct{}
	Vector  *TypeTag
	Struct  *StructTag
	U16     *struct{}
	U32     *struct{}
	U256    *struct{}
}

func (TypeTag) IsBcsEnum() {}

func TypeTagBool() TypeTag {
	return TypeTag{Bool: &struct{}{}}
}

func TypeTagU8() TypeTag {
	return TypeTag{U8: &struct{}{}}
}

func TypeTagU16() TypeTag {
	return TypeTag{U16: &struct{}{}}
}

func TypeTagU32() TypeTag {
	return TypeTag{U32: &struct{}{}}
}

func TypeTagU64() TypeTag {
	return TypeTag{U64: &struct{}{}}
}

func TypeTagU128() TypeTag {
	return TypeTag{U128: &struct{}{}}
}

func TypeTagU256() TypeTag {
	return TypeTag{U256: &struct{}{}}
}

func TypeTagAddress() TypeTag {
	return TypeTag{Address: &struct{}{}}
}

func TypeTagSigner() TypeTag {
	return TypeTag{Signer: &struct{}{}}
}

func TypeTagVector(inner TypeTag) TypeTag {
	return TypeTag{Vector: &inner}
}

func TypeTagStruct(tag StructTag) TypeTag {
	return TypeTag{Struct: &tag}
}

func NewStructTag(address types.Address, module, name string, typeParams []TypeTag) StructTag {
	return StructTag{
		Address:    address,
		Module:     module,
		Name:       name,
		TypeParams: append([]TypeTag(nil), typeParams...),
	}
}
