package graphql

import (
	"context"
	"fmt"
)

// =============================================================================
// Address & Balance Queries (equivalent to Blockvision's coin/balance methods)
// =============================================================================

// GetAllBalances returns all coin balances for an address.
// Equivalent to Blockvision's SuiXGetAllBalance.
func (c *Client) GetAllBalances(ctx context.Context, owner SuiAddress) ([]Balance, error) {
	query := `
		query GetAllBalances($address: SuiAddress!) {
			address(address: $address) {
				balances {
					nodes {
						coinType {
							repr
						}
						totalBalance
					}
				}
			}
		}
	`

	var result struct {
		Address *struct {
			Balances *Connection[Balance] `json:"balances"`
		} `json:"address"`
	}

	err := c.Execute(ctx, query, map[string]any{"address": owner}, &result)
	if err != nil {
		return nil, err
	}

	if result.Address == nil || result.Address.Balances == nil {
		return []Balance{}, nil
	}

	return result.Address.Balances.Nodes, nil
}

// GetBalance returns the balance of a specific coin type for an address.
// coinType is required (e.g., "0x2::sui::SUI").
func (c *Client) GetBalance(ctx context.Context, owner SuiAddress, coinType string) (*Balance, error) {
	query := `
		query GetBalance($address: SuiAddress!, $coinType: String!) {
			address(address: $address) {
				balance(coinType: $coinType) {
					coinType {
						repr
					}
					totalBalance
				}
			}
		}
	`

	vars := map[string]any{
		"address":  owner,
		"coinType": coinType,
	}

	var result struct {
		Address *struct {
			Balance *Balance `json:"balance"`
		} `json:"address"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Address == nil {
		return nil, nil
	}

	return result.Address.Balance, nil
}

// GetCoins returns coins of a specific type owned by an address.
// Uses objects query with type filter to get coin objects.
func (c *Client) GetCoins(ctx context.Context, owner SuiAddress, coinType *string, pagination *PaginationArgs) (*Connection[Coin], error) {
	// Default to SUI if no coin type specified
	cType := "0x2::coin::Coin<0x2::sui::SUI>"
	if coinType != nil {
		cType = fmt.Sprintf("0x2::coin::Coin<%s>", *coinType)
	}

	query := `
		query GetCoins($address: SuiAddress!, $type: String, $first: Int, $after: String, $last: Int, $before: String) {
			address(address: $address) {
				objects(filter: {type: $type}, first: $first, after: $after, last: $last, before: $before) {
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
					nodes {
						address
						version
						digest
						contents {
							type {
								repr
							}
							bcs
							json
						}
					}
				}
			}
		}
	`

	vars := map[string]any{
		"address": owner,
		"type":    cType,
	}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Address *struct {
			Objects *struct {
				PageInfo PageInfo `json:"pageInfo"`
				Nodes    []struct {
					Address  SuiAddress `json:"address"`
					Version  UInt53     `json:"version"`
					Digest   string     `json:"digest"`
					Contents *MoveValue `json:"contents"`
				} `json:"nodes"`
			} `json:"objects"`
		} `json:"address"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Address == nil || result.Address.Objects == nil {
		return nil, nil
	}

	// Convert to Coin type
	coins := make([]Coin, 0, len(result.Address.Objects.Nodes))
	for _, obj := range result.Address.Objects.Nodes {
		coin := Coin{
			Address:  obj.Address,
			Version:  obj.Version,
			Digest:   obj.Digest,
			Contents: obj.Contents,
		}
		coins = append(coins, coin)
	}

	return &Connection[Coin]{
		PageInfo: result.Address.Objects.PageInfo,
		Nodes:    coins,
	}, nil
}

// GetCoinMetadata returns metadata for a coin type.
// Equivalent to Blockvision's SuiXGetCoinMetadata.
func (c *Client) GetCoinMetadata(ctx context.Context, coinType string) (*CoinMetadata, error) {
	query := `
		query GetCoinMetadata($coinType: String!) {
			coinMetadata(coinType: $coinType) {
				address
				version
				digest
				decimals
				name
				symbol
				description
				iconUrl
				supply
			}
		}
	`

	var result struct {
		CoinMetadata *CoinMetadata `json:"coinMetadata"`
	}

	err := c.Execute(ctx, query, map[string]any{"coinType": coinType}, &result)
	if err != nil {
		return nil, err
	}

	return result.CoinMetadata, nil
}

// GetTotalSupply returns the total supply of a coin type.
// Equivalent to Blockvision's SuiXGetTotalSupply.
func (c *Client) GetTotalSupply(ctx context.Context, coinType string) (*BigInt, error) {
	metadata, err := c.GetCoinMetadata(ctx, coinType)
	if err != nil {
		return nil, err
	}
	if metadata == nil {
		return nil, nil
	}
	return metadata.Supply, nil
}

// =============================================================================
// Object Queries (equivalent to Blockvision's object methods)
// =============================================================================

// GetObject returns details for a specific object.
// Equivalent to Blockvision's SuiGetObject.
func (c *Client) GetObject(ctx context.Context, objectID SuiAddress, options *ObjectDataOptions) (*Object, error) {
	query := c.buildObjectQuery(options)

	var result struct {
		Object *Object `json:"object"`
	}

	err := c.Execute(ctx, query, map[string]any{"objectId": objectID}, &result)
	if err != nil {
		return nil, err
	}

	return result.Object, nil
}

// ObjectDataOptions controls what data is returned for objects.
type ObjectDataOptions struct {
	ShowType                bool
	ShowContent             bool
	ShowBcs                 bool
	ShowOwner               bool
	ShowPreviousTransaction bool
	ShowStorageRebate       bool
	ShowDisplay             bool
}

func (c *Client) buildObjectQuery(options *ObjectDataOptions) string {
	if options == nil {
		options = &ObjectDataOptions{
			ShowType:                true,
			ShowContent:             true,
			ShowOwner:               true,
			ShowPreviousTransaction: true,
			ShowStorageRebate:       true,
			ShowDisplay:             false, // Display often not available
		}
	}

	fields := "address version digest"

	if options.ShowStorageRebate {
		fields += " storageRebate"
	}
	if options.ShowOwner {
		fields += ` owner {
			__typename
			... on AddressOwner { address { address } }
			... on ObjectOwner { address { address } }
			... on Shared { initialSharedVersion }
		}`
	}
	if options.ShowPreviousTransaction {
		fields += " previousTransaction { digest }"
	}
	if options.ShowBcs {
		fields += " objectBcs"
	}
	if options.ShowType || options.ShowContent {
		fields += ` asMoveObject {
			address version digest hasPublicTransfer
			contents { type { repr } bcs json }
		}`
	}

	return fmt.Sprintf(`
		query GetObject($objectId: SuiAddress!) {
			object(address: $objectId) {
				%s
			}
		}
	`, fields)
}

// GetMultipleObjects returns details for multiple objects.
// Equivalent to Blockvision's SuiMultiGetObjects.
func (c *Client) GetMultipleObjects(ctx context.Context, objectIDs []SuiAddress, options *ObjectDataOptions) ([]Object, error) {
	query := `
		query MultiGetObjects($keys: [ObjectKey!]!) {
			multiGetObjects(keys: $keys) {
				address
				version
				digest
				storageRebate
				owner {
					__typename
					... on AddressOwner { address { address } }
					... on ObjectOwner { address { address } }
					... on Shared { initialSharedVersion }
				}
				previousTransaction { digest }
				asMoveObject {
					address version digest hasPublicTransfer
					contents { type { repr } bcs json }
				}
			}
		}
	`

	// Build object keys
	keys := make([]map[string]any, len(objectIDs))
	for i, id := range objectIDs {
		keys[i] = map[string]any{"address": id}
	}

	var result struct {
		MultiGetObjects []Object `json:"multiGetObjects"`
	}

	err := c.Execute(ctx, query, map[string]any{"keys": keys}, &result)
	if err != nil {
		return nil, err
	}

	if result.MultiGetObjects == nil {
		return []Object{}, nil
	}

	return result.MultiGetObjects, nil
}

// GetOwnedObjects returns objects owned by an address.
// Note: Returns MoveObject connection, not Object connection.
func (c *Client) GetOwnedObjects(ctx context.Context, owner SuiAddress, filter *ObjectFilter, pagination *PaginationArgs) (*Connection[Object], error) {
	query := `
		query GetOwnedObjects($address: SuiAddress!, $filter: ObjectFilter, $first: Int, $after: String, $last: Int, $before: String) {
			address(address: $address) {
				objects(filter: $filter, first: $first, after: $after, last: $last, before: $before) {
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
					nodes {
						address
						version
						digest
						owner {
							__typename
							... on AddressOwner { address { address } }
							... on ObjectOwner { address { address } }
							... on Shared { initialSharedVersion }
						}
						hasPublicTransfer
						contents { type { repr } bcs json }
					}
				}
			}
		}
	`

	vars := map[string]any{"address": owner}
	if filter != nil {
		vars["filter"] = filter
	}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Address *struct {
			Objects *Connection[Object] `json:"objects"`
		} `json:"address"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Address == nil {
		return nil, nil
	}

	return result.Address.Objects, nil
}

// GetDynamicFields returns dynamic fields for an object.
func (c *Client) GetDynamicFields(ctx context.Context, parentID SuiAddress, pagination *PaginationArgs) (*Connection[DynamicField], error) {
	query := `
		query GetDynamicFields($parentId: SuiAddress!, $first: Int, $after: String, $last: Int, $before: String) {
			object(address: $parentId) {
				dynamicFields(first: $first, after: $after, last: $last, before: $before) {
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
					nodes {
						name {
							type { repr }
							bcs
							json
						}
						value {
							... on MoveValue {
								type { repr }
								bcs
								json
							}
							... on MoveObject {
								address version digest hasPublicTransfer
								type { repr }
								contents { type { repr } bcs json }
							}
						}
					}
				}
			}
		}
	`

	vars := map[string]any{"parentId": parentID}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Object *struct {
			DynamicFields *Connection[DynamicField] `json:"dynamicFields"`
		} `json:"object"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Object == nil {
		return nil, nil
	}

	return result.Object.DynamicFields, nil
}

// GetDynamicFieldObject returns a specific dynamic field object.
func (c *Client) GetDynamicFieldObject(ctx context.Context, parentID SuiAddress, name DynamicFieldName) (*DynamicField, error) {
	query := `
		query GetDynamicFieldObject($parentId: SuiAddress!, $name: DynamicFieldName!) {
			object(address: $parentId) {
				dynamicField(name: $name) {
					name {
						type { repr }
						bcs
						json
					}
					value {
						... on MoveValue {
							type { repr }
							bcs
							json
						}
						... on MoveObject {
							address version digest hasPublicTransfer
							type { repr }
							contents { type { repr } bcs json }
						}
					}
				}
			}
		}
	`

	vars := map[string]any{
		"parentId": parentID,
		"name":     name,
	}

	var result struct {
		Object *struct {
			DynamicField *DynamicField `json:"dynamicField"`
		} `json:"object"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Object == nil {
		return nil, nil
	}

	return result.Object.DynamicField, nil
}

// =============================================================================
// Transaction Queries (equivalent to Blockvision's transaction methods)
// =============================================================================

// TransactionBlockOptions controls what data is returned for transactions.
type TransactionBlockOptions struct {
	ShowInput          bool
	ShowRawInput       bool
	ShowEffects        bool
	ShowEvents         bool
	ShowObjectChanges  bool
	ShowBalanceChanges bool
}

// GetTransactionBlock returns details for a transaction.
// Equivalent to Blockvision's SuiGetTransactionBlock.
func (c *Client) GetTransactionBlock(ctx context.Context, digest string, options *TransactionBlockOptions) (*Transaction, error) {
	query := c.buildTransactionQuery(options)

	var result struct {
		Transaction *Transaction `json:"transaction"`
	}

	err := c.Execute(ctx, query, map[string]any{"digest": digest}, &result)
	if err != nil {
		return nil, err
	}

	return result.Transaction, nil
}

func (c *Client) buildTransactionQuery(options *TransactionBlockOptions) string {
	if options == nil {
		options = &TransactionBlockOptions{
			ShowInput:          true,
			ShowEffects:        true,
			ShowEvents:         true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
		}
	}

	fields := "digest"

	if options.ShowInput {
		fields += `
			sender { address }
			gasInput {
				gasSponsor { address }
				gasPrice
				gasBudget
			}
			expiration { epochId }
			signatures { signatureBytes }
			kind {
				__typename
				... on ProgrammableTransaction {
					inputs {
						nodes {
							__typename
							... on Pure {
								bytes
							}
							... on OwnedOrImmutable {
								object {
									address
									version
									digest
								}
							}
							... on SharedInput {
								address
								initialSharedVersion
								mutable
							}
							... on Receiving {
								object {
									address
									version
									digest
								}
							}
						}
					}
					commands {
						nodes {
							__typename
						}
					}
				}
			}
		`
	}

	if options.ShowRawInput {
		fields += " transactionBcs"
	}

	if options.ShowEffects {
		fields += `
			effects {
				digest
				status
				executionError { message }
				lamportVersion
				gasEffects {
					gasSummary {
						computationCost
						storageCost
						storageRebate
						nonRefundableStorageFee
					}
				}
				epoch { epochId }
				checkpoint { sequenceNumber }
				timestamp
			}
		`
	}

	if options.ShowObjectChanges {
		fields += `
			effects {
				objectChanges {
					nodes {
						address
						idCreated
						idDeleted
						inputState { address version digest }
						outputState { address version digest }
					}
				}
			}
		`
	}

	if options.ShowBalanceChanges {
		fields += `
			effects {
				balanceChanges {
					nodes {
						owner {
							address
						}
						coinType { repr }
						amount
					}
				}
			}
		`
	}

	return fmt.Sprintf(`
		query GetTransaction($digest: String!) {
			transaction(digest: $digest) {
				%s
			}
		}
	`, fields)
}

// GetMultipleTransactionBlocks returns details for multiple transactions.
// Equivalent to Blockvision's SuiMultiGetTransactionBlocks.
func (c *Client) GetMultipleTransactionBlocks(ctx context.Context, digests []string, options *TransactionBlockOptions) ([]Transaction, error) {
	// Query each transaction individually since there's no batch query by digest
	transactions := make([]Transaction, 0, len(digests))
	for _, digest := range digests {
		tx, err := c.GetTransactionBlock(ctx, digest, options)
		if err != nil {
			return nil, err
		}
		if tx != nil {
			transactions = append(transactions, *tx)
		}
	}
	return transactions, nil
}

// QueryTransactionBlocks queries transactions with filters.
// Equivalent to Blockvision's SuiXQueryTransactionBlocks.
func (c *Client) QueryTransactionBlocks(ctx context.Context, filter *TransactionFilter, pagination *PaginationArgs) (*Connection[Transaction], error) {
	query := `
		query QueryTransactions($filter: TransactionFilter, $first: Int, $after: String, $last: Int, $before: String) {
			transactions(filter: $filter, first: $first, after: $after, last: $last, before: $before) {
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
				nodes {
					digest
					kind {
						__typename
						... on ProgrammableTransaction {
							inputs {
								nodes {
									__typename
									... on Pure {
										bytes
									}
									... on OwnedOrImmutable {
										object {
											address
											version
											digest
										}
									}
									... on SharedInput {
										address
										initialSharedVersion
										mutable
									}
									... on Receiving {
										object {
											address
											version
											digest
										}
									}
								}
							}
							commands {
								nodes {
									__typename
								}
							}
						}
					}
					signatures {
						signatureBytes
					}
					expiration { epochId }
					transactionBcs
					sender { address }
					gasInput {
						gasSponsor { address }
						gasPrice
						gasBudget
						gasPayment {
							nodes {
								address
								version
								digest
							}
						}
					}
					effects {
						digest
						status
						executionError { message }
						lamportVersion
						gasEffects {
							gasSummary {
								computationCost
								storageCost
								storageRebate
								nonRefundableStorageFee
							}
						}
						epoch { epochId }
						checkpoint { sequenceNumber }
						timestamp
					}
				}
			}
		}
	`

	vars := make(map[string]any)
	if filter != nil {
		vars["filter"] = filter
	}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Transactions *Connection[Transaction] `json:"transactions"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.Transactions, nil
}

// GetTotalTransactionBlocks returns the total number of transactions.
// Equivalent to Blockvision's SuiGetTotalTransactionBlocks.
func (c *Client) GetTotalTransactionBlocks(ctx context.Context) (*UInt53, error) {
	query := `
		query GetTotalTransactionBlocks {
			checkpoint {
				networkTotalTransactions
			}
		}
	`

	var result struct {
		Checkpoint *struct {
			NetworkTotalTransactions *UInt53 `json:"networkTotalTransactions"`
		} `json:"checkpoint"`
	}

	err := c.Execute(ctx, query, nil, &result)
	if err != nil {
		return nil, err
	}

	if result.Checkpoint == nil {
		return nil, nil
	}
	return result.Checkpoint.NetworkTotalTransactions, nil
}

// =============================================================================
// Event Queries (equivalent to Blockvision's event methods)
// =============================================================================

// GetEvents returns events for a transaction.
// Equivalent to Blockvision's SuiGetEvents.
func (c *Client) GetEvents(ctx context.Context, digest string) ([]Event, error) {
	query := `
		query GetEvents($digest: String!) {
			transaction(digest: $digest) {
				effects {
					events {
						nodes {
							transactionModule { name package { address } }
							sender { address }
							timestamp
							contents { type { repr } bcs json }
							eventBcs
						}
					}
				}
			}
		}
	`

	var result struct {
		Transaction *struct {
			Effects *struct {
				Events *Connection[Event] `json:"events"`
			} `json:"effects"`
		} `json:"transaction"`
	}

	err := c.Execute(ctx, query, map[string]any{"digest": digest}, &result)
	if err != nil {
		return nil, err
	}

	if result.Transaction == nil || result.Transaction.Effects == nil || result.Transaction.Effects.Events == nil {
		return []Event{}, nil
	}

	return result.Transaction.Effects.Events.Nodes, nil
}

// QueryEvents queries events with filters.
// Equivalent to Blockvision's SuiXQueryEvents.
func (c *Client) QueryEvents(ctx context.Context, filter *EventFilter, pagination *PaginationArgs) (*Connection[Event], error) {
	query := `
		query QueryEvents($filter: EventFilter, $first: Int, $after: String, $last: Int, $before: String) {
			events(filter: $filter, first: $first, after: $after, last: $last, before: $before) {
				pageInfo {
					hasNextPage
					hasPreviousPage
					startCursor
					endCursor
				}
				nodes {
					transactionModule { name package { address } }
					sender { address }
					timestamp
					contents { type { repr } bcs json }
					eventBcs
				}
			}
		}
	`

	vars := make(map[string]any)
	if filter != nil {
		vars["filter"] = filter
	}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Events *Connection[Event] `json:"events"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.Events, nil
}

// =============================================================================
// Protocol & System Queries
// =============================================================================

// GetProtocolConfig returns the protocol configuration.
// Use protocolVersion parameter to get config for a specific version, or nil for latest.
func (c *Client) GetProtocolConfig(ctx context.Context, protocolVersion *UInt53) (*ProtocolConfigs, error) {
	query := `
		query GetProtocolConfig($version: UInt53) {
			protocolConfigs(version: $version) {
				protocolVersion
				featureFlags { key value }
				configs { key value }
			}
		}
	`

	vars := make(map[string]any)
	if protocolVersion != nil {
		vars["version"] = *protocolVersion
	}

	var result struct {
		ProtocolConfigs *ProtocolConfigs `json:"protocolConfigs"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.ProtocolConfigs, nil
}

// GetChainIdentifier returns the chain identifier.
// Equivalent to Blockvision's SuiGetChainIdentifier.
func (c *Client) GetChainIdentifier(ctx context.Context) (string, error) {
	query := `
		query GetChainIdentifier {
			chainIdentifier
		}
	`

	var result struct {
		ChainIdentifier string `json:"chainIdentifier"`
	}

	err := c.Execute(ctx, query, nil, &result)
	if err != nil {
		return "", err
	}

	return result.ChainIdentifier, nil
}

// GetReferenceGasPrice returns the reference gas price for an epoch.
// Equivalent to Blockvision's SuiXGetReferenceGasPrice.
func (c *Client) GetReferenceGasPrice(ctx context.Context) (*BigInt, error) {
	query := `
		query GetReferenceGasPrice {
			epoch {
				referenceGasPrice
			}
		}
	`

	var result struct {
		Epoch *struct {
			ReferenceGasPrice *BigInt `json:"referenceGasPrice"`
		} `json:"epoch"`
	}

	err := c.Execute(ctx, query, nil, &result)
	if err != nil {
		return nil, err
	}

	if result.Epoch == nil {
		return nil, nil
	}
	return result.Epoch.ReferenceGasPrice, nil
}

// GetServiceConfig returns the GraphQL service configuration.
func (c *Client) GetServiceConfig(ctx context.Context) (*ServiceConfig, error) {
	query := `
		query GetServiceConfig {
			serviceConfig {
				maxQueryDepth
				maxQueryNodes
				maxOutputNodes
				queryTimeoutMs
				maxQueryPayloadSize
				maxTypeArgumentDepth
				maxTypeNodes
				maxMoveValueDepth
			}
		}
	`

	var result struct {
		ServiceConfig *ServiceConfig `json:"serviceConfig"`
	}

	err := c.Execute(ctx, query, nil, &result)
	if err != nil {
		return nil, err
	}

	return result.ServiceConfig, nil
}

// GetAvailableRange returns the available data range.
func (c *Client) GetAvailableRange(ctx context.Context) (*AvailableRange, error) {
	query := `
		query GetAvailableRange {
			availableRange {
				first { sequenceNumber digest }
				last { sequenceNumber digest }
			}
		}
	`

	var result struct {
		AvailableRange *AvailableRange `json:"availableRange"`
	}

	err := c.Execute(ctx, query, nil, &result)
	if err != nil {
		return nil, err
	}

	return result.AvailableRange, nil
}

// =============================================================================
// Validator & Staking Queries
// =============================================================================

// GetValidators returns all validators for an epoch.
func (c *Client) GetValidators(ctx context.Context, epochID *UInt53, pagination *PaginationArgs) (*Connection[Validator], error) {
	query := `
		query GetValidators($epochId: UInt53, $first: Int, $after: String, $last: Int, $before: String) {
			epoch(epochId: $epochId) {
				validatorSet {
					activeValidators(first: $first, after: $after, last: $last, before: $before) {
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
						nodes {
							atRisk
							contents {
								type { repr }
								json
							}
						}
					}
				}
			}
		}
	`

	vars := make(map[string]any)
	if epochID != nil {
		vars["epochId"] = *epochID
	}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Epoch *struct {
			ValidatorSet *struct {
				ActiveValidators *Connection[Validator] `json:"activeValidators"`
			} `json:"validatorSet"`
		} `json:"epoch"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Epoch == nil || result.Epoch.ValidatorSet == nil {
		return nil, nil
	}

	return result.Epoch.ValidatorSet.ActiveValidators, nil
}

// GetStakedSui returns staked SUI for an address.
func (c *Client) GetStakedSui(ctx context.Context, owner SuiAddress, pagination *PaginationArgs) (*Connection[StakedSui], error) {
	query := `
		query GetStakedSui($address: SuiAddress!, $first: Int, $after: String, $last: Int, $before: String) {
			address(address: $address) {
				stakedSuis(first: $first, after: $after, last: $last, before: $before) {
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
					nodes {
						address
						version
						digest
						principal
						stakeStatus
						activatedEpoch { epochId }
						requestedEpoch { epochId }
						estimatedReward
					}
				}
			}
		}
	`

	vars := map[string]any{"address": owner}
	if pagination != nil {
		for k, v := range pagination.ToVariables() {
			vars[k] = v
		}
	}

	var result struct {
		Address *struct {
			StakedSuis *Connection[StakedSui] `json:"stakedSuis"`
		} `json:"address"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	if result.Address == nil {
		return nil, nil
	}

	return result.Address.StakedSuis, nil
}

// =============================================================================
// Package & Module Queries
// =============================================================================

// GetPackage returns a Move package by address.
func (c *Client) GetPackage(ctx context.Context, address SuiAddress) (*MovePackage, error) {
	query := `
		query GetPackage($address: SuiAddress!) {
			object(address: $address) {
				asMovePackage {
					address
					version
					digest
					modules {
						nodes {
							name
							package { address }
							fileFormatVersion
						}
					}
					linkage {
						originalId
						upgradedId
						version
					}
					typeOrigins {
						module
						struct
						definingId
					}
				}
			}
		}
	`

	var result struct {
		Object *struct {
			AsMovePackage *MovePackage `json:"asMovePackage"`
		} `json:"object"`
	}

	err := c.Execute(ctx, query, map[string]any{"address": address}, &result)
	if err != nil {
		return nil, err
	}

	if result.Object == nil {
		return nil, nil
	}

	return result.Object.AsMovePackage, nil
}

// GetModule returns a Move module from a package.
func (c *Client) GetModule(ctx context.Context, packageAddress SuiAddress, moduleName string) (*MoveModule, error) {
	query := `
		query GetModule($address: SuiAddress!, $module: String!) {
			object(address: $address) {
				asMovePackage {
					module(name: $module) {
						name
						package { address }
						fileFormatVersion
						friends {
							nodes { name }
						}
						structs {
							nodes {
								name
								abilities
								typeParameters { constraints isPhantom }
								fields {
									name
									type {
										repr
										signature
									}
								}
							}
						}
						enums {
							nodes {
								name
								abilities
								typeParameters { constraints isPhantom }
								variants {
									name
									fields {
										name
										type {
											repr
											signature
										}
									}
								}
							}
						}
						functions {
							nodes {
								name
								visibility
								isEntry
								typeParameters { constraints }
								parameters {
									repr
									signature
								}
								return {
									repr
									signature
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Object *struct {
			AsMovePackage *struct {
				Module *MoveModule `json:"module"`
			} `json:"asMovePackage"`
		} `json:"object"`
	}

	err := c.Execute(ctx, query, map[string]any{"address": packageAddress, "module": moduleName}, &result)
	if err != nil {
		return nil, err
	}

	if result.Object == nil || result.Object.AsMovePackage == nil {
		return nil, nil
	}

	return result.Object.AsMovePackage.Module, nil
}

// GetNormalizedMoveFunction returns normalized function info for a Move function.
// Equivalent to Blockvision's SuiGetNormalizedMoveFunction.
func (c *Client) GetNormalizedMoveFunction(ctx context.Context, packageAddress SuiAddress, moduleName, functionName string) (*MoveFunction, error) {
	query := `
		query GetNormalizedMoveFunction($address: SuiAddress!, $module: String!, $function: String!) {
			object(address: $address) {
				asMovePackage {
					module(name: $module) {
						function(name: $function) {
							name
							visibility
							isEntry
							typeParameters { constraints }
							typeParameters { constraints }
							parameters {
								repr
								signature
							}
							return {
								repr
								signature
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Object *struct {
			AsMovePackage *struct {
				Module *struct {
					Function *MoveFunction `json:"function"`
				} `json:"module"`
			} `json:"asMovePackage"`
		} `json:"object"`
	}

	err := c.Execute(ctx, query, map[string]any{
		"address":  packageAddress,
		"module":   moduleName,
		"function": functionName,
	}, &result)
	if err != nil {
		return nil, err
	}

	if result.Object == nil || result.Object.AsMovePackage == nil || result.Object.AsMovePackage.Module == nil {
		return nil, nil
	}

	return result.Object.AsMovePackage.Module.Function, nil
}

// GetNormalizedMoveStruct returns normalized struct info.
// Equivalent to Blockvision's SuiGetNormalizedMoveStruct.
func (c *Client) GetNormalizedMoveStruct(ctx context.Context, packageAddress SuiAddress, moduleName, structName string) (*MoveStruct, error) {
	query := `
		query GetNormalizedMoveStruct($address: SuiAddress!, $module: String!, $struct: String!) {
			object(address: $address) {
				asMovePackage {
					module(name: $module) {
						struct(name: $struct) {
							name
							abilities
							typeParameters { constraints isPhantom }
							typeParameters { constraints isPhantom }
							fields {
								name
								type {
									repr
									signature
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Object *struct {
			AsMovePackage *struct {
				Module *struct {
					Struct *MoveStruct `json:"struct"`
				} `json:"module"`
			} `json:"asMovePackage"`
		} `json:"object"`
	}

	err := c.Execute(ctx, query, map[string]any{
		"address": packageAddress,
		"module":  moduleName,
		"struct":  structName,
	}, &result)
	if err != nil {
		return nil, err
	}

	if result.Object == nil || result.Object.AsMovePackage == nil || result.Object.AsMovePackage.Module == nil {
		return nil, nil
	}

	return result.Object.AsMovePackage.Module.Struct, nil
}

// =============================================================================
// Raw Query Execution
// =============================================================================

// RawQuery executes a custom GraphQL query with variables.
// This allows users to run any GraphQL query not covered by the built-in methods.
func (c *Client) RawQuery(ctx context.Context, query string, variables map[string]any, result any) error {
	return c.Execute(ctx, query, variables, result)
}
