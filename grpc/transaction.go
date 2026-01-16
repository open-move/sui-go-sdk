package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
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

// ExecuteAndWaitOptions configures ExecuteTransactionAndWait semantics.
type ExecuteAndWaitOptions struct {
	WaitTimeout             time.Duration
	ExecuteCallOptions      []grpc.CallOption
	SubscriptionCallOptions []grpc.CallOption
	SubscriptionRequest     *v2.SubscribeCheckpointsRequest
}

// ExecuteAndWaitRequest describes a signed transaction to submit via ExecuteSignedTransactionAndWait.
type ExecuteAndWaitRequest struct {
	Transaction *v2.Transaction
	Signatures  []*v2.UserSignature
	ReadMask    *fieldmaskpb.FieldMask
}

// CheckpointWaitError wraps the original response alongside an error encountered while waiting for the checkpoint stream.
type CheckpointWaitError struct {
	Response *v2.ExecuteTransactionResponse
	Err      error
}

func (e *CheckpointWaitError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return "checkpoint wait error"
	}
	return fmt.Sprintf("checkpoint wait error: %v", e.Err)
}

func (e *CheckpointWaitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (o *ExecuteAndWaitOptions) clone() *ExecuteAndWaitOptions {
	if o == nil {
		return &ExecuteAndWaitOptions{}
	}
	clone := &ExecuteAndWaitOptions{
		WaitTimeout:             o.WaitTimeout,
		ExecuteCallOptions:      append([]grpc.CallOption(nil), o.ExecuteCallOptions...),
		SubscriptionCallOptions: append([]grpc.CallOption(nil), o.SubscriptionCallOptions...),
	}
	if o.SubscriptionRequest != nil {
		cloned := proto.Clone(o.SubscriptionRequest)
		if cloned != nil {
			clone.SubscriptionRequest = cloned.(*v2.SubscribeCheckpointsRequest)
		}
	}
	return clone
}

type checkpointResult struct {
	checkpoint *v2.Checkpoint
	err        error
}

// ExecuteSignedTransactionAndWait submits a signed transaction and resolves when it appears in a checkpoint.
func (c *GRPCClient) ExecuteSignedTransactionAndWait(ctx context.Context, req *ExecuteAndWaitRequest, options *ExecuteAndWaitOptions) (*v2.ExecutedTransaction, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	built, err := buildExecuteTransactionRequest(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.ExecuteTransactionAndWait(ctx, built, options)
	if err != nil {
		return nil, err
	}
	tx := resp.GetTransaction()
	if tx == nil {
		return nil, &CheckpointWaitError{Response: resp, Err: ErrResponseMissingTransaction}
	}

	return tx, nil
}

// ExecuteTransactionAndWait submits an ExecuteTransactionRequest and blocks until the transaction is observed in a checkpoint or an error occurs.
func (c *GRPCClient) ExecuteTransactionAndWait(ctx context.Context, request *v2.ExecuteTransactionRequest, options *ExecuteAndWaitOptions) (*v2.ExecuteTransactionResponse, error) {
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
	execReq.ReadMask = ensureFieldMaskPaths(execReq.GetReadMask(), "transaction.digest", "effects.status", "checkpoint")

	cfg := options.clone()
	subCtx := ctx
	cancel := func() {}
	if cfg.WaitTimeout > 0 {
		subCtx, cancel = context.WithTimeout(ctx, cfg.WaitTimeout)
	} else {
		subCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	subReq := cfg.SubscriptionRequest
	if subReq == nil {
		subReq = &v2.SubscribeCheckpointsRequest{}
	} else {
		subReq = proto.Clone(subReq).(*v2.SubscribeCheckpointsRequest)
	}
	subReq.ReadMask = ensureFieldMaskPaths(subReq.GetReadMask(), "checkpoint.transactions.digest", "checkpoint.sequence_number")

	stream, err := c.SubscriptionClient().SubscribeCheckpoints(subCtx, subReq, cfg.SubscriptionCallOptions...)
	if err != nil {
		return nil, fmt.Errorf("subscribe checkpoints: %w", err)
	}

	results := make(chan checkpointResult, 1)
	go func() {
		defer close(results)
		for {
			msg, recvErr := stream.Recv()
			if recvErr != nil {
				if errors.Is(recvErr, context.Canceled) && subCtx.Err() != nil {
					return
				}
				results <- checkpointResult{err: recvErr}
				return
			}
			if msg == nil {
				continue
			}
			results <- checkpointResult{checkpoint: msg.GetCheckpoint()}
		}
	}()

	response, err := c.TransactionExecutionClient().ExecuteTransaction(ctx, execReq, cfg.ExecuteCallOptions...)
	if err != nil {
		cancel()
		return nil, err
	}

	tx := response.GetTransaction()
	if tx == nil {
		cancel()
		return response, &CheckpointWaitError{Response: response, Err: ErrResponseMissingTransaction}
	}
	digest := tx.GetDigest()
	if digest == "" {
		cancel()
		return response, &CheckpointWaitError{Response: response, Err: ErrResponseMissingDigest}
	}

	for {
		select {
		case <-ctx.Done():
			cancel()
			return response, &CheckpointWaitError{Response: response, Err: ctx.Err()}
		case res, ok := <-results:
			if !ok {
				cancel()
				return response, &CheckpointWaitError{Response: response, Err: io.EOF}
			}
			if res.err != nil {
				cancel()
				if errors.Is(res.err, io.EOF) {
					return response, &CheckpointWaitError{Response: response, Err: io.EOF}
				}
				return response, &CheckpointWaitError{Response: response, Err: res.err}
			}
			cp := res.checkpoint
			if cp == nil {
				continue
			}
			for _, t := range cp.GetTransactions() {
				if t.GetDigest() == digest {
					cancel()
					return response, nil
				}
			}
		}
	}
}

// SimulateTransactionOptions customises behaviour of SimulateTransaction.
type SimulateTransactionOptions struct {
	ReadMask       *fieldmaskpb.FieldMask
	Checks         *v2.SimulateTransactionRequest_TransactionChecks
	DoGasSelection *bool
}

// SimulateTransaction executes the SimulateTransaction RPC for the provided transaction.
func (c *GRPCClient) SimulateTransaction(ctx context.Context, tx *v2.Transaction, options *SimulateTransactionOptions, opts ...grpc.CallOption) (*v2.SimulateTransactionResponse, error) {
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

func buildExecuteTransactionRequest(req *ExecuteAndWaitRequest) (*v2.ExecuteTransactionRequest, error) {
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
