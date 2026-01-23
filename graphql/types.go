package graphql

import (
	"encoding/json"
	"math/big"
	"strconv"
)

// SuiAddress represents a Sui address (32-byte hex string with 0x prefix).
type SuiAddress string

// Base64 represents Base64-encoded binary data.
type Base64 string

// DateTime represents an ISO-8601 formatted date-time string.
type DateTime string

// BigInt represents an arbitrarily large integer as a string.
type BigInt string

// ToBigInt converts the BigInt string to a *big.Int.
func (b BigInt) ToBigInt() (*big.Int, bool) {
	n := new(big.Int)
	return n.SetString(string(b), 10)
}

// UInt53 represents a 53-bit unsigned integer (safe for JavaScript).
type UInt53 uint64

// UnmarshalJSON handles both number and string representations.
func (u *UInt53) UnmarshalJSON(data []byte) error {
	var num uint64
	if err := json.Unmarshal(data, &num); err == nil {
		*u = UInt53(num)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parsed, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return err
	}
	*u = UInt53(parsed)
	return nil
}

// PageInfo contains pagination information.
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}

// PaginationArgs contains common pagination arguments.
type PaginationArgs struct {
	First  *int    `json:"first,omitempty"`
	After  *string `json:"after,omitempty"`
	Last   *int    `json:"last,omitempty"`
	Before *string `json:"before,omitempty"`
}

// ToVariables converts pagination args to GraphQL variables.
func (p *PaginationArgs) ToVariables() map[string]any {
	if p == nil {
		return nil
	}
	vars := make(map[string]any)
	if p.First != nil {
		vars["first"] = *p.First
	}
	if p.After != nil {
		vars["after"] = *p.After
	}
	if p.Last != nil {
		vars["last"] = *p.Last
	}
	if p.Before != nil {
		vars["before"] = *p.Before
	}
	return vars
}

// Connection is a generic paginated connection type.
type Connection[T any] struct {
	PageInfo PageInfo  `json:"pageInfo"`
	Nodes    []T       `json:"nodes"`
	Edges    []Edge[T] `json:"edges,omitempty"`
}

// Edge represents an edge in a connection.
type Edge[T any] struct {
	Node   T      `json:"node"`
	Cursor string `json:"cursor"`
}

// OwnerKind represents the type of object ownership.
type OwnerKind string

const (
	OwnerKindImmutable OwnerKind = "IMMUTABLE"
	OwnerKindAddress   OwnerKind = "ADDRESS"
	OwnerKindParent    OwnerKind = "PARENT"
	OwnerKindShared    OwnerKind = "SHARED"
)

// ExecutionStatus represents the execution status of a transaction.
type ExecutionStatus string

const (
	ExecutionStatusSuccess ExecutionStatus = "SUCCESS"
	ExecutionStatusFailure ExecutionStatus = "FAILURE"
)

// MoveAbility represents Move type abilities.
type MoveAbility string

const (
	MoveAbilityCopy  MoveAbility = "COPY"
	MoveAbilityDrop  MoveAbility = "DROP"
	MoveAbilityKey   MoveAbility = "KEY"
	MoveAbilityStore MoveAbility = "STORE"
)

// MoveVisibility represents Move function visibility.
type MoveVisibility string

const (
	MoveVisibilityPublic  MoveVisibility = "PUBLIC"
	MoveVisibilityFriend  MoveVisibility = "FRIEND"
	MoveVisibilityPrivate MoveVisibility = "PRIVATE"
)

// AddressTransactionRelationship specifies transaction relationship for address queries.
type AddressTransactionRelationship string

const (
	AddressTransactionRelationshipSent     AddressTransactionRelationship = "SENT"
	AddressTransactionRelationshipAffected AddressTransactionRelationship = "AFFECTED"
)

// ZkLoginIntentScope specifies the intent scope for zkLogin verification.
type ZkLoginIntentScope string

const (
	ZkLoginIntentScopeTransactionData ZkLoginIntentScope = "TRANSACTION_DATA"
	ZkLoginIntentScopePersonalMessage ZkLoginIntentScope = "PERSONAL_MESSAGE"
)

// FieldSelector controls which fields are requested in GraphQL queries.
type FieldSelector struct {
	IncludeAll    bool     // If true, request all available fields (default: true)
	IncludeNested bool     // If true, include nested object fields (default: true)
	CustomFields  []string // Specific fields to include (overrides IncludeAll if provided)
}

// DefaultFieldSelector returns a FieldSelector that requests all fields.
func DefaultFieldSelector() *FieldSelector {
	return &FieldSelector{
		IncludeAll:    true,
		IncludeNested: true,
	}
}

// --- From mutations.go ---

// TransactionBuilder provides a fluent API for building programmable transactions.
// It helps users construct transactions that can be simulated or executed.
type TransactionBuilder struct {
	Sender       *SuiAddress
	GasObject    *ObjectRef
	GasBudget    *uint64
	GasPrice     *uint64
	GasSponsor   *SuiAddress
	Expiration   *EpochRef
	Inputs       []TransactionInputBuilder
	Transactions []ProgrammableTransactionCommand
}

// ObjectRef identifies an object by ID, version, and digest.
type ObjectRef struct {
	ObjectID SuiAddress `json:"objectId"`
	Version  UInt53     `json:"version"`
	Digest   string     `json:"digest"`
}

// EpochRef specifies an epoch for transaction expiration.
type EpochRef struct {
	EpochID uint64 `json:"epochId"`
}

// TransactionInputBuilder represents an input being built.
type TransactionInputBuilder struct {
	Kind       string      `json:"kind"`
	Type       string      `json:"type,omitempty"`       // For Pure inputs
	Value      interface{} `json:"value,omitempty"`      // For Pure inputs
	ObjectRef  *ObjectRef  `json:"objectRef,omitempty"`  // For Object inputs
	SharedInfo *SharedInfo `json:"sharedInfo,omitempty"` // For Shared object inputs
}

// SharedInfo contains information for shared object inputs.
type SharedInfo struct {
	ObjectID             SuiAddress `json:"objectId"`
	InitialSharedVersion UInt53     `json:"initialSharedVersion"`
	Mutable              bool       `json:"mutable"`
}

