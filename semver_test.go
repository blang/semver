package semver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func prstr(s string) PRVersion {
	return PRVersion{s, 0, false}
}

func prnum(i uint64) PRVersion {
	return PRVersion{"", i, true}
}

type formatTest struct {
	v      Version
	result string
}

type formatTestTolerant struct {
	v        Version
	result   string
	expected string
}

var formatTests = []formatTest{
	{Version{1, 2, 3, nil, nil}, "1.2.3"},
	{Version{0, 0, 1, nil, nil}, "0.0.1"},
	{Version{0, 0, 1, []PRVersion{prstr("alpha"), prstr("preview")}, []string{"123", "456"}}, "0.0.1-alpha.preview+123.456"},
	{Version{1, 2, 3, []PRVersion{prstr("alpha"), prnum(1)}, []string{"123", "456"}}, "1.2.3-alpha.1+123.456"},
	{Version{1, 2, 3, []PRVersion{prstr("alpha"), prnum(1)}, nil}, "1.2.3-alpha.1"},
	{Version{1, 2, 3, nil, []string{"123", "456"}}, "1.2.3+123.456"},
	// Prereleases and build metadata hyphens
	{Version{1, 2, 3, []PRVersion{prstr("alpha"), prstr("b-eta")}, []string{"123", "b-uild"}}, "1.2.3-alpha.b-eta+123.b-uild"},
	{Version{1, 2, 3, nil, []string{"123", "b-uild"}}, "1.2.3+123.b-uild"},
	{Version{1, 2, 3, []PRVersion{prstr("alpha"), prstr("b-eta")}, nil}, "1.2.3-alpha.b-eta"},
}

var tolerantFormatTests = []formatTestTolerant{
	{Version{1, 2, 3, nil, nil}, "v1.2.3", "1.2.3"},
	{Version{1, 2, 3, nil, nil}, "V1.2.3", "1.2.3"},
	{Version{1, 2, 0, []PRVersion{prstr("alpha")}, nil}, "1.2.0-alpha", "1.2.0-alpha"},
	{Version{1, 2, 0, nil, nil}, "1.2.00", "1.2.0"},
	{Version{1, 2, 3, nil, nil}, "	1.2.3 ", "1.2.3"},
	{Version{1, 2, 3, nil, nil}, "01.02.03", "1.2.3"},
	{Version{0, 0, 3, nil, nil}, "00.0.03", "0.0.3"},
	{Version{0, 0, 3, nil, nil}, "000.0.03", "0.0.3"},
	{Version{1, 2, 0, nil, nil}, "1.2", "1.2.0"},
	{Version{1, 0, 0, nil, nil}, "1", "1.0.0"},
}

func TestStringer(t *testing.T) {
	for _, test := range formatTests {
		res := test.v.String()
		require.Equal(t, test.result, res)
	}
}

func TestParse(t *testing.T) {
	for _, test := range formatTests {
		v, err := Parse(test.result)
		require.NoError(t, err)
		require.Equal(t, test.result, v.String())

		comp := v.Compare(test.v)
		require.Equal(t, 0, comp)

		err = v.Validate()
		require.NoError(t, err)
	}
}

func TestParseTolerant(t *testing.T) {
	for _, test := range tolerantFormatTests {
		v, err := ParseTolerant(test.result)
		require.NoError(t, err)
		require.Equal(t, test.expected, v.String())

		comp := v.Compare(test.v)
		require.Equal(t, 0, comp)

		err = v.Validate()
		require.NoError(t, err)
	}
}

func TestMustParse(t *testing.T) {
	require.NotPanics(t, func() {
		_ = MustParse("32.2.1-alpha")
	})
}

func TestMustParse_panic(t *testing.T) {
	require.Panics(t, func() {
		_ = MustParse("invalid version")
	})
}

func TestValidate(t *testing.T) {
	for _, test := range formatTests {
		err := test.v.Validate()
		require.NoError(t, err)
	}
}

type compareTest struct {
	v1     Version
	v2     Version
	result int
}

