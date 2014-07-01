package semver

import (
	"testing"
)

func TestParse(t *testing.T) {

}

type stringerTest struct {
	v      Version
	result string
}

var stringerTests = []stringerTest{
	{Version{1, 2, 3, nil, nil}, "1.2.3"},
	{Version{0, 0, 1, nil, nil}, "0.0.1"},
}

func TestStringer(t *testing.T) {
	for _, test := range stringerTests {
		if res := test.v.String(); res != test.result {
			t.Errorf("Stringer, expected %q but got %q", test.result, res)
		}
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
}

func TestCompare(t *testing.T) {
	for _, test := range compareTests {
		if res := test.v1.compare(&test.v2); res != test.result {
			t.Errorf("Comparing %q : %q, expected %d but got %d", test.v1, test.v2, test.result, res)
		}
		//Test counterpart
		if res := test.v2.compare(&test.v1); res != -test.result {
			t.Errorf("Comparing %q : %q, expected %d but got %d", test.v2, test.v1, -test.result, res)
		}
	}
}
