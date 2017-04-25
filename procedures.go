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
	"sort"
)

// MedianVote is used as a vote in a median procedure.
// It contains information about the weight of the voter and the value
// chosen.
type MedianVote struct {
	// Weight is the weight of the voter.
	Weight int

	// Value is the value the voter chose.
	Value int
}

// NewMedianVote returns a new MedianVote.
func NewMedianVote(weight, value int) *MedianVote {
	return &MedianVote{Weight: weight, Value: value}
}

// SortMedianVotes sorts the votes according to the voted value.
// Votes with hightest values come first.
func SortMedianVotes(votes []*MedianVote) {
	// sort votes according to value
	// if value is equal sort according to name
	valueSort := func(i, j int) bool {
		return votes[i].Value > votes[j].Value
	}
	sort.Slice(votes, valueSort)
}

// MedianResult is a result type for median votings.
type MedianResult struct {
	// Value is the value that has a majority.
	Value int
	// VotesRequired is the number of votes required for a majority.
	VotesRequired int
}

// EvaluateMedian evalues all votes given in votes and returns the
// greatest value that has a majority.
// percentRequired is a float and should be greater than 0 and lesser than
// 1. It describes how many percents of all votes are required for a majority.
// It returns 0 for value if no value was agreed upon.
func EvaluateMedian(votes []*MedianVote, percentRequired float64) *MedianResult {
	SortMedianVotes(votes)
	weightSum := 0
	for _, vote := range votes {
		weightSum += vote.Weight
	}
	// votesRequired is the number that must be reached, i.e. the sum of
	// weights for that value are *strictly* greather than this value.
	votesRequired := int(float64(weightSum) * percentRequired)
	weightSoFar := 0
	res := &MedianResult{VotesRequired: votesRequired}
	for _, vote := range votes {
		weightSoFar += vote.Weight
		if weightSoFar > votesRequired {
			res.Value = vote.Value
			break
		}
	}
	return res
}
