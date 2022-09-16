package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func NewField(field ProjectField, value string) (*Field, error) {
	f := Field{
		ID:       field.ID,
		DataType: field.DataType,
	}

	switch field.DataType {
	case "DATE":
		_, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, fmt.Errorf("invalid date for field %q: %v", field.Name, value)
		}
		f.Value.Date = value
	case "ITERATION":
		for _, iter := range field.Configuration.Iterations {
			if strings.EqualFold(value, iter.Name) {
				f.Value.IterationID = iter.ID
				return &f, nil
			}
		}
		return nil, fmt.Errorf("iteration not defined for field %q: %v", field.Name, value)
	case "NUMBER":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number for field %q: %v", field.Name, value)
		}
		f.Value.Number = v
	case "SINGLE_SELECT":
		for _, opt := range field.Options {
			if strings.EqualFold(value, opt.Name) {
				f.Value.SingleSelectOptionID = opt.ID
				return &f, nil
			}
		}
		return nil, fmt.Errorf("option not defined for field %q: %v", field.Name, value)
	default:
		f.Value.Text = value
	}

	return &f, nil
}

type Field struct {
	ID       string
	DataType string
	Value    FieldValue
}

type FieldValue struct {
	Date                 string  `json:"date,omitempty"`
	IterationID          string  `json:"iterationId,omitempty"`
	Number               float64 `json:"number,omitempty"`
	SingleSelectOptionID string  `json:"singleSelectOptionId,omitempty"`
	Text                 string  `json:"text,omitempty"`
}

type ProjectField struct {
	ID            string
	Name          string
	DataType      string
	Configuration struct {
		Iterations []struct {
			ID   string
			Name string
		}
	}
	Options []struct {
		ID   string
		Name string
	}
}
