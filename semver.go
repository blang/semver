package semver

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var SEMVER_SPEC_VERSION = Version{
	Major: 2,
	Minor: 0,
	Patch: 0,
}

type Version struct {
	Major uint64
	Minor uint64
	Patch uint64
	Pre   []*PRVersion
	Build []string //No Precendence
}

func (v *Version) String() string {
	var buf bytes.Buffer
	var DOT = []byte(".")
	var HYPHEN = []byte("-")
	var PLUS = []byte("+")
	buf.WriteString(strconv.FormatUint(v.Major, 10))
	buf.Write(DOT)
	buf.WriteString(strconv.FormatUint(v.Minor, 10))
	buf.Write(DOT)
	buf.WriteString(strconv.FormatUint(v.Patch, 10))
	if len(v.Pre) > 0 {
		buf.Write(HYPHEN)
		for i, pre := range v.Pre {
			if i > 0 {
				buf.Write(DOT)
			}
			buf.WriteString(pre.String())
		}
	}
	if len(v.Build) > 0 {
		buf.Write(PLUS)
		for i, build := range v.Build {
			if i > 0 {
				buf.Write(DOT)
			}
			buf.WriteString(build)
		}
	}
	return buf.String()
}

func (v *Version) Compare(o *Version) int {
	if v.Major != o.Major {
		if v.Major > o.Major {
			return 1
		} else {
			return -1
		}
	}
	if v.Minor != o.Minor {
		if v.Minor > o.Minor {
			return 1
		} else {
			return -1
		}
	}
	if v.Patch != o.Patch {
		if v.Patch > o.Patch {
			return 1
		} else {
			return -1
		}
	}

	// Quick comparison if a version has no prerelease versions
	if len(v.Pre) == 0 && len(o.Pre) == 0 {
		return 0
	} else if len(v.Pre) == 0 && len(o.Pre) > 0 {
		return 1
	} else if len(v.Pre) > 0 && len(o.Pre) == 0 {
		return -1
	} else {

		i := 0
		for ; i < len(v.Pre) && i < len(o.Pre); i++ {
			if comp := v.Pre[i].Compare(o.Pre[i]); comp == 0 {
				continue
			} else if comp == 1 {
				return 1
			} else {
				return -1
			}
		}

		// If all pr versions are the equal but one has further prversion, this one greater
		if i == len(v.Pre) && i == len(o.Pre) {
			return 0
		} else if i == len(v.Pre) && i < len(o.Pre) {
			return -1
		} else {
			return 1
		}

	}
}

func Parse(s string) (*Version, error) {
	if len(s) == 0 {
		return nil, errors.New("Version string empty")
	}

	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return nil, errors.New("No Major.Minor.Patch elements found")
	}

	if !containsOnly(parts[0], NUMBERS) {
		return nil, fmt.Errorf("Invalid character(s) found in major number %q", parts[0])
	}
	major, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}

	if !containsOnly(parts[1], NUMBERS) {
		return nil, fmt.Errorf("Invalid character(s) found in minor number %q", parts[1])
	}
	minor, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	preIndex := strings.Index(parts[2], "-")
	buildIndex := strings.Index(parts[2], "+")

	var subVersionIndex int
	if preIndex != -1 && buildIndex == -1 {
		subVersionIndex = preIndex
	} else if preIndex == -1 && buildIndex != -1 {
		subVersionIndex = buildIndex
	} else if preIndex == -1 && buildIndex == -1 {
		subVersionIndex = len(parts[2])
	} else {
		// if there is no actual preIndex but a hyphen inside the build meta data
		if buildIndex < preIndex {
			subVersionIndex = buildIndex
			preIndex = -1 // Build meta data before preIndex found implicates there are no prerelease versions
		} else {
			subVersionIndex = preIndex
		}
	}

	if !containsOnly(parts[2][:subVersionIndex], NUMBERS) {
		return nil, fmt.Errorf("Invalid character(s) found in patch number %q", parts[2][:subVersionIndex])
	}
	patch, err := strconv.ParseUint(parts[2][:subVersionIndex], 10, 64)
	if err != nil {
		return nil, err
	}
	v := &Version{}
	v.Major = major
	v.Minor = minor
	v.Patch = patch

	// There are PreRelease versions
	if preIndex != -1 {
		var preRels string
		if buildIndex != -1 {
			preRels = parts[2][subVersionIndex+1 : buildIndex]
		} else {
			preRels = parts[2][subVersionIndex+1:]
		}
		prparts := strings.Split(preRels, ".")
		for _, prstr := range prparts {
			parsedPR, err := NewPRVersion(prstr)
			if err != nil {
				return nil, err
			}
			v.Pre = append(v.Pre, parsedPR)
		}
	}

	// There is build meta data
	if buildIndex != -1 {
		buildStr := parts[2][buildIndex+1:]
		buildParts := strings.Split(buildStr, ".")
		for _, str := range buildParts {
			if !containsOnly(str, ALPHAS+NUMBERS) {
				return nil, fmt.Errorf("Invalid character(s) found in build meta data %q", str)
			}
			v.Build = append(v.Build, str)
		}
	}

	return v, nil
}

// PreRelease Version
type PRVersion struct {
	VersionStr string
	VersionNum uint64
	IsNum      bool
}

func NewPRVersion(s string) (*PRVersion, error) {
	v := &PRVersion{}
	if containsOnly(s, NUMBERS) {
		num, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		v.VersionNum = num
		v.IsNum = true
	} else if containsOnly(s, ALPHAS+NUMBERS) {
		v.VersionStr = s
		v.IsNum = false
	} else {
		return nil, fmt.Errorf("Invalid character(s) found in prerelease %q", s)
	}
	return v, nil
}

func (v *PRVersion) IsNumeric() bool {
	return v.IsNum
}

func (v *PRVersion) Compare(o *PRVersion) int {
	if v.IsNum && !o.IsNum {
		return -1
	} else if !v.IsNum && o.IsNum {
		return 1
	} else if v.IsNum && o.IsNum {
		if v.VersionNum == o.VersionNum {
			return 0
		} else if v.VersionNum > o.VersionNum {
			return 1
		} else {
			return -1
		}
	} else { // both are Alphas
		if v.VersionStr == o.VersionStr {
			return 0
		} else if v.VersionStr > o.VersionStr {
			return 1
		} else {
			return -1
		}
	}
}

func (v *PRVersion) String() string {
	if v.IsNum {
		return strconv.FormatUint(v.VersionNum, 10)
	}
	return v.VersionStr
}

const NUMBERS = "0123456789"
const ALPHAS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"

func containsOnly(s string, set string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(set, r)
	}) == -1
}
