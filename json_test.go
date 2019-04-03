package semver

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestJSONMarshal(t *testing.T) {
	versionString := "3.1.4-alpha.1.5.9+build.2.6.5"
	v, err := Parse(versionString)
	if err != nil {
		t.Fatal(err)
	}

	versionJSON, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	quotedVersionString := strconv.Quote(versionString)

	if string(versionJSON) != quotedVersionString {
		t.Fatalf("JSON marshaled semantic version not equal: expected %q, got %q", quotedVersionString, string(versionJSON))
	}
}

func TestJSONUnmarshal(t *testing.T) {
	versionString := "3.1.4-alpha.1.5.9+build.2.6.5"
	quotedVersionString := strconv.Quote(versionString)

	var v Version
	if err := json.Unmarshal([]byte(quotedVersionString), &v); err != nil {
		t.Fatal(err)
	}

	if v.String() != versionString {
		t.Fatalf("JSON unmarshaled semantic version not equal: expected %q, got %q", versionString, v.String())
	}

	badVersionString := strconv.Quote("3.1.4.1.5.9.2.6.5-other-digits-of-pi")
	if err := json.Unmarshal([]byte(badVersionString), &v); err == nil {
		t.Fatal("expected JSON unmarshal error, got nil")
	}

	noMinorOrPatchVersionString := strconv.Quote("3")
	if err := json.Unmarshal([]byte(noMinorOrPatchVersionString), &v); err != nil {
		t.Fatal(err)
	}

	noPatchVersionString := strconv.Quote("3.1")
	if err := json.Unmarshal([]byte(noPatchVersionString), &v); err != nil {
		t.Fatal(err)
	}

	fullVersionString := strconv.Quote("3.2.1")
	if err := json.Unmarshal([]byte(fullVersionString), &v); err != nil {
		t.Fatal(err)
	}
}