var compareTests = []compareTest{
	{Version{1, 0, 0, nil, nil}, Version{1, 0, 0, nil, nil}, 0},
	{Version{2, 0, 0, nil, nil}, Version{1, 0, 0, nil, nil}, 1},
	{Version{0, 1, 0, nil, nil}, Version{0, 1, 0, nil, nil}, 0},
	{Version{0, 2, 0, nil, nil}, Version{0, 1, 0, nil, nil}, 1},
	{Version{0, 0, 1, nil, nil}, Version{0, 0, 1, nil, nil}, 0},
	{Version{0, 0, 2, nil, nil}, Version{0, 0, 1, nil, nil}, 1},
	{Version{1, 2, 3, nil, nil}, Version{1, 2, 3, nil, nil}, 0},
	{Version{2, 2, 4, nil, nil}, Version{1, 2, 4, nil, nil}, 1},
	{Version{1, 3, 3, nil, nil}, Version{1, 2, 3, nil, nil}, 1},
	{Version{1, 2, 4, nil, nil}, Version{1, 2, 3, nil, nil}, 1},

	// Spec Examples #11
	{Version{1, 0, 0, nil, nil}, Version{2, 0, 0, nil, nil}, -1},
	{Version{2, 0, 0, nil, nil}, Version{2, 1, 0, nil, nil}, -1},
	{Version{2, 1, 0, nil, nil}, Version{2, 1, 1, nil, nil}, -1},

	// Spec Examples #9
	{Version{1, 0, 0, nil, nil}, Version{1, 0, 0, []PRVersion{prstr("alpha")}, nil}, 1},
	{Version{1, 0, 0, []PRVersion{prstr("alpha")}, nil}, Version{1, 0, 0, []PRVersion{prstr("alpha"), prnum(1)}, nil}, -1},
	{Version{1, 0, 0, []PRVersion{prstr("alpha"), prnum(1)}, nil}, Version{1, 0, 0, []PRVersion{prstr("alpha"), prstr("beta")}, nil}, -1},
	{Version{1, 0, 0, []PRVersion{prstr("alpha"), prstr("beta")}, nil}, Version{1, 0, 0, []PRVersion{prstr("beta")}, nil}, -1},
	{Version{1, 0, 0, []PRVersion{prstr("beta")}, nil}, Version{1, 0, 0, []PRVersion{prstr("beta"), prnum(2)}, nil}, -1},
	{Version{1, 0, 0, []PRVersion{prstr("beta"), prnum(2)}, nil}, Version{1, 0, 0, []PRVersion{prstr("beta"), prnum(11)}, nil}, -1},
	{Version{1, 0, 0, []PRVersion{prstr("beta"), prnum(11)}, nil}, Version{1, 0, 0, []PRVersion{prstr("rc"), prnum(1)}, nil}, -1},
	{Version{1, 0, 0, []PRVersion{prstr("rc"), prnum(1)}, nil}, Version{1, 0, 0, nil, nil}, -1},

	// Ignore Build metadata
	{Version{1, 0, 0, nil, []string{"1", "2", "3"}}, Version{1, 0, 0, nil, nil}, 0},
}

func TestCompare(t *testing.T) {
	for _, test := range compareTests {
		res := test.v1.Compare(test.v2)
		require.Equal(t, test.result, res)

		// Test counterpart
		res = test.v2.Compare(test.v1)
		require.Equal(t, -test.result, res)
	}
}

type wrongFormatTest struct {
	v   *Version
	str string
}

