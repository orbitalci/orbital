package pb

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

func (x *BuildStatus) Value() (driver.Value, error) {
	intStat := int64(*x)
	return intStat, nil
}

func (x *BuildStatus) Scan(src interface{}) error {
	fmt.Println("TRYING TO SCAN")
	fmt.Print(src)
	st, ok := src.(*BuildStatus)
	if !ok {
		return errors.New("unable to cast source to BuildStatus")
	}
	x = st
	return nil
}