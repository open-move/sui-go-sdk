package types

type Address [32]byte

type Digest [32]byte

type PersonalMessage struct {
	Message []byte
}

type ObjectRef struct {
	ObjectID Address
	Version  uint64
	Digest   Digest
}

type SharedObjectRef struct {
	ObjectID             Address
	InitialSharedVersion uint64
	Mutable              bool
}
