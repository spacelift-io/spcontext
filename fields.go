package spcontext

import (
	"fmt"
)

// Fields represents and contains structure metadata.
type Fields struct {
	previous *Fields
	keys     []string
	values   []interface{}
}

// With creates a new child Fields with additional fields.
// Panics if odd number of arguments were passed or key(first value of each pair) is not a string.
func (fields *Fields) With(kvs ...interface{}) *Fields {
	if len(kvs)%2 != 0 {
		panic("invalid Fields.With call: odd number of arguments")
	}

	keys := make([]string, len(kvs)/2)
	values := make([]interface{}, len(kvs)/2)
	for i := 0; i < len(kvs)/2; i++ {
		var ok bool
		keys[i], ok = kvs[2*i].(string)
		if !ok {
			panic(fmt.Sprintf("invalid Fields.With call: non-string log field key: %v", kvs[2*i]))
		}
		values[i] = kvs[2*i+1]
	}

	return &Fields{
		previous: fields,
		keys:     keys,
		values:   values,
	}
}

func (fields *Fields) makeFieldKVs() []interface{} {
	var out []interface{}
	if fields.previous != nil {
		out = fields.previous.makeFieldKVs()
	}

	for i := range fields.keys {
		out = append(out, fields.keys[i], fields.values[i])
	}
	return out
}

// EvaluateFields returns the fields as keys and evaluated values.
func (fields *Fields) EvaluateFields() []interface{} {
	out := fields.makeFieldKVs()
	for i := range out {
		if valuer, ok := out[i].(Valuer); ok {
			out[i] = valuer()
		}
	}
	return out
}

// Value returns the value for the given key or nil if it's not available.
func (fields *Fields) Value(key string) interface{} {
	for i := range fields.keys {
		if fields.keys[i] == key {
			return fields.values[i]
		}
	}
	if fields.previous != nil {
		return fields.previous.Value(key)
	}
	return nil
}

// Append appends new Fields to the current ones, modifying the argument.
func (fields *Fields) Append(newFields *Fields) *Fields {
	out := newFields
	for newFields.previous != nil {
		newFields = newFields.previous
	}
	newFields.previous = fields

	return out
}
