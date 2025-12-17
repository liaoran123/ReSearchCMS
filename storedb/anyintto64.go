package storedb

import (
	"math"
	"reflect"
)

// AnyToInt64 使用 reflect 包将任意数值类型转换为 int64，处理所有数值类型
func AnyToInt64(v any) int64 {
	if v == nil {
		return int64(1)
	}

	// 使用 reflect 包获取值的类型和值
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// 检查是否超出 int64 的范围
		if val.Uint() > math.MaxInt64 {
			return int64(math.MaxInt64)
		}
		return int64(val.Uint())
	case reflect.Float32, reflect.Float64:
		// 检查是否超出 int64 的范围
		floatVal := val.Float()
		if floatVal > math.MaxInt64 {
			return int64(math.MaxInt64)
		}
		if floatVal < math.MinInt64 {
			return int64(math.MinInt64)
		}
		return int64(floatVal)
	default:
		return int64(1)
	}
}
