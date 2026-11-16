package Cast

// Reference implementation and the shared check for the generated test cast_switch_test.go.
//
// The reference is written independently of the generated code: numbers are compared as exact
// values (math/big), the result is written through reflect. Only the text formats (strconv)
// are the same by definition.

import (
    "errors"
    "math"
    "math/big"
    "reflect"
    "strconv"
    "testing"
    "unsafe"
)

// Strings used as source values: valid numbers at the limits of all types, floats, special
// values, booleans and invalid text.
var test_string_values = []string{
    "0", "1", "-1", "+1", "007",
    "127", "128", "-128", "-129", "255", "256",
    "32767", "32768", "-32768", "-32769", "65535", "65536",
    "2147483647", "2147483648", "-2147483648", "-2147483649", "4294967295", "4294967296",
    "9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809",
    "18446744073709551615", "18446744073709551616",
    "1.5", "-0.5", "1e10", "1e300", "-1e300", "3.4028234663852886e38", "3.5e38", "1e-50",
    "NaN", "Inf", "-Inf", "+Inf",
    "true", "false", "TRUE", "t", "F",
    "", " 1", "1 ", "x", "0x10", "1_000",
}

func isSignedKind(kind reflect.Kind) bool {
    return kind >= reflect.Int && kind <= reflect.Int64
}

func isUnsignedKind(kind reflect.Kind) bool {
    return kind >= reflect.Uint && kind <= reflect.Uintptr
}

func isFloatKind(kind reflect.Kind) bool {
    return kind == reflect.Float32 || kind == reflect.Float64
}

// kindLimits returns the exact limits of an integer type.
func kindLimits(t reflect.Type) (min_value, max_value *big.Int) {
    bits := uint(t.Bits())
    one := big.NewInt(1)
    if isUnsignedKind(t.Kind()) {
        return big.NewInt(0), new(big.Int).Sub(new(big.Int).Lsh(one, bits), one)
    }
    return new(big.Int).Neg(new(big.Int).Lsh(one, bits-1)), new(big.Int).Sub(new(big.Int).Lsh(one, bits-1), one)
}

// testValues returns the source values for type S.
func testValues[S Castable]() (result []S) {
    t := reflect.TypeFor[S]()
    add := func(set func(v reflect.Value)) {
        var value S
        set(reflect.ValueOf(&value).Elem())
        result = append(result, value)
    }
    switch kind := t.Kind(); {
    case kind == reflect.Bool:
        add(func(v reflect.Value) { v.SetBool(false) })
        add(func(v reflect.Value) { v.SetBool(true) })
    case kind == reflect.String:
        for _, text := range test_string_values {
            add(func(v reflect.Value) { v.SetString(text) })
        }
    case isFloatKind(kind):
        for _, pattern := range test_float_patterns {
            add(func(v reflect.Value) { v.SetFloat(pattern) })
        }
    default:
        min_value, max_value := kindLimits(t)
        for _, value := range integerTestValues() {
            if value.Cmp(min_value) < 0 || value.Cmp(max_value) > 0 {
                continue
            }
            if isUnsignedKind(kind) {
                add(func(v reflect.Value) { v.SetUint(value.Uint64()) })
            } else {
                add(func(v reflect.Value) { v.SetInt(value.Int64()) })
            }
        }
    }
    return
}

// parseErrorKind maps a strconv error to the package error.
func parseErrorKind(err error) error {
    if errors.Is(err, strconv.ErrRange) {
        return ErrOutOfRange
    }
    return ErrInvalidSyntax
}

