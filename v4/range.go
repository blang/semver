package semver

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type wildcardType int

const (
	noneWildcard  wildcardType = iota
	majorWildcard wildcardType = 1
	minorWildcard wildcardType = 2
	patchWildcard wildcardType = 3
)

func wildcardTypefromInt(i int) wildcardType {
	switch i {
	case 1:
		return majorWildcard
	case 2:
		return minorWildcard
	case 3:
		return patchWildcard
	default:
		return noneWildcard
	}
}

//----------------------------------------------------------------------------------------------------------//
//
//----------------------------------------------------------------------------------------------------------//

type comparator uint8

const (
	compInvalid comparator = iota
	compEQ
	compNE
	compGT
	compGE
	compLT
	compLE
)

func parseComparator(s string) comparator {
	switch s {
	case "==", "", "=":
		return compEQ
	case ">":
		return compGT
	case ">=":
		return compGE
	case "<":
		return compLT
	case "<=":
		return compLE
	case "!", "!=":
		return compNE
	}

	return compInvalid
}

func (c comparator) compare(v1 Version, v2 Version) bool {
	switch c {
	case compEQ:
		return v1.Compare(v2) == 0
	case compNE:
		return v1.Compare(v2) != 0
	case compGT:
		return v1.Compare(v2) == 1
	case compGE:
		return v1.Compare(v2) >= 0
	case compLT:
		return v1.Compare(v2) == -1
	case compLE:
		return v1.Compare(v2) <= 0
	default:
		panic("invalid comparator")
	}
}

func (c comparator) String() string {
	switch c {
	case compEQ:
		return ""
	case compNE:
		return "!"
	case compGT:
		return ">"
	case compGE:
		return ">="
	case compLT:
		return "<"
	case compLE:
		return "<="
	default:
		panic("invalid comparator")
	}
}

//----------------------------------------------------------------------------------------------------------//
//
//----------------------------------------------------------------------------------------------------------//

type versionRange struct {
	v Version
	c comparator
}

// rangeFunc creates a Range from the given versionRange.
func (vr *versionRange) Range(v Version) bool {
	return vr.c.compare(v, vr.v)
}

func (vr *versionRange) String() string {
	return vr.c.String() + vr.v.String()
}

// Range represents a range of versions.
// Ranger is slightly different from Range, that it can be converted back to string.
// A Ranger can be used to check if a Version satisfies it:
//
//     r, err := semver.ParseRanger(">1.0.0 <2.0.0")
//     r.Range(semver.MustParse("1.1.1") // returns true
//     r.String() // returns ">1.0.0 <2.0.0" back
type Ranger interface {
	Range(b Version) bool
	fmt.Stringer
}

type andRange []Ranger

func (r andRange) Range(v Version) bool {
	for _, i := range r {
		if !i.Range(v) {
			return false
		}
	}
	return true
}

func (r andRange) String() string {
	strs := make([]string, len(r))
	for i, item := range r {
		strs[i] = item.String()
	}

	return strings.Join(strs, " ")
}

type orRange []Ranger

func (r orRange) Range(v Version) bool {
	for _, i := range r {
		if i.Range(v) {
			return true
		}
	}
	return false
}

func (r orRange) String() string {
	strs := make([]string, len(r))
	for i, item := range r {
		strs[i] = item.String()
	}

	return strings.Join(strs, " || ")
}

// MustParseRange is like ParseRange but panics if the range cannot be parsed.
func MustParseRange(s string) Ranger {
	r, err := ParseRange(s)
	if err != nil {
		panic(`semver: ParseRange(` + s + `): ` + err.Error())
	}
	return r
}