// ProgrammableTransactionCommand represents a command in a programmable transaction.
type ProgrammableTransactionCommand struct {
	Kind          string       `json:"kind"`
	Target        string       `json:"target,omitempty"`        // For MoveCall
	TypeArguments []string     `json:"typeArguments,omitempty"` // For MoveCall
	Arguments     []Argument   `json:"arguments,omitempty"`     // For various commands
	Destinations  []Argument   `json:"destinations,omitempty"`  // For SplitCoins
	Amounts       []Argument   `json:"amounts,omitempty"`       // For SplitCoins
	Objects       []Argument   `json:"objects,omitempty"`       // For MergeCoins, MakeMoveVec
	Coin          *Argument    `json:"coin,omitempty"`          // For MergeCoins, SplitCoins
	Address       *Argument    `json:"address,omitempty"`       // For TransferObjects
	Object        *Argument    `json:"object,omitempty"`        // For Publish, Upgrade
	Modules       []Base64     `json:"modules,omitempty"`       // For Publish
	Dependencies  []SuiAddress `json:"dependencies,omitempty"`  // For Publish
	Package       *SuiAddress  `json:"package,omitempty"`       // For Upgrade
	Ticket        *Argument    `json:"ticket,omitempty"`        // For Upgrade
	ElementType   string       `json:"elementType,omitempty"`   // For MakeMoveVec
}

// Argument represents a transaction argument reference.
type Argument struct {
	Kind  string `json:"kind"`
	Index uint16 `json:"index,omitempty"`
	// For NestedResult
	ResultIndex uint16 `json:"resultIndex,omitempty"`
}

// ArgumentKind constants
const (
	ArgumentKindGasCoin      = "GasCoin"
	ArgumentKindInput        = "Input"
	ArgumentKindResult       = "Result"
	ArgumentKindNestedResult = "NestedResult"
)

// TransactionData represents a complete transaction ready for simulation or signing.
type TransactionData struct {
	Sender     *SuiAddress          `json:"sender,omitempty"`
	GasConfig  *GasConfig           `json:"gasConfig,omitempty"`
	Expiration *EpochRef            `json:"expiration,omitempty"`
	Kind       *TransactionKindData `json:"kind,omitempty"`
}

// GasConfig contains gas configuration for a transaction.
type GasConfig struct {
	Budget  *uint64     `json:"budget,omitempty"`
	Price   *uint64     `json:"price,omitempty"`
	Owner   *SuiAddress `json:"owner,omitempty"`
	Payment []ObjectRef `json:"payment,omitempty"`
}

// TransactionKindData represents the transaction kind.
type TransactionKindData struct {
	ProgrammableTransaction *ProgrammableTransactionData `json:"programmableTransaction,omitempty"`
}

// ProgrammableTransactionData represents a programmable transaction.
type ProgrammableTransactionData struct {
	Inputs       []TransactionInputBuilder        `json:"inputs"`
	Transactions []ProgrammableTransactionCommand `json:"transactions"`
}

// TransferSuiParams contains parameters for a SUI transfer.
type TransferSuiParams struct {
	Sender    SuiAddress `json:"sender"`
	Recipient SuiAddress `json:"recipient"`
	Amount    uint64     `json:"amount"`
	GasBudget uint64     `json:"gasBudget"`
}

// TransferObjectParams contains parameters for transferring an object.
type TransferObjectParams struct {
	Sender    SuiAddress `json:"sender"`
	Recipient SuiAddress `json:"recipient"`
	Object    ObjectRef  `json:"object"`
	GasBudget uint64     `json:"gasBudget"`
}

// MoveCallParams contains parameters for a Move function call.
type MoveCallParams struct {
	Sender        SuiAddress    `json:"sender"`
	Package       SuiAddress    `json:"package"`
	Module        string        `json:"module"`
	Function      string        `json:"function"`
	TypeArguments []string      `json:"typeArguments,omitempty"`
	Arguments     []interface{} `json:"arguments,omitempty"`
	GasBudget     uint64        `json:"gasBudget"`
}

// SplitCoinsParams contains parameters for splitting coins.
type SplitCoinsParams struct {
	Sender    SuiAddress `json:"sender"`
	Coin      ObjectRef  `json:"coin"`
	Amounts   []uint64   `json:"amounts"`
	GasBudget uint64     `json:"gasBudget"`
}

// MergeCoinsParams contains parameters for merging coins.
type MergeCoinsParams struct {
	Sender      SuiAddress  `json:"sender"`
	Destination ObjectRef   `json:"destination"`
	Sources     []ObjectRef `json:"sources"`
	GasBudget   uint64      `json:"gasBudget"`
}

// PublishPackageParams contains parameters for publishing a package.
type PublishPackageParams struct {
	Sender       SuiAddress   `json:"sender"`
	Modules      []Base64     `json:"modules"`
	Dependencies []SuiAddress `json:"dependencies"`
	GasBudget    uint64       `json:"gasBudget"`
}

// --- From events.go ---

// Event represents a Sui event.
type Event struct {
	TransactionModule *MoveModule `json:"transactionModule"`
	Sender            *Address    `json:"sender"`
	Timestamp         *DateTime   `json:"timestamp"`
	Contents          *MoveValue  `json:"contents"`
	EventBcs          Base64      `json:"eventBcs"`
}

// EventFilter contains filters for event queries.
type EventFilter struct {
	Sender            *SuiAddress `json:"sender,omitempty"`
	TransactionDigest *string     `json:"transactionDigest,omitempty"`
	EmittingModule    *string     `json:"emittingModule,omitempty"`
	EventType         *string     `json:"eventType,omitempty"`
}

// CoinMetadata represents coin metadata.
type CoinMetadata struct {
	Address     SuiAddress `json:"address"`
	Version     UInt53     `json:"version"`
	Digest      string     `json:"digest"`
	Decimals    *int       `json:"decimals"`
	Name        *string    `json:"name"`
	Symbol      *string    `json:"symbol"`
	Description *string    `json:"description"`
	IconURL     *string    `json:"iconUrl"`
	Supply      *BigInt    `json:"supply"`
}

