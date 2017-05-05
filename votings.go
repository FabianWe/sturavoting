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
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Voter struct {
	Name   string
	Weight int
}

func NewVoter(name string, weight int) *Voter {
	return &Voter{Name: name, Weight: weight}
}

type SyntaxError struct {
	lineNumber int
	message    string
}

func NewSyntaxError(lineNumber int, message string) *SyntaxError {
	return &SyntaxError{lineNumber: lineNumber, message: message}
}

func (err *SyntaxError) Error() string {
	return fmt.Sprintf("Error in line %d: %s", err.lineNumber, err.message)
}

func ParseVoters(r io.Reader) ([]*Voter, error) {
	res := make([]*Voter, 0)
	scanner := bufio.NewScanner(r)
	lineNum := 1
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			lineNum++
			continue
		}
		// line must start with a *
		if !strings.HasPrefix(line, "*") {
			return nil, NewSyntaxError(lineNum, "Line must start with a *")
		}
		line = line[1:]
		// line must end with : som int
		lastColon := strings.LastIndex(line, ":")
		if lastColon < 0 {
			return nil, NewSyntaxError(lineNum, "Line must contain \": weight\"")
		}
		name, weightStr := strings.TrimSpace(line[:lastColon]), strings.TrimSpace(line[lastColon+1:])
		if utf8.RuneCountInString(name) > 150 {
			return nil, NewSyntaxError(lineNum, "Name must be at most 150 charachters long")
		}
		if name == "" {
			return nil, NewSyntaxError(lineNum, "Name is not allowed to be empty")
		}
		weight, parseErr := strconv.Atoi(weightStr)
		if parseErr != nil {
			return nil, NewSyntaxError(lineNum, parseErr.Error())
		}
		res = append(res, NewVoter(name, weight))
		lineNum++
	}
	return res, nil
}

type MedianVoting struct {
	Name            string
	MaxValue        int
	PercentRequired float64
}

func (voting *MedianVoting) String() string {
	return fmt.Sprintf("MedianVoting(Name=\"%s\", MaxValue=%d, PercentRequired=%.2f)",
		voting.Name, voting.MaxValue, voting.PercentRequired)
}

type SchulzeVoting struct {
	Name            string
	Options         []string
	PercentRequired float64
}

func (voting *SchulzeVoting) String() string {
	optionsStr := make([]string, len(voting.Options))
	for i, option := range voting.Options {
		optionsStr[i] = fmt.Sprintf("\"%s\"", option)
	}
	optionsRepr := strings.Join(optionsStr, ", ")
	return fmt.Sprintf("SchulzeVoting(Name=\"%s\", PercentRequired=%.2f, Options=[%s])",
		voting.Name, voting.PercentRequired, optionsRepr)
}

type VotingGroup struct {
	Name           string
	MedianVotings  []*MedianVoting
	SchulzeVotings []*SchulzeVoting
}

func (group *VotingGroup) String() string {
	var wg sync.WaitGroup
	wg.Add(2)
	medianStrings := make([]string, len(group.MedianVotings))
	schulzeStrings := make([]string, len(group.SchulzeVotings))
	var medianRepr, schulzeRepr string
	go func() {
		defer wg.Done()
		for i, voting := range group.MedianVotings {
			medianStrings[i] = fmt.Sprintf("  %s", voting.String())
		}
		medianRepr = strings.Join(medianStrings, "\n")
	}()

	go func() {
		defer wg.Done()
		for i, voting := range group.SchulzeVotings {
			schulzeStrings[i] = fmt.Sprintf("  %s", voting.String())
		}
		schulzeRepr = strings.Join(schulzeStrings, "\n")
	}()
	wg.Wait()
	return fmt.Sprintf("VotingGroup: \"%s\"\n%s\n%s", group.Name, medianRepr, schulzeRepr)
}

type VotingCollection struct {
	Name   string
	Date   time.Time
	Groups []*VotingGroup
}

func (collection *VotingCollection) String() string {
	groupStrings := make([]string, len(collection.Groups))
	var wg sync.WaitGroup
	wg.Add(len(collection.Groups))
	for i, group := range collection.Groups {
		go func(i int, group *VotingGroup) {
			defer wg.Done()
			groupStrings[i] = group.String()
		}(i, group)
	}
	wg.Wait()
	return fmt.Sprintf("VotingCollection: \"%s\" on %v:\n%s", collection.Name,
		collection.Date, strings.Join(groupStrings, "\n"))
}

type collectionParseState int

