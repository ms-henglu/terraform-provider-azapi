package types

type TypeBase interface {
	AsTypeBase() *TypeBase
	Validate(interface{}, string) []error
	GetWriteOnly(interface{}) interface{}
	TypeOfProperty(interface{}, string) *TypeBase
}

const ArrayItem = "array_item"
