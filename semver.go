package semver

import (
	"bytes"
	"errors"
	"strconv"
)

var SEMVER_SPEC_VERSION = Version{
	Major: 2,
	Minor: 0,
	Patch: 0,
}

type Version struct {
	Major int
	Minor int
	Patch int
	Pre   []PRVersion
	Build []string //No Precendence
}

func (v Version) String() string {
	var buf bytes.Buffer
	var DOT = []byte(".")
	var HYPHEN = []byte("-")
	var PLUS = []byte("+")
	buf.WriteString(strconv.Itoa(v.Major))
	buf.Write(DOT)
	buf.WriteString(strconv.Itoa(v.Minor))
	buf.Write(DOT)
	buf.WriteString(strconv.Itoa(v.Patch))
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

func (v *Version) compare(o *Version) int {
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

	if len(v.Pre) == 0 && len(o.Pre) == 0 {
		return 0
	} else if len(v.Pre) == 0 && len(o.Pre) > 0 {
		return -1
	} else if len(v.Pre) > 0 && len(o.Pre) == 0 {
		return 1
	} else {
		//Deep PreRelease Version comparison

		return -2 //TODO: Not yet implemented

	}

}

func Parse(s string) (*Version, error) {
	return nil, errors.New("Not implemented yet")
}

// PreRelease Version
type PRVersion interface {
	compare(*PRVersion) int
	String() string //fmt.Stringer
	IsNumeric() bool
}

// Alphabetical PreRelease Version
type AlphaPRVersion struct {
	Version string
}

func (v *AlphaPRVersion) IsNumeric() bool {
	return false
}

func (v *AlphaPRVersion) compare(o *AlphaPRVersion) int {
	if v.Version == o.Version {
		return 0
	} else if v.Version > o.Version {
		return 1
	} else {
		return -1
	}
}

func (v AlphaPRVersion) String() string {
	return v.Version
}

// Numeric PreRelease Version
type NumPRVersion struct {
	Version int
}

func (v *NumPRVersion) compare(o *NumPRVersion) int {
	if v.Version == o.Version {
		return 0
	} else if v.Version > o.Version {
		return 1
	} else {
		return -1
	}
}

func (v NumPRVersion) String() string {
	return strconv.Itoa(v.Version)
}

func (v *NumPRVersion) IsNumeric() bool {
	return true
}
