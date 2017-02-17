package xdr

import (
	"fmt"
	"reflect"
	"strconv"
)

func examineRecursive(objValueOf reflect.Value, maxSize uint64) (bytesNeeded uint64, err error) {
	var (
		fieldBytesNeeded   uint64
		i                  int
		objTypeOf          reflect.Type
		paddedLength       uint64
		xdrMaxSizeAsString string
		xdrMaxSizeAsUint64 uint64
		xdrName            string
	)

	// Capture reflect.Type of objValueOf

	objTypeOf = reflect.TypeOf(objValueOf.Interface())

	// First check for "encapsulating" objValueOf.Kind()'s

	if (objValueOf.Kind() == reflect.Interface) || (objValueOf.Kind() == reflect.Ptr) {
		bytesNeeded, err = examineRecursive(objValueOf.Elem(), maxSize)
		return
	}

	// Setup defaults for "non-encapsulating" objValueOf.Kind()'s

	bytesNeeded = 0
	err = nil

	// Handle specific "non-encapsulating" objValueOf.Kind()

	switch objValueOf.Kind() {
	case reflect.Bool:
		bytesNeeded = 4
	case reflect.Int32:
		bytesNeeded = 4
	case reflect.Int64:
		bytesNeeded = 8
	case reflect.Uint32:
		bytesNeeded = 4
	case reflect.Uint64:
		bytesNeeded = 8
	case reflect.Array:
		if 0 == objValueOf.Len() {
			bytesNeeded = 0 // Note: This is actually impossible
		} else {
			if reflect.Uint8 == objTypeOf.Elem().Kind() {
				paddedLength = uint64(objValueOf.Len()) + 3
				paddedLength = paddedLength / 4
				paddedLength = paddedLength * 4
				bytesNeeded = paddedLength
			} else {
				bytesNeeded, err = examineRecursive(objValueOf.Index(0), 0)
				if nil != err {
					return
				}
				bytesNeeded = uint64(objValueOf.Len()) * bytesNeeded
			}
		}
	case reflect.Slice:
		if 0 == objValueOf.Len() {
			bytesNeeded = 4
		} else {
			if 0 == maxSize {
				if 0xFFFFFFFF < objValueOf.Len() {
					err = fmt.Errorf("objValueOf slice exceeds maximum allowable length")
					return
				}
			} else {
				if maxSize < uint64(objValueOf.Len()) {
					err = fmt.Errorf("objValueOf slice exceeds XDR_MaxSize")
					return
				}
			}
			if reflect.Uint8 == objTypeOf.Elem().Kind() {
				paddedLength = uint64(objValueOf.Len()) + 3
				paddedLength = paddedLength / 4
				paddedLength = paddedLength * 4
				bytesNeeded = 4 + paddedLength
			} else {
				bytesNeeded, err = examineRecursive(objValueOf.Index(0), 0)
				if nil != err {
					return
				}
				bytesNeeded = 4 + (uint64(objValueOf.Len()) * bytesNeeded)
			}
		}
	case reflect.String:
		if 0 == objValueOf.Len() {
			bytesNeeded = 4
		} else {
			if 0 == maxSize {
				if 0xFFFFFFFF < objValueOf.Len() {
					err = fmt.Errorf("objValueOf string exceeds maximum allowable length")
					return
				}
			} else {
				if maxSize < uint64(objValueOf.Len()) {
					err = fmt.Errorf("objValueOf string exceeds XDR_MaxSize")
					return
				}
			}
			paddedLength = uint64(objValueOf.Len()) + 3
			paddedLength = paddedLength / 4
			paddedLength = paddedLength * 4
			bytesNeeded = 4 + paddedLength
		}
	case reflect.Struct:
		for i = 0; i < objValueOf.NumField(); i++ {
			xdrName = objTypeOf.Field(i).Tag.Get("XDR_Name")
			switch objValueOf.Field(i).Kind() {
			case reflect.Bool:
				if xdrName != "Boolean" {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Bool")
					return
				}
			case reflect.Int32:
				if (xdrName != "Integer") && (xdrName != "Enumeration") {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Int32")
					return
				}
			case reflect.Int64:
				if xdrName != "Hyper Integer" {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Int64")
					return
				}
			case reflect.Uint32:
				if (xdrName != "Unsigned Integer") && (xdrName != "Enumeration") {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Uint32")
					return
				}
			case reflect.Uint64:
				if xdrName != "Unsigned Hyper Integer" {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Uint64")
					return
				}
			case reflect.Array:
				if (xdrName != "Fixed-Length Opaque Data") && (xdrName != "Fixed-Length Array") {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Array")
					return
				}
			case reflect.Slice:
				if (xdrName != "Variable-Length Opaque Data") && (xdrName != "String") && (xdrName != "Variable-Length Array") {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Slice")
					return
				}
			case reflect.String:
				if xdrName != "String" {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.String")
					return
				}
			case reflect.Struct:
				if xdrName != "Structure" {
					err = fmt.Errorf("struct field missing valid XDR_Name tag for Kind() == reflect.Struct")
					return
				}
			}
			xdrMaxSizeAsString = objTypeOf.Field(i).Tag.Get("XDR_MaxSize")
			if "" == xdrMaxSizeAsString {
				xdrMaxSizeAsUint64 = 0
			} else {
				xdrMaxSizeAsUint64, err = strconv.ParseUint(xdrMaxSizeAsString, 10, 64)
				if nil != err {
					return
				}
				if 0xFFFFFFFF < xdrMaxSizeAsUint64 {
					err = fmt.Errorf("XDR_MaxSize (%v) exceeds maximum allowed (0xFFFFFFFF)", xdrMaxSizeAsUint64)
					return
				}
			}
			fieldBytesNeeded, err = examineRecursive(objValueOf.Field(i), xdrMaxSizeAsUint64)
			if nil != err {
				return
			}
			bytesNeeded += fieldBytesNeeded
		}
	default:
		err = fmt.Errorf("objValueOf is %#v; objValueOf.Kind() == %v unsupported", objValueOf, objValueOf.Kind())
		return
	}

	return
}

