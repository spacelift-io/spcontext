package internal

// DeduplicateFields removes duplicate fields from the given slice of fields.
func DeduplicateFields(fields []any) map[string]any {
	out := make(map[string]interface{})

	for i := 0; i < len(fields)/2; i++ {
		key := fields[2*i].(string)
		value := fields[2*i+1]
		out[key] = value
	}

	return out
}
