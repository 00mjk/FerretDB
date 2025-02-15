// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestQueryArraySize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "array-empty"}, {"value", bson.A{}}},
		bson.D{{"_id", "array-one"}, {"value", bson.A{"1"}}},
		bson.D{{"_id", "array-two"}, {"value", bson.A{"1", nil}}},
		bson.D{{"_id", "array-three"}, {"value", bson.A{"1", "2", math.NaN()}}},
		bson.D{{"_id", "string"}, {"value", "12"}},
		bson.D{{"_id", "document"}, {"value", bson.D{{"value", bson.A{"1", "2"}}}}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"int32": {
			filter:      bson.D{{"value", bson.D{{"$size", int32(2)}}}},
			expectedIDs: []any{"array-two"},
		},
		"int64": {
			filter:      bson.D{{"value", bson.D{{"$size", int64(2)}}}},
			expectedIDs: []any{"array-two"},
		},
		"float64": {
			filter:      bson.D{{"value", bson.D{{"$size", float64(2)}}}},
			expectedIDs: []any{"array-two"},
		},
		"Zero": {
			filter:      bson.D{{"value", bson.D{{"$size", 0}}}},
			expectedIDs: []any{"array-empty"},
		},
		"NegativeZero": {
			filter:      bson.D{{"value", bson.D{{"$size", math.Copysign(0, -1)}}}},
			expectedIDs: []any{"array-empty"},
		},
		"NotFound": {
			filter:      bson.D{{"value", bson.D{{"$size", 4}}}},
			expectedIDs: []any{},
		},
		"InvalidType": {
			filter: bson.D{{"value", bson.D{{"$size", bson.D{{"$gt", 1}}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size needs a number`,
			},
		},
		"NotWhole": {
			filter: bson.D{{"value", bson.D{{"$size", 2.1}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size must be a whole number`,
			},
		},
		"NaN": {
			filter: bson.D{{"value", bson.D{{"$size", math.NaN()}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size must be a whole number`,
			},
		},
		"Infinity": {
			filter: bson.D{{"value", bson.D{{"$size", math.Inf(+1)}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size must be a whole number`,
			},
		},
		"Negative": {
			filter: bson.D{{"value", bson.D{{"$size", -1}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `$size may not be negative`,
			},
		},
		"InvalidUse": {
			filter: bson.D{{"$size", 2}},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `unknown top level operator: $size. ` +
					`If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}
