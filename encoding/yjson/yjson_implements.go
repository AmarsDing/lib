package yjson

// MarshalJSON implements the interface MarshalJSON for json.Marshal.
func (j *Json) MarshalJSON() ([]byte, error) {
	return j.ToJson()
}

// UnmarshalJSON implements the interface UnmarshalJSON for json.Unmarshal.
func (j *Json) UnmarshalJSON(b []byte) error {
	r, err := LoadContent(b)
	if r != nil {
		// Value copy.
		*j = *r
	}
	return err
}

// UnmarshalValue is an interface implement which sets any type of value for Json.
func (j *Json) UnmarshalValue(value interface{}) error {
	if r := New(value); r != nil {
		// Value copy.
		*j = *r
	}
	return nil
}
