package xdr

import (
	"bytes"
	"reflect"
	"testing"
)

type ArrayElementStruct struct {
	BooleanInArrayElement bool `XDR_Name:"Boolean"`
}

type ChildStruct struct {
	BooleanInChild bool `XDR_Name:"Boolean"`
}

type ParentStruct struct {
	Integer                         int32                 `XDR_Name:"Integer"`
	UnsignedInteger                 uint32                `XDR_Name:"Unsigned Integer"`
	EnumerationAsInt32              int32                 `XDR_Name:"Enumeration"`
	EnumerationAsUint32             uint32                `XDR_Name:"Enumeration"`
	Boolean                         bool                  `XDR_Name:"Boolean"`
	HyperInteger                    int64                 `XDR_Name:"Hyper Integer"`
	UnsignedHyperInteger            uint64                `XDR_Name:"Unsigned Hyper Integer"`
	FixedLengthOpaqueData           [5]byte               `XDR_Name:"Fixed-Length Opaque Data"`
	VariableLengthOpaqueDataNoMax   []byte                `XDR_Name:"Variable-Length Opaque Data"`
	VariableLengthOpaqueDataWithMax []byte                `XDR_Name:"Variable-Length Opaque Data" XDR_MaxSize:"6"`
	StringAsByteSliceNoMax          []byte                `XDR_Name:"String"`
	StringAsByteSliceWithMax        []byte                `XDR_Name:"String" XDR_MaxSize:"3"`
	StringAsStringNoMax             string                `XDR_Name:"String"`
	StringAsStringWithMax           string                `XDR_Name:"String" XDR_MaxSize:"7"`
	FixedLengthArray                [3]ArrayElementStruct `XDR_Name:"Fixed-Length Array"`
	VariableLengthArrayNoMax        []ArrayElementStruct  `XDR_Name:"Variable-Length Array"`
	VariableLengthArrayWithMax      []ArrayElementStruct  `XDR_Name:"Variable-Length Array" XDR_MaxSize:"2"`
	Structure                       ChildStruct           `XDR_Name:"Structure"`
}

