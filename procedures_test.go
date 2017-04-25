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

import "testing"

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