const (
	// cStartState is the state when we haven't parsed anything yet,
	// we expect the name in the form
	// # TITLE: date in the format DD.MM.YYYY
	cStartState collectionParseState = iota
	// cTopLevelState is the state when parsing in the top level, in this state
	// we expect a voting group in the form
	// ## GROUP-NAME
	cTopLevelState
	// cGroupState is the state when parsing items inside a group.
	// We expect a voting in the form
	// ### VOTING-NAME
	cGroupState
	// cVotingState is the state when parsing options for a voting.
	// We expect either
	// 1. * VOTING-OPTION to start a Schulze voting
	// 2. - NUMBER to start a media voting
	cVotingState
	// cGroupOrVoting is the state when parsing either a new voting group
	// or a new voting is expected
	cGroupOrVoting
	// cSchulzeOptionsState is the state in which we parse more optionss for a
	// schulze voting, so we expect either
	// 1. * OPTION-NAME
	// A new voting group or a new voting
	cSchulzeOptionsState
)

type votingType int

const (
	medianVoting votingType = iota
	schulzeVoting
)

type schulzeOptionStateRes int

const (
	optionStateSchulzeOption schulzeOptionStateRes = iota
	optionStateVoting
	optionStateGroup
)

var concurrencyRegex = regexp.MustCompile(`^(\d+)([.,](\d{1,2}))?$`)

func parseConcurrency(str string) (int, error) {
	match := concurrencyRegex.FindStringSubmatch(str)
	if match == nil {
		return -1, errors.New("Not a valid number, allowed format is XXXX.XX")
	}
	firstPart, secondPart := match[1], match[3]
	switch len(secondPart) {
	case 0:
		secondPart = "00"
	case 1:
		secondPart += "0"
	}
	return strconv.Atoi(firstPart + secondPart)
}

// TODO check length of input
func ParseVotingCollection(r io.Reader) (*VotingCollection, error) {
	scanner := bufio.NewScanner(r)
	lineNumber := 1
	state := cStartState
	res := &VotingCollection{Name: "", Groups: make([]*VotingGroup, 0)}
	lastVotingName := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			lineNumber++
			continue
		}
		switch state {
		default:
			return nil, errors.New("Invalid state while parsing voting collection")
		case cStartState:
			if name, date, err := handlecStartState(line, lineNumber); err != nil {
				return nil, err
			} else {
				res.Name = name
				res.Date = date
				state = cTopLevelState
			}
		case cTopLevelState:
			if name, err := handlectopLevelState(line, lineNumber); err != nil {
				return nil, err
			} else {
				// append a new group
				group := &VotingGroup{Name: name,
					MedianVotings:  make([]*MedianVoting, 0),
					SchulzeVotings: make([]*SchulzeVoting, 0)}
				res.Groups = append(res.Groups, group)
				state = cGroupState
			}
		case cGroupState:
			if name, err := handlecGroupState(line, lineNumber); err != nil {
				return nil, err
			} else {
				lastVotingName = name
				state = cVotingState
			}
		case cVotingState:
			if str, vType, err := handlecVotingState(line, lineNumber); err != nil {
				return nil, err
			} else {
				// this check should be useless, just to be absolutely sure
				if len(res.Groups) == 0 || lastVotingName == "" {
					return nil, NewSyntaxError(lineNumber, "Got voting option without a valid group or voting name")
				}
				lastGroup := res.Groups[len(res.Groups)-1]
				switch vType {
				default:
					return nil, errors.New("Invalid voting type")
				case schulzeVoting:
					// add a new schulze voting with the last name and the new option
					newVoting := &SchulzeVoting{Name: lastVotingName, Options: []string{str}, PercentRequired: -1.0}
					lastGroup.SchulzeVotings = append(lastGroup.SchulzeVotings, newVoting)
					state = cSchulzeOptionsState
				case medianVoting:
					// value must be a valid concurrency value
					value, err := parseConcurrency(str)
					if err != nil {
						return nil, NewSyntaxError(lineNumber, err.Error())
					}
					newVoting := &MedianVoting{Name: lastVotingName, MaxValue: value, PercentRequired: -1.0}
					lastGroup.MedianVotings = append(lastGroup.MedianVotings, newVoting)
					state = cGroupOrVoting
				}
				lastVotingName = ""
			}
		case cSchulzeOptionsState:
			// expect either a schulze option, a voting or a group
			// code duplicate but anyhow
			if name, resType, err := handlecSchulzeOptionsState(line, lineNumber); err != nil {
				return nil, err
			} else {
				switch resType {
				default:
					return nil, errors.New("Invalid return state while parsing voting collection")
				case optionStateSchulzeOption:
					// add an option to the last schulze voting, assert that there is one
					// again some maybe useless checks
					if len(res.Groups) == 0 {
						return nil, NewSyntaxError(lineNumber, "Got Schulze option without a group")
					}
					lastGroup := res.Groups[len(res.Groups)-1]
					// check last schulze voting
					if len(lastGroup.SchulzeVotings) == 0 {
						return nil, NewSyntaxError(lineNumber, "Got Schulze option without a voting")
					}
					// everything ok, append new option to last voting
					lastVoting := lastGroup.SchulzeVotings[len(lastGroup.SchulzeVotings)-1]
					lastVoting.Options = append(lastVoting.Options, name)
					// state stays the same
				case optionStateVoting:
					lastVotingName = name
					state = cVotingState
				case optionStateGroup:
					group := &VotingGroup{Name: name,
						MedianVotings:  make([]*MedianVoting, 0),
						SchulzeVotings: make([]*SchulzeVoting, 0)}
					res.Groups = append(res.Groups, group)
					state = cGroupState
				}
			}
		case cGroupOrVoting:
			// expect a group or a voting
			if name, isVoting, err := handlecGroupOrVoting(line, lineNumber); err != nil {
				return nil, err
			} else {
				// if it is a voting set the name
				if isVoting {
					lastVotingName = name
					state = cVotingState
				} else {
					// it is a group, so create a new one
					group := &VotingGroup{Name: name,
						MedianVotings:  make([]*MedianVoting, 0),
						SchulzeVotings: make([]*SchulzeVoting, 0)}
					res.Groups = append(res.Groups, group)
					state = cGroupState
				}
			}
		}
		lineNumber++
	}
	return res, nil
}