func packRecursive(srcObjValueOf reflect.Value, dst []byte, oldOffset uint64) (newOffset uint64) {
	var (
		b            bool
		i            int
		i64          int64
		paddedLength uint64
		s            string
		srcObjTypeOf reflect.Type
		u64          uint64
	)

	// Capture reflect.Type of srcObjValueOf

	srcObjTypeOf = reflect.TypeOf(srcObjValueOf.Interface())

	// Handle specific srcObjValueOf.Kind()

	switch srcObjValueOf.Kind() {
	case reflect.Interface:
		newOffset = packRecursive(srcObjValueOf.Elem(), dst, oldOffset)
	case reflect.Ptr:
		newOffset = packRecursive(srcObjValueOf.Elem(), dst, oldOffset)
	case reflect.Bool:
		b = srcObjValueOf.Bool()
		dst[oldOffset+0] = 0x00
		dst[oldOffset+1] = 0x00
		dst[oldOffset+2] = 0x00
		if b {
			dst[oldOffset+3] = 0x01
		} else {
			dst[oldOffset+3] = 0x00
		}
		newOffset = oldOffset + 4
	case reflect.Int32:
		i64 = srcObjValueOf.Int()
		if 0 <= i64 {
			u64 = uint64(i64)
		} else {
			u64 = ^uint64(-i64) + 1
		}
		dst[oldOffset+3] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+2] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+1] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+0] = byte(u64 & 0xFF)
		newOffset = oldOffset + 4
	case reflect.Int64:
		i64 = srcObjValueOf.Int()
		if 0 <= i64 {
			u64 = uint64(i64)
		} else {
			u64 = ^uint64(-i64) + 1
		}
		dst[oldOffset+7] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+6] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+5] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+4] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+3] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+2] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+1] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+0] = byte(u64 & 0xFF)
		newOffset = oldOffset + 8
	case reflect.Uint32:
		u64 = srcObjValueOf.Uint()
		dst[oldOffset+3] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+2] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+1] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+0] = byte(u64 & 0xFF)
		newOffset = oldOffset + 4
	case reflect.Uint64:
		u64 = srcObjValueOf.Uint()
		dst[oldOffset+7] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+6] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+5] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+4] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+3] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+2] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+1] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+0] = byte(u64 & 0xFF)
		newOffset = oldOffset + 8
	case reflect.Array:
		if 0 == srcObjValueOf.Len() {
			newOffset = oldOffset // Note: This is actually impossible
		} else {
			if reflect.Uint8 == srcObjTypeOf.Elem().Kind() {
				paddedLength = uint64(srcObjValueOf.Len()) + 3
				paddedLength = paddedLength / 4
				paddedLength = paddedLength * 4
				for i = 0; i < srcObjValueOf.Len(); i++ {
					dst[int(oldOffset)+i] = byte(srcObjValueOf.Index(i).Uint() & 0xFF)
				}
				for i = srcObjValueOf.Len(); i < int(paddedLength); i++ {
					dst[int(oldOffset)+i] = 0x00
				}
				newOffset = oldOffset + paddedLength
			} else {
				newOffset = oldOffset
				for i = 0; i < srcObjValueOf.Len(); i++ {
					newOffset = packRecursive(srcObjValueOf.Index(i), dst, newOffset)
				}
			}
		}
	case reflect.Slice:
		u64 = uint64(srcObjValueOf.Len())
		dst[oldOffset+3] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+2] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+1] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+0] = byte(u64 & 0xFF)
		if 0 == srcObjValueOf.Len() {
			newOffset = oldOffset + 4
		} else {
			if reflect.Uint8 == srcObjTypeOf.Elem().Kind() {
				paddedLength = uint64(srcObjValueOf.Len()) + 3
				paddedLength = paddedLength / 4
				paddedLength = paddedLength * 4
				for i = 0; i < srcObjValueOf.Len(); i++ {
					dst[int(oldOffset)+4+i] = byte(srcObjValueOf.Index(i).Uint() & 0xFF)
				}
				for i = srcObjValueOf.Len(); i < int(paddedLength); i++ {
					dst[int(oldOffset)+4+i] = 0x00
				}
				newOffset = oldOffset + 4 + paddedLength
			} else {
				newOffset = oldOffset + 4
				for i = 0; i < srcObjValueOf.Len(); i++ {
					newOffset = packRecursive(srcObjValueOf.Index(i), dst, newOffset)
				}
			}
		}
	case reflect.String:
		u64 = uint64(srcObjValueOf.Len())
		dst[oldOffset+3] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+2] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+1] = byte(u64 & 0xFF)
		u64 = u64 >> 8
		dst[oldOffset+0] = byte(u64 & 0xFF)
		s = srcObjValueOf.String()
		paddedLength = uint64(len(s)) + 3
		paddedLength = paddedLength / 4
		paddedLength = paddedLength * 4
		for i = 0; i < len(s); i++ {
			dst[int(oldOffset)+4+i] = s[i] & 0xFF
		}
		for i = len(s); i < int(paddedLength); i++ {
			dst[int(oldOffset)+4+i] = 0x00
		}
		newOffset = oldOffset + 4 + paddedLength
	case reflect.Struct:
		newOffset = oldOffset
		for i = 0; i < srcObjValueOf.NumField(); i++ {
			newOffset = packRecursive(srcObjValueOf.Field(i), dst, newOffset)
		}
	}

	return
}

