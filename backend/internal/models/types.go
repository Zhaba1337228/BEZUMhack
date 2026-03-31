package models

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONMap implements driver.Valuer and sql.Scanner for JSONB fields
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}
