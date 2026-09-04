package filter

import "strings"

type Filter struct {
	Mutate Mutate
}

type Mutate struct {
	CopyValue       []CopyValue     `xml:"copy-value"`
	RemoveFields    RemoveFields    `xml:"remove-fields"`
	AddFields       AddFields       `xml:"add-fields"`
	LowercaseFields LowercaseFields `xml:"lowercase-fields"`
	Row             string          `xml:"row,attr"`
}

type AddFields struct {
	Field []Field `xml:"field"`
}

type Field struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type CopyValue struct {
	From string `xml:"from,attr"`
	To   string `xml:"to,attr"`
	In   string `xml:"in,attr"`
}

type LowercaseFields struct {
	Field         []string `xml:"field"`
	For           string   `xml:"for,attr"`
	CaseSentitive string   `xml:"case-sentitive,attr"`
}

type RemoveFields struct {
	Field      []string `xml:"field"`
	IgnoreCase string   `xml:"ignore-case-attr"`
}

// Apply takes a row and a Filter, returns a new map with all mutations applied.
// CopyValue copies from one field to another.
// RemoveFields deletes fields (case-insensitive when IgnoreCase is "true").
// AddFields adds new fields with configured values.
// LowercaseFields converts string values to lowercase.
func Apply(row map[string]any, filter Filter) map[string]any {
	result := make(map[string]any, len(row))
	for k, v := range row {
		result[k] = v
	}

	mut := filter.Mutate

	for _, cv := range mut.CopyValue {
		if val, ok := result[cv.From]; ok {
			result[cv.To] = val
		}
	}

	for _, field := range mut.RemoveFields.Field {
		if mut.RemoveFields.IgnoreCase == "true" {
			for key := range result {
				if strings.EqualFold(key, field) {
					delete(result, key)
					break
				}
			}
		} else {
			delete(result, field)
		}
	}

	for _, f := range mut.AddFields.Field {
		result[f.Key] = f.Value
	}

	for _, field := range mut.LowercaseFields.Field {
		if val, ok := result[field]; ok {
			if s, ok := val.(string); ok {
				result[field] = strings.ToLower(s)
			}
		}
	}

	return result
}

// FilterFromModel converts a dbetls.Filter model to a filter.Filter.
func FilterFromModel(m Mutate) Filter {
	return Filter{Mutate: m}
}
