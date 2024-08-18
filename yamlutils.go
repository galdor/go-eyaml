package yamlutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"go.n16f.net/ejson"
	"gopkg.in/yaml.v3"
)

type Decoder struct {
	*yaml.Decoder
}

func NewDecoder(data []byte) *Decoder {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	return &Decoder{Decoder: decoder}
}

func (d *Decoder) Decode(dest any) error {
	var yamlValue any
	if err := d.Decoder.Decode(&yamlValue); err != nil {
		if err == io.EOF {
			return io.EOF
		}

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

func Load(data []byte, dest any) error {
	decoder := NewDecoder(data)
	return decoder.Decode(dest)
}

func LoadDocuments(data []byte, dest any) error {
	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Pointer {
		return fmt.Errorf("destination value is not a pointer")
	}

	pointedDestType := destType.Elem()
	if pointedDestType.Kind() != reflect.Slice {
		return fmt.Errorf("destination value is not a pointer to a slice")
	}

	eltType := pointedDestType.Elem()

	decoder := NewDecoder(data)
	docs := reflect.MakeSlice(pointedDestType, 0, 0)

	for {
		doc := reflect.New(eltType)
		if err := decoder.Decode(doc.Interface()); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		docs = reflect.Append(docs, doc.Elem())
	}

	reflect.ValueOf(dest).Elem().Set(docs)
	return nil
}

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
