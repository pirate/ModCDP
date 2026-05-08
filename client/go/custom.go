package modcdp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type CustomCommandHandle[P any, R any] struct {
	client *ModCDPClient
	name   string
}

func AddCustomCommand[P any, R any](client *ModCDPClient, name string, expression ...string) (CustomCommandHandle[P, R], error) {
	paramsSchema := JSONSchemaFor[P]()
	resultSchema := JSONSchemaFor[R]()
	payload := map[string]any{
		"name":         name,
		"paramsSchema": paramsSchema,
		"resultSchema": resultSchema,
	}
	if len(expression) > 0 && expression[0] != "" {
		payload["expression"] = expression[0]
	}
	if _, err := client.Send("Mod.addCustomCommand", payload); err != nil {
		return CustomCommandHandle[P, R]{}, err
	}
	return CustomCommandHandle[P, R]{client: client, name: name}, nil
}

func (h CustomCommandHandle[P, R]) Send(params P) (R, error) {
	return SendTyped[P, R](h.client, h.name, params)
}

func SendTyped[P any, R any](client *ModCDPClient, name string, params P) (R, error) {
	var typed R
	rawParams, err := cdpParamsMap(params)
	if err != nil {
		return typed, err
	}
	result, err := client.Send(name, rawParams)
	if err != nil {
		return typed, err
	}
	body, err := json.Marshal(result)
	if err != nil {
		return typed, err
	}
	if err := json.Unmarshal(body, &typed); err != nil {
		return typed, fmt.Errorf("%s result did not match typed result shape: %w", name, err)
	}
	return typed, nil
}

func AddCustomEvent[E any](client *ModCDPClient, name string) error {
	_, err := client.Send("Mod.addCustomEvent", map[string]any{
		"name":        name,
		"eventSchema": JSONSchemaFor[E](),
	})
	return err
}

func OnTyped[E any](client *ModCDPClient, name string, handler func(E)) {
	client.On(name, func(data any) {
		var typed E
		body, err := json.Marshal(data)
		if err != nil {
			return
		}
		if err := json.Unmarshal(body, &typed); err != nil {
			return
		}
		handler(typed)
	})
}

func JSONSchemaFor[T any]() map[string]any {
	var zero T
	return jsonSchemaForType(reflect.TypeOf(zero))
}

func jsonSchemaForType(t reflect.Type) map[string]any {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return map[string]any{"type": "object"}
	}
	switch t.Kind() {
	case reflect.Struct:
		properties := map[string]any{}
		required := []any{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name, omitEmpty, skip := jsonFieldName(field)
			if skip {
				continue
			}
			properties[name] = jsonSchemaForType(field.Type)
			if !omitEmpty && !isOptionalType(field.Type) {
				required = append(required, name)
			}
		}
		schema := map[string]any{
			"type":                 "object",
			"properties":           properties,
			"additionalProperties": false,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": jsonSchemaForType(t.Elem())}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": jsonSchemaForType(t.Elem())}
	default:
		return map[string]any{}
	}
}

func jsonFieldName(field reflect.StructField) (string, bool, bool) {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		name = field.Name
	}
	omitEmpty := false
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitEmpty = true
		}
	}
	return name, omitEmpty, false
}

func isOptionalType(t reflect.Type) bool {
	return t.Kind() == reflect.Pointer || (t.Kind() == reflect.Slice) || (t.Kind() == reflect.Map)
}