// ServiceConfig represents the GraphQL service configuration.
type ServiceConfig struct {
	MaxQueryDepth        int `json:"maxQueryDepth"`
	MaxQueryNodes        int `json:"maxQueryNodes"`
	MaxOutputNodes       int `json:"maxOutputNodes"`
	QueryTimeoutMs       int `json:"queryTimeoutMs"`
	MaxQueryPayloadSize  int `json:"maxQueryPayloadSize"`
	MaxTypeArgumentDepth int `json:"maxTypeArgumentDepth"`
	MaxTypeNodes         int `json:"maxTypeNodes"`
	MaxMoveValueDepth    int `json:"maxMoveValueDepth"`
}

// AvailableRange represents the available data range.
type AvailableRange struct {
	First *Checkpoint `json:"first"`
	Last  *Checkpoint `json:"last"`
}

// ZkLoginVerifyResult represents the result of zkLogin verification.
type ZkLoginVerifyResult struct {
	Success bool    `json:"success"`
	Error   *string `json:"error"`
}

// ExecuteTransactionResult represents the result of executing a transaction.
type ExecuteTransactionResult struct {
	Effects *TransactionEffects `json:"effects"`
	Errors  []string            `json:"errors"`
}

// SimulationResult represents the result of simulating a transaction.
type SimulationResult struct {
	Effects *TransactionEffects `json:"effects"`
	Outputs []CommandResult     `json:"outputs"`
	Error   *string             `json:"error"`
}

// CommandResult represents the result of a command in a programmable transaction.
type CommandResult struct {
	// Each element is a JSON representation of the returned value
	Results []MoveValueResult `json:"results"`
}

// MoveValueResult represents a Move value returned from simulation.
type MoveValueResult struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// PackageCheckpointFilter filters packages by checkpoint.
type PackageCheckpointFilter struct {
	AfterCheckpoint  *UInt53 `json:"afterCheckpoint,omitempty"`
	BeforeCheckpoint *UInt53 `json:"beforeCheckpoint,omitempty"`
}

// --- From objects.go ---

// Object represents a Sui object.
type Object struct {
	Address                  SuiAddress      `json:"address"`
	Version                  UInt53          `json:"version"`
	Digest                   string          `json:"digest"`
	StorageRebate            *BigInt         `json:"storageRebate,omitempty"`
	Owner                    *ObjectOwner    `json:"owner,omitempty"`
	PreviousTransactionBlock *TransactionRef `json:"previousTransactionBlock,omitempty"`
	Display                  []DisplayEntry  `json:"display,omitempty"`
	ObjectBcs                *Base64         `json:"objectBcs,omitempty"`
	HasPublicTransfer        *bool           `json:"hasPublicTransfer,omitempty"`
	AsMoveObject             *MoveObject     `json:"asMoveObject,omitempty"`
	AsMovePackage            *MovePackage    `json:"asMovePackage,omitempty"`
}

// TransactionRef is a minimal transaction reference.
type TransactionRef struct {
	Digest string `json:"digest"`
}

// DisplayEntry represents a single display field.
type DisplayEntry struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
	Error *string `json:"error,omitempty"`
}

// OwnerAddress represents the address of an owner.
type OwnerAddress struct {
	Address SuiAddress `json:"address"`
}

// ObjectOwner represents ownership information for an object.
// This is a union type that can be AddressOwner, ObjectOwner, Shared, or Immutable.
type ObjectOwner struct {
	// For address ownership (AddressOwner or ObjectOwner)
	Address *OwnerAddress `json:"address,omitempty"`
	// For shared objects
	InitialSharedVersion *UInt53 `json:"initialSharedVersion,omitempty"`
	// Raw typename for owner kind detection
	Typename string `json:"__typename,omitempty"`
}

// MoveObject represents a Move object with type information.
type MoveObject struct {
	Address           SuiAddress `json:"address"`
	Version           UInt53     `json:"version"`
	Digest            string     `json:"digest"`
	Contents          *MoveValue `json:"contents,omitempty"`
	HasPublicTransfer bool       `json:"hasPublicTransfer"`
	Type              *MoveType  `json:"type,omitempty"`
}

// MovePackage represents a published Move package.
type MovePackage struct {
	Address     SuiAddress              `json:"address"`
	Version     UInt53                  `json:"version"`
	Digest      string                  `json:"digest"`
	Modules     *Connection[MoveModule] `json:"modules,omitempty"`
	Linkage     []Linkage               `json:"linkage,omitempty"`
	TypeOrigins []TypeOrigin            `json:"typeOrigins,omitempty"`
}

// MoveModule represents a Move module.
type MoveModule struct {
	Name              string                    `json:"name"`
	Package           *MovePackageRef           `json:"package,omitempty"`
	FileFormatVersion int                       `json:"fileFormatVersion"`
	Friends           *Connection[MoveModule]   `json:"friends,omitempty"`
	Structs           *Connection[MoveStruct]   `json:"structs,omitempty"`
	Enums             *Connection[MoveEnum]     `json:"enums,omitempty"`
	Functions         *Connection[MoveFunction] `json:"functions,omitempty"`
	Bytes             *Base64                   `json:"bytes,omitempty"`
	Disassembly       *string                   `json:"disassembly,omitempty"`
}

// MovePackageRef is a reference to a Move package.
type MovePackageRef struct {
	Address SuiAddress `json:"address"`
}

// MoveStruct represents a Move struct definition.
type MoveStruct struct {
	Name           string                      `json:"name"`
	Abilities      []MoveAbility               `json:"abilities"`
	TypeParameters []MoveDatatypeTypeParameter `json:"typeParameters"`
	Fields         []MoveField                 `json:"fields"`
}

// MoveEnum represents a Move enum definition.
type MoveEnum struct {
	Name           string                      `json:"name"`
	Abilities      []MoveAbility               `json:"abilities"`
	TypeParameters []MoveDatatypeTypeParameter `json:"typeParameters"`
	Variants       []MoveEnumVariant           `json:"variants"`
}

