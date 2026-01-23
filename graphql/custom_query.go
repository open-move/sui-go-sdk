package graphql

import (
	"context"
	"encoding/json"
	"fmt"
)

// =============================================================================
// Custom Query/Mutation Support (TypeScript gql-style API)
// =============================================================================

// GraphQLDocument represents a parsed GraphQL query or mutation document.
// It provides a type-safe way to define and execute custom queries.
type GraphQLDocument struct {
	query string
	name  string
}

// Query creates a new GraphQL query document from a query string.
// This provides a similar experience to the TypeScript graphql“ tagged template.
func Query(query string) *GraphQLDocument {
	return &GraphQLDocument{
		query: query,
		name:  extractOperationName(query),
	}
}

// Mutation creates a new GraphQL mutation document from a mutation string.
//
// Example:
//
//	executeTx := Mutation(`
//		mutation ExecuteTx($txBytes: String!, $signatures: [String!]!) {
//			executeTransaction(txBytes: $txBytes, signatures: $signatures) {
//				effects { digest status }
//				errors
//			}
//		}
//	`)
func Mutation(query string) *GraphQLDocument {
	return &GraphQLDocument{
		query: query,
		name:  extractOperationName(query),
	}
}

// String returns the raw query string.
func (d *GraphQLDocument) String() string {
	return d.query
}

// Name returns the operation name if present.
func (d *GraphQLDocument) Name() string {
	return d.name
}

// Vars is a shorthand for map[string]any for query variables.
type Vars map[string]any

// =============================================================================
// Client Methods for Custom Queries
// =============================================================================

// Query executes a custom GraphQL query document.
func (c *Client) Query(ctx context.Context, doc *GraphQLDocument, variables Vars, result any) error {
	return c.Execute(ctx, doc.query, variables, result)
}

// MustQuery executes a query and panics on error. Useful for initialization.
func (c *Client) MustQuery(ctx context.Context, doc *GraphQLDocument, variables Vars, result any) {
	if err := c.Query(ctx, doc, variables, result); err != nil {
		panic(fmt.Sprintf("graphql query failed: %v", err))
	}
}

// Mutate executes a custom GraphQL mutation document.
//
// Example:
//
//	executeTx := Mutation(`
//		mutation ExecuteTx($txBytes: String!, $signatures: [String!]!) {
//			executeTransaction(txBytes: $txBytes, signatures: $signatures) {
//				effects { digest status }
//				errors
//			}
//		}
//	`)
//
//	var result struct {
//		ExecuteTransaction *ExecuteTransactionResult `json:"executeTransaction"`
//	}
//	err := client.Mutate(ctx, executeTx, Vars{
//		"txBytes": txBytes,
//		"signatures": signatures,
//	}, &result)
func (c *Client) Mutate(ctx context.Context, doc *GraphQLDocument, variables Vars, result any) error {
	return c.Execute(ctx, doc.query, variables, result)
}

// MustMutate executes a mutation and panics on error. Useful for initialization.
func (c *Client) MustMutate(ctx context.Context, doc *GraphQLDocument, variables Vars, result any) {
	if err := c.Mutate(ctx, doc, variables, result); err != nil {
		panic(fmt.Sprintf("graphql mutation failed: %v", err))
	}
}

// =============================================================================
// Typed Query Helpers
// =============================================================================

// TypedQuery provides a type-safe wrapper for queries with specific result types.
// This allows defining queries with their expected result types upfront.
type TypedQuery[T any] struct {
	doc *GraphQLDocument
}

// NewTypedQuery creates a new typed query.
func NewTypedQuery[T any](query string) *TypedQuery[T] {
	return &TypedQuery[T]{
		doc: Query(query),
	}
}

