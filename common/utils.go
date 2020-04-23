package common

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func InterfaceToType(val interface{}, t string) (ret interface{}, err error) {
	switch t {
	case "string":
		ret = val.(string)
	case "int":
		if reflect.TypeOf(val).Name() == "string" {
			ret, err = strconv.Atoi(val.(string))
		} else {
			ret = val
		}
	case "uint":
		if reflect.TypeOf(val).Name() == "string" {
			ret, err = strconv.ParseUint(val.(string), 10, 10)
		} else if reflect.TypeOf(val).Name() == "float64" {
			ret = uint(val.(float64))
		} else {
			ret = val
		}
	case "boolean":
		ret = val.(bool)
	case "timestamp":
		// ret, err = time.Parse("2006-01-02", val.(string))
		ret, err = time.Parse(time.RFC3339, val.(string))
	default:
		err = fmt.Errorf("unknown value is of incompatible type")
	}
	return
}
