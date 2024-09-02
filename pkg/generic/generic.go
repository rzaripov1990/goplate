package generic

import (
	"reflect"
)

func Ptr[T any](val T) *T {
	return &val
}

func In[T comparable](x T, arr []T) bool {
	for i := range arr {
		if arr[i] == x {
			return true
		}
	}
	return false
}

func If[T any](value bool, vtrue, vfalse T) T {
	if value {
		return vtrue
	}
	return vfalse
}

func Equal[E1, E2 comparable](comp1 E1, comp2 E2) bool {
	return reflect.DeepEqual(comp1, comp2)
}
