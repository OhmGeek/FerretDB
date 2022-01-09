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
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/jackc/pgx/v4"
)

func (h *storage) MsgListIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	// TODO does this really apply to the sql storage plugin (rather than jsonb)?
	h.l.Info("Entered MsgListIndexes")
	document, err := msg.Document()
	if err != nil {
		h.l.Error("Error when trying to create indexes")
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	collection := m["listIndexes"].(string)
	db := m["$db"].(string)

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	rows, err := h.pgPool.Query(ctx, "SELECT indexname, indexdef FROM pg_indexes where schemaname=$1 AND tablename=$2", pgx.Identifier{db}.Sanitize(), pgx.Identifier{collection}.Sanitize())

	if err != nil {
		// TODO handle this properly.
		return nil, lazyerrors.Error(err)
	}

	defer rows.Close()

	var indexes []types.Document

	for rows.Next() {
		var idxName string
		var idxDef string

		rows.Scan(&idxName, &idxDef)

		var idxUnique = strings.Contains(idxDef, "UNIQUE")
		var idxBackground = strings.Contains(idxDef, "BACKGROUND")

		indexes = append(indexes, types.Document(types.MustMakeDocument(
			"name", idxName,
			"ns", db+"."+collection,
			"background", idxBackground,
			"unique", idxUnique,
			// TODO support hidden indexes
			"hidden", false,
		)))
	}
	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"ok", float64(1),
			"cursor", types.MustMakeDocument(
				"ns", db+"."+collection,
				"firstBatch", indexes,
				"id", 0,
			),
		)},
	})

	return &reply, err
}
