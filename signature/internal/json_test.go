package internal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// implementsUnmarshalJSON is a minimalistic type used to detect that
// paranoidUnmarshalJSONObject uses the json.Unmarshaler interface of resolved
// pointers.
type implementsUnmarshalJSON bool

// Compile-time check that Policy implements json.Unmarshaler.
var _ json.Unmarshaler = (*implementsUnmarshalJSON)(nil)

func (dest *implementsUnmarshalJSON) UnmarshalJSON(data []byte) error {
	_ = data     // We don't care, not really.
	*dest = true // Mark handler as called
	return nil
}

func TestParanoidUnmarshalJSONObject(t *testing.T) {
	type testStruct struct {
		A string
		B int
	}
	ts := testStruct{}
	var unmarshalJSONCalled implementsUnmarshalJSON
	tsResolver := func(key string) interface{} {
		switch key {
		case "a":
			return &ts.A
		case "b":
			return &ts.B
		case "implementsUnmarshalJSON":
			return &unmarshalJSONCalled
		default:
			return nil
		}
	}

	// Empty object
	ts = testStruct{}
	err := ParanoidUnmarshalJSONObject([]byte(`{}`), tsResolver)
	require.NoError(t, err)
	assert.Equal(t, testStruct{}, ts)

	// Success
	ts = testStruct{}
	err = ParanoidUnmarshalJSONObject([]byte(`{"a":"x", "b":2}`), tsResolver)
	require.NoError(t, err)
	assert.Equal(t, testStruct{A: "x", B: 2}, ts)

	// json.Unmarshaler is used for decoding values
	ts = testStruct{}
	unmarshalJSONCalled = implementsUnmarshalJSON(false)
	err = ParanoidUnmarshalJSONObject([]byte(`{"implementsUnmarshalJSON":true}`), tsResolver)
	require.NoError(t, err)
	assert.Equal(t, unmarshalJSONCalled, implementsUnmarshalJSON(true))

	// Various kinds of invalid input
	for _, input := range []string{
		``,                       // Empty input
		`&`,                      // Entirely invalid JSON
		`1`,                      // Not an object
		`{&}`,                    // Invalid key JSON
		`{1:1}`,                  // Key not a string
		`{"b":1, "b":1}`,         // Duplicate key
		`{"thisdoesnotexist":1}`, // Key rejected by resolver
		`{"a":&}`,                // Invalid value JSON
		`{"a":1}`,                // Type mismatch
		`{"a":"value"}{}`,        // Extra data after object
	} {
		ts = testStruct{}
		err := ParanoidUnmarshalJSONObject([]byte(input), tsResolver)
		assert.Error(t, err, input)
	}
}

func TestParanoidUnmarshalJSONObjectExactFields(t *testing.T) {
	var stringValue string
	var float64Value float64
	var rawValue json.RawMessage
	var unmarshallCalled implementsUnmarshalJSON
	exactFields := map[string]interface{}{
		"string":       &stringValue,
		"float64":      &float64Value,
		"raw":          &rawValue,
		"unmarshaller": &unmarshallCalled,
	}

	// Empty object
	err := ParanoidUnmarshalJSONObjectExactFields([]byte(`{}`), map[string]interface{}{})
	require.NoError(t, err)

	// Success
	err = ParanoidUnmarshalJSONObjectExactFields([]byte(`{"string": "a", "float64": 3.5, "raw": {"a":"b"}, "unmarshaller": true}`), exactFields)
	require.NoError(t, err)
	assert.Equal(t, "a", stringValue)
	assert.Equal(t, 3.5, float64Value)
	assert.Equal(t, json.RawMessage(`{"a":"b"}`), rawValue)
	assert.Equal(t, implementsUnmarshalJSON(true), unmarshallCalled)

	// Various kinds of invalid input
	for _, input := range []string{
		``,      // Empty input
		`&`,     // Entirely invalid JSON
		`1`,     // Not an object
		`{&}`,   // Invalid key JSON
		`{1:1}`, // Key not a string
		`{"string": "a", "string": "a", "float64": 3.5, "raw": {"a":"b"}, "unmarshaller": true}`,      // Duplicate key
		`{"string": "a", "float64": 3.5, "raw": {"a":"b"}, "unmarshaller": true, "thisisunknown", 1}`, // Unknown key
		`{"string": &, "float64": 3.5, "raw": {"a":"b"}, "unmarshaller": true}`,                       // Invalid value JSON
		`{"string": 1, "float64": 3.5, "raw": {"a":"b"}, "unmarshaller": true}`,                       // Type mismatch
		`{"string": "a", "float64": 3.5, "raw": {"a":"b"}, "unmarshaller": true}{}`,                   // Extra data after object
	} {
		err := ParanoidUnmarshalJSONObjectExactFields([]byte(input), exactFields)
		assert.Error(t, err, input)
	}
}