var (
	badParentStruct = ParentStruct{
		Integer:                         -1000000,
		UnsignedInteger:                 1000000,
		EnumerationAsInt32:              -1000,
		EnumerationAsUint32:             1000,
		Boolean:                         true,
		HyperInteger:                    -1000000000000,
		UnsignedHyperInteger:            1000000000000,
		FixedLengthOpaqueData:           [5]byte{0x01, 0x02, 0x03, 0x04, 0x05},
		VariableLengthOpaqueDataNoMax:   []byte{0x01, 0x02, 0x03},
		VariableLengthOpaqueDataWithMax: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}, // Exceeds XDR_MaxSize
		StringAsByteSliceNoMax:          []byte{'H', 'i'},
		StringAsByteSliceWithMax:        []byte{'B', 'y', 'e'},
		StringAsStringNoMax:             "Hi",
		StringAsStringWithMax:           "Bye",
		FixedLengthArray:                [3]ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}, {BooleanInArrayElement: true}},
		VariableLengthArrayNoMax:        []ArrayElementStruct{{BooleanInArrayElement: true}},
		VariableLengthArrayWithMax:      []ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}},
		Structure:                       ChildStruct{BooleanInChild: true},
	}

	badParentStructPtr = &badParentStruct

	badParentStructPacked = []byte{
		0xFF, 0xF0, 0xBD, 0xC0, //                                                 -1000000
		0x00, 0x0F, 0x42, 0x40, //                                                 1000000
		0xFF, 0xFF, 0xFC, 0x18, //                                                 -1000
		0x00, 0x00, 0x03, 0xE8, //                                                 1000
		0x00, 0x00, 0x00, 0x01, //                                                 true
		0xFF, 0xFF, 0xFF, 0x17, 0x2B, 0x5A, 0xF0, 0x00, //                         -1000000000000
		0x00, 0x00, 0x00, 0xE8, 0xD4, 0xA5, 0x10, 0x00, //                         1000000000000
		0x01, 0x02, 0x03, 0x04, 0x05, 0x00, 0x00, 0x00, //                         [5]byte{0x01, 0x02, 0x03, 0x04, 0x05}
		0x00, 0x00, 0x00, 0x03, 0x01, 0x02, 0x03, 0x00, //                         []byte{0x01, 0x02, 0x03}
		0x00, 0x00, 0x00, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, // []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08} - exceeds XDR_MaxSize
		0x00, 0x00, 0x00, 0x02, 0x48, 0x69, 0x00, 0x00, //                         []byte{'H', 'i'}
		0x00, 0x00, 0x00, 0x03, 0x42, 0x79, 0x65, 0x00, //                         []byte{B', 'y', 'e'}
		0x00, 0x00, 0x00, 0x02, 0x48, 0x69, 0x00, 0x00, //                         "Hi"
		0x00, 0x00, 0x00, 0x03, 0x42, 0x79, 0x65, 0x00, //                         "Bye"
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // [3]ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}, {BooleanInArrayElement: true}}
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, //                         []ArrayElementStruct{{BooleanInArrayElement: true}}
		0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // []ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}}
		0x00, 0x00, 0x00, 0x01, //                                                 ChildStruct{BooleanInChild: true}
	}

	goodParentStruct = ParentStruct{
		Integer:                         -1000000,
		UnsignedInteger:                 1000000,
		EnumerationAsInt32:              -1000,
		EnumerationAsUint32:             1000,
		Boolean:                         true,
		HyperInteger:                    -1000000000000,
		UnsignedHyperInteger:            1000000000000,
		FixedLengthOpaqueData:           [5]byte{0x01, 0x02, 0x03, 0x04, 0x05},
		VariableLengthOpaqueDataNoMax:   []byte{0x01, 0x02, 0x03},
		VariableLengthOpaqueDataWithMax: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		StringAsByteSliceNoMax:          []byte{'H', 'i'},
		StringAsByteSliceWithMax:        []byte{'B', 'y', 'e'},
		StringAsStringNoMax:             "Hi",
		StringAsStringWithMax:           "Bye",
		FixedLengthArray:                [3]ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}, {BooleanInArrayElement: true}},
		VariableLengthArrayNoMax:        []ArrayElementStruct{{BooleanInArrayElement: true}},
		VariableLengthArrayWithMax:      []ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}},
		Structure:                       ChildStruct{BooleanInChild: true},
	}

	goodParentStructPtr = &goodParentStruct

	goodParentStructPacked = []byte{
		0xFF, 0xF0, 0xBD, 0xC0, //                                                 -1000000
		0x00, 0x0F, 0x42, 0x40, //                                                 1000000
		0xFF, 0xFF, 0xFC, 0x18, //                                                 -1000
		0x00, 0x00, 0x03, 0xE8, //                                                 1000
		0x00, 0x00, 0x00, 0x01, //                                                 true
		0xFF, 0xFF, 0xFF, 0x17, 0x2B, 0x5A, 0xF0, 0x00, //                         -1000000000000
		0x00, 0x00, 0x00, 0xE8, 0xD4, 0xA5, 0x10, 0x00, //                         1000000000000
		0x01, 0x02, 0x03, 0x04, 0x05, 0x00, 0x00, 0x00, //                         [5]byte{0x01, 0x02, 0x03, 0x04, 0x05}
		0x00, 0x00, 0x00, 0x03, 0x01, 0x02, 0x03, 0x00, //                         []byte{0x01, 0x02, 0x03}
		0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x00, 0x00, // []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
		0x00, 0x00, 0x00, 0x02, 0x48, 0x69, 0x00, 0x00, //                         []byte{'H', 'i'}
		0x00, 0x00, 0x00, 0x03, 0x42, 0x79, 0x65, 0x00, //                         []byte{B', 'y', 'e'}
		0x00, 0x00, 0x00, 0x02, 0x48, 0x69, 0x00, 0x00, //                         "Hi"
		0x00, 0x00, 0x00, 0x03, 0x42, 0x79, 0x65, 0x00, //                         "Bye"
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // [3]ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}, {BooleanInArrayElement: true}}
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, //                         []ArrayElementStruct{{BooleanInArrayElement: true}}
		0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, // []ArrayElementStruct{{BooleanInArrayElement: true}, {BooleanInArrayElement: false}}
		0x00, 0x00, 0x00, 0x01, //                                                 ChildStruct{BooleanInChild: true}
	}

	goodParentStructPackedLen = uint64(len(goodParentStructPacked))
)