// MoveEnumVariant represents a variant of a Move enum.
type MoveEnumVariant struct {
	Name   string      `json:"name"`
	Fields []MoveField `json:"fields"`
}

// MoveDatatypeTypeParameter represents a type parameter.
type MoveDatatypeTypeParameter struct {
	Constraints []MoveAbility `json:"constraints"`
	IsPhantom   bool          `json:"isPhantom"`
}

// MoveField represents a field in a Move struct or enum variant.
type MoveField struct {
	Name string        `json:"name"`
	Type *OpenMoveType `json:"type"`
}

// MoveFunction represents a Move function.
type MoveFunction struct {
	Name           string                      `json:"name"`
	Visibility     MoveVisibility              `json:"visibility"`
	IsEntry        bool                        `json:"isEntry"`
	TypeParameters []MoveFunctionTypeParameter `json:"typeParameters"`
	Parameters     []OpenMoveType              `json:"parameters"`
	Return         []OpenMoveType              `json:"return"`
}

// MoveFunctionTypeParameter represents a function type parameter.
type MoveFunctionTypeParameter struct {
	Constraints []MoveAbility `json:"constraints"`
}

// MoveType represents a concrete Move type.
type MoveType struct {
	Repr      string             `json:"repr"`
	Signature *MoveTypeSignature `json:"signature,omitempty"`
	Layout    *MoveTypeLayout    `json:"layout,omitempty"`
	Abilities []MoveAbility      `json:"abilities,omitempty"`
}

// MoveTypeSignature represents a type signature.
type MoveTypeSignature struct {
	// Contains nested type information
	json.RawMessage
}

// MoveTypeLayout represents type layout information.
type MoveTypeLayout struct {
	// Contains nested layout information
	json.RawMessage
}

// OpenMoveType represents a type that may have unbound type parameters.
type OpenMoveType struct {
	Repr      string             `json:"repr"`
	Signature *MoveTypeSignature `json:"signature,omitempty"`
}

// MoveValue represents a Move value.
type MoveValue struct {
	Type MoveType        `json:"type"`
	Bcs  Base64          `json:"bcs"`
	Json json.RawMessage `json:"json,omitempty"`
}

// Linkage represents package linkage information.
type Linkage struct {
	OriginalID SuiAddress `json:"originalId"`
	UpgradedID SuiAddress `json:"upgradedId"`
	Version    UInt53     `json:"version"`
}

// TypeOrigin represents the origin of a type.
type TypeOrigin struct {
	Module     string     `json:"module"`
	Struct     string     `json:"struct"`
	DefiningId SuiAddress `json:"definingId"`
}

// DynamicField represents a dynamic field on an object.
type DynamicField struct {
	Name  *MoveValue         `json:"name"`
	Value *DynamicFieldValue `json:"value"`
}

// DynamicFieldValue represents the value of a dynamic field.
type DynamicFieldValue struct {
	// Can be MoveValue for dynamic fields or MoveObject for dynamic object fields
	AsMoveValue  *MoveValue  `json:"asMoveValue,omitempty"`
	AsMoveObject *MoveObject `json:"asMoveObject,omitempty"`
}

// DynamicFieldName is an input type for querying dynamic fields.
type DynamicFieldName struct {
	Type string `json:"type"`
	Bcs  Base64 `json:"bcs"`
}

// ObjectFilter contains filters for object queries.
type ObjectFilter struct {
	Type       *string      `json:"type,omitempty"`
	Owner      *SuiAddress  `json:"owner,omitempty"`
	ObjectID   *SuiAddress  `json:"objectId,omitempty"`
	ObjectIDs  []SuiAddress `json:"objectIds,omitempty"`
	ObjectKeys []ObjectKey  `json:"objectKeys,omitempty"`
}

// ObjectKey identifies an object by address and optional version.
type ObjectKey struct {
	Address      SuiAddress `json:"address"`
	Version      *UInt53    `json:"version,omitempty"`
	RootVersion  *UInt53    `json:"rootVersion,omitempty"`
	AtCheckpoint *UInt53    `json:"atCheckpoint,omitempty"`
}

// VersionFilter filters object versions.
type VersionFilter struct {
	AfterVersion  *UInt53 `json:"afterVersion,omitempty"`
	BeforeVersion *UInt53 `json:"beforeVersion,omitempty"`
}

// --- From transactions.go ---

// Transaction represents a Sui transaction.
type Transaction struct {
	Digest         string              `json:"digest"`
	Sender         *Address            `json:"sender"`
	GasInput       *GasInput           `json:"gasInput"`
	Kind           *TransactionKind    `json:"kind"`
	Signatures     []UserSignature     `json:"signatures,omitempty"`
	Effects        *TransactionEffects `json:"effects,omitempty"`
	Expiration     *Epoch              `json:"expiration,omitempty"`
	TransactionBcs *Base64             `json:"transactionBcs,omitempty"`
}

// Address represents a Sui address with associated data.
type Address struct {
	Address            SuiAddress                     `json:"address"`
	Balance            *Balance                       `json:"balance,omitempty"`
	Balances           *Connection[Balance]           `json:"balances,omitempty"`
	StakedSuis         *Connection[StakedSui]         `json:"stakedSuis,omitempty"`
	Objects            *Connection[MoveObject]        `json:"objects,omitempty"`
	SuinsRegistrations *Connection[SuinsRegistration] `json:"suinsRegistrations,omitempty"`
	TransactionBlocks  *Connection[Transaction]       `json:"transactionBlocks,omitempty"`
}

// Balance represents a coin balance.
type Balance struct {
	CoinType     *MoveType `json:"coinType,omitempty"`
	TotalBalance BigInt    `json:"totalBalance"`
}

// Coin represents a coin object.
type Coin struct {
	CoinBalance BigInt     `json:"coinBalance"`
	Address     SuiAddress `json:"address"`
	Version     UInt53     `json:"version"`
	Digest      string     `json:"digest"`
	Contents    *MoveValue `json:"contents,omitempty"`
}

