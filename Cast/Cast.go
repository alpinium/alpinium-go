// Package Cast converts values between the basic types bool, all integers, float32, float64 and
// string, and types derived from them (e.g. type MyInt int). The main function is Cast: a value
// that does not fit into the destination type returns ErrOutOfRange, a string that cannot be
// parsed ErrInvalidSyntax. The package assumes a 64-bit platform (see Platform64.go).
package Cast

import (
    "errors"
    "fmt"
    "strconv"
    "unsafe"
)

// Limits of the integer types.
const MinUInt    uint   = uint(0)                          // 0x0000000000000000  0 on 64-bit systems
const MaxUInt    uint   = ^uint(0)                         // 0xFFFFFFFFFFFFFFFF  18446744073709551615 on 64-bit systems
const MinUInt8   uint8  = uint8(0)                         // 0x00                0
const MaxUInt8   uint8  = ^uint8(0)                        // 0xFF                255
const MinUInt16  uint16 = uint16(0)                        // 0x0000              0
const MaxUInt16  uint16 = ^uint16(0)                       // 0xFFFF              65535
const MinUInt32  uint32 = uint32(0)                        // 0x00000000          0
const MaxUInt32  uint32 = ^uint32(0)                       // 0xFFFFFFFF          4294967295
const MinUInt64  uint64 = uint64(0)                        // 0x0000000000000000  0
const MaxUInt64  uint64 = ^uint64(0)                       // 0xFFFFFFFFFFFFFFFF  18446744073709551615
const MinUIntPtr uintptr = uintptr(0)                      // 0x0000000000000000  0 on 64-bit systems
const MaxUIntPtr uintptr = ^uintptr(0)                     // 0xFFFFFFFFFFFFFFFF  18446744073709551615 on 64-bit systems
const MinInt     int    = ^int(^uint(0) >> 1)              // 0x8000000000000000 -9223372036854775808 on 64-bit systems
const MaxInt     int    = int(^uint(0) >> 1)               // 0x7FFFFFFFFFFFFFFF  9223372036854775807 on 64-bit systems
const MinInt8    int8   = ^int8(^uint8(0) >> 1)            // 0x80               -128
const MaxInt8    int8   = int8(^uint8(0) >> 1)             // 0x7F                127
const MinInt16   int16  = ^int16(^uint16(0) >> 1)          // 0x8000             -32768
const MaxInt16   int16  = int16(^uint16(0) >> 1)           // 0x7FFF              32767
const MinInt32   int32  = ^int32(^uint32(0) >> 1)          // 0x80000000         -2147483648
const MaxInt32   int32  = int32(^uint32(0) >> 1)           // 0x7FFFFFFF          2147483647
const MinInt64   int64  = ^int64(^uint64(0) >> 1)          // 0x8000000000000000 -9223372036854775808
const MaxInt64   int64  = int64(^uint64(0) >> 1)           // 0x7FFFFFFFFFFFFFFF  9223372036854775807

// Limits of the float types (the largest finite values).
const MinFloat32 float32 = -0x1.fffffep+127                // 0xFF7FFFFF         -3.40282346638528859811704183484516925440e+38
const MaxFloat32 float32 =  0x1.fffffep+127                // 0x7F7FFFFF          3.40282346638528859811704183484516925440e+38
const MinFloat64 float64 = -0x1.fffffffffffffp+1023        // 0xFFEFFFFFFFFFFFFF -1.797693134862315708145274237317043567981e+308
const MaxFloat64 float64 =  0x1.fffffffffffffp+1023        // 0x7FEFFFFFFFFFFFFF  1.797693134862315708145274237317043567981e+308

// Limits of a float32 value that converts to the integer type without overflow: the smallest
// and the largest float32 inside the range of the integer type.
const MinFloat32ForUInt32 float32 =  0x0p+00               // 0x00000000          0
const MaxFloat32ForUInt32 float32 =  0x1.fffffep+31        // 0xFFFFFF00          4294967040
const MinFloat32ForUInt64 float32 =  0x0p+00               // 0x0000000000000000  0
const MaxFloat32ForUInt64 float32 =  0x1.fffffep+63        // 0xFFFFFF0000000000  18446742974197923840
const MinFloat32ForInt32  float32 = -0x1p+31               // 0x80000000         -2147483648
const MaxFloat32ForInt32  float32 =  0x1.fffffep+30        // 0x7FFFFF80          2147483520
const MinFloat32ForInt64  float32 = -0x1p+63               // 0x8000000000000000 -9223372036854775808
const MaxFloat32ForInt64  float32 =  0x1.fffffep+62        // 0x7FFFFF8000000000  9223371487098961920

