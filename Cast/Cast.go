package Cast

import (
    "fmt"
    "reflect"
    "strconv"
    "time"
    "unsafe"
)

const MinUInt   uint   = uint(0)                           // 0x0000000000000000  0 on 64-bit systems
const MaxUInt   uint   = ^uint(0)                          // 0xFFFFFFFFFFFFFFFF  18446744073709551615 on 64-bit systems
const MinUInt8  uint8  = uint8(0)                          // 0x00                0
const MaxUInt8  uint8  = ^uint8(0)                         // 0xFF                255
const MinUInt16 uint16 = uint16(0)                         // 0x0000              0
const MaxUInt16 uint16 = ^uint16(0)                        // 0xFFFF              65535
const MinUInt32 uint32 = uint32(0)                         // 0x00000000          0
const MaxUInt32 uint32 = ^uint32(0)                        // 0xFFFFFFFF          4294967295
const MinUInt64 uint64 = uint64(0)                         // 0x0000000000000000  0
const MaxUInt64 uint64 = ^uint64(0)                        // 0xFFFFFFFFFFFFFFFF  18446744073709551615
const MinInt    int    = ^int(^uint(0) >> 1)               // 0x8000000000000000 -9223372036854775808 on 64-bit systems
const MaxInt    int    = int(^uint(0) >> 1)                // 0x7FFFFFFFFFFFFFFF  9223372036854775807 on 64-bit systems
const MinInt8   int8   = ^int8(^uint8(0) >> 1)             // 0x80               -128
const MaxInt8   int8   = int8(^uint8(0) >> 1)              // 0x7F                127
const MinInt16  int16  = ^int16(^uint16(0) >> 1)           // 0x8000             -32768
const MaxInt16  int16  = int16(^uint16(0) >> 1)            // 0x7FFF              32767
const MinInt32  int32  = ^int32(^uint32(0) >> 1)           // 0x80000000         -2147483648
const MaxInt32  int32  = int32(^uint32(0) >> 1)            // 0x7FFFFFFF          2147483647
const MinInt64  int64  = ^int64(^uint64(0) >> 1)           // 0x8000000000000000 -9223372036854775808
const MaxInt64  int64  = int64(^uint64(0) >> 1)            // 0x7FFFFFFFFFFFFFFF  9223372036854775807

const MinFloat32 float32 = -0x1.fffffep+127                // 0xFF7FFFFF         -3.40282346638528859811704183484516925440e+38
const MaxFloat32 float32 =  0x1.fffffep+127                // 0x7F7FFFFF          3.40282346638528859811704183484516925440e+38
const MinFloat64 float64 = -0x1.fffffffffffffp+1023        // 0xFFEFFFFFFFFFFFFF -1.797693134862315708145274237317043567981e+308
const MaxFloat64 float64 =  0x1.fffffffffffffp+1023        // 0x7FEFFFFFFFFFFFFF  1.797693134862315708145274237317043567981e+308

const MinFloat32ForUInt32 float32 =  0x0p+00               // 0x00000000          0
const MaxFloat32ForUInt32 float32 =  0x1.fffffep+31        // 0xFFFFFF00          4294967040
const MinFloat32ForUInt64 float32 =  0x0p+00               // 0x0000000000000000  0
const MaxFloat32ForUInt64 float32 =  0x1.fffffep+63        // 0xFFFFFF0000000000  18446742974197923840
const MinFloat32ForInt32 float32  = -0x1p+31               // 0x80000000         -2147483648
const MaxFloat32ForInt32 float32  =  0x1.fffffep+30        // 0x7FFFFF80          2147483520
const MinFloat32ForInt64 float32  = -0x1p+63               // 0x8000000000000000 -9223372036854775808
const MaxFloat32ForInt64 float32  =  0x1.fffffep+62        // 0x7FFFFF8000000000  9223371487098961920

