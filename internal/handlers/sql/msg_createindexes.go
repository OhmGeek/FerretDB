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

package sql

import (
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/jackc/pgx/v4"
)

func (h *storage) MsgCreateIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	h.l.Info("Entered MsgCreateIndexes")
	document, err := msg.Document()
	if err != nil {
		h.l.Error("Error when trying to create indexes")
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	collection := m[document.Command()].(string)
	db := m["$db"].(string)
	indexes, _ := m["indexes"].(*types.Array)

	// Get the index count BEFORE!
	var indexCountBefore int
	err = h.pgPool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM indexes WHERE tablename=$1", collection)).Scan(indexCountBefore)

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.l.Info("Trying to create index db=%s, coll=%s, idx=%s", db, collection, indexes)

	// create an index for each specified
	var sql string
	for i := 0; i < indexes.Len(); i++ {
		idx, err := indexes.Get(i)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		i := idx.(*types.Document).Map()

		sql := fmt.Sprintf("CREATE INDEX %s ON %s (", i["name"], pgx.Identifier{db, collection})

		keys, _ := i["key"].(*types.Document)

		// This is wrong. We need to improve this.
		var args []any

		for v, k := range keys.Keys() {
			if len(args) != 0 {
				sql += ", "
			}

			sql += pgx.Identifier{k}.Sanitize()
			args = append(args, v)
		}

		sql += ")"

		fmt.Printf("sql: %v\n", sql)
		_, err = h.pgPool.Exec(ctx, sql, args...)

		if err != nil {
			return nil, err
		}
	}

	var res wire.OpMsg
	err = res.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"ok", float64(1),
			"numIndexesBefore", indexCountBefore,
			"note", fmt.Sprintf("sql: %s", sql),
		)},
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return &res, nil
}