var wrongFormatTests = []wrongFormatTest{
	{nil, ""},
	{nil, "."},
	{nil, "1."},
	{nil, ".1"},
	{nil, "a.b.c"},
	{nil, "1.a.b"},
	{nil, "1.1.a"},
	{nil, "1.a.1"},
	{nil, "a.1.1"},
	{nil, ".."},
	{nil, "1.."},
	{nil, "1.1."},
	{nil, "1..1"},
	{nil, "1.1.+123"},
	{nil, "1.1.-beta"},
	{nil, "-1.1.1"},
	{nil, "1.-1.1"},
	{nil, "1.1.-1"},
	// giant numbers
	{nil, "20000000000000000000.1.1"},
	{nil, "1.20000000000000000000.1"},
	{nil, "1.1.20000000000000000000"},
	{nil, "1.1.1-20000000000000000000"},
	// Leading zeroes
	{nil, "01.1.1"},
	{nil, "001.1.1"},
	{nil, "1.01.1"},
	{nil, "1.001.1"},
	{nil, "1.1.01"},
	{nil, "1.1.001"},
	{nil, "1.1.1-01"},
	{nil, "1.1.1-001"},
	{nil, "1.1.1-beta.01"},
	{nil, "1.1.1-beta.001"},
	{&Version{0, 0, 0, []PRVersion{prstr("!")}, nil}, "0.0.0-!"},
	{&Version{0, 0, 0, nil, []string{"!"}}, "0.0.0+!"},
	// empty prerelease version
	{&Version{0, 0, 0, []PRVersion{prstr(""), prstr("alpha")}, nil}, "0.0.0-.alpha"},
	// empty build metadata
	{&Version{0, 0, 0, []PRVersion{prstr("alpha")}, []string{""}}, "0.0.0-alpha+"},
	{&Version{0, 0, 0, []PRVersion{prstr("alpha")}, []string{"test", ""}}, "0.0.0-alpha+test."},
}

func TestWrongFormat(t *testing.T) {
	for _, test := range wrongFormatTests {
		_, err := Parse(test.str)
		require.Error(t, err)

		if test.v != nil {
			err = test.v.Validate()
			require.Error(t, err)
		}
	}
}

var wrongTolerantFormatTests = []wrongFormatTest{
	{nil, "1.0+abc"},
	{nil, "1.0-rc.1"},
}

func TestWrongTolerantFormat(t *testing.T) {
	for _, test := range wrongTolerantFormatTests {
		_, err := ParseTolerant(test.str)
		require.Error(t, err)
	}
}

func TestCompareHelper(t *testing.T) {
	v := Version{1, 0, 0, []PRVersion{prstr("alpha")}, nil}
	v1 := Version{1, 0, 0, nil, nil}

	require.True(t, v.EQ(v), "should equal to self")
	require.True(t, v.Equals(v), "should equal to self")
	require.True(t, v1.NE(v), "should not be equal with right-hand side")
	require.True(t, v.GTE(v), "should be greater than or equal to self")
	require.True(t, v.LTE(v), "should be less than or equal to self")
	require.True(t, v.LT(v1), "should be less than right-hand side")
	require.True(t, v.LTE(v1), "should be less than or equal right-hand side")
	require.True(t, v.LE(v1), "should be less than right-hand side")
	require.True(t, v1.GT(v), "should be greater than right-hand side")
	require.True(t, v1.GTE(v), "should be greater than or equal right-hand side")
	require.True(t, v1.GE(v), "should be greater than right-hand side")
}

const (
	MAJOR = iota
	MINOR
	PATCH
)

type incrementTest struct {
	version         Version
	incrementType   int
	expectedVersion Version
}

var incrementTests = []incrementTest{
	{Version{1, 2, 3, nil, nil}, PATCH, Version{1, 2, 4, nil, nil}},
	{Version{1, 2, 3, nil, nil}, MINOR, Version{1, 3, 0, nil, nil}},
	{Version{1, 2, 3, nil, nil}, MAJOR, Version{2, 0, 0, nil, nil}},
	{Version{0, 1, 2, nil, nil}, PATCH, Version{0, 1, 3, nil, nil}},
	{Version{0, 1, 2, nil, nil}, MINOR, Version{0, 2, 0, nil, nil}},
	{Version{0, 1, 2, nil, nil}, MAJOR, Version{1, 0, 0, nil, nil}},
}

var incrementOOBTests = []incrementTest{
	{Version{1, 2, ^uint64(0), nil, nil}, PATCH, Version{}},
	{Version{1, ^uint64(0), 3, nil, nil}, MINOR, Version{}},
	{Version{^uint64(0), 2, 3, nil, nil}, MAJOR, Version{}},
}

