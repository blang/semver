package semver

import (
	"gopkg.in/yaml.v3"
)

var _ yaml.Marshaler = (*Version)(nil)
var _ yaml.Unmarshaler = (*Version)(nil)

// MarshalYAML implements the encoding/json.Marshaler interface.
func (v Version) MarshalYAML() (interface{}, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}

	return v.String(), nil
}

// UnmarshalYAML implements the encoding/json.Unmarshaler interface.
func (v *Version) UnmarshalYAML(value *yaml.Node) error {
	var err error

	if *v, err = Parse(value.Value); err != nil {
		return err
	}

	return nil
}
