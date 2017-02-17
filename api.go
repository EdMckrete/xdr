package xdr

import (
	"reflect"
)

// Examine may be used to determine the size of a []byte needed by Pack() (passed by value or reference).
func Examine(objIF interface{}) (bytesNeeded uint64, err error) {
	var (
		objValueOf reflect.Value
	)

	objValueOf = reflect.ValueOf(objIF)

	bytesNeeded, err = examineRecursive(objValueOf, 0)

	return
}

// Pack is used to serialize the supplied struct (passed by value or reference).
func Pack(srcObjIF interface{}) (dst []byte, err error) {
	var (
		bytesNeeded   uint64
		srcObjValueOf reflect.Value
	)

	srcObjValueOf = reflect.ValueOf(srcObjIF)

	bytesNeeded, err = examineRecursive(srcObjValueOf, 0)
	if nil != err {
		return
	}

	dst = make([]byte, bytesNeeded)

	_ = packRecursive(srcObjValueOf, dst, 0)

	return
}

// Unpack is used to deserialize into the supplied struct (passed by reference).
func Unpack(src []byte, dstObjIF interface{}) (bytesConsumed uint64, err error) {
	var (
		dstObjValueOf reflect.Value
	)

	dstObjValueOf = reflect.ValueOf(dstObjIF)

	_, err = examineRecursive(dstObjValueOf, 0)
	if nil != err {
		return
	}

	bytesConsumed, err = unpackRecursive(src, 0, 0, dstObjValueOf)

	return
}
