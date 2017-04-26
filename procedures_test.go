// The MIT License (MIT)

// Copyright (c) 2017 Fabian Wenzelmann

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package sturavoting

import (
	"testing"
)

func TestMedianOne(t *testing.T) {
	v1 := NewMedianVote(4, 200)
	v2 := NewMedianVote(3, 1000)
	v3 := NewMedianVote(2, 700)
	v4 := NewMedianVote(2, 500)

	res := EvaluateMedian([]*MedianVote{v1, v2, v3, v4}, 0.5)
	if res.VotesRequired != 5 {
		t.Errorf("Expected 5 required votes in median, got %d", res.VotesRequired)
	}

	if res.Value != 500 {
		t.Errorf("Expected value of 500 in median, got %d", res.Value)
	}
}

func TestMedianTwo(t *testing.T) {
	// Example from stura.org
	v1 := NewMedianVote(1, 0)
	v2 := NewMedianVote(2, 150)
	v3 := NewMedianVote(3, 200)

	res := EvaluateMedian([]*MedianVote{v1, v2, v3}, 0.5)
	if res.VotesRequired != 3 {
		t.Errorf("Expected 3 required votes in median, got %d", res.VotesRequired)
	}

	if res.Value != 150 {
		t.Errorf("Expected value of 150 in median, got %d", res.Value)
	}
}

func TestSchulzeOne(t *testing.T) {
	// example from stura.org
	v1 := NewSchulzeVote(1, []int{0, 0, 0, 0, 0, 1})
	v2 := NewSchulzeVote(2, []int{0, 0, 1, 3, 0, 2})
	v3 := NewSchulzeVote(3, []int{1, 1, 0, 2, 2, 3})
	res, err := EvaluateSchulze([]*SchulzeVote{v1, v2, v3}, 6, 0.5)
	if err != nil {
		t.Error(err)
		return
	}
	expected := [][]int{[]int{0, 0, 2, 5, 3, 6},
		[]int{0, 0, 2, 5, 3, 6},
		[]int{3, 3, 0, 5, 3, 6},
		[]int{0, 0, 0, 0, 0, 4},
		[]int{0, 0, 2, 2, 0, 6},
		[]int{0, 0, 0, 2, 0, 0}}

	if !res.D.Equals(expected) {
		t.Error("Matrix d is wrong")
	}
}

func compareSlices(a, b [][]int) bool {
	n1, n2 := len(a), len(b)
	if n1 != n2 {
		return false
	}
	n := n1
	for i := 0; i < n; i++ {
		row1, row2 := a[i], b[i]
		m1, m2 := len(row1), len(row2)
		if m1 != m2 {
			return false
		}
		m := m1
		for j := 0; j < m; j++ {
			if row1[j] != row2[j] {
				return false
			}
		}
	}
	return true
}

func TestSchulzeTwo(t *testing.T) {
	// Example from Wikipedia:
	// http://de.wikipedia.org/wiki/Schulze-Methode#Beispiel_1
	v1 := NewSchulzeVote(5, []int{0, 2, 1, 4, 3})
	v2 := NewSchulzeVote(5, []int{0, 4, 3, 1, 2})
	v3 := NewSchulzeVote(8, []int{3, 0, 4, 2, 1})
	v4 := NewSchulzeVote(3, []int{1, 2, 0, 4, 3})
	v5 := NewSchulzeVote(7, []int{1, 3, 0, 4, 2})
	v6 := NewSchulzeVote(2, []int{2, 1, 0, 3, 4})
	v7 := NewSchulzeVote(7, []int{4, 3, 1, 0, 2})
	v8 := NewSchulzeVote(8, []int{2, 1, 4, 3, 0})

	res, err := EvaluateSchulze([]*SchulzeVote{v1, v2, v3, v4, v5, v6, v7, v8}, 5, 0.5)
	if err != nil {
		t.Error(err)
		return
	}
	expectedD := [][]int{[]int{0, 20, 26, 30, 22},
		[]int{25, 0, 16, 33, 18},
		[]int{19, 29, 0, 17, 24},
		[]int{15, 12, 28, 0, 14},
		[]int{23, 27, 21, 31, 0}}

	if !res.D.Equals(expectedD) {
		t.Error("Matrix d is wrong")
		return
	}

	expectedP := [][]int{[]int{0, 28, 28, 30, 24},
		[]int{25, 0, 28, 33, 24},
		[]int{25, 29, 0, 29, 24},
		[]int{25, 28, 28, 0, 24},
		[]int{25, 28, 28, 31, 0}}
	if !res.P.Equals(expectedP) {
		t.Error("Matrix p is wrong")
		return
	}
	expectedRanks := [][]int{[]int{4}, []int{0}, []int{2}, []int{1}, []int{3}}
	if !compareSlices(res.Ranked, expectedRanks) {
		t.Error("Ranking is wrong")
	}
}

func TestSchuleThree(t *testing.T) {
	// Example from Wikipedia:
	// http://de.wikipedia.org/wiki/Schulze-Methode#Beispiel_2
	v1 := NewSchulzeVote(3, []int{0, 1, 2, 3})
	v2 := NewSchulzeVote(2, []int{1, 2, 3, 0})
	v3 := NewSchulzeVote(2, []int{3, 1, 2, 0})
	v4 := NewSchulzeVote(2, []int{3, 1, 0, 2})

	res, err := EvaluateSchulze([]*SchulzeVote{v1, v2, v3, v4}, 4, 0.5)
	if err != nil {
		t.Error(err)
		return
	}
	expectedD := [][]int{[]int{0, 5, 5, 3},
		[]int{4, 0, 7, 5},
		[]int{4, 2, 0, 5},
		[]int{6, 4, 4, 0}}

	if !res.D.Equals(expectedD) {
		t.Error("Matrix d is wrong")
		return
	}

	expectedP := [][]int{[]int{0, 5, 5, 5},
		[]int{5, 0, 7, 5},
		[]int{5, 5, 0, 5},
		[]int{6, 5, 5, 0}}
	if !res.P.Equals(expectedP) {
		t.Error("Matrix p is wrong")
		return
	}
}
