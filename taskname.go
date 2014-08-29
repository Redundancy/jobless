package jobless

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	QUOTE           = `"`
	SEPARATOR       = `.`
	MATCH_ANYTHING  = `*`
	MATCH_RECURSIVE = `**`
)

type TaskName []string

func (t TaskName) String() string {
	return strings.Join(t, SEPARATOR)
}

// JSON saving support
func (t TaskName) MarshalJSON() ([]byte, error) {
	return []byte(QUOTE + t.String() + QUOTE), nil
}

// YAML saving support
func (t TaskName) GetYAML() (tag string, value interface{}) {
	return "", t.String()
}

// JSON loading suppor
func (t *TaskName) UnmarshalJSON(d []byte) error {
	asString := string(d)

	if !strings.HasPrefix(asString, QUOTE) {
		return fmt.Errorf("%v must begin with a quotation mark", asString)
	} else if !strings.HasSuffix(asString, QUOTE) {
		return fmt.Errorf("%v must end with a quotation mark", asString)
	}

	*t = TaskNameFromString(asString[1 : len(asString)-1])
	return nil
}

// Yaml loading support
func (t *TaskName) SetYAML(tag string, value interface{}) bool {
	if str, ok := value.(string); !ok {
		return false
	} else {
		*t = TaskNameFromString(str)
	}
	return true
}

func TaskNameFromString(s string) TaskName {
	return TaskName(strings.Split(s, SEPARATOR))
}

func (t TaskName) IsAncestorOf(o TaskName) bool {
	for i, v := range t {
		if o[i] != v {
			return false
		}
	}

	return true
}

func (t TaskName) Parent() TaskName {
	if t.IsRoot() {
		return t
	} else {
		return t[:len(t)-1]
	}
}

func (t TaskName) IsRoot() bool {
	return len(t) <= 1
}

// Returns true, if this TaskName is the direct parent of o
func (t TaskName) IsParentOf(o TaskName) bool {
	return len(t) == len(o)-1 && t.IsAncestorOf(o)
}

func (t TaskName) IsChildOf(o TaskName) bool {
	return o.IsParentOf(t)
}

// Returns true if this task matches a pattern in another task
// * is a wildcard, ** is a recursive wildcard
// for example:
//  *.*.*test matches A.B.unittest
//  **.unittest matches it too
func (t TaskName) Matches(pattern string) bool {
	return match(t, TaskNameFromString(pattern))
}

func match(toMatch, pattern TaskName) bool {
	len_pattern := len(pattern)
	len_toMatch := len(toMatch)

	switch {
	case len_pattern == 0 && len_toMatch == 0:
		return true
	case len_pattern == 0:
		return false
	case len_toMatch == 0:
		return false

	case len_pattern == 1 && pattern[0] == MATCH_RECURSIVE:
		// universal match
		return true

	case itemMatch(toMatch[0], pattern[0]):
		return match(toMatch[1:], pattern[1:])

	case itemMatch(toMatch[len_toMatch-1], pattern[len_pattern-1]):
		return match(toMatch[:len_toMatch-1], pattern[:len_pattern-1])

	case pattern[0] == MATCH_RECURSIVE:
		// most expensive recursion
		remainingPattern := pattern[1:]
		pattern_len_without_stars := len_pattern - 1

		// match progressively more items, but always
		// leave enough for the remaining pattern items to match
		for i := 1; len_toMatch-i >= pattern_len_without_stars; i++ {
			remainingMatch := toMatch[i:]
			if match(remainingMatch, remainingPattern) {
				return true
			}
		}

		return false
	default:
		return false
	}
}

func itemMatch(a, pattern string) bool {
	switch {
	case a == pattern:
		return true
	case pattern == MATCH_ANYTHING:
		return true
	case pattern == MATCH_RECURSIVE:
		// should not be handled in itemMatch
		return false
	case regularExpressionMatch(a, pattern):
		return true
	default:
		return false
	}
}

func regularExpressionMatch(a, pattern string) bool {
	fixedPattern := strings.Replace(pattern, MATCH_ANYTHING, ".*", -1)
	m, e := regexp.MatchString(fixedPattern, a)
	return e == nil && m
}
