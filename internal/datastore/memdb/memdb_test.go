package memdb

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/authzed/spicedb/internal/datastore"
	pb "github.com/authzed/spicedb/pkg/REDACTEDapi/api"
	"github.com/authzed/spicedb/pkg/tuple"
)

const (
	testUserNamespace     = "test/user"
	testResourceNamespace = "rest/resource"
	testReaderRelation    = "reader"
	ellipsis              = "..."
)

var noPreconditions = []*pb.RelationTuple{}

func makeTestTuple(resourceID, userID string) *pb.RelationTuple {
	return &pb.RelationTuple{
		ObjectAndRelation: &pb.ObjectAndRelation{
			Namespace: testResourceNamespace,
			ObjectId:  resourceID,
			Relation:  testReaderRelation,
		},
		User: &pb.User{
			UserOneof: &pb.User_Userset{
				Userset: &pb.ObjectAndRelation{
					Namespace: testUserNamespace,
					ObjectId:  userID,
					Relation:  ellipsis,
				},
			},
		},
	}
}

func TestSimple(t *testing.T) {
	testCases := []int{1, 2, 4, 32, 1024}

	for _, numTuples := range testCases {
		t.Run(strconv.Itoa(numTuples), func(t *testing.T) {
			require := require.New(t)

			ds, err := NewMemdbTupleDatastore()
			require.NoError(err)

			tRequire := tupleChecker{require, ds}

			var testTuples []*pb.RelationTuple

			var writtenAt uint64
			for i := 0; i < numTuples; i++ {
				resourceName := fmt.Sprintf("resource%d", i)
				userName := fmt.Sprintf("user%d", i)

				newTuple := makeTestTuple(resourceName, userName)
				testTuples = append(testTuples, newTuple)

				writtenAt, err = ds.WriteTuples(noPreconditions, []*pb.RelationTupleUpdate{c(newTuple)})
				require.NoError(err)
				require.Greater(writtenAt, uint64(0))

				tRequire.tupleExists(newTuple, writtenAt)
				tRequire.tupleExists(newTuple, writtenAt+100)
				tRequire.noTupleExists(newTuple, writtenAt-1)
			}

			for _, tupleToFind := range testTuples {
				// Check that we can find the tuple a number of ways
				q := ds.QueryTuples(tupleToFind.ObjectAndRelation.Namespace, writtenAt)

				queries := []datastore.TupleQuery{
					q.WithObjectID(tupleToFind.ObjectAndRelation.ObjectId),
					q.WithUserset(tupleToFind.User.GetUserset()),
					q.WithObjectID(tupleToFind.ObjectAndRelation.ObjectId).WithRelation(tupleToFind.ObjectAndRelation.Relation),
					q.WithObjectID(tupleToFind.ObjectAndRelation.ObjectId).WithUserset(tupleToFind.User.GetUserset()),
					q.WithRelation(tupleToFind.ObjectAndRelation.Relation).WithUserset(tupleToFind.User.GetUserset()),
				}
				for _, query := range queries {
					iter, err := query.Execute()
					require.NoError(err)
					tRequire.verifyIteratorResults(iter, tupleToFind)
				}
			}

			// Check that we can find the group of tuples too
			q := ds.QueryTuples(testTuples[0].ObjectAndRelation.Namespace, writtenAt)

			queries := []datastore.TupleQuery{
				q,
				q.WithRelation(testTuples[0].ObjectAndRelation.Relation),
			}
			for _, query := range queries {
				iter, err := query.Execute()
				require.NoError(err)
				tRequire.verifyIteratorResults(iter, testTuples...)
			}

			// Try some bad queries
			badQueries := []datastore.TupleQuery{
				q.WithObjectID("fakeobjectid"),
				q.WithUserset(&pb.ObjectAndRelation{
					Namespace: "test/user",
					ObjectId:  "fakeuser",
					Relation:  ellipsis,
				}),
			}
			for _, badQuery := range badQueries {
				iter, err := badQuery.Execute()
				require.NoError(err)
				tRequire.verifyIteratorResults(iter)
			}

			// Delete the first tuple
			deletedAt, err := ds.WriteTuples(noPreconditions, []*pb.RelationTupleUpdate{d(testTuples[0])})
			require.NoError(err)

			// Verify it can still be read at the old revision
			tRequire.tupleExists(testTuples[0], deletedAt-1)

			// Verify that it does not show up at the new revision
			tRequire.noTupleExists(testTuples[0], deletedAt)
			alreadyDeletedIter, err := ds.QueryTuples(testTuples[0].ObjectAndRelation.Namespace, deletedAt).Execute()
			require.NoError(err)
			tRequire.verifyIteratorResults(alreadyDeletedIter, testTuples[1:]...)
		})
	}
}

func c(tpl *pb.RelationTuple) *pb.RelationTupleUpdate {
	return &pb.RelationTupleUpdate{
		Operation: pb.RelationTupleUpdate_CREATE,
		Tuple:     tpl,
	}
}

func t(tpl *pb.RelationTuple) *pb.RelationTupleUpdate {
	return &pb.RelationTupleUpdate{
		Operation: pb.RelationTupleUpdate_TOUCH,
		Tuple:     tpl,
	}
}

func d(tpl *pb.RelationTuple) *pb.RelationTupleUpdate {
	return &pb.RelationTupleUpdate{
		Operation: pb.RelationTupleUpdate_DELETE,
		Tuple:     tpl,
	}
}

type tupleChecker struct {
	require *require.Assertions
	ds      datastore.TupleDatastore
}

func (tc tupleChecker) exactTupleIterator(tpl *pb.RelationTuple, rev uint64) datastore.TupleIterator {
	iter, err := tc.ds.QueryTuples(tpl.ObjectAndRelation.Namespace, rev).
		WithObjectID(tpl.ObjectAndRelation.ObjectId).
		WithRelation(tpl.ObjectAndRelation.Relation).
		WithUserset(tpl.User.GetUserset()).
		Execute()

	tc.require.NoError(err)
	return iter
}

func (tc tupleChecker) verifyIteratorResults(iter datastore.TupleIterator, tpls ...*pb.RelationTuple) {
	defer iter.Close()

	toFind := make(map[string]interface{}, 1024)

	for _, tpl := range tpls {
		toFind[tuple.String(tpl)] = struct{}{}
	}

	for found := iter.Next(); found != nil; found = iter.Next() {
		tc.require.NoError(iter.Err())
		foundStr := tuple.String(found)
		_, ok := toFind[foundStr]
		tc.require.True(ok)
		delete(toFind, foundStr)
	}
	tc.require.NoError(iter.Err())

	tc.require.Zero(len(toFind), "Should not be any extra to find")
}

func (tc tupleChecker) tupleExists(tpl *pb.RelationTuple, rev uint64) {
	iter := tc.exactTupleIterator(tpl, rev)
	tc.verifyIteratorResults(iter, tpl)
}

func (tc tupleChecker) noTupleExists(tpl *pb.RelationTuple, rev uint64) {
	iter := tc.exactTupleIterator(tpl, rev)
	tc.verifyIteratorResults(iter)
}
