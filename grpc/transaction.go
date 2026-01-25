package grpc

import (
	"context"
	"errors"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/transaction"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var (
	// ErrMissingTransaction indicates the request lacked the required transaction payload.
	ErrMissingTransaction = errors.New("execute transaction request missing transaction")
	// ErrResponseMissingTransaction indicates the RPC response omitted the executed transaction payload.
	ErrResponseMissingTransaction = errors.New("execute transaction response missing transaction data")
	// ErrResponseMissingDigest indicates the executed transaction did not include its digest.
	ErrResponseMissingDigest = errors.New("execute transaction response missing digest")
)

// ExecuteOptions configures ExecuteTransaction semantics.
type ExecuteOptions struct {
	ExecuteCallOptions []grpc.CallOption
}

// ExecuteRequest describes a signed transaction to submit via ExecuteSignedTransaction.
type ExecuteRequest struct {
	Transaction *v2.Transaction
	Signatures  []*v2.UserSignature
	ReadMask    *fieldmaskpb.FieldMask
}

func (o *ExecuteOptions) clone() *ExecuteOptions {
	if o == nil {
		return &ExecuteOptions{}
	}
	return &ExecuteOptions{
		ExecuteCallOptions: append([]grpc.CallOption(nil), o.ExecuteCallOptions...),
	}
}

// ExecuteSignedTransaction submits a signed transaction and returns its immediate response.
func (c *Client) ExecuteSignedTransaction(ctx context.Context, req *ExecuteRequest, options *ExecuteOptions) (*v2.ExecutedTransaction, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	built, err := buildExecuteTransactionRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeTransaction(ctx, built, options)
	if err != nil {
		return nil, err
	}
	tx := resp.GetTransaction()
	if tx == nil {
		return nil, ErrResponseMissingTransaction
	}
	if tx.GetDigest() == "" {
		return nil, ErrResponseMissingDigest
	}

	return tx, nil
}

// SignAndExecute resolves, signs, and submits the provided transaction builder.
func (c *Client) SignAndExecuteTransaction(ctx context.Context, tx *transaction.Builder, signer transaction.TransactionSigner, options *ExecuteOptions) (*v2.ExecutedTransaction, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if tx == nil {
		return nil, errors.New("nil transaction builder")
	}
	if signer == nil {
		return nil, errors.New("nil signer")
	}

	if !tx.HasSender() {
		addr, err := signer.SuiAddress()
		if err != nil {
			return nil, err
		}
		tx.SetSender(addr)
	}

	if err := tx.Err(); err != nil {
		return nil, err
	}

	resolver := NewResolver(c)
	result, err := tx.Build(ctx, transaction.BuildOptions{Resolver: resolver, GasResolver: resolver})
	if err != nil {
		return nil, err
	}
	if result.Transaction == nil || len(result.TransactionBytes) == 0 {
		return nil, errors.New("built transaction missing data")
	}

	signature, err := signer.SignTransaction(result.TransactionBytes)
	if err != nil {
		return nil, err
	}
	userSig, err := transaction.UserSignatureFromSerialized(signature)
	if err != nil {
		return nil, err
	}

	return c.ExecuteSignedTransaction(ctx, &ExecuteRequest{
		Transaction: result.Transaction,
		Signatures:  []*v2.UserSignature{userSig},
	}, options)
}

// ExecuteTransaction submits an ExecuteTransactionRequest and returns its immediate response.
func (c *Client) executeTransaction(ctx context.Context, request *v2.ExecuteTransactionRequest, options *ExecuteOptions) (*v2.ExecuteTransactionResponse, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if request == nil {
		return nil, errors.New("nil request")
	}

	execReqProto := proto.Clone(request)
	if execReqProto == nil {
		return nil, errors.New("failed to clone execute transaction request")
	}
	execReq := execReqProto.(*v2.ExecuteTransactionRequest)
	if execReq.GetTransaction() == nil {
		return nil, ErrMissingTransaction
	}
	execReq.ReadMask = ensureFieldMaskPaths(execReq.GetReadMask(), "digest", "effects.status", "checkpoint")

	cfg := options.clone()
	return c.TransactionExecutionClient().ExecuteTransaction(ctx, execReq, cfg.ExecuteCallOptions...)
}

// SimulateTransactionOptions customises behaviour of SimulateTransaction.
type SimulateTransactionOptions struct {
	ReadMask       *fieldmaskpb.FieldMask
	Checks         *v2.SimulateTransactionRequest_TransactionChecks
	DoGasSelection *bool
}

// SimulateTransaction executes the SimulateTransaction RPC for the provided transaction.
func (c *Client) SimulateTransaction(ctx context.Context, tx *v2.Transaction, options *SimulateTransactionOptions, opts ...grpc.CallOption) (*v2.SimulateTransactionResponse, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if tx == nil {
		return nil, errors.New("nil transaction")
	}

	req, err := buildSimulateTransactionRequest(tx, options)
	if err != nil {
		return nil, err
	}

	return c.TransactionExecutionClient().SimulateTransaction(ctx, req, opts...)
}

func buildExecuteTransactionRequest(req *ExecuteRequest) (*v2.ExecuteTransactionRequest, error) {
	if req == nil {
		return nil, errors.New("nil execute request")
	}
	if req.Transaction == nil {
		return nil, ErrMissingTransaction
	}

	cloned := proto.Clone(req.Transaction)
	if cloned == nil {
		return nil, errors.New("failed to clone transaction")
	}
	clonedTx := cloned.(*v2.Transaction)
	clonedSigs := cloneUserSignatures(req.Signatures)
	readMask := ensureFieldMaskPaths(req.ReadMask)

	return &v2.ExecuteTransactionRequest{
		Transaction: clonedTx,
		Signatures:  clonedSigs,
		ReadMask:    readMask,
	}, nil
}

func buildSimulateTransactionRequest(tx *v2.Transaction, options *SimulateTransactionOptions) (*v2.SimulateTransactionRequest, error) {
	clonedTx := proto.Clone(tx)
	if clonedTx == nil {
		return nil, errors.New("failed to clone transaction")
	}

	req := &v2.SimulateTransactionRequest{
		Transaction: clonedTx.(*v2.Transaction),
	}

	if options != nil {
		req.ReadMask = ensureFieldMaskPaths(options.ReadMask)
		if options.Checks != nil {
			value := *options.Checks
			req.Checks = value.Enum()
		}
		if options.DoGasSelection != nil {
			val := *options.DoGasSelection
			req.DoGasSelection = &val
		}
	}

	return req, nil
}

func cloneUserSignatures(sig []*v2.UserSignature) []*v2.UserSignature {
	if len(sig) == 0 {
		return nil
	}
	cloned := make([]*v2.UserSignature, len(sig))
	for i, s := range sig {
		if s == nil {
			continue
		}
		if clonedSig := proto.Clone(s); clonedSig != nil {
			cloned[i] = clonedSig.(*v2.UserSignature)
		}
	}
	return cloned
}
