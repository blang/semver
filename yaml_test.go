package semver

import (
	"testing"

	"gopkg.in/yaml.v3"
)

const (
	testYaml        = "version: 3.1.4-alpha.1.5.9+build.2.6.5\n"
	testYamlInvalid = "version: 3.1.4.1.5.9.2.6.5-other-digits-of-pi\n"
	versionString   = "3.1.4-alpha.1.5.9+build.2.6.5"
)

type testYamlObject struct {
	Version Version `yaml:"version"`
}

func TestYAMLMarshal(t *testing.T) {
	var v testYamlObject
	var err error

	v.Version, err = Parse(versionString)
	if err != nil {
		t.Fatal(err)
	}

	var versionYAML []byte

	if versionYAML, err = yaml.Marshal(&v); err != nil {
		t.Fatal(err)
	}

	if string(versionYAML) != testYaml {
		t.Fatalf("YAML marshaled semantic version not equal: expected %q, got %q", testYaml, string(versionYAML))
	}
}

func TestYAMLUnmarshalValid(t *testing.T) {
	var v testYamlObject

	if err := yaml.Unmarshal([]byte(testYaml), &v); err != nil {
		t.Fatal(err)
	}

	if v.Version.String() != versionString {
		t.Fatalf("JSON unmarshaled semantic version not equal: expected %q, got %q", versionString, v.Version.String())
	}
}

func TestYAMLUnmarshalInValid(t *testing.T) {
	var v testYamlObject

	if err := yaml.Unmarshal([]byte(testYamlInvalid), &v); err == nil {
		t.Fatal("expected YAML unmarshal error, got nil")
	}
}
