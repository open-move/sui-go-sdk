package transaction

import (
	"context"
	"errors"
	"reflect"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

type stubResolver struct {
	objects map[string]ObjectMetadata
	move    *MoveFunction
}

func (s stubResolver) ResolveObjects(_ context.Context, objectIDs []string) ([]ObjectMetadata, error) {
	refs := make([]ObjectMetadata, len(objectIDs))
	for i, id := range objectIDs {
		normalized, err := utils.NormalizeAddress(id)
		if err != nil {
			return nil, err
		}
		meta, ok := s.objects[normalized]
		if !ok {
			return nil, errors.New("missing object metadata")
		}
		refs[i] = meta
	}
	return refs, nil
}

func (s stubResolver) ResolveMoveFunction(_ context.Context, _, _, _ string) (*MoveFunction, error) {
	if s.move == nil {
		return nil, errors.New("missing move function")
	}
	return s.move, nil
}

type stubGasResolver struct {
	calls       []string
	price       uint64
	budget      uint64
	payment     []types.ObjectRef
	budgetInput GasBudgetInput
	paymentArgs struct {
		owner  types.Address
		budget uint64
	}
}

func (s *stubGasResolver) ResolveGasPrice(_ context.Context) (uint64, error) {
	s.calls = append(s.calls, "price")
	return s.price, nil
}

func (s *stubGasResolver) ResolveGasBudget(_ context.Context, input GasBudgetInput) (uint64, error) {
	s.calls = append(s.calls, "budget")
	s.budgetInput = input
	return s.budget, nil
}

func (s *stubGasResolver) ResolveGasPayment(_ context.Context, owner types.Address, budget uint64) ([]types.ObjectRef, error) {
	s.calls = append(s.calls, "payment")
	s.paymentArgs.owner = owner
	s.paymentArgs.budget = budget
	return s.payment, nil
}

func TestResolveInputsSharedMutable(t *testing.T) {
	sharedVersion := uint64(1)
	digest := types.Digest(make([]byte, 32))
	for i := range digest {
		digest[i] = 1
	}

	sharedID := mustNormalize(t, "0x1")
	resolver := stubResolver{
		objects: map[string]ObjectMetadata{
			sharedID: {
				ID:           mustAddress(t, "0x1"),
				Version:      10,
				Digest:       digest,
				OwnerKind:    OwnerShared,
				OwnerVersion: &sharedVersion,
			},
		},
		move: &MoveFunction{
			Parameters: []MoveParameter{{Reference: ReferenceMutable, TypeName: "0x2::foo::Thing"}},
		},
	}

	tx := New()
	tx.MoveCall(MoveCall{
		Target:    "0x2::foo::bar",
		Arguments: []Argument{tx.Object("0x1")},
	})

	result, err := tx.Build(context.Background(), BuildOptions{Resolver: resolver})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if len(result.ResolvedInputArgs) != 1 {
		t.Fatalf("expected 1 resolved input")
	}
	arg := result.ResolvedInputArgs[0]
	if arg.Object == nil || arg.Object.SharedObject == nil {
		t.Fatalf("expected shared object input")
	}
	if !arg.Object.SharedObject.Mutable {
		t.Fatalf("expected shared object to be mutable")
	}
	if arg.Object.SharedObject.InitialSharedVersion != sharedVersion {
		t.Fatalf("unexpected shared version")
	}
}

func TestResolveInputsReceiving(t *testing.T) {
	digest := types.Digest(make([]byte, 32))
	for i := range digest {
		digest[i] = 2
	}

	objectID := mustNormalize(t, "0x1")
	resolver := stubResolver{
		objects: map[string]ObjectMetadata{
			objectID: {
				ID:        mustAddress(t, "0x1"),
				Version:   7,
				Digest:    digest,
				OwnerKind: OwnerAddress,
			},
		},
		move: &MoveFunction{
			Parameters: []MoveParameter{{Reference: ReferenceImmutable, TypeName: "0x2::transfer::Receiving"}},
		},
	}

	tx := New()
	tx.MoveCall(MoveCall{
		Target:    "0x2::foo::bar",
		Arguments: []Argument{tx.Object("0x1")},
	})

	result, err := tx.Build(context.Background(), BuildOptions{Resolver: resolver})
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	arg := result.ResolvedInputArgs[0]
	if arg.Object == nil || arg.Object.Receiving == nil {
		t.Fatalf("expected receiving object input")
	}
	if !reflect.DeepEqual(arg.Object.Receiving.Digest, digest) {
		t.Fatalf("unexpected receiving digest")
	}
}

func TestResolveGasOrder(t *testing.T) {
	digest := types.Digest(make([]byte, 32))
	for i := range digest {
		digest[i] = 3
	}

	payment := types.ObjectRef{
		ObjectID: mustAddress(t, "0x2"),
		Version:  12,
		Digest:   digest,
	}
	resolver := &stubGasResolver{
		price:   7,
		budget:  42,
		payment: []types.ObjectRef{payment},
	}

	tx := New()
	tx.SetSender("0x1")
	tx.MoveCall(MoveCall{Target: "0x2::foo::bar"})

	result, err := tx.Build(context.Background(), BuildOptions{GasResolver: resolver})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(result.TransactionBytes) == 0 {
		t.Fatalf("expected transaction bytes")
	}

	if !reflect.DeepEqual(resolver.calls, []string{"price", "budget", "payment"}) {
		t.Fatalf("unexpected call order: %v", resolver.calls)
	}
	if resolver.paymentArgs.budget != resolver.budget {
		t.Fatalf("payment budget mismatch")
	}

	var data TransactionData
	if _, err := bcs.UnmarshalInto(result.TransactionBytes, &data); err != nil {
		t.Fatalf("unmarshal transaction data: %v", err)
	}
	if data.V1 == nil {
		t.Fatalf("expected v1 transaction data")
	}
	if data.V1.GasData.Price != resolver.price || data.V1.GasData.Budget != resolver.budget {
		t.Fatalf("unexpected gas data")
	}
	if len(data.V1.GasData.Payment) != 1 {
		t.Fatalf("expected gas payment")
	}
	if !reflect.DeepEqual(data.V1.GasData.Payment[0], payment) {
		t.Fatalf("unexpected gas payment")
	}
}

func mustNormalize(t *testing.T, value string) string {
	t.Helper()
	normalized, err := utils.NormalizeAddress(value)
	if err != nil {
		t.Fatalf("normalize address %q: %v", value, err)
	}
	return normalized
}
