package cfg

import (
	"github.com/spf13/cast"
	"strconv"
	"time"
)

// ToTime converts in to a time.Time value.
func ToTime(in interface{}) (time.Time, error) {
	switch val := in.(type) {
	case int64:
		return time.Unix(val, 0), nil
	default:
		t, err := cast.ToTimeE(val)
		if err != nil {
			return t, ErrBadType
		}
		return t, nil
	}
}

// ToBool converts in to a bool value.
func ToBool(in interface{}) (bool, error) {
	switch val := in.(type) {
	case string:
		sval, err := strconv.ParseBool(val)
		if err != nil {
			return false, ErrBadType
		}
		return sval, nil
	default:
		value, err := cast.ToBoolE(val)
		if err != nil {
			return false, ErrBadType
		}
		return value, nil
	}
}

// ToInt converts in to a int value.
func ToInt(in interface{}) (int, error) {
	if val, err := cast.ToIntE(in); err == nil {
		return val, nil
	}
	return 0, ErrBadType
}

// ToString converts in to a string value.
func ToString(in interface{}) (string, error) {
	if val, err := cast.ToStringE(in); err == nil {
		return val, nil
	}
	return "", ErrBadType
}
