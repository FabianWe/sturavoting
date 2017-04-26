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
	"fmt"
	"sort"
	"sync"
)

//// Median ////

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

//// Schulze ////

// SchulzeVote is a vote used in the Schulze procedure.
type SchulzeVote struct {
	// Weight is the weight of the voter.
	Weight int
	// Ranking is the ordering for all options.
	// It must be a list of n elements if n is the number of possible options
	// where Ordering[i] is the value in the ranking.
	// Smaller values mean that the option is voted higher (comes first) in the
	// ranking.
	// For example: If there are three options and the first and third one should
	// be equally preferred to option two the ranking would be
	// [0, 1, 0].
	Ranking []int
}

// NewSchulzeVote returns a new SchulzeVote. See struct documentation for
// details.
func NewSchulzeVote(weight int, ranking []int) *SchulzeVote {
	return &SchulzeVote{Weight: weight, Ranking: ranking}
}

// IntMatrix is a quadratic matrix of integer values.
type IntMatrix [][]int

// NewIntMatrix returns a new n x n IntMatrix.
func NewIntMatrix(n int) IntMatrix {
	var res IntMatrix = make([][]int, n)
	for i := 0; i < n; i++ {
		res[i] = make([]int, n)
	}
	return res
}

// Equals compares to n x n IntMatrix instances.
func (m IntMatrix) Equals(other IntMatrix) bool {
	n1, n2 := len(m), len(other)
	if n1 != n2 {
		return false
	}
	n := n1
	for i := 0; i < n; i++ {
		row1, row2 := m[i], other[i]
		for j := 0; j < n; j++ {
			if row1[j] != row2[j] {
				return false
			}
		}
	}
	return true
}

// SchulzeRes is the result returned by the schulze method.
type SchulzeRes struct {
	// VotesRequired is the number of votes required for a majority.
	VotesRequired int
	// D is the matrix d as described in Wikipedia.
	D IntMatrix
	// P is the matrix p as described in Wikipedia.
	P IntMatrix
	// Ranked contains the result of the ranking algorithm.
	// It contains a list of list of integers.
	// The first list contains all options that are winners,
	// the second list contains all options that are on the second place etc.
	Ranked [][]int
	// Percents is a list of length n - 1 containing the percentage of votes
	// that voted the options 0...n-1 before no (no being the last option).
	// So if there are three options and 50% voted option 1 before no and
	// 75% voted option 2 before no this slice will be [0.5, 0.75].
	Percents []float64
}

// EvaluateSchulze evaluates the Schulze method.
// votes contains all votes to be evaluated, n is the number of options in the
// voting (so all votes must have a Ranking slice of length n) and
// percentRequired is a float and should be greater than 0 and lesser than
// 1. It describes how many percents of all votes are required for a majority.
func EvaluateSchulze(votes []*SchulzeVote, n int, percentRequired float64) (*SchulzeRes, error) {
	// first compute votes required, check length of each result while doing this
	weightSum := 0
	for _, vote := range votes {
		weightSum += vote.Weight
		if len(vote.Ranking) != n {
			return nil, fmt.Errorf("Expected ranking of length %d, got length %d", n, len(vote.Ranking))
		}
	}
	votesRequred := int(float64(weightSum) * percentRequired)

	d := computeD(votes, n)
	// compute p and percents
	var p IntMatrix
	var percents []float64
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		p = computeP(d, n)
	}()
	go func() {
		defer wg.Done()
		percents = computePercentage(d, n, weightSum)
	}()
	wg.Wait()
	ranked := rankP(p, n)
	res := &SchulzeRes{VotesRequired: votesRequred, D: d, P: p,
		Ranked: ranked, Percents: percents}
	return res, nil
}

// computeD computes the matrix d as described here:
// http://de.wikipedia.org/wiki/Schulze-Methode#Implementierung
func computeD(votes []*SchulzeVote, n int) IntMatrix {
	res := NewIntMatrix(n)
	for _, vote := range votes {
		w := vote.Weight
		ranking := vote.Ranking
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				switch {
				case ranking[i] < ranking[j]:
					res[i][j] += w
				case ranking[j] < ranking[i]:
					res[j][i] += w
				}
			}
		}
	}
	return res
}

// computePercentage returns the percents slice as defined in SchulzeResult.
func computePercentage(d IntMatrix, n, weightSum int) []float64 {
	if n == 0 {
		return nil
	}
	res := make([]float64, n-1)
	if weightSum == 0 {
		return res
	}
	weightSumF := float64(weightSum)
	for i, row := range d[:n-1] {
		res[i] = float64(row[n-1]) / weightSumF
	}
	return res
}

// computeP computes the matrix p as described here:
// http://de.wikipedia.org/wiki/Schulze-Methode#Implementierung
func computeP(d IntMatrix, n int) IntMatrix {
	res := NewIntMatrix(n)
	// first part: initialize p[i][j]
	// we start a gourtine for each i and set p[i][j] for all j
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < n; j++ {
				if i == j {
					continue
				}
				if d[i][j] > d[j][i] {
					res[i][j] = d[i][j]
				}
			}
		}(i)
	}
	wg.Wait()
	// TODO: is there a concurrent apporach?
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			for k := 0; k < n; k++ {
				if i == k || j == k {
					continue
				}
				res[j][k] = IntMax(res[j][k], IntMin(res[j][i], res[i][k]))
			}
		}
	}
	return res
}

// rankP ranks the matrix p, inspired by
// https://github.com/mgp/schulze-method/blob/master/schulze.py
func rankP(p IntMatrix, n int) [][]int {
	// wait for all i
	var wg sync.WaitGroup
	wg.Add(n)
	candidateWins := make(map[int][]int)
	// mutex used to sync writes to candidateWins
	var mutex sync.Mutex
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			numWins := 0
			for j := 0; j < n; j++ {
				if i == j {
					continue
				}
				if p[i][j] > p[j][i] {
					numWins++
				}
			}
			// get list from the dict and append to it
			mutex.Lock()
			candidateWins[numWins] = append(candidateWins[numWins], i)
			mutex.Unlock()
		}(i)
	}
	wg.Wait()
	// get the keys from the dictionary and sort them (in reverse)
	keys := make([]int, 0, len(candidateWins))
	for key, _ := range candidateWins {
		keys = append(keys, key)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	res := make([][]int, len(keys))
	for i, key := range keys {
		res[i] = candidateWins[key]
	}
	return res
}