// Limits of a float64 value that converts to the integer type without overflow: the smallest
// and the largest float64 inside the range of the integer type.
const MinFloat64ForUInt32 float64 =  0x0p+00               // 0x00000000          0
const MaxFloat64ForUInt32 float64 =  0x1.fffffffep+31      // 0xFFFFFFFF          4294967295
const MinFloat64ForUInt64 float64 =  0x0p+00               // 0x0000000000000000  0
const MaxFloat64ForUInt64 float64 =  0x1.fffffffffffffp+63 // 0xFFFFFFFFFFFFF800  18446744073709549568
const MinFloat64ForInt32  float64 = -0x1p+31               // 0x80000000         -2147483648
const MaxFloat64ForInt32  float64 =  0x1.fffffffcp+30      // 0x7FFFFFFF          2147483647
const MinFloat64ForInt64  float64 = -0x1p+63               // 0x8000000000000000 -9223372036854775808
const MaxFloat64ForInt64  float64 =  0x1.fffffffffffffp+62 // 0x7FFFFFFFFFFFFC00  9223372036854774784

// Errors returned by the conversion functions. Test for them with errors.Is,
// e.g. errors.Is(err, Cast.ErrOutOfRange).
// ErrOutOfRange: the value does not fit into the destination type or the allowed range.
var ErrOutOfRange = errors.New("value out of range")
// ErrInvalidSyntax: the string is not a valid representation of the destination type.
var ErrInvalidSyntax = errors.New("invalid syntax")
// ErrUnsupportedType: the source or destination data type is not supported.
var ErrUnsupportedType = errors.New("unsupported data type")

// parseError keeps the original strconv error (its message and errors.Is(err, strconv.ErrRange)
// keep working) and adds the matching package error (ErrOutOfRange or ErrInvalidSyntax).
type parseError struct {
    kind error
    err  error
}

func (e *parseError) Error() string   { return e.err.Error() }
func (e *parseError) Unwrap() []error { return []error{e.kind, e.err} }

// wrapParseError maps a strconv error to ErrOutOfRange or ErrInvalidSyntax.
func wrapParseError(err error) error {
    switch {
    case err == nil:
        return nil
    case errors.Is(err, strconv.ErrRange):
        return &parseError{kind: ErrOutOfRange, err: err}
    case errors.Is(err, strconv.ErrSyntax):
        return &parseError{kind: ErrInvalidSyntax, err: err}
    }
    return err
}

