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

package fjson

import (
	"testing"

	"github.com/AlekSi/pointer"
)

var boolTestCases = []testCase{{
	name: "false",
	v:    pointer.To(Bool(false)),
	j:    `false`,
}, {
	name: "true",
	v:    pointer.To(Bool(true)),
	j:    `true`,
}}

func TestBool(t *testing.T) {
	t.Parallel()
	testJSON(t, boolTestCases, func() fjsontype { return new(Bool) })
}

func FuzzBool(f *testing.F) {
	fuzzJSON(f, boolTestCases, func() fjsontype { return new(Bool) })
}

func BenchmarkBool(b *testing.B) {
	benchmark(b, boolTestCases, func() fjsontype { return new(Bool) })
}
