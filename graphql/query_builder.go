package graphql

import (
	"context"
	"fmt"
	"strings"
)

// =============================================================================
// QueryBuilder - Flexible GraphQL Query Construction
// =============================================================================

// QueryBuilder provides a fluent API for building custom GraphQL queries.
// It allows users to construct queries dynamically with full control over
// which fields are requested.
type QueryBuilder struct {
	operationType string // "query" or "mutation"
	operationName string
	variables     []variableDef
	selections    []selectionBuilder
	variableIndex int
}

// variableDef represents a GraphQL variable definition.
type variableDef struct {
	name     string
	typeName string
	value    any
}

// selectionBuilder represents a GraphQL field selection.
type selectionBuilder struct {
	name       string
	alias      string
	arguments  []argumentBuilder
	selections []selectionBuilder
	inline     bool   // for inline fragments
	typeName   string // for inline fragments (...on Type)
}

// argumentBuilder represents a GraphQL field argument.
type argumentBuilder struct {
	name     string
	value    any
	variable string // if using a variable reference
}

// NewQueryBuilder creates a new query builder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		operationType: "query",
		variables:     make([]variableDef, 0),
		selections:    make([]selectionBuilder, 0),
	}
}

// NewMutationBuilder creates a new mutation builder.
func NewMutationBuilder() *QueryBuilder {
	return &QueryBuilder{
		operationType: "mutation",
		variables:     make([]variableDef, 0),
		selections:    make([]selectionBuilder, 0),
	}
}

// Name sets the operation name (optional but recommended for debugging).
func (qb *QueryBuilder) Name(name string) *QueryBuilder {
	qb.operationName = name
	return qb
}

// Variable adds a variable definition and returns the variable reference name.
func (qb *QueryBuilder) Variable(name, typeName string, value any) string {
	qb.variables = append(qb.variables, variableDef{
		name:     name,
		typeName: typeName,
		value:    value,
	})
	return "$" + name
}

// AutoVariable adds a variable with an auto-generated name.
func (qb *QueryBuilder) AutoVariable(typeName string, value any) string {
	qb.variableIndex++
	name := fmt.Sprintf("v%d", qb.variableIndex)
	return qb.Variable(name, typeName, value)
}

// Field adds a root-level field selection.
func (qb *QueryBuilder) Field(name string) *FieldBuilder {
	fb := &FieldBuilder{
		parent: qb,
		selection: selectionBuilder{
			name:       name,
			arguments:  make([]argumentBuilder, 0),
			selections: make([]selectionBuilder, 0),
		},
	}
	return fb
}

// FieldBuilder helps construct field selections.
type FieldBuilder struct {
	parent    *QueryBuilder
	selection selectionBuilder
}

// Alias sets an alias for the field.
func (fb *FieldBuilder) Alias(alias string) *FieldBuilder {
	fb.selection.alias = alias
	return fb
}

// Arg adds an argument to the field.
func (fb *FieldBuilder) Arg(name string, value any) *FieldBuilder {
	fb.selection.arguments = append(fb.selection.arguments, argumentBuilder{
		name:  name,
		value: value,
	})
	return fb
}

// ArgVar adds an argument using a variable reference.
func (fb *FieldBuilder) ArgVar(name, variableRef string) *FieldBuilder {
	fb.selection.arguments = append(fb.selection.arguments, argumentBuilder{
		name:     name,
		variable: variableRef,
	})
	return fb
}

// Fields adds scalar field selections.
func (fb *FieldBuilder) Fields(fields ...string) *FieldBuilder {
	for _, f := range fields {
		fb.selection.selections = append(fb.selection.selections, selectionBuilder{
			name: f,
		})
	}
	return fb
}

// SubField adds a nested field selection with its own fields.
func (fb *FieldBuilder) SubField(name string) *SubFieldBuilder {
	return &SubFieldBuilder{
		parent: fb,
		selection: selectionBuilder{
			name:       name,
			arguments:  make([]argumentBuilder, 0),
			selections: make([]selectionBuilder, 0),
		},
	}
}

// InlineFragment adds an inline fragment (...on Type).
func (fb *FieldBuilder) InlineFragment(typeName string) *SubFieldBuilder {
	return &SubFieldBuilder{
		parent: fb,
		selection: selectionBuilder{
			inline:     true,
			typeName:   typeName,
			selections: make([]selectionBuilder, 0),
		},
	}
}

// Done finalizes the field and adds it to the query.
func (fb *FieldBuilder) Done() *QueryBuilder {
	fb.parent.selections = append(fb.parent.selections, fb.selection)
	return fb.parent
}

// SubFieldBuilder helps construct nested field selections.
type SubFieldBuilder struct {
	parent    *FieldBuilder
	selection selectionBuilder
}

