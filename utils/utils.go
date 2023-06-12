package utils

import (
	"strconv"
	"strings"
)

const urlSeparator = "/"

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func ConvertSlice[E any](in []any) (out []E) {
	out = make([]E, 0, len(in))
	for _, v := range in {
		out = append(out, v.(E))
	}
	return
}

func MakeURL(path ...string) string {
	return strings.Join(path, urlSeparator)
}

func MakeInt(input string) (result int, err error) {
	result, err = strconv.Atoi(input)
	return
}

func BoolP(b bool) *bool {
	return &b
}

func IntP(i int) *int {
	return &i
}

func Int64P(i int64) *int64 {
	return &i
}

func Int32P(i int32) *int32 {
	return &i
}

func Int8P(i int8) *int8 {
	return &i
}

func Float32P(f float32) *float32 {
	return &f
}

func Float64P(f float64) *float64 {
	return &f
}

func StringP(s string) *string {
	return &s
}