// reference converts value to D. On error it returns the expected package error
// (ErrOutOfRange or ErrInvalidSyntax) and the zero value.
func reference[S Castable, D Castable](value S) (result D, err error) {
    src := reflect.ValueOf(value)
    dst := reflect.ValueOf(&result).Elem()
    src_kind, dst_kind := src.Kind(), dst.Kind()

    switch {
    // string -> *
    case src_kind == reflect.String:
        text := src.String()
        switch {
        case dst_kind == reflect.String:
            dst.SetString(text)
        case dst_kind == reflect.Bool:
            parsed, parse_err := strconv.ParseBool(text)
            if parse_err != nil {
                return result, parseErrorKind(parse_err)
            }
            dst.SetBool(parsed)
        case isFloatKind(dst_kind):
            parsed, parse_err := strconv.ParseFloat(text, dst.Type().Bits())
            if parse_err != nil {
                return result, parseErrorKind(parse_err)
            }
            dst.SetFloat(parsed)
        case isUnsignedKind(dst_kind):
            parsed, parse_err := strconv.ParseUint(text, 10, dst.Type().Bits())
            if parse_err != nil {
                return result, parseErrorKind(parse_err)
            }
            dst.SetUint(parsed)
        default:
            parsed, parse_err := strconv.ParseInt(text, 10, dst.Type().Bits())
            if parse_err != nil {
                return result, parseErrorKind(parse_err)
            }
            dst.SetInt(parsed)
        }

    // bool -> *
    case src_kind == reflect.Bool:
        switch {
        case dst_kind == reflect.Bool:
            dst.SetBool(src.Bool())
        case dst_kind == reflect.String:
            dst.SetString(strconv.FormatBool(src.Bool()))
        case src.Bool() && isFloatKind(dst_kind):
            dst.SetFloat(1)
        case src.Bool() && isUnsignedKind(dst_kind):
            dst.SetUint(1)
        case src.Bool():
            dst.SetInt(1)
        }

    // number -> string
    case dst_kind == reflect.String:
        switch {
        case isFloatKind(src_kind):
            dst.SetString(strconv.FormatFloat(src.Float(), 'g', -1, src.Type().Bits()))
        case isUnsignedKind(src_kind):
            dst.SetString(strconv.FormatUint(src.Uint(), 10))
        default:
            dst.SetString(strconv.FormatInt(src.Int(), 10))
        }

    // number -> bool
    case dst_kind == reflect.Bool:
        switch {
        case isFloatKind(src_kind):
            dst.SetBool(src.Float() != 0) // NaN != 0 is true
        case isUnsignedKind(src_kind):
            dst.SetBool(src.Uint() != 0)
        default:
            dst.SetBool(src.Int() != 0)
        }

    // number -> number, compared as exact values
    default:
        exact := new(big.Float).SetPrec(1100)
        switch {
        case isFloatKind(src_kind):
            float := src.Float()
            if math.IsNaN(float) || math.IsInf(float, 0) {
                if isFloatKind(dst_kind) {
                    dst.SetFloat(float) // NaN and infinities are kept
                    return
                }
                return result, ErrOutOfRange
            }
            exact.SetFloat64(float)
        case isUnsignedKind(src_kind):
            exact.SetUint64(src.Uint())
        default:
            exact.SetInt64(src.Int())
        }
        if isFloatKind(dst_kind) {
            if dst_kind == reflect.Float32 && exact.Cmp(new(big.Float).SetFloat64(math.MaxFloat32)) > 0 ||
                dst_kind == reflect.Float32 && exact.Cmp(new(big.Float).SetFloat64(-math.MaxFloat32)) < 0 {
                return result, ErrOutOfRange
            }
            if dst_kind == reflect.Float32 {
                float, _ := exact.Float32() // rounds to nearest even, as the Go conversion
                dst.SetFloat(float64(float))
            } else {
                float, _ := exact.Float64()
                dst.SetFloat(float)
            }
            return
        }
        min_value, max_value := kindLimits(dst.Type())
        if exact.Cmp(new(big.Float).SetInt(min_value)) < 0 || exact.Cmp(new(big.Float).SetInt(max_value)) > 0 {
            return result, ErrOutOfRange
        }
        integer, _ := exact.Int(nil) // truncates toward zero, as the Go conversion
        if isUnsignedKind(dst_kind) {
            dst.SetUint(integer.Uint64())
        } else {
            dst.SetInt(integer.Int64())
        }
    }
    return
}

// sameValue compares two values, NaN equals NaN.
func sameValue[D Castable](a, b D) bool {
    if a == b {
        return true
    }
    va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
    return isFloatKind(va.Kind()) && math.IsNaN(va.Float()) && math.IsNaN(vb.Float())
}

// checkPair checks all conversion paths S -> D against the reference:
//   - Cast[S, D]: fast path,
//   - Cast[DS, DD], Cast[DS, D], Cast[S, DD]: derived types, castDerived and castKind,
//   - direct: the exported non-generic function (Int8ToUInt64 ...).
//
// DS and DD are types derived from S and D (same memory layout, no methods).
func checkPair[S Castable, D Castable, DS Castable, DD Castable](t *testing.T, direct func(S) (D, error)) {
    t.Helper()
    for _, value := range testValues[S]() {
        want, want_err := reference[S, D](value)
        check := func(path string, got D, got_err error) {
            t.Helper()
            switch {
            case want_err != nil && !errors.Is(got_err, want_err):
                t.Errorf("%s(%#v): expected error %v, got %v, %v", path, value, want_err, got, got_err)
            case want_err != nil && got != *new(D):
                t.Errorf("%s(%#v): expected zero result on error, got %v", path, value, got)
            case want_err == nil && got_err != nil:
                t.Errorf("%s(%#v): unexpected error %v", path, value, got_err)
            case want_err == nil && !sameValue(got, want):
                t.Errorf("%s(%#v): expected %#v, got %#v", path, value, want, got)
            }
        }
        name := func(s, d string) string { return "Cast[" + s + ", " + d + "]" }
        s_name, d_name := reflect.TypeFor[S]().Name(), reflect.TypeFor[D]().Name()
        ds_name, dd_name := reflect.TypeFor[DS]().Name(), reflect.TypeFor[DD]().Name()
        derived := *(*DS)(unsafe.Pointer(&value))

        got, got_err := Cast[S, D](value)
        check(name(s_name, d_name), got, got_err)

        got_dd, got_err := Cast[DS, DD](derived)
        check(name(ds_name, dd_name), *(*D)(unsafe.Pointer(&got_dd)), got_err)

        got, got_err = Cast[DS, D](derived)
        check(name(ds_name, d_name), got, got_err)

        got_dd, got_err = Cast[S, DD](value)
        check(name(s_name, dd_name), *(*D)(unsafe.Pointer(&got_dd)), got_err)

        got, got_err = direct(value)
        check("direct "+s_name+" -> "+d_name, got, got_err)
    }
}
