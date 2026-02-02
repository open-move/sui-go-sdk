package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/utils"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const defaultCoinPageSize = uint32(500)

// ErrInsufficientBalance indicates the requested amount could not be covered by the selected coins.
var ErrInsufficientBalance = errors.New("insufficient balance to satisfy requested amount")

// CoinSelectionOption customises behaviour of SelectCoins helpers.
type CoinSelectionOption func(*coinSelectionConfig)

type coinSelectionConfig struct {
	pageSize   uint32
	excludeIDs map[string]struct{}
	readMask   *fieldmaskpb.FieldMask
}

func newCoinSelectionConfig() *coinSelectionConfig {
	return &coinSelectionConfig{pageSize: defaultCoinPageSize}
}

// WithCoinPageSize bounds the number of objects requested per page when scanning owned coins.
func WithCoinPageSize(size uint32) CoinSelectionOption {
	return func(cfg *coinSelectionConfig) {
		if size == 0 {
			cfg.pageSize = defaultCoinPageSize
			return
		}
		if size > 1000 {
			size = 1000
		}
		cfg.pageSize = size
	}
}

// WithCoinExclusions skips any coins whose IDs match the provided list.
func WithCoinExclusions(ids ...string) CoinSelectionOption {
	return func(cfg *coinSelectionConfig) {
		if len(ids) == 0 {
			return
		}
		if cfg.excludeIDs == nil {
			cfg.excludeIDs = make(map[string]struct{}, len(ids))
		}
		for _, id := range ids {
			normalized := normalizeObjectID(id)
			if normalized == "" {
				continue
			}
			cfg.excludeIDs[normalized] = struct{}{}
		}
	}
}

// WithCoinReadMask overrides the FieldMask used when paginating owned coins.
func WithCoinReadMask(mask *fieldmaskpb.FieldMask) CoinSelectionOption {
	return func(cfg *coinSelectionConfig) {
		if mask == nil {
			cfg.readMask = nil
			return
		}
		cfg.readMask = cloneFieldMask(mask)
	}
}

// SelectCoins returns enough Coin<T> objects owned by owner to meet the requested amount.
func (c *Client) SelectCoins(ctx context.Context, owner string, coinType string, amount uint64, opts ...CoinSelectionOption) ([]*v2.Object, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if owner == "" {
		return nil, errors.New("owner address is empty")
	}
	if coinType == "" {
		return nil, errors.New("coin type is empty")
	}

	cfg := newCoinSelectionConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	req := &v2.ListOwnedObjectsRequest{
		Owner:      utils.Ptr(owner),
		ObjectType: utils.Ptr("0x2::coin::Coin<" + coinType + ">"),
	}
	if cfg.pageSize > 0 {
		size := cfg.pageSize
		req.PageSize = &size
	}
	req.ReadMask = ensureFieldMaskPaths(cfg.readMask,
		"object_id", "version", "digest", "balance", "owner",
	)

	pager, err := c.OwnedObjectsPager(req)
	if err != nil {
		return nil, err
	}

	var (
		total    uint64
		selected []*v2.Object
	)

	for {
		batch, err := pager.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		for _, obj := range batch {
			if obj == nil {
				continue
			}
			if shouldExclude(cfg.excludeIDs, obj.GetObjectId()) {
				continue
			}
			balance := obj.GetBalance()
			if total+balance < total {
				total = ^uint64(0)
			} else {
				total += balance
			}
			selected = append(selected, obj)
			if total >= amount {
				return selected, nil
			}
		}
	}

	return nil, fmt.Errorf("%w: required %d, available %d", ErrInsufficientBalance, amount, total)
}

// SelectUpToNLargestCoins returns up to n Coin<T> objects owned by owner, preserving the iteration order provided by the RPC.
func (c *Client) SelectUpToNLargestCoins(ctx context.Context, owner string, coinType string, n int, opts ...CoinSelectionOption) ([]*v2.Object, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if n <= 0 {
		return nil, nil
	}
	if owner == "" {
		return nil, errors.New("owner address is empty")
	}
	if coinType == "" {
		return nil, errors.New("coin type is empty")
	}

	cfg := newCoinSelectionConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	req := &v2.ListOwnedObjectsRequest{
		Owner:      utils.Ptr(owner),
		ObjectType: utils.Ptr(coinType),
	}
	if cfg.pageSize > 0 {
		size := cfg.pageSize
		req.PageSize = &size
	}
	req.ReadMask = ensureFieldMaskPaths(cfg.readMask,
		"object_id", "version", "digest", "balance", "owner",
	)

	pager, err := c.OwnedObjectsPager(req)
	if err != nil {
		return nil, err
	}

	selected := make([]*v2.Object, 0, n)

	for len(selected) < n {
		batch, err := pager.Next(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		for _, obj := range batch {
			if obj == nil {
				continue
			}
			if shouldExclude(cfg.excludeIDs, obj.GetObjectId()) {
				continue
			}
			selected = append(selected, obj)
			if len(selected) >= n {
				break
			}
		}
	}

	return selected, nil
}

func shouldExclude(exclusions map[string]struct{}, id string) bool {
	if len(exclusions) == 0 {
		return false
	}
	_, ok := exclusions[normalizeObjectID(id)]
	return ok
}

func normalizeObjectID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}
