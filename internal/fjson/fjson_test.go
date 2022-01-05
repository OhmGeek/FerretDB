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
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name   string
	v      fjsontype
	j      string
	canonJ string // canonical form without extra object fields, zero values, etc.
	jErr   string // unwrapped
}

func testJSON(t *testing.T, testCases []testCase, newFunc func() fjsontype) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotEmpty(t, tc.name, "name should not be empty")
			require.NotEmpty(t, tc.j, "j should not be empty")

			t.Parallel()

			var dst bytes.Buffer
			require.NoError(t, json.Compact(&dst, []byte(tc.j)))
			require.Equal(t, tc.j, dst.String(), "j should be compacted")
			if tc.canonJ != "" {
				dst.Reset()
				require.NoError(t, json.Compact(&dst, []byte(tc.canonJ)))
				require.Equal(t, tc.canonJ, dst.String(), "canonJ should be compacted")
			}

			t.Run("UnmarshalJSON", func(t *testing.T) {
				t.Parallel()

				v := newFunc()
				err := v.UnmarshalJSON([]byte(tc.j))

				if tc.jErr == "" {
					require.NoError(t, err)

					if d, ok := tc.v.(*Double); ok && math.IsNaN(float64(*d)) {
						// NaN != NaN, do special handling
						d, ok = v.(*Double)
						require.True(t, ok, "%#v", v)
						assert.True(t, math.IsNaN(float64(*d)))
					} else {
						assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v)
					}
					return
				}

				require.Error(t, err)
				for {
					e := errors.Unwrap(err)
					if e == nil {
						break
					}
					err = e
				}
				require.Equal(t, tc.jErr, err.Error())
			})

			t.Run("Unmarshal", func(t *testing.T) {
				if tc.jErr != "" {
					t.Skip("tc.jErr is not empty")
				}

				t.Parallel()

				v, err := Unmarshal([]byte(tc.j))
				require.NoError(t, err)
				v = toFJSON(v)

				if d, ok := tc.v.(*Double); ok && math.IsNaN(float64(*d)) {
					// NaN != NaN, do special handling
					d, ok = v.(*Double)
					require.True(t, ok)
					assert.True(t, math.IsNaN(float64(*d)))
				} else {
					assert.Equal(t, tc.v, v, "expected: %s\nactual  : %s", tc.v, v)
					assert.Equal(t, tc.v, v)
				}
			})

			t.Run("MarshalJSON", func(t *testing.T) {
				t.Parallel()

				actualJ, err := tc.v.MarshalJSON()
				require.NoError(t, err)
				expectedJ := tc.j
				if tc.canonJ != "" {
					expectedJ = tc.canonJ
				}
				assert.Equal(t, expectedJ, string(actualJ))
			})

			t.Run("Marshal", func(t *testing.T) {
				t.Parallel()

				actualJ, err := Marshal(fromFJSON(tc.v))
				require.NoError(t, err)
				expectedJ := tc.j
				if tc.canonJ != "" {
					expectedJ = tc.canonJ
				}
				assert.Equal(t, expectedJ, string(actualJ))
			})
		})
	}
}

func fuzzJSON(f *testing.F, testCases []testCase, newFunc func() fjsontype) {
	for _, tc := range testCases {
		f.Add(tc.j)
		if tc.canonJ != "" {
			f.Add(tc.canonJ)
		}
	}

	f.Fuzz(func(t *testing.T, j string) {
		t.Parallel()

		// raw "null" should never reach UnmarshalJSON due to the way encoding/json works
		if j == "null" {
			t.Skip(j)
		}

		// j may not be a canonical form.
		// We can't compare it with MarshalJSON() result directly.
		// Instead, we compare second results.

		v := newFunc()
		if err := v.UnmarshalJSON([]byte(j)); err != nil {
			t.Skip(err)
		}

		// test MarshalJSON
		{
			b, err := v.MarshalJSON()
			require.NoError(t, err)
			j = string(b)
		}

		// test UnmarshalJSON
		{
			actualV := newFunc()
			err := actualV.UnmarshalJSON([]byte(j))
			require.NoError(t, err)
			if d, ok := v.(*Double); ok && math.IsNaN(float64(*d)) {
				// NaN != NaN, do special handling
				d, ok = actualV.(*Double)
				assert.True(t, ok)
				assert.True(t, math.IsNaN(float64(*d)))
			} else {
				assert.Equal(t, v, actualV, "expected: %s\nactual  : %s", v, actualV)
			}
		}
	})
}

func benchmark(b *testing.B, testCases []testCase, newFunc func() fjsontype) {
	for _, tc := range testCases {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.Run("UnmarshalJSON", func(b *testing.B) {
				data := []byte(tc.j)
				var v fjsontype
				var err error

				b.ReportAllocs()
				b.SetBytes(int64(len(data)))
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v = newFunc()
					err = v.UnmarshalJSON(data)
				}

				b.StopTimer()

				assert.NoError(b, err)
				assert.Equal(b, tc.v, v)
			})
		})
	}
}