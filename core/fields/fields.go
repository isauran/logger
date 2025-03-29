package fields

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"time"
)

// Field represents a structured log field
type Field struct {
	Key   string
	Value any
}

// Fields is a collection of Field entries
type Fields []Field

// Add creates a new field with automatic type inference
func Add(key string, value any) Field {
	return Field{Key: key, Value: inferValue(value)}
}

// Group creates a new field group
func Group(key string, fields ...Field) Field {
	attrs := make([]slog.Attr, len(fields))
	for i, f := range fields {
		attrs[i] = f.ToAttr()
	}
	return Field{Key: key, Value: slog.GroupValue(attrs...)}
}

// ToAttr converts a Field to a slog.Attr
func (f Field) ToAttr() slog.Attr {
	return slog.Any(f.Key, f.Value)
}

// ToAttrs converts Fields to slog.Attrs
func (f Fields) ToAttrs() []slog.Attr {
	attrs := make([]slog.Attr, len(f))
	for i, field := range f {
		attrs[i] = field.ToAttr()
	}
	return attrs
}

// inferValue performs type inference and conversion for values
func inferValue(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case bool, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64,
		time.Time, time.Duration, fmt.Stringer, error:
		return v
	case []byte:
		return string(v)
	case json.Marshaler:
		if data, err := v.MarshalJSON(); err == nil {
			return string(data)
		}
		return fmt.Sprintf("%+v", v)
	default:
		rvalue := reflect.ValueOf(value)
		switch rvalue.Kind() {
		case reflect.Array, reflect.Slice:
			return formatArray(rvalue)
		case reflect.Map:
			return formatMap(rvalue)
		case reflect.Struct:
			return formatStruct(rvalue)
		case reflect.Ptr:
			if rvalue.IsNil() {
				return nil
			}
			return inferValue(rvalue.Elem().Interface())
		default:
			return fmt.Sprintf("%+v", value)
		}
	}
}

// formatArray formats array/slice values
func formatArray(v reflect.Value) any {
	length := v.Len()
	array := make([]any, length)
	for i := 0; i < length; i++ {
		array[i] = inferValue(v.Index(i).Interface())
	}
	return array
}

// formatMap formats map values
func formatMap(v reflect.Value) any {
	m := make(map[string]any)
	for _, key := range v.MapKeys() {
		m[fmt.Sprint(key.Interface())] = inferValue(v.MapIndex(key).Interface())
	}
	return m
}

// formatStruct formats struct values
func formatStruct(v reflect.Value) any {
	t := v.Type()
	m := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			value := v.Field(i)
			m[field.Name] = inferValue(value.Interface())
		}
	}
	return m
}

// Common field constructors
func String(key, value string) Field                 { return Add(key, value) }
func Int(key string, value int) Field                { return Add(key, value) }
func Int64(key string, value int64) Field            { return Add(key, value) }
func Float64(key string, value float64) Field        { return Add(key, value) }
func Bool(key string, value bool) Field              { return Add(key, value) }
func Time(key string, value time.Time) Field         { return Add(key, value) }
func Duration(key string, value time.Duration) Field { return Add(key, value) }
func Error(err error) Field                          { return Add("error", err) }
func Stringer(key string, value fmt.Stringer) Field  { return Add(key, value) }

// Stack creates a stack trace field
func Stack(key string, skip int) Field {
	return Add(key, formatStack(skip+1))
}

// Err creates an error field with optional wrapped error context
func Err(err error) Fields {
	if err == nil {
		return nil
	}

	fields := Fields{Error(err)}

	type causer interface {
		Cause() error
	}
	type wrapper interface {
		Unwrap() error
	}

	// Add cause chain
	for {
		var cause error
		switch e := err.(type) {
		case causer:
			cause = e.Cause()
		case wrapper:
			cause = e.Unwrap()
		}
		if cause == nil {
			break
		}
		fields = append(fields, Add("error.cause", cause.Error()))
		err = cause
	}

	return fields
}
