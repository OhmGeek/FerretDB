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

package bson

import (
	"bufio"
	"bytes"
	"encoding/binary"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Timestamp represents BSON Timestamp data type.
type Timestamp types.Timestamp

func (ts *Timestamp) bsontype() {}

// ReadFrom implements bsontype interface.
func (ts *Timestamp) ReadFrom(r *bufio.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, ts); err != nil {
		return lazyerrors.Errorf("bson.Timestamp.ReadFrom (binary.Read): %w", err)
	}

	return nil
}

// WriteTo implements bsontype interface.
func (ts Timestamp) WriteTo(w *bufio.Writer) error {
	v, err := ts.MarshalBinary()
	if err != nil {
		return lazyerrors.Errorf("bson.Timestamp.WriteTo: %w", err)
	}

	_, err = w.Write(v)
	if err != nil {
		return lazyerrors.Errorf("bson.Timestamp.WriteTo: %w", err)
	}

	return nil
}

// MarshalBinary implements bsontype interface.
func (ts Timestamp) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, ts)

	return buf.Bytes(), nil
}

// UnmarshalJSON implements bsontype interface.
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	var tsJ fjson.Timestamp
	if err := tsJ.UnmarshalJSON(data); err != nil {
		return err
	}

	*ts = Timestamp(tsJ)
	return nil
}

// MarshalJSON implements bsontype interface.
func (ts Timestamp) MarshalJSON() ([]byte, error) {
	return fjson.Marshal(fromBSON(&ts))
}

// check interfaces
var (
	_ bsontype = (*Timestamp)(nil)
)