// Arg adds an argument to the sub-field.
func (sfb *SubFieldBuilder) Arg(name string, value any) *SubFieldBuilder {
	sfb.selection.arguments = append(sfb.selection.arguments, argumentBuilder{
		name:  name,
		value: value,
	})
	return sfb
}

// ArgVar adds an argument using a variable reference.
func (sfb *SubFieldBuilder) ArgVar(name, variableRef string) *SubFieldBuilder {
	sfb.selection.arguments = append(sfb.selection.arguments, argumentBuilder{
		name:     name,
		variable: variableRef,
	})
	return sfb
}

// Fields adds scalar field selections.
func (sfb *SubFieldBuilder) Fields(fields ...string) *SubFieldBuilder {
	for _, f := range fields {
		sfb.selection.selections = append(sfb.selection.selections, selectionBuilder{
			name: f,
		})
	}
	return sfb
}

// SubField adds a nested field.
func (sfb *SubFieldBuilder) SubField(name string) *NestedSubFieldBuilder {
	return &NestedSubFieldBuilder{
		parent: sfb,
		selection: selectionBuilder{
			name:       name,
			arguments:  make([]argumentBuilder, 0),
			selections: make([]selectionBuilder, 0),
		},
	}
}

// InlineFragment adds an inline fragment.
func (sfb *SubFieldBuilder) InlineFragment(typeName string) *NestedSubFieldBuilder {
	return &NestedSubFieldBuilder{
		parent: sfb,
		selection: selectionBuilder{
			inline:     true,
			typeName:   typeName,
			selections: make([]selectionBuilder, 0),
		},
	}
}

// End finalizes the sub-field and returns to the parent.
func (sfb *SubFieldBuilder) End() *FieldBuilder {
	sfb.parent.selection.selections = append(sfb.parent.selection.selections, sfb.selection)
	return sfb.parent
}

// NestedSubFieldBuilder handles deeply nested selections.
type NestedSubFieldBuilder struct {
	parent    *SubFieldBuilder
	selection selectionBuilder
}

// Arg adds an argument.
func (nsfb *NestedSubFieldBuilder) Arg(name string, value any) *NestedSubFieldBuilder {
	nsfb.selection.arguments = append(nsfb.selection.arguments, argumentBuilder{
		name:  name,
		value: value,
	})
	return nsfb
}

// ArgVar adds an argument using a variable reference.
func (nsfb *NestedSubFieldBuilder) ArgVar(name, variableRef string) *NestedSubFieldBuilder {
	nsfb.selection.arguments = append(nsfb.selection.arguments, argumentBuilder{
		name:     name,
		variable: variableRef,
	})
	return nsfb
}

// Fields adds scalar field selections.
func (nsfb *NestedSubFieldBuilder) Fields(fields ...string) *NestedSubFieldBuilder {
	for _, f := range fields {
		nsfb.selection.selections = append(nsfb.selection.selections, selectionBuilder{
			name: f,
		})
	}
	return nsfb
}

// SubField adds a nested field.
func (nsfb *NestedSubFieldBuilder) SubField(name string) *DeepNestedSubFieldBuilder {
	return &DeepNestedSubFieldBuilder{
		parent: nsfb,
		selection: selectionBuilder{
			name:       name,
			arguments:  make([]argumentBuilder, 0),
			selections: make([]selectionBuilder, 0),
		},
	}
}

// InlineFragment adds an inline fragment.
func (nsfb *NestedSubFieldBuilder) InlineFragment(typeName string) *DeepNestedSubFieldBuilder {
	return &DeepNestedSubFieldBuilder{
		parent: nsfb,
		selection: selectionBuilder{
			inline:     true,
			typeName:   typeName,
			selections: make([]selectionBuilder, 0),
		},
	}
}

// End finalizes and returns to parent.
func (nsfb *NestedSubFieldBuilder) End() *SubFieldBuilder {
	nsfb.parent.selection.selections = append(nsfb.parent.selection.selections, nsfb.selection)
	return nsfb.parent
}

// DeepNestedSubFieldBuilder handles even deeper nested selections.
type DeepNestedSubFieldBuilder struct {
	parent    *NestedSubFieldBuilder
	selection selectionBuilder
}

// Arg adds an argument.
func (dnsfb *DeepNestedSubFieldBuilder) Arg(name string, value any) *DeepNestedSubFieldBuilder {
	dnsfb.selection.arguments = append(dnsfb.selection.arguments, argumentBuilder{
		name:  name,
		value: value,
	})
	return dnsfb
}

// Fields adds scalar field selections.
func (dnsfb *DeepNestedSubFieldBuilder) Fields(fields ...string) *DeepNestedSubFieldBuilder {
	for _, f := range fields {
		dnsfb.selection.selections = append(dnsfb.selection.selections, selectionBuilder{
			name: f,
		})
	}
	return dnsfb
}