func unpackRecursive(src []byte, oldOffset uint64, maxSize uint64, dstObjValueOf reflect.Value) (newOffset uint64, err error) {
	var (
		actualLength       uint64
		dstObjTypeOf       reflect.Type
		i                  int
		i64                int64
		paddedLength       uint64
		u64                uint64
		xdrMaxSizeAsString string
		xdrMaxSizeAsUint64 uint64
	)

	// Capture reflect.Type & reflect.Value of dstObjValueOf

	dstObjTypeOf = reflect.TypeOf(dstObjValueOf.Interface())

	// Handle specific srcObjValueOf.Kind()

	switch dstObjValueOf.Kind() {
	case reflect.Interface:
		newOffset, err = unpackRecursive(src, oldOffset, maxSize, dstObjValueOf.Elem())
		if nil != err {
			return
		}
	case reflect.Ptr:
		newOffset, err = unpackRecursive(src, oldOffset, maxSize, dstObjValueOf.Elem())
		if nil != err {
			return
		}
	case reflect.Bool:
		if uint64(len(src)) < (oldOffset + 4) {
			err = fmt.Errorf("No room for reflect.Bool field in src []byte")
			return
		}
		if (0 != src[oldOffset+0]) || (0 != src[oldOffset+1]) || (0 != src[oldOffset+2]) || (1 < src[oldOffset+3]) {
			err = fmt.Errorf("Invalid bytes for reflect.Bool field in src []byte at offset 0x%X", oldOffset)
			return
		}
		dstObjValueOf.SetBool(0x01 == src[oldOffset+3])
		newOffset = oldOffset + 4
	case reflect.Int32:
		if uint64(len(src)) < (oldOffset + 4) {
			err = fmt.Errorf("No room for reflect.Int32 field in src []byte")
			return
		}
		u64 = uint64(src[oldOffset+0])
		u64 = (u64 << 8) + uint64(src[oldOffset+1])
		u64 = (u64 << 8) + uint64(src[oldOffset+2])
		u64 = (u64 << 8) + uint64(src[oldOffset+3])
		if 0 == ((u64 >> 0x1F) & 0x01) {
			i64 = int64(u64)
		} else {
			i64 = -int64(^((u64 - 1) | uint64(0xFFFFFFFF00000000)))
		}
		dstObjValueOf.SetInt(i64)
		newOffset = oldOffset + 4
	case reflect.Int64:
		if uint64(len(src)) < (oldOffset + 8) {
			err = fmt.Errorf("No room for reflect.Int64 field in src []byte")
			return
		}
		u64 = uint64(src[oldOffset+0])
		u64 = (u64 << 8) + uint64(src[oldOffset+1])
		u64 = (u64 << 8) + uint64(src[oldOffset+2])
		u64 = (u64 << 8) + uint64(src[oldOffset+3])
		u64 = (u64 << 8) + uint64(src[oldOffset+4])
		u64 = (u64 << 8) + uint64(src[oldOffset+5])
		u64 = (u64 << 8) + uint64(src[oldOffset+6])
		u64 = (u64 << 8) + uint64(src[oldOffset+7])
		if 0 == ((u64 >> 0x3F) & 0x01) {
			i64 = int64(u64)
		} else {
			i64 = -int64(^(u64 - 1))
		}
		dstObjValueOf.SetInt(i64)
		newOffset = oldOffset + 8
	case reflect.Uint32:
		if uint64(len(src)) < (oldOffset + 4) {
			err = fmt.Errorf("No room for reflect.Uint32 field in src []byte")
			return
		}
		u64 = uint64(src[oldOffset+0])
		u64 = (u64 << 8) + uint64(src[oldOffset+1])
		u64 = (u64 << 8) + uint64(src[oldOffset+2])
		u64 = (u64 << 8) + uint64(src[oldOffset+3])
		dstObjValueOf.SetUint(u64)
		newOffset = oldOffset + 4
	case reflect.Uint64:
		if uint64(len(src)) < (oldOffset + 8) {
			err = fmt.Errorf("No room for reflect.Uint64 field in src []byte")
			return
		}
		u64 = uint64(src[oldOffset+0])
		u64 = (u64 << 8) + uint64(src[oldOffset+1])
		u64 = (u64 << 8) + uint64(src[oldOffset+2])
		u64 = (u64 << 8) + uint64(src[oldOffset+3])
		u64 = (u64 << 8) + uint64(src[oldOffset+4])
		u64 = (u64 << 8) + uint64(src[oldOffset+5])
		u64 = (u64 << 8) + uint64(src[oldOffset+6])
		u64 = (u64 << 8) + uint64(src[oldOffset+7])
		dstObjValueOf.SetUint(u64)
		newOffset = oldOffset + 8
	case reflect.Array:
		if 0 == dstObjValueOf.Len() {
			newOffset = oldOffset // Note: This is actually impossible
		} else {
			if reflect.Uint8 == dstObjTypeOf.Elem().Kind() {
				paddedLength = uint64(dstObjValueOf.Len()) + 3
				paddedLength = paddedLength / 4
				paddedLength = paddedLength * 4
				if uint64(len(src)) < (oldOffset + paddedLength) {
					err = fmt.Errorf("No room for refelct.Array field in src []byte")
					return
				}
				for i = 0; i < dstObjValueOf.Len(); i++ {
					dstObjValueOf.Index(i).SetUint(uint64(src[int(oldOffset)+i]))
				}
				for i = dstObjValueOf.Len(); i < int(paddedLength); i++ {
					if 0x00 != src[int(oldOffset)+i] {
						err = fmt.Errorf("Non-zero pad bytes in src []byte")
						return
					}
				}
				newOffset = oldOffset + paddedLength
			} else {
				newOffset = oldOffset
				for i = 0; i < dstObjValueOf.Len(); i++ {
					newOffset, err = unpackRecursive(src, newOffset, 0, dstObjValueOf.Index(i))
					if nil != err {
						return
					}
				}
			}
		}
	case reflect.Slice:
		if uint64(len(src)) < (oldOffset + 4) {
			err = fmt.Errorf("No room for reflect.Slice length field in src []byte")
			return
		}
		actualLength = uint64(src[oldOffset+0])
		actualLength = (actualLength << 8) + uint64(src[oldOffset+1])
		actualLength = (actualLength << 8) + uint64(src[oldOffset+2])
		actualLength = (actualLength << 8) + uint64(src[oldOffset+3])
		if 0 == actualLength {
			dstObjValueOf.Set(reflect.MakeSlice(dstObjTypeOf, 0, 0))
			newOffset = oldOffset + 4
		} else {
			if 0 == maxSize {
				if 0xFFFFFFFF < actualLength {
					err = fmt.Errorf("dstObjValueOf slice exceeds maximum allowable length")
					return
				}
			} else {
				if maxSize < actualLength {
					err = fmt.Errorf("dstObjValueOf slice exceeds XDR_MaxSize")
					return
				}
			}
			if reflect.Uint8 == dstObjTypeOf.Elem().Kind() {
				paddedLength = actualLength + 3
				paddedLength = paddedLength / 4
				paddedLength = paddedLength * 4
				if (oldOffset + 4 + paddedLength) > uint64(len(src)) {
					err = fmt.Errorf("No room for byte reflect.Slice padded length in src []byte")
					return
				}
				dstObjValueOf.SetBytes(src[(oldOffset + 4):(oldOffset + 4 + actualLength)])
				for i = int(oldOffset + 4 + actualLength); i < int(oldOffset+4+paddedLength); i++ {
					if 0x00 != src[i] {
						err = fmt.Errorf("Non-zero pad bytes in src []byte")
						return
					}
				}
				newOffset = oldOffset + 4 + paddedLength
			} else {
				dstObjValueOf.Set(reflect.MakeSlice(dstObjTypeOf, int(actualLength), int(actualLength)))
				newOffset = oldOffset + 4
				for i = 0; i < dstObjValueOf.Len(); i++ {
					newOffset, err = unpackRecursive(src, newOffset, 0, dstObjValueOf.Index(i))
					if nil != err {
						return
					}
				}
			}
		}
	case reflect.String:
		if uint64(len(src)) < (oldOffset + 4) {
			err = fmt.Errorf("No room for reflect.String length field in src []byte")
			return
		}
		actualLength = uint64(src[oldOffset+0])
		actualLength = (actualLength << 8) + uint64(src[oldOffset+1])
		actualLength = (actualLength << 8) + uint64(src[oldOffset+2])
		actualLength = (actualLength << 8) + uint64(src[oldOffset+3])
		if 0 == actualLength {
			dstObjValueOf.SetString("")
			newOffset = oldOffset + 4
		} else {
			if 0 == maxSize {
				if 0xFFFFFFFF < actualLength {
					err = fmt.Errorf("dstObjValueOf string exceeds maximum allowable length")
					return
				}
			} else {
				if maxSize < actualLength {
					err = fmt.Errorf("dstObjValueOf string exceeds XDR_MaxSize")
					return
				}
			}
			paddedLength = actualLength + 3
			paddedLength = paddedLength / 4
			paddedLength = paddedLength * 4
			if (oldOffset + 4 + paddedLength) > uint64(len(src)) {
				err = fmt.Errorf("No room for reflect.String padded length in src []byte")
				return
			}
			dstObjValueOf.SetString(string(src[(oldOffset + 4):(oldOffset + 4 + actualLength)]))
			for i = int(oldOffset + 4 + actualLength); i < int(oldOffset+4+paddedLength); i++ {
				if 0x00 != src[i] {
					err = fmt.Errorf("Non-zero pad bytes in src []byte")
					return
				}
			}
			newOffset = oldOffset + 4 + paddedLength
		}
	case reflect.Struct:
		newOffset = oldOffset
		for i = 0; i < dstObjValueOf.NumField(); i++ {
			xdrMaxSizeAsString = dstObjTypeOf.Field(i).Tag.Get("XDR_MaxSize")
			if "" == xdrMaxSizeAsString {
				xdrMaxSizeAsUint64 = 0
			} else {
				// Note: The strconv.ParseUint() & bounds check failures would have been already caught by examineRecursive()
				xdrMaxSizeAsUint64, err = strconv.ParseUint(xdrMaxSizeAsString, 10, 64)
				if nil != err {
					return
				}
				if 0xFFFFFFFF < xdrMaxSizeAsUint64 {
					err = fmt.Errorf("XDR_MaxSize (%v) exceeds maximum allowed (0xFFFFFFFF)", xdrMaxSizeAsUint64)
					return
				}
			}
			newOffset, err = unpackRecursive(src, newOffset, xdrMaxSizeAsUint64, dstObjValueOf.Field(i))
			if nil != err {
				return
			}
		}
	}

	err = nil
	return
}
