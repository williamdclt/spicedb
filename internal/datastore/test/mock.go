package test

import (
	"context"

	v0 "github.com/authzed/authzed-go/proto/authzed/api/v0"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/stretchr/testify/mock"

	"github.com/authzed/spicedb/internal/datastore"
	"github.com/authzed/spicedb/internal/datastore/options"
)

// MockedDatastore is a mock implementation of the datastore.
type MockedDatastore struct {
	mock.Mock
}

func (md *MockedDatastore) Close() error {
	args := md.Called()
	return args.Error(0)
}

func (md *MockedDatastore) IsReady(ctx context.Context) (bool, error) {
	args := md.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (md *MockedDatastore) DeleteRelationships(ctx context.Context, preconditions []*v1.Precondition, filter *v1.RelationshipFilter) (datastore.Revision, error) {
	args := md.Called(ctx, preconditions, filter)
	return args.Get(0).(datastore.Revision), args.Error(1)
}

func (md *MockedDatastore) WriteTuples(ctx context.Context, preconditions []*v1.Precondition, updates []*v1.RelationshipUpdate) (datastore.Revision, error) {
	args := md.Called(ctx, preconditions, updates)
	return args.Get(0).(datastore.Revision), args.Error(1)
}

func (md *MockedDatastore) OptimizedRevision(ctx context.Context) (datastore.Revision, error) {
	args := md.Called(ctx)
	return args.Get(0).(datastore.Revision), args.Error(1)
}

func (md *MockedDatastore) HeadRevision(ctx context.Context) (datastore.Revision, error) {
	args := md.Called(ctx)
	return args.Get(0).(datastore.Revision), args.Error(1)
}

func (md *MockedDatastore) Watch(ctx context.Context, afterRevision datastore.Revision) (<-chan *datastore.RevisionChanges, <-chan error) {
	args := md.Called(ctx, afterRevision)
	return args.Get(0).(<-chan *datastore.RevisionChanges), args.Get(1).(<-chan error)
}

func (md *MockedDatastore) WriteNamespace(ctx context.Context, newConfig *v0.NamespaceDefinition) (datastore.Revision, error) {
	args := md.Called(ctx, newConfig)
	return args.Get(0).(datastore.Revision), args.Error(1)
}

func (md *MockedDatastore) ReadNamespace(ctx context.Context, nsName string, revision datastore.Revision) (*v0.NamespaceDefinition, datastore.Revision, error) {
	args := md.Called(ctx, nsName, revision)
	return args.Get(0).(*v0.NamespaceDefinition), args.Get(1).(datastore.Revision), args.Error(2)
}

func (md *MockedDatastore) DeleteNamespace(ctx context.Context, nsName string) (datastore.Revision, error) {
	args := md.Called(ctx, nsName)
	return args.Get(0).(datastore.Revision), args.Error(1)
}

func (md *MockedDatastore) QueryTuples(
	ctx context.Context,
	filter *v1.RelationshipFilter,
	revision datastore.Revision,
	options ...options.QueryOptionsOption,
) (iter datastore.TupleIterator, err error) {
	callArgs := make([]interface{}, 0, len(options)+2)
	callArgs = append(callArgs, filter, revision)
	for _, option := range options {
		callArgs = append(callArgs, option)
	}

	args := md.Called(callArgs...)
	return args.Get(0).(datastore.TupleIterator), args.Error(1)
}

func (md *MockedDatastore) ReverseQueryTuplesFromSubjectNamespace(subjectNamespace string, revision datastore.Revision) datastore.ReverseTupleQuery {
	args := md.Called(subjectNamespace, revision)
	return args.Get(0).(datastore.ReverseTupleQuery)
}

func (md *MockedDatastore) ReverseQueryTuplesFromSubject(subject *v0.ObjectAndRelation, revision datastore.Revision) datastore.ReverseTupleQuery {
	args := md.Called(subject, revision)
	return args.Get(0).(datastore.ReverseTupleQuery)
}

func (md *MockedDatastore) ReverseQueryTuplesFromSubjectRelation(subjectNamespace, subjectRelation string, revision datastore.Revision) datastore.ReverseTupleQuery {
	args := md.Called(subjectNamespace, subjectRelation, revision)
	return args.Get(0).(datastore.ReverseTupleQuery)
}

func (md *MockedDatastore) CheckRevision(ctx context.Context, revision datastore.Revision) error {
	args := md.Called(ctx, revision)
	return args.Error(0)
}

func (md *MockedDatastore) ListNamespaces(ctx context.Context, revision datastore.Revision) ([]*v0.NamespaceDefinition, error) {
	args := md.Called(ctx, revision)
	return args.Get(0).([]*v0.NamespaceDefinition), args.Error(1)
}

// MockedReverseTupleQuery is a mock implementation of the reverse tuple query.
type MockedReverseTupleQuery struct {
	mock.Mock
}

func (mrtq *MockedReverseTupleQuery) WithObjectRelation(namespace string, relation string) datastore.ReverseTupleQuery {
	args := mrtq.Called(namespace, relation)
	return args.Get(0).(datastore.ReverseTupleQuery)
}

func (mrtq *MockedReverseTupleQuery) Execute(ctx context.Context) (datastore.TupleIterator, error) {
	args := mrtq.Called(ctx)
	return args.Get(0).(datastore.TupleIterator), args.Error(1)
}

func (mrtq *MockedReverseTupleQuery) Limit(limit uint64) datastore.ReverseTupleQuery {
	args := mrtq.Called(limit)
	return args.Get(0).(datastore.ReverseTupleQuery)
}

var (
	_ datastore.Datastore         = &MockedDatastore{}
	_ datastore.ReverseTupleQuery = &MockedReverseTupleQuery{}
)
