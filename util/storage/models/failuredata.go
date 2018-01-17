package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)
type FailureData map[string]interface{}


func (f FailureData) Value() (driver.Value, error) {
	j, err := json.Marshal(f)
	return j, err
}

// source: http://coussej.github.io/2016/02/16/Handling-JSONB-in-Go-Structs/
func (f *FailureData) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*f, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("Type assertion .(map[string]interface{}) failed.")
	}

	return nil
}
