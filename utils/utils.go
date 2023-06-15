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

func P[E any](in any) (out *E) {
	r := in.(E)
	return &r
}