// SubField adds a nested field.
func (dnsfb *DeepNestedSubFieldBuilder) SubField(name string) *DeepNestedSubFieldBuilder {
	// For deeply nested selections, we add the selection immediately
	subSel := selectionBuilder{
		name:       name,
		arguments:  make([]argumentBuilder, 0),
		selections: make([]selectionBuilder, 0),
	}
	dnsfb.selection.selections = append(dnsfb.selection.selections, subSel)
	return dnsfb
}

// End finalizes and returns to parent.
func (dnsfb *DeepNestedSubFieldBuilder) End() *NestedSubFieldBuilder {
	dnsfb.parent.selection.selections = append(dnsfb.parent.selection.selections, dnsfb.selection)
	return dnsfb.parent
}

// Build generates the GraphQL query string.
func (qb *QueryBuilder) Build() (string, map[string]any) {
	var sb strings.Builder

	// Operation type and name
	sb.WriteString(qb.operationType)
	if qb.operationName != "" {
		sb.WriteString(" ")
		sb.WriteString(qb.operationName)
	}

	// Variable definitions
	if len(qb.variables) > 0 {
		sb.WriteString("(")
		for i, v := range qb.variables {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("$")
			sb.WriteString(v.name)
			sb.WriteString(": ")
			sb.WriteString(v.typeName)
		}
		sb.WriteString(")")
	}

	sb.WriteString(" {\n")

	// Selections
	for _, sel := range qb.selections {
		qb.writeSelection(&sb, sel, 1)
	}

	sb.WriteString("}")

	// Build variables map
	vars := make(map[string]any)
	for _, v := range qb.variables {
		vars[v.name] = v.value
	}

	return sb.String(), vars
}

// writeSelection writes a field selection to the string builder with proper indentation.
func (qb *QueryBuilder) writeSelection(sb *strings.Builder, sel selectionBuilder, indent int) {
	indentStr := strings.Repeat("  ", indent)

	if sel.inline {
		sb.WriteString(indentStr)
		sb.WriteString("... on ")
		sb.WriteString(sel.typeName)
		sb.WriteString(" {\n")
		for _, sub := range sel.selections {
			qb.writeSelection(sb, sub, indent+1)
		}
		sb.WriteString(indentStr)
		sb.WriteString("}\n")
		return
	}

	sb.WriteString(indentStr)

	// Alias
	if sel.alias != "" {
		sb.WriteString(sel.alias)
		sb.WriteString(": ")
	}

	sb.WriteString(sel.name)

	// Arguments
	if len(sel.arguments) > 0 {
		sb.WriteString("(")
		for i, arg := range sel.arguments {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(arg.name)
			sb.WriteString(": ")
			if arg.variable != "" {
				sb.WriteString(arg.variable)
			} else {
				sb.WriteString(formatValue(arg.value))
			}
		}
		sb.WriteString(")")
	}

	// Sub-selections
	if len(sel.selections) > 0 {
		sb.WriteString(" {\n")
		for _, sub := range sel.selections {
			qb.writeSelection(sb, sub, indent+1)
		}
		sb.WriteString(indentStr)
		sb.WriteString("}")
	}

	sb.WriteString("\n")
}

// formatValue formats a value for inclusion in a GraphQL query.
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Execute runs the built query against the client.
func (qb *QueryBuilder) Execute(ctx context.Context, client *Client, result any) error {
	query, vars := qb.Build()
	return client.Execute(ctx, query, vars, result)
}

// =============================================================================
// Pre-built Query Templates (using raw queries for complex nested structures)
// =============================================================================

// RawQueryTemplate holds a pre-built query template with variable definitions.
type RawQueryTemplate struct {
	Query     string
	Variables map[string]any
}

// Execute runs the query template against the client.
func (rqt *RawQueryTemplate) Execute(ctx context.Context, client *Client, result any) error {
	return client.Execute(ctx, rqt.Query, rqt.Variables, result)
}

// ObjectQueryTemplate creates a query template for object queries.
func ObjectQueryTemplate(objectID SuiAddress) *RawQueryTemplate {
	return &RawQueryTemplate{
		Query: `
			query GetObject($objectId: SuiAddress!) {
				object(address: $objectId) {
					address
					version
					digest
					storageRebate
					owner {
						__typename
						... on AddressOwner { address { address } }
						... on Shared { initialSharedVersion }
					}
					asMoveObject {
						hasPublicTransfer
						type { repr }
						contents { type { repr } bcs json }
					}
				}
			}
		`,
		Variables: map[string]any{"objectId": objectID},
	}
}

// BalanceQueryTemplate creates a query template for balance queries.
func BalanceQueryTemplate(owner SuiAddress) *RawQueryTemplate {
	return &RawQueryTemplate{
		Query: `
			query GetBalances($address: SuiAddress!) {
				address(address: $address) {
					balances {
						nodes {
							totalBalance
							coinType { repr }
						}
					}
				}
			}
		`,
		Variables: map[string]any{"address": owner},
	}
}

