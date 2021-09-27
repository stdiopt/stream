package drow

import (
	"fmt"
	"reflect"
)

var (
	int64Typ   = reflect.TypeOf(int64(0))
	float64Typ = reflect.TypeOf(float64(0))
)

func Int(v interface{}) int {
	return int(Int64(v))
}

func Int64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	val := reflect.Indirect(reflect.ValueOf(v))
	if !val.IsValid() {
		return 0
	}
	if !val.CanConvert(int64Typ) {
		return 0
	}
	return val.Convert(int64Typ).Int()
}

func Float64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	val := reflect.Indirect(reflect.ValueOf(v))
	if !val.IsValid() {
		return 0
	}
	if !val.CanConvert(float64Typ) {
		return 0
	}
	return val.Convert(float64Typ).Float()
}

func String(v interface{}) string {
	if v == nil {
		return ""
	}
	switch v := v.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
	}
	return fmt.Sprint(v)
}

func Int64x(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch v := v.(type) {
	case uint8:
		return int64(v)
	case int8:
		return int64(v)
	case uint16:
		return int64(v)
	case int16:
		return int64(v)
	case uint32:
		return int64(v)
	case int32:
		return int64(v)
	case uint64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case *uint8:
		if v != nil {
			return int64(*v)
		}
	case *int8:
		if v != nil {
			return int64(*v)
		}
	case *uint16:
		if v != nil {
			return int64(*v)
		}
	case *int16:
		if v != nil {
			return int64(*v)
		}
	case *uint32:
		if v != nil {
			return int64(*v)
		}
	case *int32:
		if v != nil {
			return int64(*v)
		}
	case *uint64:
		if v != nil {
			return int64(*v)
		}
	case *int64:
		if v != nil {
			return int64(*v)
		}
	case *int:
		if v != nil {
			return int64(*v)
		}
	case *float32:
		if v != nil {
			return int64(*v)
		}
	case *float64:
		if v != nil {
			return int64(*v)
		}
	default:
		// panic(fmt.Sprintf("cannot convert %T to int", v))
	}
	return 0
}