func TestIncrements(t *testing.T) {
	for _, test := range incrementTests {
		var originalVersion = Version{
			test.version.major,
			test.version.minor,
			test.version.patch,
			test.version.pre,
			test.version.build,
		}

		var err error
		switch test.incrementType {
		case PATCH:
			err = test.version.IncrementPatch()
		case MINOR:
			err = test.version.IncrementMinor()
		case MAJOR:
			err = test.version.IncrementMajor()
		}

		require.NoError(t, err)
		require.True(t, test.version.NE(originalVersion))
		require.True(t, test.version.GT(originalVersion))
		require.True(t, test.expectedVersion.GT(originalVersion))
	}
}

func TestOOBIncrements(t *testing.T) {
	for _, test := range incrementOOBTests {
		var err error
		switch test.incrementType {
		case PATCH:
			err = test.version.IncrementPatch()
		case MINOR:
			err = test.version.IncrementMinor()
		case MAJOR:
			err = test.version.IncrementMajor()
		}

		require.EqualError(t, err, ErrOutOfBound.Error())
	}
}

func TestPreReleaseVersions(t *testing.T) {
	p, err := NewPRVersion("123")
	require.NoError(t, err)
	require.True(t, p.IsNumeric(), "Expected numeric prerelease")
	require.Equal(t, uint64(123), p.VersionNum)

	p, err = NewPRVersion("alpha")
	require.NoError(t, err)
	require.False(t, p.IsNumeric(), "Expected non-numeric prerelease")
	require.Equal(t, "alpha", p.VersionStr)
}

func TestBuildMetaDataVersions(t *testing.T) {
	_, err := NewBuildVersion("123")
	require.NoError(t, err)

	_, err = NewBuildVersion("build")
	require.NoError(t, err)

	_, err = NewBuildVersion("test?")
	require.Error(t, err)

	_, err = NewBuildVersion("")
	require.Error(t, err)
}

func TestNewHelper(t *testing.T) {
	v, err := New("1.2.3")
	require.NoError(t, err)
	require.NotNil(t, v)
	require.Equal(t, 0, v.Compare(Version{1, 2, 3, nil, nil}))
}

func TestMakeHelper(t *testing.T) {
	v, err := Make("1.2.3")
	require.NoError(t, err)
	require.Equal(t, 0, v.Compare(Version{1, 2, 3, nil, nil}))
}

func BenchmarkParseSimple(b *testing.B) {
	const VERSION = "0.0.1"
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Parse(VERSION)
	}
}

func BenchmarkParseComplex(b *testing.B) {
	const VERSION = "0.0.1-alpha.preview+123.456"
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Parse(VERSION)
	}
}

func BenchmarkParseAverage(b *testing.B) {
	l := len(formatTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Parse(formatTests[n%l].result)
	}
}

func BenchmarkParseTolerantAverage(b *testing.B) {
	l := len(tolerantFormatTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = ParseTolerant(tolerantFormatTests[n%l].result)
	}
}

func BenchmarkStringSimple(b *testing.B) {
	const VERSION = "0.0.1"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = v.String()
	}
}

func BenchmarkStringLarger(b *testing.B) {
	const VERSION = "11.15.2012"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = v.String()
	}
}

func BenchmarkStringComplex(b *testing.B) {
	const VERSION = "0.0.1-alpha.preview+123.456"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = v.String()
	}
}

func BenchmarkStringAverage(b *testing.B) {
	l := len(formatTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = formatTests[n%l].v.String()
	}
}

func BenchmarkValidateSimple(b *testing.B) {
	const VERSION = "0.0.1"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = v.Validate()
	}
}

func BenchmarkValidateComplex(b *testing.B) {
	const VERSION = "0.0.1-alpha.preview+123.456"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = v.Validate()
	}
}

func BenchmarkValidateAverage(b *testing.B) {
	l := len(formatTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = formatTests[n%l].v.Validate()
	}
}

func BenchmarkCompareSimple(b *testing.B) {
	const VERSION = "0.0.1"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v.Compare(v)
	}
}

func BenchmarkCompareComplex(b *testing.B) {
	const VERSION = "0.0.1-alpha.preview+123.456"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v.Compare(v)
	}
}

func BenchmarkCompareAverage(b *testing.B) {
	l := len(compareTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		compareTests[n%l].v1.Compare(compareTests[n%l].v2)
	}
}