// TransactionQueryTemplate creates a query template for transaction queries.
func TransactionQueryTemplate(digest string) *RawQueryTemplate {
	return &RawQueryTemplate{
		Query: `
			query GetTransaction($digest: String!) {
				transactionBlock(digest: $digest) {
					digest
					sender { address }
					effects {
						digest
						status
						timestamp
						epoch { epochId }
						checkpoint { sequenceNumber }
						gasEffects {
							gasSummary {
								computationCost
								storageCost
								storageRebate
								nonRefundableStorageFee
							}
						}
					}
				}
			}
		`,
		Variables: map[string]any{"digest": digest},
	}
}

// CoinsQueryTemplate creates a query template for coin queries.
func CoinsQueryTemplate(owner SuiAddress, coinType *string, first int) *RawQueryTemplate {
	vars := map[string]any{
		"address": owner,
		"first":   first,
	}
	if coinType != nil {
		vars["coinType"] = *coinType
	}

	return &RawQueryTemplate{
		Query: `
			query GetCoins($address: SuiAddress!, $first: Int, $coinType: String) {
				address(address: $address) {
					coins(first: $first, type: $coinType) {
						pageInfo {
							hasNextPage
							endCursor
						}
						nodes {
							coinBalance
							address
							version
							digest
						}
					}
				}
			}
		`,
		Variables: vars,
	}
}

// EventsQueryTemplate creates a query template for event queries.
func EventsQueryTemplate(filter *EventFilter, first int) *RawQueryTemplate {
	vars := map[string]any{"first": first}
	if filter != nil {
		vars["filter"] = filter
	}

	return &RawQueryTemplate{
		Query: `
			query GetEvents($first: Int, $filter: EventFilter) {
				events(first: $first, filter: $filter) {
					pageInfo {
						hasNextPage
						endCursor
					}
					nodes {
						timestamp
						eventBcs
						sender { address }
						transactionModule { name package { address } }
						contents { type { repr } bcs json }
					}
				}
			}
		`,
		Variables: vars,
	}
}

// CheckpointQueryTemplate creates a query template for checkpoint queries.
func CheckpointQueryTemplate(sequenceNumber *UInt53) *RawQueryTemplate {
	vars := map[string]any{}
	if sequenceNumber != nil {
		vars["sequenceNumber"] = *sequenceNumber
	}

	return &RawQueryTemplate{
		Query: `
			query GetCheckpoint($sequenceNumber: UInt53) {
				checkpoint(id: { sequenceNumber: $sequenceNumber }) {
					sequenceNumber
					digest
					timestamp
					previousCheckpointDigest
					networkTotalTransactions
					rollingGasSummary {
						computationCost
						storageCost
						storageRebate
						nonRefundableStorageFee
					}
					epoch { epochId }
				}
			}
		`,
		Variables: vars,
	}
}

// =============================================================================
// Connection Helpers for Pagination
// =============================================================================

// PagedQuery wraps a query builder to handle pagination automatically.
type PagedQuery[T any] struct {
	client     *Client
	baseQuery  func(cursor *string) *QueryBuilder
	resultPath func(any) (*Connection[T], error)
}

// NewPagedQuery creates a new paged query.
func NewPagedQuery[T any](
	client *Client,
	baseQuery func(cursor *string) *QueryBuilder,
	resultPath func(any) (*Connection[T], error),
) *PagedQuery[T] {
	return &PagedQuery[T]{
		client:     client,
		baseQuery:  baseQuery,
		resultPath: resultPath,
	}
}

// FetchPage fetches a single page of results.
func (pq *PagedQuery[T]) FetchPage(ctx context.Context, cursor *string) (*Connection[T], error) {
	qb := pq.baseQuery(cursor)
	query, vars := qb.Build()

	var rawResult map[string]any
	if err := pq.client.Execute(ctx, query, vars, &rawResult); err != nil {
		return nil, err
	}

	return pq.resultPath(rawResult)
}

// FetchAll fetches all pages and returns combined results.
func (pq *PagedQuery[T]) FetchAll(ctx context.Context, maxPages int) ([]T, error) {
	var allNodes []T
	var cursor *string
	pages := 0

	for {
		conn, err := pq.FetchPage(ctx, cursor)
		if err != nil {
			return nil, err
		}

		if conn == nil {
			break
		}

		allNodes = append(allNodes, conn.Nodes...)

		if !conn.PageInfo.HasNextPage {
			break
		}

		cursor = conn.PageInfo.EndCursor
		pages++

		if maxPages > 0 && pages >= maxPages {
			break
		}
	}

	return allNodes, nil
}
