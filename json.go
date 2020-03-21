package semver

import (
	"encoding/json"
)

var _ json.Marshaler = (*Version)(nil)
var _ json.Unmarshaler = (*Version)(nil)

// MarshalJSON implements the encoding/json.Marshaler interface.
func (v Version) MarshalJSON() ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}

	return json.Marshal(v.String())
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface.
func (v *Version) UnmarshalJSON(data []byte) error {
	var versionString string
	var err error

	if err = json.Unmarshal(data, &versionString); err != nil {
		return err
	}

	if *v, err = Parse(versionString); err != nil {
		return err
	}

	return nil
}
