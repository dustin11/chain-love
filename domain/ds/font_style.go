package ds

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// FontStyle 是 Book 的值对象 实现 Value/Scan 以便直接存为 JSON 列
type FontStyle struct {
	Size       float64 `json:"Size"`
	Url        *string `json:"Url,omitempty"`
	Family     string  `json:"Family,omitempty"`
	LineHeight float64 `json:"LineHeight"`
	Bold       bool    `json:"Bold"`
}

func (f FontStyle) Value() (driver.Value, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (f *FontStyle) Scan(src interface{}) error {
	if src == nil {
		*f = FontStyle{}
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported scan type %T for FontStyle", src)
	}
	return json.Unmarshal(b, f)
}