// StakedSui represents staked SUI.
type StakedSui struct {
	Address         SuiAddress `json:"address"`
	Version         UInt53     `json:"version"`
	Digest          string     `json:"digest"`
	Principal       BigInt     `json:"principal"`
	StakeStatus     string     `json:"stakeStatus"`
	ActivatedEpoch  *Epoch     `json:"activatedEpoch"`
	RequestedEpoch  *Epoch     `json:"requestedEpoch"`
	EstimatedReward *BigInt    `json:"estimatedReward"`
}

// SuinsRegistration represents a SuiNS registration.
type SuinsRegistration struct {
	Domain  string     `json:"domain"`
	Address SuiAddress `json:"address"`
}

// GasInput represents gas configuration for a transaction.
type GasInput struct {
	GasSponsor *Address            `json:"gasSponsor,omitempty"`
	GasPayment *Connection[Object] `json:"gasPayment,omitempty"`
	GasPrice   *BigInt             `json:"gasPrice,omitempty"`
	GasBudget  *BigInt             `json:"gasBudget,omitempty"`
}

// UserSignature represents a user's signature.
type UserSignature struct {
	SignatureBytes Base64 `json:"signatureBytes"`
}

// TransactionKind represents the kind of transaction.
type TransactionKind struct {
	Typename string `json:"__typename"`
	// For ProgrammableTransaction
	Inputs       *Connection[TransactionInput]    `json:"inputs,omitempty"`
	Transactions *Connection[ProgrammableCommand] `json:"commands,omitempty"`
}

// TransactionInput represents an input to a transaction.
type TransactionInput struct {
	Typename    string       `json:"__typename"`
	Pure        *PureInput   `json:"pure,omitempty"`
	Object      *ObjectInput `json:"object,omitempty"`
	SharedInput *SharedInput `json:"sharedInput,omitempty"`
	MoveValue   *MoveValue   `json:"moveValue,omitempty"`
}

// SharedInput represents a shared object input.
type SharedInput struct {
	Address              SuiAddress `json:"address"`
	InitialSharedVersion UInt53     `json:"initialSharedVersion"`
	Mutable              bool       `json:"mutable"`
}

// PureInput represents a pure value input.
type PureInput struct {
	Bytes Base64 `json:"bytes"`
}

// ObjectInput represents an object input.
type ObjectInput struct {
	Address SuiAddress `json:"address"`
	Version *UInt53    `json:"version,omitempty"`
	Digest  *string    `json:"digest,omitempty"`
}

// ProgrammableCommand represents a command in a programmable transaction.
type ProgrammableCommand struct {
	Typename string `json:"__typename"`
	// Different fields depending on command type
	json.RawMessage
}

// TransactionEffects represents the effects of a transaction.
type TransactionEffects struct {
	Digest         string                     `json:"digest"`
	Status         ExecutionStatus            `json:"status"`
	ExecutionError *ExecutionError            `json:"executionError,omitempty"`
	Lamport        UInt53                     `json:"lamportVersion"`
	Dependencies   *Connection[Transaction]   `json:"dependencies,omitempty"`
	BalanceChanges *Connection[BalanceChange] `json:"balanceChanges,omitempty"`
	ObjectChanges  *Connection[ObjectChange]  `json:"objectChanges,omitempty"`
	GasEffects     *GasEffects                `json:"gasEffects,omitempty"`
	Epoch          *Epoch                     `json:"epoch,omitempty"`
	Checkpoint     *Checkpoint                `json:"checkpoint,omitempty"`
	Timestamp      *DateTime                  `json:"timestamp,omitempty"`
	EffectsBcs     *Base64                    `json:"effectsBcs,omitempty"`
	EffectsJson    any                        `json:"effectsJson,omitempty"`
	EffectsDigest  *string                    `json:"effectsDigest,omitempty"`
}

// ExecutionError represents an execution error.
type ExecutionError struct {
	Message           string      `json:"message"`
	AbortCode         *BigInt     `json:"abortCode"`
	SourceLineNumber  *int        `json:"sourceLineNumber"`
	InstructionOffset *int        `json:"instructionOffset"`
	Identifier        *string     `json:"identifier"`
	Constant          *string     `json:"constant"`
	Module            *MoveModule `json:"module"`
}

// ExecutionResult represents the execution status.
type ExecutionResult struct {
	Status ExecutionStatus `json:"status"`
	// Error message if failed
	Error *string `json:"error,omitempty"`
}

// BalanceChange represents a balance change from a transaction.
type BalanceChange struct {
	Owner    *Address  `json:"owner"`
	CoinType *MoveType `json:"coinType"`
	Amount   BigInt    `json:"amount"`
}

// ObjectChange represents an object change from a transaction.
type ObjectChange struct {
	Address     SuiAddress `json:"address"`
	IDCreated   *bool      `json:"idCreated,omitempty"`
	IDDeleted   *bool      `json:"idDeleted,omitempty"`
	InputState  *Object    `json:"inputState,omitempty"`
	OutputState *Object    `json:"outputState,omitempty"`
}

// GasEffects represents gas usage information.
type GasEffects struct {
	GasSummary *GasCostSummary `json:"gasSummary"`
	GasObject  *Object         `json:"gasObject,omitempty"`
}

// GasCostSummary summarizes gas costs.
type GasCostSummary struct {
	ComputationCost         UInt53 `json:"computationCost"`
	StorageCost             UInt53 `json:"storageCost"`
	StorageRebate           UInt53 `json:"storageRebate"`
	NonRefundableStorageFee UInt53 `json:"nonRefundableStorageFee"`
}

// TransactionFilter contains filters for transaction queries.
type TransactionFilter struct {
	Function         *string     `json:"function,omitempty"`
	Kind             *string     `json:"kind,omitempty"`
	AfterCheckpoint  *UInt53     `json:"afterCheckpoint,omitempty"`
	BeforeCheckpoint *UInt53     `json:"beforeCheckpoint,omitempty"`
	AtCheckpoint     *UInt53     `json:"atCheckpoint,omitempty"`
	SignAddress      *SuiAddress `json:"signAddress,omitempty"`
	SentAddress      *SuiAddress `json:"sentAddress,omitempty"`
	RecvAddress      *SuiAddress `json:"recvAddress,omitempty"`
	PaidAddress      *SuiAddress `json:"paidAddress,omitempty"`
	InputObject      *SuiAddress `json:"inputObject,omitempty"`
	ChangedObject    *SuiAddress `json:"changedObject,omitempty"`
	TransactionIDs   []string    `json:"transactionIds,omitempty"`
}

