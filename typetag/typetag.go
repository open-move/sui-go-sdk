// Package typetag provides types and utilities for representing Move type tags.
package typetag

import (
	"fmt"
	"strings"

	"github.com/open-move/sui-go-sdk/types"
)

// StructTag represents a Move struct type tag.
type StructTag struct {
	Address    types.Address
	Module     string
	Name       string
	TypeParams []TypeTag
}

// TypeTag represents a Move type tag.
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

// TypeTagBool creates a TypeTag for the bool type.
func TypeTagBool() TypeTag {
	return TypeTag{Bool: &struct{}{}}
}

// TypeTagU8 creates a TypeTag for the u8 type.
func TypeTagU8() TypeTag {
	return TypeTag{U8: &struct{}{}}
}

// TypeTagU16 creates a TypeTag for the u16 type.
func TypeTagU16() TypeTag {
	return TypeTag{U16: &struct{}{}}
}

// TypeTagU32 creates a TypeTag for the u32 type.
func TypeTagU32() TypeTag {
	return TypeTag{U32: &struct{}{}}
}

// TypeTagU64 creates a TypeTag for the u64 type.
func TypeTagU64() TypeTag {
	return TypeTag{U64: &struct{}{}}
}

// TypeTagU128 creates a TypeTag for the u128 type.
func TypeTagU128() TypeTag {
	return TypeTag{U128: &struct{}{}}
}

// TypeTagU256 creates a TypeTag for the u256 type.
func TypeTagU256() TypeTag {
	return TypeTag{U256: &struct{}{}}
}

// TypeTagAddress creates a TypeTag for the address type.
func TypeTagAddress() TypeTag {
	return TypeTag{Address: &struct{}{}}
}

// TypeTagSigner creates a TypeTag for the signer type.
func TypeTagSigner() TypeTag {
	return TypeTag{Signer: &struct{}{}}
}

// TypeTagVector creates a TypeTag for a vector of the given inner type.
func TypeTagVector(inner TypeTag) TypeTag {
	return TypeTag{Vector: &inner}
}

// TypeTagStruct creates a TypeTag for a struct.
func TypeTagStruct(tag StructTag) TypeTag {
	return TypeTag{Struct: &tag}
}

// NewStructTag creates a new StructTag.
func NewStructTag(address types.Address, module, name string, typeParams []TypeTag) StructTag {
	return StructTag{
		Address:    address,
		Module:     module,
		Name:       name,
		TypeParams: append([]TypeTag(nil), typeParams...),
	}
}

// String returns the string representation of the TypeTag.
func (t TypeTag) String() string {
	switch {
	case t.Bool != nil:
		return "bool"
	case t.U8 != nil:
		return "u8"
	case t.U16 != nil:
		return "u16"
	case t.U32 != nil:
		return "u32"
	case t.U64 != nil:
		return "u64"
	case t.U128 != nil:
		return "u128"
	case t.U256 != nil:
		return "u256"
	case t.Address != nil:
		return "address"
	case t.Signer != nil:
		return "signer"
	case t.Vector != nil:
		return fmt.Sprintf("vector<%s>", t.Vector.String())
	case t.Struct != nil:
		structTag := t.Struct
		params := make([]string, len(structTag.TypeParams))
		for i, param := range structTag.TypeParams {
			params[i] = param.String()
		}
		base := fmt.Sprintf("%s::%s::%s", structTag.Address.String(), structTag.Module, structTag.Name)
		if len(params) == 0 {
			return base
		}
		return fmt.Sprintf("%s<%s>", base, strings.Join(params, ", "))
	default:
		panic("invalid type tag")
	}
}
