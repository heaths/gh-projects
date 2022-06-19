package models

import (
	"encoding/json"
	"fmt"
)

func NewField(id, dataType, settings, value string) Field {
	return Field{
		ID:       id,
		DataType: dataType,
		Value:    value,
		settings: settings,
	}
}

type Field struct {
	ID       string
	DataType string
	Value    string

	settings string
}

func (f *Field) Settings() (settings FieldSettings, err error) {
	switch f.DataType {
	case "ITERATION":
		settings = &IterationFieldSettings{}
		err = json.Unmarshal([]byte(f.settings), &settings)
	case "SINGLE_SELECT":
		settings = &SingleSelectFieldSettings{}
		err = json.Unmarshal([]byte(f.settings), &settings)
	default:
		err = fmt.Errorf("unsupported data type %q", f.DataType)
	}
	return
}

type FieldSettings interface{}

type IterationFieldSettings struct {
	FieldSettings
	Configuration struct {
		Iterations []struct {
			ID    string
			Title string
		}
	}
}

type SingleSelectFieldSettings struct {
	FieldSettings
	Options []struct {
		ID   string
		Name string
	}
}
