package repository

import (
	"context"
	"fmt"

	"github.com/pushp314/opencode/athena/repository/parse"
)

// Symbol is a durable symbol fact resolved to the snapshot that contains it.
type Symbol struct {
	parse.Symbol
	Path   string
	SHA256 string
}

// SymbolQuery filters the symbols of a snapshot. Empty fields match anything.
type SymbolQuery struct {
	Name string
	Kind string
	Path string
}

// Import is a durable module reference resolved to the snapshot that contains
// it.
type Import struct {
	parse.Import
	Path   string
	SHA256 string
}

// ImportQuery filters the imports of a snapshot. Empty fields match anything.
type ImportQuery struct {
	Kind string
	Spec string
	Path string
}

// Query reads durable repository facts through the Store.
type Query struct {
	store Store
}

func NewQuery(store Store) *Query {
	return &Query{store: store}
}

// Symbols returns the symbols of the latest snapshot of a repository, ordered
// by path and byte offset. repositoryID is the canonical repository ID.
func (q *Query) Symbols(ctx context.Context, repositoryID string, filter SymbolQuery) ([]Symbol, error) {
	snapshot, _, found, err := q.store.Latest(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no snapshot for repository %s", repositoryID)
	}
	return q.store.Symbols(ctx, snapshot.ID, filter)
}

// Imports returns the module references of the latest snapshot of a repository,
// ordered by path and byte offset.
func (q *Query) Imports(ctx context.Context, repositoryID string, filter ImportQuery) ([]Import, error) {
	snapshot, _, found, err := q.store.Latest(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no snapshot for repository %s", repositoryID)
	}
	return q.store.Imports(ctx, snapshot.ID, filter)
}

// Edges returns the resolved import edges of the latest snapshot of a
// repository, ordered by source path and byte offset.
func (q *Query) Edges(ctx context.Context, repositoryID string, filter EdgeQuery) ([]Edge, error) {
	snapshot, _, found, err := q.store.Latest(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("no snapshot for repository %s", repositoryID)
	}
	return q.store.Edges(ctx, snapshot.ID, filter)
}