// --- From epochs.go ---

// Epoch represents a Sui epoch.
type Epoch struct {
	EpochID             UInt53                   `json:"epochId"`
	ReferenceGasPrice   *BigInt                  `json:"referenceGasPrice,omitempty"`
	StartTimestamp      *DateTime                `json:"startTimestamp,omitempty"`
	EndTimestamp        *DateTime                `json:"endTimestamp,omitempty"`
	TotalCheckpoints    *UInt53                  `json:"totalCheckpoints,omitempty"`
	TotalTransactions   *UInt53                  `json:"totalTransactions,omitempty"`
	TotalGasFees        *BigInt                  `json:"totalGasFees,omitempty"`
	TotalStakeRewards   *BigInt                  `json:"totalStakeRewards,omitempty"`
	TotalStakeSubsidies *BigInt                  `json:"totalStakeSubsidies,omitempty"`
	FundSize            *BigInt                  `json:"fundSize,omitempty"`
	FundInflow          *BigInt                  `json:"fundInflow,omitempty"`
	FundOutflow         *BigInt                  `json:"fundOutflow,omitempty"`
	NetInflow           *BigInt                  `json:"netInflow,omitempty"`
	ProtocolConfigs     *ProtocolConfigs         `json:"protocolConfigs,omitempty"`
	ValidatorSet        *ValidatorSet            `json:"validatorSet,omitempty"`
	Checkpoints         *Connection[Checkpoint]  `json:"checkpoints,omitempty"`
	TransactionBlocks   *Connection[Transaction] `json:"transactionBlocks,omitempty"`
	FirstCheckpoint     *Checkpoint              `json:"firstCheckpoint,omitempty"`
	LastCheckpoint      *Checkpoint              `json:"lastCheckpoint,omitempty"`
	SafeMode            *SafeMode                `json:"safeMode,omitempty"`
	SystemStateVersion  *UInt53                  `json:"systemStateVersion,omitempty"`
	StorageFund         *StorageFund             `json:"storageFund,omitempty"`
	SystemParameters    *SystemParameters        `json:"systemParameters,omitempty"`
	StakeSubsidy        *StakeSubsidy            `json:"stakeSubsidy,omitempty"`
}

// Checkpoint represents a Sui checkpoint.
type Checkpoint struct {
	SequenceNumber           UInt53                   `json:"sequenceNumber"`
	Digest                   string                   `json:"digest"`
	Timestamp                *DateTime                `json:"timestamp,omitempty"`
	PreviousCheckpointDigest *string                  `json:"previousCheckpointDigest,omitempty"`
	NetworkTotalTransactions *UInt53                  `json:"networkTotalTransactions,omitempty"`
	RollingGasSummary        *GasCostSummary          `json:"rollingGasSummary,omitempty"`
	Epoch                    *Epoch                   `json:"epoch,omitempty"`
	TransactionBlocks        *Connection[Transaction] `json:"transactionBlocks,omitempty"`
	EndOfEpoch               *EndOfEpochData          `json:"endOfEpoch,omitempty"`
}

// EndOfEpochData contains end of epoch information.
type EndOfEpochData struct {
	NextProtocolVersion *UInt53 `json:"nextProtocolVersion,omitempty"`
}

// ValidatorSet represents the set of validators for an epoch.
type ValidatorSet struct {
	TotalStake                  BigInt                 `json:"totalStake,omitempty"`
	PendingActiveValidatorsSize *UInt53                `json:"pendingActiveValidatorsSize,omitempty"`
	StakingPoolMappingsSize     *UInt53                `json:"stakingPoolMappingsSize,omitempty"`
	InactivePoolsSize           *UInt53                `json:"inactivePoolsSize,omitempty"`
	ValidatorCandidatesSize     *UInt53                `json:"validatorCandidatesSize,omitempty"`
	ActiveValidators            *Connection[Validator] `json:"activeValidators,omitempty"`
}

// Validator represents a Sui validator.
type Validator struct {
	Address                    SuiAddress            `json:"address"`
	Name                       *string               `json:"name,omitempty"`
	Description                *string               `json:"description,omitempty"`
	ImageURL                   *string               `json:"imageUrl,omitempty"`
	ProjectURL                 *string               `json:"projectUrl,omitempty"`
	OperationCap               *Object               `json:"operationCap,omitempty"`
	StakingPool                *Object               `json:"stakingPool,omitempty"`
	ExchangeRates              *Object               `json:"exchangeRates,omitempty"`
	ExchangeRatesSize          *UInt53               `json:"exchangeRatesSize,omitempty"`
	StakingPoolActivationEpoch *UInt53               `json:"stakingPoolActivationEpoch,omitempty"`
	StakingPoolSuiBalance      *BigInt               `json:"stakingPoolSuiBalance,omitempty"`
	RewardsPool                *BigInt               `json:"rewardsPool,omitempty"`
	PoolTokenBalance           *BigInt               `json:"poolTokenBalance,omitempty"`
	PendingStake               *BigInt               `json:"pendingStake,omitempty"`
	PendingTotalSuiWithdraw    *BigInt               `json:"pendingTotalSuiWithdraw,omitempty"`
	PendingPoolTokenWithdraw   *BigInt               `json:"pendingPoolTokenWithdraw,omitempty"`
	VotingPower                *UInt53               `json:"votingPower,omitempty"`
	GasPrice                   *BigInt               `json:"gasPrice,omitempty"`
	CommissionRate             *UInt53               `json:"commissionRate,omitempty"`
	NextEpochStake             *BigInt               `json:"nextEpochStake,omitempty"`
	NextEpochGasPrice          *BigInt               `json:"nextEpochGasPrice,omitempty"`
	NextEpochCommissionRate    *UInt53               `json:"nextEpochCommissionRate,omitempty"`
	AtRisk                     *UInt53               `json:"atRisk,omitempty"`
	ReportRecords              []SuiAddress          `json:"reportRecords,omitempty"`
	Credentials                *ValidatorCredentials `json:"credentials,omitempty"`
}

