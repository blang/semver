package semver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	numbers  string = "0123456789"
	alphas          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-"
	alphanum        = alphas + numbers
)

// SpecVersion is the latest fully supported spec version of semver
var (
	SpecVersion = Version{2, 0, 0, nil, nil}

	ErrOutOfBound = errors.New("semver: out-of-bound")
)

// PRVersion represents a PreRelease Version
type PRVersion struct {
	VersionStr string
	VersionNum uint64
	IsNum      bool
}

// Version represents a semver compatible version
type Version struct {
	major uint64
	minor uint64
	patch uint64
	pre   []PRVersion
	build []string // No Precedence
}

// New is an alias for Parse and returns a pointer, parses version string and returns a validated Version or error
func New(s string) (*Version, error) {
	v, err := Parse(s)

	if err != nil {
		return nil, err
	}

	return &v, nil
}

// Make is an alias for Parse, parses version string and returns a validated Version or error
func Make(s string) (Version, error) {
	return Parse(s)
}

// ParseTolerant allows for certain version specifications that do not strictly adhere to semver
// specs to be parsed by this library. It does so by normalizing versions before passing them to
// Parse(). It currently trims spaces, removes a "v" prefix, adds a 0 patch number to versions
// with only major and minor components specified, and removes leading 0s.
func ParseTolerant(s string) (Version, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")

	// Split into major.minor.(patch+pr+meta)
	parts := strings.SplitN(s, ".", 3)
	// Remove leading zeros.
	for i, p := range parts {
		if len(p) > 1 {
			p = strings.TrimLeft(p, "0")
			if len(p) == 0 || !strings.ContainsAny(p[0:1], "0123456789") {
				p = "0" + p
			}
			parts[i] = p
		}
	}
	// Fill up shortened versions.
	if len(parts) < 3 {
		if strings.ContainsAny(parts[len(parts)-1], "+-") {
			return Version{}, errors.New("short version cannot contain PreRelease/Build metadata")
		}
		for len(parts) < 3 {
			parts = append(parts, "0")
		}
	}
	s = strings.Join(parts, ".")

	return Parse(s)
}

// Parse parses version string and returns a validated Version or error
func Parse(s string) (Version, error) {
	if len(s) == 0 {
		return Version{}, errors.New("version string empty")
	}

	var err error
	v := Version{}

	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")

	// Split into major.minor.(patch+pr+meta)
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return Version{}, errors.New("no Major.Minor.Patch elements found")
	}

	// Major
	if !containsOnly(parts[0], numbers) {
		return Version{}, fmt.Errorf("invalid character(s) found in major number %q", parts[0])
	}
	if hasLeadingZeroes(parts[0]) {
		return Version{}, fmt.Errorf("major number must not contain leading zeroes %q", parts[0])
	}

	if v.major, err = strconv.ParseUint(parts[0], 10, 64); err != nil {
		return Version{}, err
	}

	// Minor
	if !containsOnly(parts[1], numbers) {
		return Version{}, fmt.Errorf("invalid character(s) found in minor number %q", parts[1])
	}
	if hasLeadingZeroes(parts[1]) {
		return Version{}, fmt.Errorf("minor number must not contain leading zeroes %q", parts[1])
	}

	if v.minor, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
		return Version{}, err
	}

	var build, prerelease []string
	patchStr := parts[2]

	if buildIndex := strings.IndexRune(patchStr, '+'); buildIndex != -1 {
		build = strings.Split(patchStr[buildIndex+1:], ".")
		patchStr = patchStr[:buildIndex]
	}

	if preIndex := strings.IndexRune(patchStr, '-'); preIndex != -1 {
		prerelease = strings.Split(patchStr[preIndex+1:], ".")
		patchStr = patchStr[:preIndex]
	}

	if !containsOnly(patchStr, numbers) {
		return Version{}, fmt.Errorf("invalid character(s) found in patch number %q", patchStr)
	}
	if hasLeadingZeroes(patchStr) {
		return Version{}, fmt.Errorf("patch number must not contain leading zeroes %q", patchStr)
	}

	if v.patch, err = strconv.ParseUint(patchStr, 10, 64); err != nil {
		return Version{}, err
	}

	// Prerelease
	for _, prstr := range prerelease {
		parsedPR, err := NewPRVersion(prstr)
		if err != nil {
			return Version{}, err
		}
		v.pre = append(v.pre, parsedPR)
	}

	// Build meta data
	for _, str := range build {
		if len(str) == 0 {
			return Version{}, errors.New("build metadata is empty")
		}
		if !containsOnly(str, alphanum) {
			return Version{}, fmt.Errorf("invalid character(s) found in build metadata %q", str)
		}
		v.build = append(v.build, str)
	}

	return v, nil
}

// MustParse is like Parse but panics if the version cannot be parsed.
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(`semver: Parse(` + s + `): ` + err.Error())
	}
	return v
}