const MinFloat64ForUInt32 float64 =  0x0p+00               // 0x00000000          0
const MaxFloat64ForUInt32 float64 =  0x1.fffffffep+31      // 0xFFFFFFFF          4294967295
const MinFloat64ForUInt64 float64 =  0x0p+00               // 0x0000000000000000  0
const MaxFloat64ForUInt64 float64 =  0x1.fffffffffffffp+63 // 0xFFFFFFFFFFFFF800  18446744073709549568
const MinFloat64ForInt32 float64  = -0x1p+31               // 0x80000000         -2147483648
const MaxFloat64ForInt32 float64  =  0x1.fffffffcp+30      // 0x7FFFFFFF          2147483647
const MinFloat64ForInt64 float64  = -0x1p+63               // 0x8000000000000000 -9223372036854775808
const MaxFloat64ForInt64 float64  =  0x1.fffffffffffffp+62 // 0x7FFFFFFFFFFFFC00  9223372036854774784

type CastableUnsignedInteger interface {
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type CastableSignedInteger interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type CastableInteger interface {
    CastableUnsignedInteger | CastableSignedInteger
}

type CastableFloat interface {
    ~float32 | ~float64
}

type CastableNumber interface {
    CastableInteger | CastableFloat
}

type Castable interface {
    ~bool | CastableNumber | ~string
}

// Bool

func BoolToNumber[T CastableNumber](value bool) (result T, err error) {
    if value {
        result = 1
    }
    return
}

func BoolToString(value bool) (result string, err error) {
    result = strconv.FormatBool(value)
    return
}

// Number

func NumberToBool[T CastableNumber](value T) (result bool, err error) {
    if value != 0 {
        result = true
    }
    return
}

func NumberToNumber[T CastableNumber, TT CastableNumber](value T) (result TT, err error) {
    var src_temp T
    var dst_temp TT
    src_size, dst_size := unsafe.Sizeof(src_temp), unsafe.Sizeof(dst_temp)
    src_is_integer, dst_is_integer := (T(1) / T(2)) == T(0), (TT(1) / TT(2)) == TT(0)
    src_is_unsigned, dst_is_unsigned := (T(0) - T(1)) > T(0), (TT(0) - TT(1)) > TT(0)
    var src_min_value, src_max_value T
    var dst_min_value, dst_max_value TT

    if dst_is_integer {
        if dst_is_unsigned {
            switch dst_size {
                case 1:
                    min_uint8_value, max_uint8_value := MinUInt8, MaxUInt8
                    src_min_value, src_max_value = T(min_uint8_value), T(max_uint8_value)
                    dst_min_value, dst_max_value = TT(min_uint8_value), TT(max_uint8_value)
                case 2:
                    min_uint16_value, max_uint16_value := MinUInt16, MaxUInt16
                    src_min_value, src_max_value = T(min_uint16_value), T(max_uint16_value)
                    dst_min_value, dst_max_value = TT(min_uint16_value), TT(max_uint16_value)
                case 4:
                    if src_is_integer {
                        min_uint32_value, max_uint32_value := MinUInt32, MaxUInt32
                        src_min_value, src_max_value = T(min_uint32_value), T(max_uint32_value) 
                        dst_min_value, dst_max_value = TT(min_uint32_value), TT(max_uint32_value)
                    } else {
                        switch src_size {
                            case 4:
                                min_float32_value, max_float32_value := MinFloat32ForUInt32, MaxFloat32ForUInt32
                                src_min_value, src_max_value = T(min_float32_value), T(max_float32_value)
                                dst_min_value, dst_max_value = TT(min_float32_value), TT(max_float32_value)
                            case 8:
                                min_float64_value, max_float64_value := MinFloat64ForUInt32, MaxFloat64ForUInt32
                                src_min_value, src_max_value = T(min_float64_value), T(max_float64_value)
                                dst_min_value, dst_max_value = TT(min_float64_value), TT(max_float64_value)
                        }
                    }
                case 8:
                    if src_is_integer {
                        min_uint64_value, max_uint64_value := MinUInt64, MaxUInt64
                        src_min_value, src_max_value = T(min_uint64_value), T(max_uint64_value)
                        dst_min_value, dst_max_value = TT(min_uint64_value), TT(max_uint64_value)
                    } else {
                        switch src_size {
                            case 4:
                                min_float32_value, max_float32_value := MinFloat32ForUInt64, MaxFloat32ForUInt64
                                src_min_value, src_max_value = T(min_float32_value), T(max_float32_value)
                                dst_min_value, dst_max_value = TT(min_float32_value), TT(max_float32_value)
                            case 8:
                                min_float64_value, max_float64_value := MinFloat64ForUInt64, MaxFloat64ForUInt64
                                src_min_value, src_max_value = T(min_float64_value), T(max_float64_value)
                                dst_min_value, dst_max_value = TT(min_float64_value), TT(max_float64_value)
                        }
                    }
            }
        } else {
            switch dst_size {
                case 1:
                    min_int8_value, max_int8_value := MinInt8, MaxInt8
                    src_min_value, src_max_value = T(min_int8_value), T(max_int8_value)
                    dst_min_value, dst_max_value = TT(min_int8_value), TT(max_int8_value)
                case 2:
                    min_int16_value, max_int16_value := MinInt16, MaxInt16
                    src_min_value, src_max_value = T(min_int16_value), T(max_int16_value)
                    dst_min_value, dst_max_value = TT(min_int16_value), TT(max_int16_value)
                case 4:
                    if src_is_integer {
                        min_int32_value, max_int32_value := MinInt32, MaxInt32
                        src_min_value, src_max_value = T(min_int32_value), T(max_int32_value)
                        dst_min_value, dst_max_value = TT(min_int32_value), TT(max_int32_value)
                    } else {
                        switch src_size {
                            case 4:
                                min_float32_value, max_float32_value := MinFloat32ForInt32, MaxFloat32ForInt32
                                src_min_value, src_max_value = T(min_float32_value), T(max_float32_value)
                                dst_min_value, dst_max_value = TT(min_float32_value), TT(max_float32_value)
                            case 8:
                                min_float64_value, max_float64_value := MinFloat64ForInt32, MaxFloat64ForInt32
                                src_min_value, src_max_value = T(min_float64_value), T(max_float64_value)
                                dst_min_value, dst_max_value = TT(min_float64_value), TT(max_float64_value)
                        }
                    }
                case 8:
                    if src_is_integer {
                        min_int64_value, max_int64_value := MinInt64, MaxInt64
                        src_min_value, src_max_value = T(min_int64_value), T(max_int64_value)
                        dst_min_value, dst_max_value = TT(min_int64_value), TT(max_int64_value)
                    } else {
                        switch src_size {
                            case 4:
                                min_float32_value, max_float32_value := MinFloat32ForInt64, MaxFloat32ForInt64
                                src_min_value, src_max_value = T(min_float32_value), T(max_float32_value)
                                dst_min_value, dst_max_value = TT(min_float32_value), TT(max_float32_value)
                            case 8:
                                min_float64_value, max_float64_value := MinFloat64ForInt64, MaxFloat64ForInt64
                                src_min_value, src_max_value = T(min_float64_value), T(max_float64_value)
                                dst_min_value, dst_max_value = TT(min_float64_value), TT(max_float64_value)
                        }
                    }
            }
        }
    } else {
        switch dst_size {
            case 4:
                min_float32_value, max_float32_value := MinFloat32, MaxFloat32
                src_min_value, src_max_value = T(min_float32_value), T(max_float32_value)
                dst_min_value, dst_max_value = TT(min_float32_value), TT(max_float32_value)
            case 8:
                min_float64_value, max_float64_value := MinFloat64, MaxFloat64
                src_min_value, src_max_value = T(min_float64_value), T(max_float64_value)
                dst_min_value, dst_max_value = TT(min_float64_value), TT(max_float64_value)
        }
    }

    if src_is_integer {
        if dst_is_integer {
            var invalid bool
            switch {
                case src_size < dst_size:
                    switch {
                        case !src_is_unsigned && dst_is_unsigned:
                            invalid = (value < 0)
                    }
                case src_size == dst_size:
                    switch {
                        case src_is_unsigned != dst_is_unsigned:
                            invalid = (uint64(value) > ((uint64(1) << (dst_size*8 - 1)) - 1))
                    }
                default:
                    switch {
                        case src_is_unsigned && dst_is_unsigned:
                            invalid = (uint64(value) > ((uint64(1) << (dst_size * 8)) - 1))
                        case src_is_unsigned && !dst_is_unsigned:
                            invalid = (uint64(value) > ((uint64(1) << ((dst_size * 8) - 1)) - 1))
                        case !src_is_unsigned && dst_is_unsigned:
                            invalid = (value < 0) || (int64(value) > (int64(1)<<(dst_size*8))-1)
                        case !src_is_unsigned && !dst_is_unsigned:
                            invalid = (int64(value) < (int64(-1) << ((dst_size * 8) - 1))) || (int64(value) > (int64(1)<<((dst_size*8)-1))-1)
                    }
            }
            if invalid {
                err = fmt.Errorf("Value %v is not in allowed range < %v ... %v >", value, dst_min_value, dst_max_value)
            } else {
                result = TT(value)
            }
        } else {
            result = TT(value)
        }
    } else {
        if dst_is_integer {
            var src_is_valid bool
            if src_size == 4 {
                temp := float32(value)
                src_is_valid = (((*(*uint32)(unsafe.Pointer(&temp)) >> 23) & 0xFF) != 0xFF)
            } else {
                temp := float64(value)
                src_is_valid = (((*(*uint64)(unsafe.Pointer(&temp)) >> 52) & 0x7FF) != 0x7FF)
            }
            if !src_is_valid || (value < src_min_value) || (value > src_max_value) {
                err = fmt.Errorf("Value %v is not in allowed range < %v ... %v >", value, src_min_value, src_max_value)
            } else {
                result = TT(value)
            }
        } else {
            invalid := false
            if src_size > dst_size {
                temp := float64(value)
                src_is_valid := (((*(*uint64)(unsafe.Pointer(&temp)) >> 52) & 0x7FF) != 0x7FF)
                if src_is_valid {
                    if (value < src_min_value) || (value > src_max_value) {
                        err = fmt.Errorf("Value %v is not in allowed range < %v ... %v >", value, src_min_value, src_max_value)
                        invalid = true
                    }
                }
            }
            if !invalid {
                result = TT(value)
            }
        }
    }
    return
}

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
        result = strconv.FormatFloat(float64(value), 'G', -1, int(src_size)*8)
    }
    return
}

