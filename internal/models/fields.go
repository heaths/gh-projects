package models

import (
	"encoding/json"
	"fmt"
)

type Field struct {
	ID string
	// Name     string
	DataType string
	Settings string
}

func (f *Field) UnmarshalSettings() (v interface{}, err error) {
	switch f.DataType {
	case "SINGLE_SELECT":
		v = &SingleSelectFieldSettings{}
		err = json.Unmarshal([]byte(f.Settings), &v)
	default:
		err = fmt.Errorf("unsupported data type %q", f.DataType)
	}
	return
}

type SingleSelectFieldSettings struct {
	Options []struct {
		ID   string
		Name string
	}
}