// ParseRange parses a range and returns a Range.
// If the range could not be parsed an error is returned.
//
// Valid ranges are:
//   - "<1.0.0"
//   - "<=1.0.0"
//   - ">1.0.0"
//   - ">=1.0.0"
//   - "1.0.0", "=1.0.0", "==1.0.0"
//   - "!1.0.0", "!=1.0.0"
//
// A Range can consist of multiple ranges separated by space:
// Ranges can be linked by logical AND:
//   - ">1.0.0 <2.0.0" would match between both ranges, so "1.1.1" and "1.8.7" but not "1.0.0" or "2.0.0"
//   - ">1.0.0 <3.0.0 !2.0.3-beta.2" would match every version between 1.0.0 and 3.0.0 except 2.0.3-beta.2
//
// Ranges can also be linked by logical OR:
//   - "<2.0.0 || >=3.0.0" would match "1.x.x" and "3.x.x" but not "2.x.x"
//
// AND has a higher precedence than OR. It's not possible to use brackets.
//
// Ranges can be combined by both AND and OR
//
//  - `>1.0.0 <2.0.0 || >3.0.0 !4.2.1` would match `1.2.3`, `1.9.9`, `3.1.1`, but not `4.2.1`, `2.1.1`
func ParseRange(s string) (Ranger, error) {
	parts := splitAndTrim(s)
	orParts, err := splitORParts(parts)
	if err != nil {
		return nil, err
	}
	expandedParts, err := expandWildcardVersion(orParts)
	if err != nil {
		return nil, err
	}
	var andRanges []Ranger
	for _, parts := range expandedParts {
		var ranges []Ranger
		for _, andParts := range parts {
			opStr, vStr, err := splitComparatorVersion(andParts)
			if err != nil {
				return nil, err
			}
			vr, err := buildVersionRange(opStr, vStr)
			if err != nil {
				return nil, fmt.Errorf("Could not parse Range %q: %s", andParts, err)
			}

			ranges = append(ranges, vr)
		}
		switch len(ranges) {
		case 0:
			return nil, errors.New("empty range")
		case 1:
			andRanges = append(andRanges, ranges[0])
		default:
			andRanges = append(andRanges, andRange(ranges))
		}
	}
	switch len(andRanges) {
	case 0:
		return nil, errors.New("empty range")
	case 1:
		return andRanges[0], nil
	default:
		return orRange(andRanges), nil
	}
}

// splitORParts splits the already cleaned parts by '||'.
// Checks for invalid positions of the operator and returns an
// error if found.
func splitORParts(parts []string) ([][]string, error) {
	var ORparts [][]string
	last := 0
	for i, p := range parts {
		if p == "||" {
			if i == 0 {
				return nil, fmt.Errorf("First element in range is '||'")
			}
			ORparts = append(ORparts, parts[last:i])
			last = i + 1
		}
	}
	if last == len(parts) {
		return nil, fmt.Errorf("Last element in range is '||'")
	}
	ORparts = append(ORparts, parts[last:])
	return ORparts, nil
}

// buildVersionRange takes a slice of 2: operator and version
// and builds a versionRange, otherwise an error.
func buildVersionRange(opStr, vStr string) (*versionRange, error) {
	c := parseComparator(opStr)
	if c == compInvalid {
		return nil, fmt.Errorf("Could not parse comparator %q in %q", opStr, strings.Join([]string{opStr, vStr}, ""))
	}
	v, err := Parse(vStr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse version %q in %q: %s", vStr, strings.Join([]string{opStr, vStr}, ""), err)
	}

	return &versionRange{
		v: v,
		c: c,
	}, nil

}

func containsByte(b []byte, c byte) bool {
	return bytes.IndexByte(b, c) >= 0
}

// splitAndTrim splits a range string by spaces and cleans whitespaces
func splitAndTrim(s string) (result []string) {
	last := 0
	var lastChar byte
	excludeFromSplit := []byte{'>', '<', '='}
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' && !containsByte(excludeFromSplit, lastChar) {
			if last < i-1 {
				result = append(result, s[last:i])
			}
			last = i + 1
		} else if s[i] != ' ' {
			lastChar = s[i]
		}
	}
	if last < len(s)-1 {
		result = append(result, s[last:])
	}

	for i, v := range result {
		result[i] = strings.Replace(v, " ", "", -1)
	}

	// parts := strings.Split(s, " ")
	// for _, x := range parts {
	// 	if s := strings.TrimSpace(x); len(s) != 0 {
	// 		result = append(result, s)
	// 	}
	// }
	return
}

// splitComparatorVersion splits the comparator from the version.
// Input must be free of leading or trailing spaces.
func splitComparatorVersion(s string) (string, string, error) {
	i := strings.IndexFunc(s, unicode.IsDigit)
	if i == -1 {
		return "", "", fmt.Errorf("Could not get version from string: %q", s)
	}
	return strings.TrimSpace(s[0:i]), s[i:], nil
}