// CastableUnsignedInteger is the constraint of the unsigned integer types and types derived from them.
type CastableUnsignedInteger interface {
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// CastableSignedInteger is the constraint of the signed integer types and types derived from them.
type CastableSignedInteger interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

// CastableInteger is the constraint of all integer types and types derived from them.
type CastableInteger interface {
    CastableUnsignedInteger | CastableSignedInteger
}

// CastableFloat is the constraint of float32, float64 and types derived from them.
type CastableFloat interface {
    ~float32 | ~float64
}

// CastableNumber is the constraint of all number types and types derived from them.
type CastableNumber interface {
    CastableInteger | CastableFloat
}

// Castable is the constraint of all types supported by Cast: bool, numbers, string and types
// derived from them.
type Castable interface {
    ~bool | CastableNumber | ~string
}

// Bool

// BoolToNumber converts true to 1 and false to 0.
func BoolToNumber[T CastableNumber](value bool) (result T, err error) {
    if value {
        result = 1
    }
    return
}

// BoolToString converts the value to "true" or "false".
func BoolToString(value bool) (result string, err error) {
    result = strconv.FormatBool(value)
    return
}

// Number

// NumberToBool returns true for a non-zero value (NaN is non-zero, so it returns true).
func NumberToBool[T CastableNumber](value T) (result bool, err error) {
    if value != 0 {
        result = true
    }
    return
}

// NumberToString formats the value as a decimal string. A float is formatted in the shortest
// form that parses back to the same value ('g' format, e.g. 1.5, 1e+21).
func NumberToString[T CastableNumber](value T) (result string, err error) {
    var src_temp T
    src_size := unsafe.Sizeof(src_temp)
    src_is_integer := (T(1) / T(2)) == T(0)
    if src_is_integer {
        src_is_unsigned := (T(0) - T(1)) > T(0)
        if src_is_unsigned {
            result = strconv.FormatUint(uint64(value), 10)
        } else {
            result = strconv.FormatInt(int64(value), 10)
        }
    } else {
        result = strconv.FormatFloat(float64(value), 'g', -1, int(src_size)*8)
    }
    return
}

// String

// StringToBool parses the string with strconv.ParseBool (1, t, T, TRUE, true, True, 0, f, F,
// FALSE, false, False). Invalid syntax returns ErrInvalidSyntax and false.
func StringToBool(value string) (result bool, err error) {
    result, err = strconv.ParseBool(value)
    err = wrapParseError(err)
    return
}

// StringToNumber parses a decimal string into the number type T. Invalid syntax returns
// ErrInvalidSyntax, a value out of range of T returns ErrOutOfRange; on error the result is 0.
func StringToNumber[T CastableNumber](value string) (result T, err error) {
    var dst_temp T
    dst_size := unsafe.Sizeof(dst_temp)
    dst_is_integer := (T(1) / T(2)) == T(0)
    if dst_is_integer {
        dst_is_unsigned := (T(0) - T(1)) > T(0)
        if dst_is_unsigned {
            var v uint64
            if v, err = strconv.ParseUint(value, 10, int(dst_size)*8); err == nil {
                result = T(v)
            }
        } else {
            var v int64
            if v, err = strconv.ParseInt(value, 10, int(dst_size)*8); err == nil {
                result = T(v)
            }
        }
    } else {
        var v float64
        if v, err = strconv.ParseFloat(value, int(dst_size)*8); err == nil {
            result = T(v)
        }
    }
    err = wrapParseError(err)
    return
}

// DecodeScalar

// ParseScalar converts a string to the first matching scalar type: uint64, int64, float64, bool
// (in this order, e.g. "1" is uint64, "-1" int64, "1.5" float64, "true" bool). If value is not
// a string or the string is none of these types, it returns nil and false.
func ParseScalar(value any) (result any, ok bool) {

    var value_string string
    if value_string, ok = (value).(string); !ok {
        return
    }

    if value_uint64, err := strconv.ParseUint(value_string, 10, 64); err == nil {
        return value_uint64, true
    }

    if value_int64, err := strconv.ParseInt(value_string, 10, 64); err == nil {
        return value_int64, true
    }

    if value_float64, err := strconv.ParseFloat(value_string, 64); err == nil {
        return value_float64, true
    }

    if value_bool, err := strconv.ParseBool(value_string); err == nil {
        return value_bool, true
    }

    return nil, false
}

// Cast

// castError builds the range error of Cast. Cast checks the range inline and calls castError
// only for values out of range, so building the error (boxing, fmt) stays out of the fast path.
//
//go:noinline
func castError(value, min_value, max_value any) error {
    return fmt.Errorf("%w: %v is not in allowed range < %v ... %v >", ErrOutOfRange, value, min_value, max_value)
}

// The Cast function is generated into cast_switch.go by cast_gen.go.
// After changing the generator run: go generate ./...
//go:generate go run cast_gen.go

// CastDef converts the value with Cast; on error it returns default_value.
func CastDef[T Castable, TT Castable](value T, default_value TT) (result TT) {
    var err error
    if result, err = Cast[T, TT](value); err != nil {
        result = default_value
    }
    return
}

// CastIndirect converts source with Cast and stores the result in *destination. A nil destination
// is ignored (nothing is converted, the error is nil).
func CastIndirect[T Castable, TT Castable](source T, destination *TT) (err error) {
    if destination != nil {
        *destination, err = Cast[T, TT](source)
    }
    return
}

// CastIndirectDef converts source with Cast and stores the result in *destination; on error it
// stores default_value. A nil destination is ignored.
func CastIndirectDef[T Castable, TT Castable](source T, destination *TT, default_value TT) {
    if destination != nil {
        var err error
        if *destination, err = Cast[T, TT](source); err != nil {
            *destination = default_value
        }
    }
}