// String

func StringToBool(value string) (result bool, err error) {
    result, err = strconv.ParseBool(value)
    return
}

func StringToNumber[T CastableNumber](value string) (result T, err error) {
    var dst_temp T
    dst_size := unsafe.Sizeof(dst_temp)
    dst_is_integer := (T(1) / T(2)) == T(0)
    if dst_is_integer {
        dst_is_unsigned := (T(0) - T(1)) > T(0)
        if dst_is_unsigned {
            var v uint64
            v, err = strconv.ParseUint(value, 10, int(dst_size)*8)
            if err == nil {
                result = T(v)
            }
        } else {
            var v int64
            v, err = strconv.ParseInt(value, 10, int(dst_size)*8)
            if err == nil {
                result = T(v)
            }
        }
    } else {
        var v float64
        v, err = strconv.ParseFloat(value, int(dst_size)*8)
        if err == nil {
            result = T(v)
        }
    }
    return
}

// Misc

func IsNil[T any](value *T) bool {
    if value == nil {
        return true
    }

    switch reflect.TypeOf(*value).Kind() {
    case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
        return reflect.ValueOf(*value).IsNil()
    }

    return false
}

// Timestamp

func TimestampToTime(value time.Time) time.Time {
    return value.Add(-value.Sub(value.Truncate(24 * time.Hour)))
}