// getWildcardType will return the type of wildcard that the
// passed version contains
func getWildcardType(vStr string) wildcardType {
	parts := strings.Split(vStr, ".")
	nparts := len(parts)
	wildcard := parts[nparts-1]

	possibleWildcardType := wildcardTypefromInt(nparts)
	if wildcard == "x" {
		return possibleWildcardType
	}

	return noneWildcard
}

// createVersionFromWildcard will convert a wildcard version
// into a regular version, replacing 'x's with '0's, handling
// special cases like '1.x.x' and '1.x'
func createVersionFromWildcard(vStr string) string {
	// handle 1.x.x
	vStr2 := strings.Replace(vStr, ".x.x", ".x", 1)
	vStr2 = strings.Replace(vStr2, ".x", ".0", 1)
	parts := strings.Split(vStr2, ".")

	// handle 1.x
	if len(parts) == 2 {
		return vStr2 + ".0"
	}

	return vStr2
}

// incrementMajorVersion will increment the major version
// of the passed version
func incrementMajorVersion(vStr string) (string, error) {
	parts := strings.Split(vStr, ".")
	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", err
	}
	parts[0] = strconv.Itoa(i + 1)

	return strings.Join(parts, "."), nil
}

// incrementMajorVersion will increment the minor version
// of the passed version
func incrementMinorVersion(vStr string) (string, error) {
	parts := strings.Split(vStr, ".")
	i, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", err
	}
	parts[1] = strconv.Itoa(i + 1)

	return strings.Join(parts, "."), nil
}

// expandWildcardVersion will expand wildcards inside versions
// following these rules:
//
// * when dealing with patch wildcards:
// >= 1.2.x    will become    >= 1.2.0
// <= 1.2.x    will become    <  1.3.0
// >  1.2.x    will become    >= 1.3.0
// <  1.2.x    will become    <  1.2.0
// != 1.2.x    will become    <  1.2.0 >= 1.3.0
//
// * when dealing with minor wildcards:
// >= 1.x      will become    >= 1.0.0
// <= 1.x      will become    <  2.0.0
// >  1.x      will become    >= 2.0.0
// <  1.0      will become    <  1.0.0
// != 1.x      will become    <  1.0.0 >= 2.0.0
//
// * when dealing with wildcards without
// version operator:
// 1.2.x       will become    >= 1.2.0 < 1.3.0
// 1.x         will become    >= 1.0.0 < 2.0.0
func expandWildcardVersion(parts [][]string) ([][]string, error) {
	var expandedParts [][]string
	for _, p := range parts {
		var newParts []string
		for _, ap := range p {
			if strings.Contains(ap, "x") {
				opStr, vStr, err := splitComparatorVersion(ap)
				if err != nil {
					return nil, err
				}

				versionWildcardType := getWildcardType(vStr)
				flatVersion := createVersionFromWildcard(vStr)

				var resultOperator string
				var shouldIncrementVersion bool
				switch opStr {
				case ">":
					resultOperator = ">="
					shouldIncrementVersion = true
				case ">=":
					resultOperator = ">="
				case "<":
					resultOperator = "<"
				case "<=":
					resultOperator = "<"
					shouldIncrementVersion = true
				case "", "=", "==":
					newParts = append(newParts, ">="+flatVersion)
					resultOperator = "<"
					shouldIncrementVersion = true
				case "!=", "!":
					newParts = append(newParts, "<"+flatVersion)
					resultOperator = ">="
					shouldIncrementVersion = true
				}

				var resultVersion string
				if shouldIncrementVersion {
					switch versionWildcardType {
					case patchWildcard:
						resultVersion, _ = incrementMinorVersion(flatVersion)
					case minorWildcard:
						resultVersion, _ = incrementMajorVersion(flatVersion)
					}
				} else {
					resultVersion = flatVersion
				}

				ap = resultOperator + resultVersion
			}
			newParts = append(newParts, ap)
		}
		expandedParts = append(expandedParts, newParts)
	}

	return expandedParts, nil
}

type wildcardSemver struct {
	Major int
	Minor int
	Patch int
	PR    []PRVersion
}