func TestExamine(t *testing.T) {
	var (
		bytesNeeded uint64
		err         error
	)

	_, err = Examine(badParentStruct)
	if nil == err {
		t.Fatalf("Examine(badParentStruct) should have failed")
	}

	_, err = Examine(badParentStructPtr)
	if nil == err {
		t.Fatalf("Examine(badParentStructPtr) should have failed")
	}

	bytesNeeded, err = Examine(goodParentStruct)
	if nil != err {
		t.Fatalf("Examine(goodParentStruct) received unexpected error: %v", err)
	}
	if goodParentStructPackedLen != bytesNeeded {
		t.Fatalf("Examine(goodParentStruct) received unexpected bytesNeeded (0x%X) - should have been 0x%X", bytesNeeded, goodParentStructPackedLen)
	}

	bytesNeeded, err = Examine(goodParentStructPtr)
	if nil != err {
		t.Fatalf("Examine(goodParentStructPtr) received unexpected error: %v", err)
	}
	if goodParentStructPackedLen != bytesNeeded {
		t.Fatalf("Examine(goodParentStructPtr) received unexpected bytesNeeded (0x%X) - should have been 0x%X", bytesNeeded, goodParentStructPackedLen)
	}
}

func TestPack(t *testing.T) {
	var (
		err                            error
		goodParentStructPackedReturned []byte
	)

	_, err = Pack(badParentStruct)
	if nil == err {
		t.Fatalf("Pack(badParentStruct) should have failed")
	}

	_, err = Pack(badParentStructPtr)
	if nil == err {
		t.Fatalf("Pack(badParentStructPtr) should have failed")
	}

	goodParentStructPackedReturned, err = Pack(goodParentStruct)
	if nil != err {
		t.Fatalf("Pack(goodParentStruct) received unexpected error: %v", err)
	}
	if 0 != bytes.Compare(goodParentStructPacked, goodParentStructPackedReturned) {
		t.Fatalf("Pack(goodParentStruct) received unexpected goodParentStructPackedReturned")
	}

	goodParentStructPackedReturned, err = Pack(goodParentStructPtr)
	if nil != err {
		t.Fatalf("Pack(goodParentStructPtr) received unexpected error: %v", err)
	}
	if 0 != bytes.Compare(goodParentStructPacked, goodParentStructPackedReturned) {
		t.Fatalf("Pack(goodParentStructPtr) received unexpected goodParentStructPackedReturned")
	}
}

func TestUnpack(t *testing.T) {
	var (
		badParentStructReturned  ParentStruct
		bytesConsumed            uint64
		err                      error
		goodParentStructReturned ParentStruct
	)

	_, err = Unpack(badParentStructPacked, &badParentStructReturned)
	if nil == err {
		t.Fatalf("Unpack(badParentStructPacked, &badParentStructReturned) should have failed")
	}

	bytesConsumed, err = Unpack(goodParentStructPacked, &goodParentStructReturned)
	if nil != err {
		t.Fatalf("Unpack(goodParentStructPacked, &goodParentStructReturned) received unexpected error: %v", err)
	}
	if goodParentStructPackedLen != bytesConsumed {
		t.Fatalf("Unpack(goodParentStructPacked, &goodParentStructReturned) received unexpected bytesConsumed (0x%X) - should have been 0x%X", bytesConsumed, goodParentStructPackedLen)
	}
	if !reflect.DeepEqual(goodParentStruct, goodParentStructReturned) {
		t.Fatalf("Unpack(goodParentStructPacked, &goodParentStructReturned) received unexpected goodParentStructReturned")
	}
}