func TimestampToDate(value time.Time) time.Time {
    return value.Truncate(24 * time.Hour)
}

// Hex

func HexStringToUInt(value string) (result uint, err error) {
    temp, err := strconv.ParseUint(value, 16, strconv.IntSize)
    result = uint(temp)
    return
}

func UIntToHexString(value uint) string {
    return strconv.FormatUint(uint64(value), 16)
}

func HexStringToUInt64(value string) (result uint64, err error) {
    return strconv.ParseUint(value, 16, 64)
}

func UInt64ToHexString(value uint64) string {
    return strconv.FormatUint(uint64(value), 16)
}

// Float

func FloatToString(value float64) string {
    return strconv.FormatFloat(value, 'G', -1, 64)
}

func StringToFloat(value string) (result float64, err error) {
    return strconv.ParseFloat(value, 64)
}

func RangeCheckFloat64(value float64, min_value float64, max_value float64) error {
    if !((value >= min_value) && (value <= max_value)) {
        return fmt.Errorf("Value %G is not in allowed range < %G ... %G >", value, min_value, max_value)
    }

    return nil
}

func Float64ToFloat64RangeCheck(value float64, min_value float64, max_value float64) (result float64, err error) {
    err = RangeCheckFloat64(value, min_value, max_value)
    if err == nil {
        result = value
    }

    return
}

