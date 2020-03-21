package semver

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

var _ sql.Scanner = (*Version)(nil)
var _ driver.Value = (*Version)(nil)

// Scan implements the database/sql.Scanner interface.
func (v *Version) Scan(src interface{}) error {
	var str string
	switch tSrc := src.(type) {
	case string:
		str = tSrc
	case []byte:
		str = string(tSrc)
	default:
		return fmt.Errorf("version.Scan: cannot convert %T to string", src)
	}

	t, err := Parse(str)

	if err != nil {
		return err
	}

	*v = t

	return nil
}

// Value implements the database/sql/driver.Valuer interface.
func (v Version) Value() (driver.Value, error) {
	return v.String(), nil
}
