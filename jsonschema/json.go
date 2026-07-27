// Package jsonschema provides very simple functionality for representing a JSON schema as a
// (nested) struct. This struct can be used with the chat completion "function call" feature.
// For more complicated schemas, it is recommended to use a dedicated JSON schema library
// and/or pass in the schema in []byte format.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type DataType string

const (
	Object  DataType = "object"
	Number  DataType = "number"
	Integer DataType = "integer"
	String  DataType = "string"
	Array   DataType = "array"
	Null    DataType = "null"
	Boolean DataType = "boolean"
)

// Definition is a struct for describing a JSON Schema.
// It is fairly limited, and you may have better luck using a third-party library.
type Definition struct {
	// Type is the JSON Schema type (object, integer, string, ...).
	Type DataType `json:"type,omitempty"`
	// Title is a short human label for the node.
	Title string `json:"title,omitempty"`
	// Description explains the field to the model.
	Description string `json:"description,omitempty"`
	// Enum limits the value to a fixed set of strings.
	Enum []string `json:"enum,omitempty"`
	// Default is the value suggested when the instance omits the field.
	Default any `json:"default,omitempty"`
	// Properties holds object fields when Type is object.
	Properties map[string]Definition `json:"properties,omitempty"`
	// Required lists property names that must be present on an object.
	Required []string `json:"required,omitempty"`
	// AdditionalProperties controls undeclared object keys (bool or a schema).
	AdditionalProperties any `json:"additionalProperties,omitempty"`
	// Items is the element schema when Type is array.
	Items *Definition `json:"items,omitempty"`
	// PrefixItems is a tuple-style list of per-index array item schemas.
	PrefixItems []Definition `json:"prefixItems,omitempty"`
	// MinItems is the minimum array length.
	MinItems *int `json:"minItems,omitempty"`
	// MaxItems is the maximum array length.
	MaxItems *int `json:"maxItems,omitempty"`
	// UniqueItems requires all array elements to be unique.
	UniqueItems bool `json:"uniqueItems,omitempty"`
	// Minimum is the inclusive numeric lower bound.
	Minimum *float64 `json:"minimum,omitempty"`
	// Maximum is the inclusive numeric upper bound.
	Maximum *float64 `json:"maximum,omitempty"`
	// MinLength is the minimum string length in characters.
	MinLength *int `json:"minLength,omitempty"`
	// MaxLength is the maximum string length in characters.
	MaxLength *int `json:"maxLength,omitempty"`
	// Pattern is a regex the string must match.
	Pattern string `json:"pattern,omitempty"`
	// Format is a semantic string format hint (e.g. date-time, email).
	Format string `json:"format,omitempty"`
	// Nullable marks the value as allowed to be null (OpenAI-style extension).
	Nullable bool `json:"nullable,omitempty"`
}

func (d *Definition) MarshalJSON() ([]byte, error) {
	if d.Properties == nil {
		d.Properties = make(map[string]Definition)
	}
	type Alias Definition
	return json.Marshal(struct {
		Alias
	}{
		Alias: (Alias)(*d),
	})
}

func (d *Definition) Unmarshal(content string, v any) error {
	return VerifySchemaAndUnmarshal(*d, []byte(content), v)
}

func GenerateSchema[T any]() (*Definition, error) {
	var v T
	return reflectSchema(reflect.TypeOf(v))
}

func GenerateSchemaForType(v any) (*Definition, error) {
	return reflectSchema(reflect.TypeOf(v))
}

func reflectSchema(t reflect.Type) (*Definition, error) {
	var d Definition
	switch t.Kind() {
	case reflect.String:
		d.Type = String
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d.Type = Integer
	case reflect.Float32, reflect.Float64:
		d.Type = Number
	case reflect.Bool:
		d.Type = Boolean
	case reflect.Slice, reflect.Array:
		d.Type = Array
		items, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d.Items = items
	case reflect.Struct:
		d.Type = Object
		d.AdditionalProperties = false
		object, err := reflectSchemaObject(t)
		if err != nil {
			return nil, err
		}
		d = *object
	case reflect.Ptr:
		definition, err := reflectSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		d = *definition
	case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.UnsafePointer:
		return nil, fmt.Errorf("unsupported type: %s", t.Kind().String())
	default:
	}
	return &d, nil
}

func reflectSchemaObject(t reflect.Type) (*Definition, error) {
	var d = Definition{
		Type:                 Object,
		AdditionalProperties: false,
	}
	properties := make(map[string]Definition)
	var requiredFields []string
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		skipSchema := field.Tag.Get("skipschema")
		if skipSchema == "true" {
			continue
		}
		jsonTag := field.Tag.Get("json")
		var required = true
		switch {
		case jsonTag == "-":
			continue
		case jsonTag == "":
			jsonTag = field.Name
		case strings.HasSuffix(jsonTag, ",omitempty"):
			jsonTag = strings.TrimSuffix(jsonTag, ",omitempty")
			required = false
		}

		item, err := reflectSchema(field.Type)
		if err != nil {
			return nil, err
		}
		description := field.Tag.Get("description")
		if description != "" {
			item.Description = description
		}
		enum := field.Tag.Get("enum")
		if enum != "" {
			item.Enum = strings.Split(enum, ",")
		}

		if n := field.Tag.Get("nullable"); n != "" {
			nullable, _ := strconv.ParseBool(n)
			item.Nullable = nullable
		}

		properties[jsonTag] = *item

		if s := field.Tag.Get("required"); s != "" {
			required, _ = strconv.ParseBool(s)
		}
		if required {
			requiredFields = append(requiredFields, jsonTag)
		}
	}
	d.Required = requiredFields
	d.Properties = properties
	return &d, nil
}
