# xdr

Go package implementing XDR as specified in RFC 4506 pack/unpack.

The mapping between Go and XDR data types is as follows:

| XDR                         | Go                 | Go struct field tags                                           |
| --------------------------- | ------------------ | -------------------------------------------------------------- |
| Integer                     | int32              | \`XDR_Name:"Integer"\`                                         |
| Unsigned Integer            | uint32             | \`XDR_Name:"Unsigned Integer"\`                                |
| Enumeration                 | int32 \|\| uint32  | \`XDR_Name:"Enumeration"\`                                     |
| Boolean                     | bool               | \`XDR_Name:"Boolean"\`                                         |
| Hyper Integer               | int64              | \`XDR_Name:"Hyper Integer"\`                                   |
| Unsigned Hyper Integer      | uint64             | \`XDR_Name:"Unsigned Hyper Integer"\`                          |
| Fixed-Length Opaque Data    | [\<n\>]byte        | \`XDR_Name:"Fixed-Length Opaque Data"\`                        |
| Variable-Length Opaque Data | []byte             | \`XDR_Name:"Variable-Length Opaque Data" XDR_MaxSize:"\<n\>"\` |
| String                      | []byte \|\| string | \`XDR_Name:"String" XDR_MaxSize:"\<n\>"\`                      |
| Fixed-Length Array          | [\<n\>]\<type\>    | \`XDR_Name:"Fixed-Length Array"\`                              |
| Variable-Length Array       | []\<type\>         | \`XDR_Name:"Variable-Length Array" XDR_MaxSize:"\<n\>"\`       |
| Structure                   | \<struct\>         | \`XDR_Name:"Structure"\`                                       |

The **XDR_MaxSize** tags refer to the maximum number of elements of the implicit array and are optional and default to 2\^32-1.

As unions aren't really a *thing* in Go, users are required to break up their structures
such that the preceding structure ends in the "tag" (a.k.a. "discriminant-declaration")
field of the XDR Discriminated Union.

## API Reference
```
// Examine may be used to determine the size of a []byte needed by Pack() (passed by value or reference).
func Examine(objIF interface{}) (bytesNeeded uint64, err error)

// Pack is used to serialize the supplied struct (passed by value or reference).
func Pack(srcObjIF interface{}) (dst []byte, err error)

// Unpack is used to deserialize into the supplied struct (passed by reference).
func Unpack(src []byte, dstObjIF interface{}) (bytesConsumed uint64, err error)
```

## Contributors

 * ed@swiftstack.com

## License

TBD