// ValidatorCredentials represents validator credentials.
type ValidatorCredentials struct {
	ProtocolPubKey    *Base64 `json:"protocolPubKey,omitempty"`
	NetworkPubKey     *Base64 `json:"networkPubKey,omitempty"`
	WorkerPubKey      *Base64 `json:"workerPubKey,omitempty"`
	ProofOfPossession *Base64 `json:"proofOfPossession,omitempty"`
	NetAddress        *string `json:"netAddress,omitempty"`
	P2PAddress        *string `json:"p2PAddress,omitempty"`
	PrimaryAddress    *string `json:"primaryAddress,omitempty"`
	WorkerAddress     *string `json:"workerAddress,omitempty"`
}

// ProtocolConfigs represents protocol configuration.
type ProtocolConfigs struct {
	ProtocolVersion UInt53           `json:"protocolVersion"`
	FeatureFlags    []FeatureFlag    `json:"featureFlags,omitempty"`
	Configs         []ProtocolConfig `json:"configs,omitempty"`
}

// FeatureFlag represents a protocol feature flag.
type FeatureFlag struct {
	Key   string `json:"key"`
	Value bool   `json:"value"`
}

// ProtocolConfig represents a protocol configuration value.
type ProtocolConfig struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

// SafeMode represents safe mode configuration.
type SafeMode struct {
	Enabled                 bool            `json:"enabled"`
	GasSummary              *GasCostSummary `json:"gasSummary,omitempty"`
	NonRefundableStorageFee *BigInt         `json:"nonRefundableStorageFee,omitempty"`
}

// StorageFund represents the storage fund.
type StorageFund struct {
	TotalObjectStorageRebates BigInt `json:"totalObjectStorageRebates"`
	NonRefundableBalance      BigInt `json:"nonRefundableBalance"`
}

// SystemParameters represents system parameters.
type SystemParameters struct {
	DurationMs                     *BigInt `json:"durationMs,omitempty"`
	StakeSubsidyStartEpoch         *UInt53 `json:"stakeSubsidyStartEpoch,omitempty"`
	MinValidatorCount              *UInt53 `json:"minValidatorCount,omitempty"`
	MaxValidatorCount              *UInt53 `json:"maxValidatorCount,omitempty"`
	MinValidatorJoiningStake       *BigInt `json:"minValidatorJoiningStake,omitempty"`
	ValidatorLowStakeThreshold     *BigInt `json:"validatorLowStakeThreshold,omitempty"`
	ValidatorVeryLowStakeThreshold *BigInt `json:"validatorVeryLowStakeThreshold,omitempty"`
	ValidatorLowStakeGracePeriod   *UInt53 `json:"validatorLowStakeGracePeriod,omitempty"`
}

// StakeSubsidy represents stake subsidy information.
type StakeSubsidy struct {
	Balance                   BigInt `json:"balance"`
	DistributionCounter       UInt53 `json:"distributionCounter"`
	CurrentDistributionAmount BigInt `json:"currentDistributionAmount"`
	PeriodLength              UInt53 `json:"periodLength"`
	DecreasRate               UInt53 `json:"decreaseRate"`
}

// CheckpointFilter contains filters for checkpoint queries.
type CheckpointFilter struct {
	AfterCheckpoint  *UInt53 `json:"afterCheckpoint,omitempty"`
	BeforeCheckpoint *UInt53 `json:"beforeCheckpoint,omitempty"`
}

// --- From execute.go ---

// SimulationOptions configures transaction simulation.
type SimulationOptions struct {
	ChecksEnabled  *bool `json:"checksEnabled,omitempty"`
	DoGasSelection *bool `json:"doGasSelection,omitempty"`
}

// =============================================================================
// Comprehensive Transaction Result Types
// =============================================================================

// TransactionResult represents a complete transaction result with all details.
// This matches the comprehensive JSON structure returned by Sui RPC.
type TransactionResult struct {
	// Transaction digest (unique identifier)
	Digest string `json:"digest"`

	// Transaction data
	Transaction *TransactionData_ `json:"transaction,omitempty"`

	// Transaction effects
	Effects *TransactionEffectsResult `json:"effects,omitempty"`

	// Events emitted by the transaction
	Events []EventResult `json:"events,omitempty"`

	// Object changes
	ObjectChanges []ObjectChangeResult `json:"objectChanges,omitempty"`

	// Balance changes
	BalanceChanges []BalanceChangeResult `json:"balanceChanges,omitempty"`

	// Timestamp in milliseconds
	TimestampMs *string `json:"timestampMs,omitempty"`

	// Whether the transaction was confirmed locally
	ConfirmedLocalExecution bool `json:"confirmedLocalExecution,omitempty"`

	// Checkpoint number
	Checkpoint *string `json:"checkpoint,omitempty"`

	// Errors from execution (if any)
	Errors []string `json:"errors,omitempty"`
}

// TransactionData_ represents the transaction data portion of a result.
type TransactionData_ struct {
	// Raw transaction data
	Data *TransactionDataContent `json:"data,omitempty"`

	// Transaction signatures
	TxSignatures []string `json:"txSignatures,omitempty"`
}

// TransactionDataContent represents the content of transaction data.
type TransactionDataContent struct {
	MessageVersion string `json:"messageVersion,omitempty"`

	// Transaction kind details
	Transaction *TransactionKindContent `json:"transaction,omitempty"`

	// Sender address
	Sender string `json:"sender,omitempty"`

	// Gas data
	GasData *GasData `json:"gasData,omitempty"`
}

// TransactionKindContent represents the kind of transaction.
type TransactionKindContent struct {
	Kind string `json:"kind,omitempty"`

	// Inputs for programmable transactions
	Inputs []InputContent `json:"inputs,omitempty"`

	// Commands/transactions
	Transactions []json.RawMessage `json:"transactions,omitempty"`
}

