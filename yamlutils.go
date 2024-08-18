package yamlutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"go.n16f.net/ejson"
	"gopkg.in/yaml.v3"
)

// Load parses a YAML stream and copy the first document to an arbitrary value.
// The document must be a valid JSON value, meaning that object keys must be
// strings.
//
// The data conversion process uses "json" structure tags so that the same
// structures can be encoded and decoding using either YAML or JSON. This also
// ensures that the semantics of encoding/json are used everywhere.
//
// JSON decoding using the go.n16f.net/ejson module, meaning that ValidateJSON
// methods will be called on objects which define it.
func Load(data []byte, dest any) error {
	yamlDecoder := yaml.NewDecoder(bytes.NewReader(data))

	var yamlValue any
	if err := yamlDecoder.Decode(&yamlValue); err != nil && err != io.EOF {
		return fmt.Errorf("cannot decode YAML data: %w", err)
	}

	jsonValue, err := YAMLValueToJSONValue(yamlValue)
	if err != nil {
		return fmt.Errorf("invalid YAML data: %w", err)
	}

	jsonData, err := json.Marshal(jsonValue)
	if err != nil {
		return fmt.Errorf("cannot generate JSON data: %w", err)
	}

	if err := ejson.Unmarshal(jsonData, dest); err != nil {
		return fmt.Errorf("cannot decode JSON data: %w", err)
	}

	return nil
}

// YAMLValueToJSONValue converts a value returned by the YAML parser into a
// value that can be safely encoded to JSON data. The function fails if the
// value is an object containing a non-string key.
func YAMLValueToJSONValue(yamlValue any) (any, error) {
	// For some reason, gopkg.in/yaml.v3 will return objects as map[string]any
	// if all keys are strings, and as map[any]any if not. So we have to handle
	// both.

	var jsonValue any

	switch v := yamlValue.(type) {
	case []any:
		array := make([]any, len(v))

		for i, yamlElement := range v {
			jsonElement, err := YAMLValueToJSONValue(yamlElement)
			if err != nil {
				return nil, err
			}

			array[i] = jsonElement
		}

		jsonValue = array

	case map[any]any:
		object := make(map[string]any)

		for key, yamlEntry := range v {
			keyString, ok := key.(string)
			if !ok {
				return nil,
					fmt.Errorf("object key \"%v\" is not a string", key)
			}

			jsonEntry, err := YAMLValueToJSONValue(yamlEntry)
			if err != nil {
				return nil, err
			}

			object[keyString] = jsonEntry
		}

		jsonValue = object

	case map[string]any:
		object := make(map[string]any)

		for key, yamlEntry := range v {
			jsonEntry, err := YAMLValueToJSONValue(yamlEntry)
			if err != nil {
				return nil, err
			}

			object[key] = jsonEntry
		}

		jsonValue = object

	default:
		jsonValue = yamlValue
	}

	return jsonValue, nil
}
