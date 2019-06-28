package pb

import (
	"errors"
	"strings"
)

// MarshalYAML implements a YAML Marshaler for SubCredType
func (i StageResultVal) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

// UnmarshalYAML implements a YAML Unmarshaler for SubCredType
func (i *StageResultVal) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	var err error
	sct, ok := StageResultVal_value[strings.ToUpper(s)]
	if !ok {
		return errors.New("not found in StageResultVal_value map")
	}
	*i = StageResultVal(sct)
	return err
}