func StringToFloat64RangeCheck(value string, min_value float64, max_value float64) (result float64, err error) {
    result, err = strconv.ParseFloat(value, 64)
    if err == nil {
        err = RangeCheckFloat64(result, min_value, max_value)
    }

    return
}

// DecodeScalar

func ParseScalar(value *any) bool {

    value_string, ok := (*value).(string)
    if !ok {
        return false
    }

    var err error
    var value_uint64 uint64
    value_uint64, err = strconv.ParseUint(value_string, 10, 64)
    if err == nil {
        *value = value_uint64
        return true
    }

    var value_int64 int64
    value_int64, err = strconv.ParseInt(value_string, 10, 64)
    if err == nil {
        *value = value_int64
        return true
    }

    var value_float64 float64
    value_float64, err = strconv.ParseFloat(value_string, 64)
    if err == nil {
        *value = value_float64
        return true
    }

    var value_bool bool
    value_bool, err = strconv.ParseBool(value_string)
    if err == nil {
        *value = value_bool
        return true
    }

    return false
}

// Cast

func Cast[T Castable, TT Castable](value T) (result TT, err error) {
    switch typed_value := any(value).(type) {
    case bool:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result = typed_value
        case *uint:
            *typed_result, err = BoolToNumber[uint](typed_value)
        case *uint8:
            *typed_result, err = BoolToNumber[uint8](typed_value)
        case *uint16:
            *typed_result, err = BoolToNumber[uint16](typed_value)
        case *uint32:
            *typed_result, err = BoolToNumber[uint32](typed_value)
        case *uint64:
            *typed_result, err = BoolToNumber[uint64](typed_value)
        case *int:
            *typed_result, err = BoolToNumber[int](typed_value)
        case *int8:
            *typed_result, err = BoolToNumber[int8](typed_value)
        case *int16:
            *typed_result, err = BoolToNumber[int16](typed_value)
        case *int32:
            *typed_result, err = BoolToNumber[int32](typed_value)
        case *int64:
            *typed_result, err = BoolToNumber[int64](typed_value)
        case *float32:
            *typed_result, err = BoolToNumber[float32](typed_value)
        case *float64:
            *typed_result, err = BoolToNumber[float64](typed_value)
        case *string:
            *typed_result, err = BoolToString(typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case uint:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[uint, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[uint, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[uint, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[uint, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[uint, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[uint, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[uint, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[uint, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[uint, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[uint, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[uint, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[uint, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[uint](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case uint8:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[uint8, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[uint8, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[uint8, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[uint8, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[uint8, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[uint8, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[uint8, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[uint8, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[uint8, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[uint8, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[uint8, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[uint8, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[uint8](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case uint16:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[uint16, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[uint16, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[uint16, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[uint16, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[uint16, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[uint16, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[uint16, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[uint16, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[uint16, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[uint16, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[uint16, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[uint16, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[uint16](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case uint32:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[uint32, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[uint32, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[uint32, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[uint32, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[uint32, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[uint32, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[uint32, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[uint32, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[uint32, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[uint32, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[uint32, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[uint32, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[uint32](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case uint64:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[uint64, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[uint64, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[uint64, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[uint64, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[uint64, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[uint64, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[uint64, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[uint64, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[uint64, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[uint64, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[uint64, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[uint64, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[uint64](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case int:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[int, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[int, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[int, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[int, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[int, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[int, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[int, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[int, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[int, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[int, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[int, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[int, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[int](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case int8:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[int8, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[int8, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[int8, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[int8, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[int8, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[int8, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[int8, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[int8, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[int8, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[int8, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[int8, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[int8, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[int8](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case int16:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[int16, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[int16, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[int16, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[int16, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[int16, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[int16, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[int16, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[int16, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[int16, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[int16, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[int16, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[int16, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[int16](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case int32:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[int32, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[int32, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[int32, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[int32, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[int32, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[int32, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[int32, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[int32, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[int32, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[int32, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[int32, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[int32, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[int32](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case int64:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[int64, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[int64, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[int64, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[int64, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[int64, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[int64, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[int64, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[int64, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[int64, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[int64, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[int64, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[int64, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[int64](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case float32:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[float32, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[float32, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[float32, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[float32, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[float32, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[float32, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[float32, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[float32, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[float32, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[float32, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[float32, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[float32, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[float32](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case float64:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = NumberToBool(typed_value)
        case *uint:
            *typed_result, err = NumberToNumber[float64, uint](typed_value)
        case *uint8:
            *typed_result, err = NumberToNumber[float64, uint8](typed_value)
        case *uint16:
            *typed_result, err = NumberToNumber[float64, uint16](typed_value)
        case *uint32:
            *typed_result, err = NumberToNumber[float64, uint32](typed_value)
        case *uint64:
            *typed_result, err = NumberToNumber[float64, uint64](typed_value)
        case *int:
            *typed_result, err = NumberToNumber[float64, int](typed_value)
        case *int8:
            *typed_result, err = NumberToNumber[float64, int8](typed_value)
        case *int16:
            *typed_result, err = NumberToNumber[float64, int16](typed_value)
        case *int32:
            *typed_result, err = NumberToNumber[float64, int32](typed_value)
        case *int64:
            *typed_result, err = NumberToNumber[float64, int64](typed_value)
        case *float32:
            *typed_result, err = NumberToNumber[float64, float32](typed_value)
        case *float64:
            *typed_result, err = NumberToNumber[float64, float64](typed_value)
        case *string:
            *typed_result, err = NumberToString[float64](typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    case string:
        switch typed_result := any(&result).(type) {
        case *bool:
            *typed_result, err = StringToBool(typed_value)
        case *uint:
            *typed_result, err = StringToNumber[uint](typed_value)
        case *uint8:
            *typed_result, err = StringToNumber[uint8](typed_value)
        case *uint16:
            *typed_result, err = StringToNumber[uint16](typed_value)
        case *uint32:
            *typed_result, err = StringToNumber[uint32](typed_value)
        case *uint64:
            *typed_result, err = StringToNumber[uint64](typed_value)
        case *int:
            *typed_result, err = StringToNumber[int](typed_value)
        case *int8:
            *typed_result, err = StringToNumber[int8](typed_value)
        case *int16:
            *typed_result, err = StringToNumber[int16](typed_value)
        case *int32:
            *typed_result, err = StringToNumber[int32](typed_value)
        case *int64:
            *typed_result, err = StringToNumber[int64](typed_value)
        case *float32:
            *typed_result, err = StringToNumber[float32](typed_value)
        case *float64:
            *typed_result, err = StringToNumber[float64](typed_value)
        case *string:
            *typed_result = string(typed_value)
        default:
            err = fmt.Errorf("Unsupported destination data type %T", typed_result)
        }
    default:
        err = fmt.Errorf("Unsupported source data type %T", typed_value)
    }
    return
}

func CastDef[T Castable, TT Castable](value T, default_value TT) (result TT) {
    var err error
    if result, err = Cast[T, TT](value); err != nil {
        result = default_value
    }
    return
}

func CastIndirect[T Castable, TT Castable](source T, destination *TT) (err error) {
    if destination != nil {
        *destination, err = Cast[T, TT](source)
    }
    return
}

func CastIndirectDef[T Castable, TT Castable](source T, destination *TT, default_value TT) {
    if destination != nil {
        var err error
        if *destination, err = Cast[T, TT](source); err != nil {
            *destination = default_value
        }
    }
}
