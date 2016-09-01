package codegen

// Ref is a reference of an object
type Ref interface {
	String() string
	Size() int32
	RegSizeAlign() bool
}