// Execute runs the typed query and returns the result.
func (tq *TypedQuery[T]) Execute(ctx context.Context, client *Client, variables Vars) (*T, error) {
	var result T
	if err := client.Query(ctx, tq.doc, variables, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MustExecute runs the typed query and panics on error.
func (tq *TypedQuery[T]) MustExecute(ctx context.Context, client *Client, variables Vars) *T {
	result, err := tq.Execute(ctx, client, variables)
	if err != nil {
		panic(fmt.Sprintf("typed query failed: %v", err))
	}
	return result
}

// TypedMutation provides a type-safe wrapper for mutations with specific result types.
type TypedMutation[T any] struct {
	doc *GraphQLDocument
}

// NewTypedMutation creates a new typed mutation.
func NewTypedMutation[T any](mutation string) *TypedMutation[T] {
	return &TypedMutation[T]{
		doc: Mutation(mutation),
	}
}

// Execute runs the typed mutation and returns the result.
func (tm *TypedMutation[T]) Execute(ctx context.Context, client *Client, variables Vars) (*T, error) {
	var result T
	if err := client.Mutate(ctx, tm.doc, variables, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MustExecute runs the typed mutation and panics on error.
func (tm *TypedMutation[T]) MustExecute(ctx context.Context, client *Client, variables Vars) *T {
	result, err := tm.Execute(ctx, client, variables)
	if err != nil {
		panic(fmt.Sprintf("typed mutation failed: %v", err))
	}
	return result
}

// =============================================================================
// Query Request Builder (fluent API for single queries)
// =============================================================================

// QueryRequest provides a fluent API for building and executing a single query.
type QueryRequest struct {
	client    *Client
	doc       *GraphQLDocument
	variables Vars
}

// NewQueryRequest creates a new query request.
func (c *Client) NewQueryRequest(doc *GraphQLDocument) *QueryRequest {
	return &QueryRequest{
		client:    c,
		doc:       doc,
		variables: make(Vars),
	}
}

// Var adds a single variable to the query.
func (qr *QueryRequest) Var(name string, value any) *QueryRequest {
	qr.variables[name] = value
	return qr
}

// Variables sets all variables at once.
func (qr *QueryRequest) Variables(vars Vars) *QueryRequest {
	qr.variables = vars
	return qr
}

// Execute runs the query and unmarshals into result.
func (qr *QueryRequest) Execute(ctx context.Context, result any) error {
	return qr.client.Query(ctx, qr.doc, qr.variables, result)
}

// ExecuteRaw runs the query and returns the raw JSON response.
func (qr *QueryRequest) ExecuteRaw(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := qr.client.Query(ctx, qr.doc, qr.variables, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MutationRequest provides a fluent API for building and executing a single mutation.
type MutationRequest struct {
	client    *Client
	doc       *GraphQLDocument
	variables Vars
}

// NewMutationRequest creates a new mutation request.
func (c *Client) NewMutationRequest(doc *GraphQLDocument) *MutationRequest {
	return &MutationRequest{
		client:    c,
		doc:       doc,
		variables: make(Vars),
	}
}

// Var adds a single variable to the mutation.
func (mr *MutationRequest) Var(name string, value any) *MutationRequest {
	mr.variables[name] = value
	return mr
}

// Variables sets all variables at once.
func (mr *MutationRequest) Variables(vars Vars) *MutationRequest {
	mr.variables = vars
	return mr
}

// Execute runs the mutation and unmarshals into result.
func (mr *MutationRequest) Execute(ctx context.Context, result any) error {
	return mr.client.Mutate(ctx, mr.doc, mr.variables, result)
}

// ExecuteRaw runs the mutation and returns the raw JSON response.
func (mr *MutationRequest) ExecuteRaw(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	if err := mr.client.Mutate(ctx, mr.doc, mr.variables, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// extractOperationName extracts the operation name from a GraphQL query string.
func extractOperationName(query string) string {
	// Simple extraction - looks for "query Name" or "mutation Name"
	// This is a basic implementation; a full parser would be more robust
	var inWord bool
	var wordStart int
	var foundKeyword bool
	var keyword string

	for i, c := range query {
		isLetter := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
		isAlphanumeric := isLetter || (c >= '0' && c <= '9')

		if !inWord && isLetter {
			inWord = true
			wordStart = i
		} else if inWord && !isAlphanumeric {
			word := query[wordStart:i]
			if !foundKeyword {
				if word == "query" || word == "mutation" || word == "subscription" {
					foundKeyword = true
					keyword = word
				}
			} else {
				// This is the operation name
				return word
			}
			inWord = false
		}

		// Stop at opening parenthesis or brace
		if foundKeyword && (c == '(' || c == '{') {
			if inWord {
				return query[wordStart:i]
			}
			break
		}
	}

	// Handle case where name is at the end
	if inWord && foundKeyword {
		return query[wordStart:]
	}

	return keyword // Return keyword if no name found
}

// =============================================================================
// Pre-defined Common Queries
// =============================================================================

// Common query documents that can be reused.
var (
	// GetBalanceQuery fetches the balance of a specific coin type for an address.
	GetBalanceQuery = Query(`
		query getBalance($address: SuiAddress!, $coinType: String) {
			address(address: $address) {
				balance(type: $coinType) {
					coinType { repr }
					totalBalance
				}
			}
		}
	`)

	// GetObjectQuery fetches an object by its address.
	GetObjectQuery = Query(`
		query getObject($objectId: SuiAddress!) {
			object(address: $objectId) {
				address
				version
				digest
				owner {
					__typename
					... on AddressOwner { address { address } }
					... on Shared { initialSharedVersion }
				}
				asMoveObject {
					type { repr }
					contents { json }
				}
			}
		}
	`)

	// GetTransactionQuery fetches a transaction by its digest.
	GetTransactionQuery = Query(`
		query getTransaction($digest: String!) {
			transactionBlock(digest: $digest) {
				digest
				sender { address }
				effects {
					status
					timestamp
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
						}
					}
				}
			}
		}
	`)

	// GetEpochQuery fetches the current epoch information.
	GetEpochQuery = Query(`
		query getEpoch {
			epoch {
				epochId
				referenceGasPrice
				startTimestamp
				endTimestamp
			}
		}
	`)

	// SimulateTransactionMutation simulates a transaction.
	SimulateTransactionMutation = Mutation(`
		mutation simulateTransaction($txBytes: String!, $skipChecks: Boolean) {
			simulateTransaction(txBytes: $txBytes, skipChecks: $skipChecks) {
				effects {
					status
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
						}
					}
				}
				error
			}
		}
	`)

	// ExecuteTransactionMutation executes a signed transaction.
	ExecuteTransactionMutation = Mutation(`
		mutation executeTransaction($txBytes: String!, $signatures: [String!]!) {
			executeTransaction(txBytes: $txBytes, signatures: $signatures) {
				effects {
					digest
					status
					timestamp
				}
				errors
			}
		}
	`)
)