// InputContent represents an input in a transaction.
type InputContent struct {
	Type      string `json:"type,omitempty"`
	ValueType string `json:"valueType,omitempty"`
	Value     any    `json:"value,omitempty"`
	ObjectID  string `json:"objectId,omitempty"`
	Version   string `json:"version,omitempty"`
	Digest    string `json:"digest,omitempty"`
}

// GasData represents gas payment data.
type GasData struct {
	Payment []ObjectRefResult `json:"payment,omitempty"`
	Owner   string            `json:"owner,omitempty"`
	Price   string            `json:"price,omitempty"`
	Budget  string            `json:"budget,omitempty"`
}

// ObjectRefResult represents an object reference in results.
type ObjectRefResult struct {
	ObjectID string `json:"objectId"`
	Version  string `json:"version"`
	Digest   string `json:"digest"`
}

// TransactionEffectsResult represents transaction effects.
type TransactionEffectsResult struct {
	MessageVersion     string              `json:"messageVersion,omitempty"`
	Status             *StatusResult       `json:"status,omitempty"`
	ExecutedEpoch      string              `json:"executedEpoch,omitempty"`
	GasUsed            *GasUsedResult      `json:"gasUsed,omitempty"`
	ModifiedAtVersions []ModifiedAtVersion `json:"modifiedAtVersions,omitempty"`
	TransactionDigest  string              `json:"transactionDigest,omitempty"`
	Created            []ObjectOwnerResult `json:"created,omitempty"`
	Mutated            []ObjectOwnerResult `json:"mutated,omitempty"`
	Deleted            []ObjectRefResult   `json:"deleted,omitempty"`
	GasObject          *ObjectOwnerResult  `json:"gasObject,omitempty"`
	Dependencies       []string            `json:"dependencies,omitempty"`
}

// StatusResult represents execution status.
type StatusResult struct {
	Status string  `json:"status"`
	Error  *string `json:"error,omitempty"`
}

// GasUsedResult represents gas usage details.
type GasUsedResult struct {
	ComputationCost         string `json:"computationCost"`
	StorageCost             string `json:"storageCost"`
	StorageRebate           string `json:"storageRebate"`
	NonRefundableStorageFee string `json:"nonRefundableStorageFee"`
}

// ModifiedAtVersion represents an object modified at a specific version.
type ModifiedAtVersion struct {
	ObjectID       string `json:"objectId"`
	SequenceNumber string `json:"sequenceNumber"`
}

// ObjectOwnerResult represents an object with its owner.
type ObjectOwnerResult struct {
	Owner     any              `json:"owner"` // Can be "Immutable" string or {"AddressOwner": "0x..."} object
	Reference *ObjectRefResult `json:"reference,omitempty"`
}

// EventResult represents an event emitted by a transaction.
type EventResult struct {
	ID                *EventID `json:"id,omitempty"`
	PackageID         string   `json:"packageId,omitempty"`
	TransactionModule string   `json:"transactionModule,omitempty"`
	Sender            string   `json:"sender,omitempty"`
	Type              string   `json:"type,omitempty"`
	ParsedJSON        any      `json:"parsedJson,omitempty"`
	Bcs               string   `json:"bcs,omitempty"`
	TimestampMs       string   `json:"timestampMs,omitempty"`
}

// EventID represents an event identifier.
type EventID struct {
	TxDigest string `json:"txDigest"`
	EventSeq string `json:"eventSeq"`
}

// ObjectChangeResult represents an object change from a transaction.
type ObjectChangeResult struct {
	Type            string `json:"type"` // "created", "mutated", "deleted", "published", "transferred", "wrapped"
	Sender          string `json:"sender,omitempty"`
	Owner           any    `json:"owner,omitempty"` // Can be string or object
	ObjectType      string `json:"objectType,omitempty"`
	ObjectID        string `json:"objectId,omitempty"`
	Version         string `json:"version,omitempty"`
	PreviousVersion string `json:"previousVersion,omitempty"`
	Digest          string `json:"digest,omitempty"`
	// For published packages
	PackageID string   `json:"packageId,omitempty"`
	Modules   []string `json:"modules,omitempty"`
}

// BalanceChangeResult represents a balance change from a transaction.
type BalanceChangeResult struct {
	Owner    any    `json:"owner"` // Can be {"AddressOwner": "0x..."} or other owner types
	CoinType string `json:"coinType"`
	Amount   string `json:"amount"`
}

// =============================================================================
// Helper Methods for TransactionResult
// =============================================================================

// IsSuccess returns true if the transaction executed successfully.
func (r *TransactionResult) IsSuccess() bool {
	if r.Effects != nil && r.Effects.Status != nil {
		return r.Effects.Status.Status == "success"
	}
	return false
}

// GetError returns the execution error if any.
func (r *TransactionResult) GetError() *string {
	if r.Effects != nil && r.Effects.Status != nil {
		return r.Effects.Status.Error
	}
	return nil
}

// GetGasUsed returns the total gas used (computation + storage - rebate).
func (r *TransactionResult) GetGasUsed() (uint64, error) {
	if r.Effects == nil || r.Effects.GasUsed == nil {
		return 0, nil
	}
	gas := r.Effects.GasUsed
	comp, _ := strconv.ParseUint(gas.ComputationCost, 10, 64)
	storage, _ := strconv.ParseUint(gas.StorageCost, 10, 64)
	rebate, _ := strconv.ParseUint(gas.StorageRebate, 10, 64)
	return comp + storage - rebate, nil
}

// GetCreatedObjects returns the IDs of created objects.
func (r *TransactionResult) GetCreatedObjects() []string {
	var created []string
	for _, change := range r.ObjectChanges {
		if change.Type == "created" {
			created = append(created, change.ObjectID)
		}
	}
	return created
}

// GetPublishedPackages returns the IDs of published packages.
func (r *TransactionResult) GetPublishedPackages() []string {
	var packages []string
	for _, change := range r.ObjectChanges {
		if change.Type == "published" {
			packages = append(packages, change.PackageID)
		}
	}
	return packages
}
