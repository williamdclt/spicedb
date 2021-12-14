package options

import (
	v0 "github.com/authzed/authzed-go/proto/authzed/api/v0"
)

//go:generate go run github.com/ecordell/optgen -output zz_generated.query_options.go . QueryOptions

// QueryOptions are the options that can affect the results of a normal forward query.
type QueryOptions struct {
	Limit    *uint64
	Usersets []*v0.ObjectAndRelation
}

var (
	one = uint64(1)

	// LimitOne is a constant *uint64 that can be used with WithLimit requests.
	LimitOne = &one
)
