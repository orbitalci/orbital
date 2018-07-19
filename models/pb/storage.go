package pb

import (
	"database/sql/driver"
	"errors"
)

func (x *BuildStatus) Value() (driver.Value, error) {
	intStat := int64(*x)
	return intStat, nil
}

func (x *BuildStatus) Scan(src interface{}) error {
	inty, ok := src.(int64)
	if !ok {
		return errors.New("can't cast to int64")
	}
	*x = BuildStatus(inty)
	return nil
}