// NewPRVersion creates a new valid prerelease version
func NewPRVersion(s string) (PRVersion, error) {
	if len(s) == 0 {
		return PRVersion{}, errors.New("prerelease is empty")
	}

	v := PRVersion{}
	if containsOnly(s, numbers) {
		if hasLeadingZeroes(s) {
			return PRVersion{}, fmt.Errorf("numeric PreRelease version must not contain leading zeroes %q", s)
		}
		num, err := strconv.ParseUint(s, 10, 64)

		// Might never be hit, but just in case
		if err != nil {
			return PRVersion{}, err
		}
		v.VersionNum = num
		v.IsNum = true
	} else if containsOnly(s, alphanum) {
		v.VersionStr = s
		v.IsNum = false
	} else {
		return PRVersion{}, fmt.Errorf("invalid character(s) found in prerelease %q", s)
	}
	return v, nil
}

// Version to string
func (v Version) String() string {
	b := make([]byte, 0, 5)
	b = strconv.AppendUint(b, v.major, 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, v.minor, 10)
	b = append(b, '.')
	b = strconv.AppendUint(b, v.patch, 10)

	if pre := v.PrerelString(); pre != "" {
		b = append(b, '-')
		b = append(b, pre...)
	}

	if build := v.BuildString(); build != "" {
		b = append(b, '+')
		b = append(b, build...)
	}

	return string(b)
}

func (v Version) Major() uint64 {
	return v.major
}

func (v Version) Minor() uint64 {
	return v.minor
}

func (v Version) Patch() uint64 {
	return v.patch
}

func (v *Version) SetMajor(val uint64) {
	v.major = val
}

func (v *Version) SetMinor(val uint64) {
	v.minor = val
}

func (v *Version) SetPatch(val uint64) {
	v.patch = val
}

func (v *Version) SetPrerel(val []PRVersion) {
	v.pre = make([]PRVersion, len(val))
	copy(v.pre, val)
}

func (v *Version) SetBuild(val []string) {
	v.build = make([]string, len(val))
	copy(v.build, val)
}

func (v Version) Prerel() []PRVersion {
	res := make([]PRVersion, len(v.pre))
	copy(res, v.pre)
	return res
}

func (v Version) Build() []string {
	res := make([]string, len(v.build))
	copy(res, v.build)
	return res
}

func (v Version) PrerelString() string {
	if len(v.pre) == 0 {
		return ""
	}

	res := v.pre[0].String()

	for _, pre := range v.pre[1:] {
		res += "." + pre.String()
	}

	return res
}

func (v Version) BuildString() string {
	if len(v.build) == 0 {
		return ""
	}

	res := v.build[0]

	for _, build := range v.build[1:] {
		res += "." + build
	}

	return res
}

// IncrementPatch increments the patch version
func (v *Version) IncrementPatch() error {
	if v.patch == ^uint64(0) {
		return ErrOutOfBound
	}
	v.patch += 1
	return nil
}

// IncrementMinor increments the minor version
func (v *Version) IncrementMinor() error {
	if v.minor == ^uint64(0) {
		return ErrOutOfBound
	}

	v.minor += 1
	v.patch = 0
	return nil
}

// IncrementMajor increments the major version
func (v *Version) IncrementMajor() error {
	if v.major == ^uint64(0) {
		return ErrOutOfBound
	}

	v.major += 1
	v.minor = 0
	v.patch = 0
	return nil
}

// Validate validates v and returns error in case
func (v Version) Validate() error {
	// Major, Minor, Patch already validated using uint64

	for _, pre := range v.pre {
		if !pre.IsNum { // Numeric prerelease versions already uint64
			if len(pre.VersionStr) == 0 {
				return fmt.Errorf("prerelease cannot be empty %q", pre.VersionStr)
			}
			if !containsOnly(pre.VersionStr, alphanum) {
				return fmt.Errorf("invalid character(s) found in prerelease %q", pre.VersionStr)
			}
		}
	}

	for _, build := range v.build {
		if len(build) == 0 {
			return fmt.Errorf("build metadata cannot be empty %q", build)
		}

		if !containsOnly(build, alphanum) {
			return fmt.Errorf("invalid character(s) found in build metadata %q", build)
		}
	}

	return nil
}

// IsNumeric checks if prerelease-version is numeric
func (v PRVersion) IsNumeric() bool {
	return v.IsNum
}

// PreRelease version to string
func (v PRVersion) String() string {
	if v.IsNum {
		return strconv.FormatUint(v.VersionNum, 10)
	}
	return v.VersionStr
}

func containsOnly(s string, set string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(set, r)
	}) == -1
}

func hasLeadingZeroes(s string) bool {
	return len(s) > 1 && s[0] == '0'
}

// NewBuildVersion creates a new valid build version
func NewBuildVersion(s string) (string, error) {
	if len(s) == 0 {
		return "", errors.New("build version is empty")
	}

	if !containsOnly(s, alphanum) {
		return "", fmt.Errorf("invalid character(s) found in build metadata %q", s)
	}

	return s, nil
}
