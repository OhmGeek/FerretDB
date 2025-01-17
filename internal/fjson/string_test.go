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

var stringTestCases = []testCase{{
	name: "foo",
	v:    pointer.To(String("foo")),
	j:    `"foo"`,
}, {
	name: "empty",
	v:    pointer.To(String("")),
	j:    `""`,
}, {
	name: "zero",
	v:    pointer.To(String("\x00")),
	j:    `"\u0000"`,
}}

func TestString(t *testing.T) {
	t.Parallel()
	testJSON(t, stringTestCases, func() fjsontype { return new(String) })
}

func FuzzString(f *testing.F) {
	fuzzJSON(f, stringTestCases, func() fjsontype { return new(String) })
}

func BenchmarkString(b *testing.B) {
	benchmark(b, stringTestCases, func() fjsontype { return new(String) })
}
