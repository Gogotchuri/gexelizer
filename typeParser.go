package gexelizer

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func parseStringIntoType(s string, t reflect.Type) (any, error) {
	switch t.Kind() {
	case reflect.String:
		return s, nil
	case reflect.Uint:
		return parseUint(s)
	case reflect.Int:
		return parseInt(s)
	case reflect.Int64:
		return parseInt64(s)
	case reflect.Float64:
		return parseFloat(s)
	case reflect.Bool:
		return parseBool(s)
	//case reflect.Struct: //TODO add Time and Decimal support
	default:
		return nil, fmt.Errorf("unsupported type %s", t.Kind())
	}
}

func parseUint(s string) (uint, error) {
	i, err := strconv.Atoi(s)
	return uint(i), err
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseBool(s string) (bool, error) {
	trues := []string{"true", "t", "1", "yes", "y"}
	falses := []string{"false", "f", "0", "no", "n"}
	s = strings.ToLower(s)
	for _, t := range trues {
		if t == s {
			return true, nil
		}
	}
	for _, f := range falses {
		if f == s {
			return false, nil
		}
	}
	return false, fmt.Errorf("invalid bool value: %s", s)
}