func validateVotingsString(s string) error {
	if s == "" || utf8.RuneCountInString(s) > 150 {
		return errors.New("Name must be not empty and at most 150 characters long")
	}
	return nil
}

func handlecStartState(line string, lineNumber int) (string, time.Time, error) {
	// in the start state we expect the title for the group
	if !strings.HasPrefix(line, "# ") {
		return "", time.Now(), NewSyntaxError(lineNumber, "Expected a title starting with #")
	}
	// now we expect a colon with a date
	line = strings.TrimSpace(line[1:])
	lastColon := strings.LastIndex(line, ":")
	if lastColon < 0 {
		return "", time.Now(), NewSyntaxError(lineNumber, "Line must contain \": date\"")
	}
	name, dateStr := strings.TrimSpace(line[:lastColon]), strings.TrimSpace(line[lastColon+1:])
	if valErr := validateVotingsString(name); valErr != nil {
		return "", time.Now(), valErr
	}
	date, timeErr := time.Parse("02.01.2006", dateStr)
	if timeErr != nil {
		return "", time.Now(), timeErr
	}
	return name, date, nil
}

func handlectopLevelState(line string, lineNumber int) (string, error) {
	if !strings.HasPrefix(line, "## ") {
		return "", NewSyntaxError(lineNumber, "Expected a group starting with ## ")
	}
	line = strings.TrimSpace(line[2:])
	if valErr := validateVotingsString(line); valErr != nil {
		return "", valErr
	}
	return line, nil
}

func handlecGroupState(line string, lineNumber int) (string, error) {
	if !strings.HasPrefix(line, "### ") {
		return "", NewSyntaxError(lineNumber, "Expected a voting starting with ### ")
	}
	line = strings.TrimSpace(line[3:])
	if valErr := validateVotingsString(line); valErr != nil {
		return "", valErr
	}
	return line, nil
}

func handlecVotingState(line string, lineNumber int) (string, votingType, error) {
	switch {
	default:
		return "", -1, NewSyntaxError(lineNumber, "Expected either * OPTION (for Schulze voting) or - VALUE (for median voting)")
	case strings.HasPrefix(line, "* "):
		line = strings.TrimSpace(line[1:])
		if valErr := validateVotingsString(line); valErr != nil {
			return "", -1, valErr
		}
		return line, schulzeVoting, nil
	case strings.HasPrefix(line, "- "):
		line = strings.TrimSpace(line[1:])
		if valErr := validateVotingsString(line); valErr != nil {
			return "", -1, valErr
		}
		return line, medianVoting, nil
	}
}

// returns true if it is a voting
func handlecGroupOrVoting(line string, lineNumber int) (string, bool, error) {
	// first try to parse a voting
	if name, err := handlecGroupState(line, lineNumber); err == nil {
		return name, true, nil
	}
	// now try to parse a group
	if name, err := handlectopLevelState(line, lineNumber); err == nil {
		return name, false, nil
	}
	return "", false, NewSyntaxError(lineNumber, "Expected a voting or a voting group")
}

func handlecSchulzeOptionsState(line string, lineNumber int) (string, schulzeOptionStateRes, error) {
	// first try to parse as schulze option
	if name, vType, err := handlecVotingState(line, lineNumber); err == nil {
		// check if vtype is schulze, otherwise return error
		if vType != schulzeVoting {
			return "", -1, NewSyntaxError(lineNumber, "Expected a schulze option, found median option")
		}
		// fine, return it
		return name, optionStateSchulzeOption, nil
	}
	// now try to parse as group or voting
	if name, isVoting, err := handlecGroupOrVoting(line, lineNumber); err == nil {
		if isVoting {
			return name, optionStateVoting, nil
		} else {
			return name, optionStateGroup, nil
		}
	}
	return "", -1, NewSyntaxError(lineNumber, "Expected either a schulze option, a voting or a voting group")
}
