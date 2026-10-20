package Cast

import (
    "math"
    "math/big"
    "reflect"
    "strconv"
    "testing"
)

type integer interface {
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type float interface {
    ~float32 | ~float64
}

func isUnsigned(kind reflect.Kind) bool {
    switch kind {
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        return true
    }
    return false
}

func integerLimits[T integer]() (min_value *big.Int, max_value *big.Int) {
    t := reflect.TypeOf(T(0))
    bits := uint(t.Bits())
    one := big.NewInt(1)
    if isUnsigned(t.Kind()) {
        min_value = big.NewInt(0)
        max_value = new(big.Int).Sub(new(big.Int).Lsh(one, bits), one)
    } else {
        min_value = new(big.Int).Neg(new(big.Int).Lsh(one, bits-1))
        max_value = new(big.Int).Sub(new(big.Int).Lsh(one, bits-1), one)
    }
    return
}

func toBig[T integer](value T) *big.Int {
    v := reflect.ValueOf(value)
    if isUnsigned(v.Kind()) {
        return new(big.Int).SetUint64(v.Uint())
    }
    return big.NewInt(v.Int())
}

func floatLimits[T float]() (min_value *big.Float, max_value *big.Float) {
    t := reflect.TypeOf(T(0))
    bits := uint(t.Bits())
    if bits == 32 {
        min_value = new(big.Float).SetFloat64(-math.MaxFloat32)
        max_value = new(big.Float).SetFloat64(math.MaxFloat32)
    } else {
        min_value = new(big.Float).SetFloat64(-math.MaxFloat64)
        max_value = new(big.Float).SetFloat64(math.MaxFloat64)
    }
    return
}

func floatPrecision[T float]() (result uint) {
    t := reflect.TypeOf(T(0))
    bits := uint(t.Bits())
    if bits == 32 {
        result = 24
    } else {
        result = 53
    }
    return
}

var test_integer_patterns = [...]uint64 {
    0x00,
    0x1, 0x2, 0x4, 0x8,
    0x10, 0x20, 0x40, 0x80,
    0x100, 0x200, 0x400, 0x800,
    0x1000, 0x2000, 0x4000, 0x8000,
    0x10000, 0x20000, 0x40000, 0x80000,
    0x100000, 0x200000, 0x400000, 0x800000,
    0x1000000, 0x2000000, 0x4000000, 0x8000000,
    0x10000000, 0x20000000, 0x40000000, 0x80000000,
    0x100000000, 0x200000000, 0x400000000, 0x800000000,
    0x1000000000, 0x2000000000, 0x4000000000, 0x8000000000,
    0x10000000000, 0x20000000000, 0x40000000000, 0x80000000000,
    0x100000000000, 0x200000000000, 0x400000000000, 0x800000000000,
    0x1000000000000, 0x2000000000000, 0x4000000000000, 0x8000000000000,
    0x10000000000000, 0x20000000000000, 0x40000000000000, 0x80000000000000,
    0x100000000000000, 0x200000000000000, 0x400000000000000, 0x800000000000000,
    0x1000000000000000, 0x2000000000000000, 0x4000000000000000, 0x8000000000000000,
    0x7F, 0x80, 0x81, 0xFF,
    0x7FFF, 0x8000, 0x8001, 0xFFFF,
    0x7FFFFFFF, 0x80000000, 0x80000001, 0xFFFFFFFF,
    0x7FFFFFFFFFFFFFFF, 0x8000000000000000, 0x8000000000000001, 0xFFFFFFFFFFFFFFFF,
}

var test_float_patterns = [...]float64 {
    0, math.Copysign(0, -1), math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat64,
    0.5, -0.5, 1, -1, 1.5, -1.5,
    127, 127.5, 128, -128, -128.5, -129, 255, 255.5, 256, -256,
    32767, 32767.5, 32768, -32768, -32768.5, -32769, 65535, 65535.5, 65536,
    2147483647, 2147483647.5, 2147483648, -2147483648, -2147483648.5, -2147483649, 4294967295, 4294967295.5, 4294967296,
    0x1p63, 0x1.fffffffffffffp62, -0x1p63, -0x1.0000000000001p63, 0x1p64, 0x1.fffffffffffffp63,
    math.MaxFloat32, 0x1.fffffe0000001p127, 0x1p128, math.MaxFloat64,
    -math.MaxFloat32, -0x1.fffffe0000001p127, -0x1p128, -math.MaxFloat64,
    math.Inf(1), math.Inf(-1), math.NaN(),
}

// Every pattern is used twice: once as uint64 and once as int64 (the same bits).
func integerTestValues() (result []*big.Int) {
    for _, bits := range test_integer_patterns {
        result = append(result, new(big.Int).SetUint64(bits), big.NewInt(int64(bits)))
    }
    return
}

func fromBig[T integer](value *big.Int) T {
    if value.Sign() < 0 {
        return T(value.Int64())
    }
    return T(value.Uint64())
}

func testIntegerCast[T integer, TT integer](t *testing.T) {
    t.Helper()

    source_min_value, source_max_value := integerLimits[T]()
    dest_min_value, dest_max_value := integerLimits[TT]()

    for _, test_value := range integerTestValues() {
        if (test_value.Cmp(source_min_value) >= 0) && (test_value.Cmp(source_max_value) <= 0) {
            source := fromBig[T](test_value)
            assumed_result := (test_value.Cmp(dest_min_value) >= 0) && (test_value.Cmp(dest_max_value) <= 0)
            test_result, err := Cast[T, TT](source)
            if assumed_result {
                if err != nil {
                    t.Errorf("Cast[%T, %T](%v): unexpected error: %v", T(0), TT(0), source, err)
                } else if toBig(test_result).Cmp(test_value) != 0 {
                    t.Errorf("Cast[%T, %T](%v): expected %v, got %v", T(0), TT(0), source, test_value, test_result)
                }
            } else if err == nil {
                t.Errorf("Cast[%T, %T](%v): expected out of range error, got result %v", T(0), TT(0), source, test_result)
            }
        }
    }
}

// The integer must be converted to the nearest float (round half to even).
func testIntegerToFloatCast[T integer, TT float](t *testing.T) {
    t.Helper()

    source_min_value, source_max_value := integerLimits[T]()

    for _, test_value := range integerTestValues() {
        if (test_value.Cmp(source_min_value) >= 0) && (test_value.Cmp(source_max_value) <= 0) {
            source := fromBig[T](test_value)
            result, err := Cast[T, TT](source)
            if err != nil {
                t.Errorf("Cast[%T, %T](%v): unexpected error: %v", T(0), TT(0), source, err)
                continue
            }
            assumed_result, _ := new(big.Float).SetPrec(floatPrecision[TT]()).SetMode(big.ToNearestEven).SetInt(test_value).Float64()
            if float64(result) != assumed_result {
                t.Errorf("Cast[%T, %T](%v): expected %v, got %v", T(0), TT(0), source, assumed_result, result)
            }
        }
    }
}

func testFloatToIntegerCast[T float, TT integer](t *testing.T) {
    t.Helper()

    dest_min_value, dest_max_value := integerLimits[TT]()

    for _, pattern := range test_float_patterns {
        source := T(pattern)
        result, err := Cast[T, TT](source)
        if math.IsNaN(float64(source)) || math.IsInf(float64(source), 0) {
            if err == nil {
                t.Errorf("Cast[%T, %T](%v): expected error, got result %v", T(0), TT(0), source, result)
            }
            continue
        }
        test_value := new(big.Float).SetFloat64(float64(source))
        assumed_result := (test_value.Cmp(new(big.Float).SetInt(dest_min_value)) >= 0) && (test_value.Cmp(new(big.Float).SetInt(dest_max_value)) <= 0)
        if assumed_result {
            if err != nil {
                t.Errorf("Cast[%T, %T](%v): unexpected error: %v", T(0), TT(0), source, err)
                continue
            }
            expected_value, _ := test_value.Int(nil)
            if toBig(result).Cmp(expected_value) != 0 {
                t.Errorf("Cast[%T, %T](%v): expected %v, got %v", T(0), TT(0), source, expected_value, result)
            }
        } else if err == nil {
            t.Errorf("Cast[%T, %T](%v): expected out of range error, got result %v", T(0), TT(0), source, result)
        }
    }
}

func testFloatToFloatCast[T float, TT float](t *testing.T) {
    t.Helper()

    dest_min_value, dest_max_value := floatLimits[TT]()

    for _, pattern := range test_float_patterns {
        source := T(pattern)
        result, err := Cast[T, TT](source)
        if math.IsNaN(float64(source)) || math.IsInf(float64(source), 0) {
            if err != nil {
                t.Errorf("Cast[%T, %T](%v): unexpected error: %v", T(0), TT(0), source, err)
            } else if math.IsNaN(float64(source)) != math.IsNaN(float64(result)) || (!math.IsNaN(float64(source)) && float64(source) != float64(result)) {
                t.Errorf("Cast[%T, %T](%v): got %v", T(0), TT(0), source, result)
            }
            continue
        }
        test_value := new(big.Float).SetFloat64(float64(source))
        assumed_result := (test_value.Cmp(dest_min_value) >= 0) && (test_value.Cmp(dest_max_value) <= 0)
        if assumed_result {
            if err != nil {
                t.Errorf("Cast[%T, %T](%v): unexpected error: %v", T(0), TT(0), source, err)
            } else if float64(result) != float64(TT(source)) {
                t.Errorf("Cast[%T, %T](%v): got %v", T(0), TT(0), source, result)
            }
        } else if err == nil {
            t.Errorf("Cast[%T, %T](%v): expected out of range error, got result %v", T(0), TT(0), source, result)
        }
    }
}


func testIntegerCastFrom[T integer](t *testing.T) {
    t.Run("to uint", testIntegerCast[T, uint])
    t.Run("to uint8", testIntegerCast[T, uint8])
    t.Run("to uint16", testIntegerCast[T, uint16])
    t.Run("to uint32", testIntegerCast[T, uint32])
    t.Run("to uint64", testIntegerCast[T, uint64])
    t.Run("to int", testIntegerCast[T, int])
    t.Run("to int8", testIntegerCast[T, int8])
    t.Run("to int16", testIntegerCast[T, int16])
    t.Run("to int32", testIntegerCast[T, int32])
    t.Run("to int64", testIntegerCast[T, int64])
    t.Run("to float32", testIntegerToFloatCast[T, float32])
    t.Run("to float64", testIntegerToFloatCast[T, float64])
}

func TestCastIntegerFromUint(t *testing.T) {
    testIntegerCastFrom[uint](t)
}
func TestCastIntegerFromUint8(t *testing.T) {
    testIntegerCastFrom[uint8](t)
}
func TestCastIntegerFromUint16(t *testing.T) {
    testIntegerCastFrom[uint16](t)
}
func TestCastIntegerFromUint32(t *testing.T) {
    testIntegerCastFrom[uint32](t)
}
func TestCastIntegerFromUint64(t *testing.T) {
    testIntegerCastFrom[uint64](t)
}
func TestCastIntegerFromInt(t *testing.T) {
    testIntegerCastFrom[int](t)
}
func TestCastIntegerFromInt8(t *testing.T) {
    testIntegerCastFrom[int8](t)
}
func TestCastIntegerFromInt16(t *testing.T) {
    testIntegerCastFrom[int16](t)
}
func TestCastIntegerFromInt32(t *testing.T) {
    testIntegerCastFrom[int32](t)
}
func TestCastIntegerFromInt64(t *testing.T) {
    testIntegerCastFrom[int64](t)
}

func testFloatCastFrom[T float](t *testing.T) {
    t.Run("to uint", testFloatToIntegerCast[T, uint])
    t.Run("to uint8", testFloatToIntegerCast[T, uint8])
    t.Run("to uint16", testFloatToIntegerCast[T, uint16])
    t.Run("to uint32", testFloatToIntegerCast[T, uint32])
    t.Run("to uint64", testFloatToIntegerCast[T, uint64])
    t.Run("to int", testFloatToIntegerCast[T, int])
    t.Run("to int8", testFloatToIntegerCast[T, int8])
    t.Run("to int16", testFloatToIntegerCast[T, int16])
    t.Run("to int32", testFloatToIntegerCast[T, int32])
    t.Run("to int64", testFloatToIntegerCast[T, int64])
    t.Run("to float32", testFloatToFloatCast[T, float32])
    t.Run("to float64", testFloatToFloatCast[T, float64])
}

func TestCastFloatFromFloat32(t *testing.T) {
    testFloatCastFrom[float32](t)
}
func TestCastFloatFromFloat64(t *testing.T) {
    testFloatCastFrom[float64](t)
}

type number interface {
    integer | float
}

// Bool

func testBoolToNumber[TT number](t *testing.T) {
    t.Helper()

    if result, err := Cast[bool, TT](true); err != nil || result != TT(1) {
        t.Errorf("Cast[bool, %T](true): expected 1, got %v, error %v", TT(0), result, err)
    }
    if result, err := Cast[bool, TT](false); err != nil || result != TT(0) {
        t.Errorf("Cast[bool, %T](false): expected 0, got %v, error %v", TT(0), result, err)
    }
}

func TestCastBoolTo(t *testing.T) {
    t.Run("to uint", testBoolToNumber[uint])
    t.Run("to uint8", testBoolToNumber[uint8])
    t.Run("to uint16", testBoolToNumber[uint16])
    t.Run("to uint32", testBoolToNumber[uint32])
    t.Run("to uint64", testBoolToNumber[uint64])
    t.Run("to int", testBoolToNumber[int])
    t.Run("to int8", testBoolToNumber[int8])
    t.Run("to int16", testBoolToNumber[int16])
    t.Run("to int32", testBoolToNumber[int32])
    t.Run("to int64", testBoolToNumber[int64])
    t.Run("to float32", testBoolToNumber[float32])
    t.Run("to float64", testBoolToNumber[float64])
    t.Run("to bool", func(t *testing.T) {
        for _, value := range []bool{false, true} {
            if result, err := Cast[bool, bool](value); err != nil || result != value {
                t.Errorf("Cast[bool, bool](%v): got %v, error %v", value, result, err)
            }
        }
    })
    t.Run("to string", func(t *testing.T) {
        for _, value := range []bool{false, true} {
            if result, err := Cast[bool, string](value); err != nil || result != strconv.FormatBool(value) {
                t.Errorf("Cast[bool, string](%v): got %q, error %v", value, result, err)
            }
        }
    })
}

// Number to bool: everything except zero is true.

func testIntegerToBool[T integer](t *testing.T) {
    t.Helper()

    source_min_value, source_max_value := integerLimits[T]()

    for _, test_value := range integerTestValues() {
        if (test_value.Cmp(source_min_value) >= 0) && (test_value.Cmp(source_max_value) <= 0) {
            source := fromBig[T](test_value)
            assumed_result := test_value.Sign() != 0
            if result, err := Cast[T, bool](source); err != nil || result != assumed_result {
                t.Errorf("Cast[%T, bool](%v): expected %v, got %v, error %v", T(0), source, assumed_result, result, err)
            }
        }
    }
}

func testFloatToBool[T float](t *testing.T) {
    t.Helper()

    for _, pattern := range test_float_patterns {
        source := T(pattern)
        assumed_result := math.IsNaN(float64(source)) || (source != 0)
        if result, err := Cast[T, bool](source); err != nil || result != assumed_result {
            t.Errorf("Cast[%T, bool](%v): expected %v, got %v, error %v", T(0), source, assumed_result, result, err)
        }
    }
}

func TestCastNumberToBool(t *testing.T) {
    t.Run("from uint", testIntegerToBool[uint])
    t.Run("from uint8", testIntegerToBool[uint8])
    t.Run("from uint16", testIntegerToBool[uint16])
    t.Run("from uint32", testIntegerToBool[uint32])
    t.Run("from uint64", testIntegerToBool[uint64])
    t.Run("from int", testIntegerToBool[int])
    t.Run("from int8", testIntegerToBool[int8])
    t.Run("from int16", testIntegerToBool[int16])
    t.Run("from int32", testIntegerToBool[int32])
    t.Run("from int64", testIntegerToBool[int64])
    t.Run("from float32", testFloatToBool[float32])
    t.Run("from float64", testFloatToBool[float64])
}

// Number to string

func testIntegerToString[T integer](t *testing.T) {
    t.Helper()

    source_min_value, source_max_value := integerLimits[T]()

    for _, test_value := range integerTestValues() {
        if (test_value.Cmp(source_min_value) >= 0) && (test_value.Cmp(source_max_value) <= 0) {
            source := fromBig[T](test_value)
            if result, err := Cast[T, string](source); err != nil || result != test_value.String() {
                t.Errorf("Cast[%T, string](%v): expected %q, got %q, error %v", T(0), source, test_value.String(), result, err)
            }
        }
    }
}

func testFloatToString[T float](t *testing.T) {
    t.Helper()

    bits := reflect.TypeOf(T(0)).Bits()

    for _, pattern := range test_float_patterns {
        source := T(pattern)
        assumed_result := strconv.FormatFloat(float64(source), 'G', -1, bits)
        if result, err := Cast[T, string](source); err != nil || result != assumed_result {
            t.Errorf("Cast[%T, string](%v): expected %q, got %q, error %v", T(0), source, assumed_result, result, err)
        }
    }
}

func TestCastNumberToString(t *testing.T) {
    t.Run("from uint", testIntegerToString[uint])
    t.Run("from uint8", testIntegerToString[uint8])
    t.Run("from uint16", testIntegerToString[uint16])
    t.Run("from uint32", testIntegerToString[uint32])
    t.Run("from uint64", testIntegerToString[uint64])
    t.Run("from int", testIntegerToString[int])
    t.Run("from int8", testIntegerToString[int8])
    t.Run("from int16", testIntegerToString[int16])
    t.Run("from int32", testIntegerToString[int32])
    t.Run("from int64", testIntegerToString[int64])
    t.Run("from float32", testFloatToString[float32])
    t.Run("from float64", testFloatToString[float64])
}

// String to bool

func TestCastStringToBool(t *testing.T) {
    for value, assumed_result := range map[string]bool{
        "1": true, "t": true, "T": true, "true": true, "TRUE": true, "True": true,
        "0": false, "f": false, "F": false, "false": false, "FALSE": false, "False": false,
    } {
        if result, err := Cast[string, bool](value); err != nil || result != assumed_result {
            t.Errorf("Cast[string, bool](%q): expected %v, got %v, error %v", value, assumed_result, result, err)
        }
    }
    for _, value := range []string{"", " ", "2", "-1", "yes", "no", "tRuE", " true", "true "} {
        if result, err := Cast[string, bool](value); err == nil {
            t.Errorf("Cast[string, bool](%q): expected error, got result %v", value, result)
        }
    }
}

// String to number

var test_invalid_numbers = [...]string{"", " ", "abc", "1,5", "0x10", " 1", "1 ", "--1", "1e", "."}

func testStringToInteger[TT integer](t *testing.T) {
    t.Helper()

    dest_min_value, dest_max_value := integerLimits[TT]()
    one := big.NewInt(1)

    for _, test_value := range integerTestValues() {
        value := test_value.String()
        result, err := Cast[string, TT](value)
        if (test_value.Cmp(dest_min_value) >= 0) && (test_value.Cmp(dest_max_value) <= 0) {
            if err != nil {
                t.Errorf("Cast[string, %T](%q): unexpected error: %v", TT(0), value, err)
            } else if toBig(result).Cmp(test_value) != 0 {
                t.Errorf("Cast[string, %T](%q): expected %v, got %v", TT(0), value, test_value, result)
            }
        } else if err == nil {
            t.Errorf("Cast[string, %T](%q): expected out of range error, got result %v", TT(0), value, result)
        }
    }

    for _, test_value := range []*big.Int{new(big.Int).Sub(dest_min_value, one), new(big.Int).Add(dest_max_value, one)} {
        value := test_value.String()
        if result, err := Cast[string, TT](value); err == nil {
            t.Errorf("Cast[string, %T](%q): expected out of range error, got result %v", TT(0), value, result)
        }
    }

    for _, value := range test_invalid_numbers {
        if result, err := Cast[string, TT](value); err == nil {
            t.Errorf("Cast[string, %T](%q): expected error, got result %v", TT(0), value, result)
        }
    }
    for _, value := range []string{"1_0", "1.5"} {
        if result, err := Cast[string, TT](value); err == nil {
            t.Errorf("Cast[string, %T](%q): expected error, got result %v", TT(0), value, result)
        }
    }
}

func testStringToFloat[TT float](t *testing.T) {
    t.Helper()

    for value, assumed_result := range map[string]float64{
        "0": 0, "-0": math.Copysign(0, -1), "1": 1, "-1": -1, "1.5": 1.5, "-0.5": -0.5, "1e10": 1e10,
        "4294967296": 4294967296, "-2147483648": -2147483648, "Inf": math.Inf(1), "-Inf": math.Inf(-1),
    } {
        result, err := Cast[string, TT](value)
        if err != nil {
            t.Errorf("Cast[string, %T](%q): unexpected error: %v", TT(0), value, err)
        } else if float64(result) != assumed_result || math.Signbit(float64(result)) != math.Signbit(assumed_result) {
            t.Errorf("Cast[string, %T](%q): expected %v, got %v", TT(0), value, assumed_result, result)
        }
    }

    if result, err := Cast[string, TT]("NaN"); err != nil || !math.IsNaN(float64(result)) {
        t.Errorf("Cast[string, %T](\"NaN\"): expected NaN, got %v, error %v", TT(0), result, err)
    }

    _, dest_max_value := floatLimits[TT]()
    max_value, _ := dest_max_value.Float64()
    max_string := dest_max_value.Text('g', -1)
    if result, err := Cast[string, TT](max_string); err != nil || float64(result) != max_value {
        t.Errorf("Cast[string, %T](%q): expected the maximum, got %v, error %v", TT(0), max_string, result, err)
    }

    for _, value := range test_invalid_numbers {
        if result, err := Cast[string, TT](value); err == nil {
            t.Errorf("Cast[string, %T](%q): expected error, got result %v", TT(0), value, result)
        }
    }

    // Beyond the range of the type.
    out_of_range := "1e39"
    if reflect.TypeOf(TT(0)).Bits() == 64 {
        out_of_range = "1e309"
    }
    for _, value := range []string{out_of_range, "-" + out_of_range} {
        if result, err := Cast[string, TT](value); err == nil {
            t.Errorf("Cast[string, %T](%q): expected out of range error, got result %v", TT(0), value, result)
        }
    }
}

func TestCastStringToNumber(t *testing.T) {
    t.Run("to uint", testStringToInteger[uint])
    t.Run("to uint8", testStringToInteger[uint8])
    t.Run("to uint16", testStringToInteger[uint16])
    t.Run("to uint32", testStringToInteger[uint32])
    t.Run("to uint64", testStringToInteger[uint64])
    t.Run("to int", testStringToInteger[int])
    t.Run("to int8", testStringToInteger[int8])
    t.Run("to int16", testStringToInteger[int16])
    t.Run("to int32", testStringToInteger[int32])
    t.Run("to int64", testStringToInteger[int64])
    t.Run("to float32", testStringToFloat[float32])
    t.Run("to float64", testStringToFloat[float64])
}

func TestCastStringToString(t *testing.T) {
    for _, value := range []string{"", " ", "abc", "0", "true", "1.5", "\x00\xff"} {
        if result, err := Cast[string, string](value); err != nil || result != value {
            t.Errorf("Cast[string, string](%q): got %q, error %v", value, result, err)
        }
    }
}

func TestParseScalar(t *testing.T) {
    for _, test := range []struct {
        text    string
        assumed any
    }{
        {"0", uint64(0)}, {"18446744073709551615", uint64(math.MaxUint64)},
        {"-1", int64(-1)}, {"-9223372036854775808", int64(math.MinInt64)},
        {"1.5", 1.5}, {"1e3", 1000.0},
        {"true", true}, {"False", false}, {"t", true},
    } {
        value := any(test.text)
        if !ParseScalar(&value) || value != test.assumed {
            t.Errorf("ParseScalar(%q): got %v (%T), expected %v (%T)", test.text, value, value, test.assumed, test.assumed)
        }
    }

    // Text that is not a scalar stays unchanged and is reported as not parsed.
    for _, text := range []string{"", "abc", "1,5", "0x10", " 1"} {
        value := any(text)
        if ParseScalar(&value) || value != any(text) {
            t.Errorf("ParseScalar(%q): expected false and unchanged value, got %v (%T)", text, value, value)
        }
    }

    // Not a string at all.
    for _, original := range []any{1, 1.5, true, nil} {
        value := original
        if ParseScalar(&value) || value != original {
            t.Errorf("ParseScalar(%v): expected false and unchanged value, got %v", original, value)
        }
    }
}

func TestHexString(t *testing.T) {
    for _, bits := range test_integer_patterns {
        text := strconv.FormatUint(bits, 16)
        if result := UInt64ToHexString(bits); result != text {
            t.Errorf("UInt64ToHexString(%#x): expected %q, got %q", bits, text, result)
        }
        if result, err := HexStringToUInt64(text); err != nil || result != bits {
            t.Errorf("HexStringToUInt64(%q): expected %#x, got %#x, error %v", text, bits, result, err)
        }
        if bits <= uint64(MaxUInt) {
            if result := UIntToHexString(uint(bits)); result != text {
                t.Errorf("UIntToHexString(%#x): expected %q, got %q", bits, text, result)
            }
            if result, err := HexStringToUInt(text); err != nil || uint64(result) != bits {
                t.Errorf("HexStringToUInt(%q): expected %#x, got %#x, error %v", text, bits, result, err)
            }
        }
    }

    // One more bit than fits in 64 bit.
    if result, err := HexStringToUInt64("10000000000000000"); err == nil {
        t.Errorf("HexStringToUInt64: expected out of range error, got %#x", result)
    }
    for _, text := range []string{"", "g", "-1", "0x10", " 1"} {
        if result, err := HexStringToUInt64(text); err == nil {
            t.Errorf("HexStringToUInt64(%q): expected error, got %#x", text, result)
        }
        if result, err := HexStringToUInt(text); err == nil {
            t.Errorf("HexStringToUInt(%q): expected error, got %#x", text, result)
        }
    }
}

// A failed conversion must return zero, not a partial result (e.g. the maximum for an out of range value).
func testStringToNumberZeroOnError[TT number](t *testing.T) {
    t.Helper()

    for _, value := range []string{"", "abc", "1,5", "-1", "300", "99999999999999999999", "-99999999999999999999", "1e999", "-1e999"} {
        result, err := Cast[string, TT](value)
        if err != nil && result != TT(0) {
            t.Errorf("Cast[string, %T](%q): error %v, but result is %v instead of 0", TT(0), value, err, result)
        }
    }
}

func TestCastStringToNumberZeroOnError(t *testing.T) {
    t.Run("to uint", testStringToNumberZeroOnError[uint])
    t.Run("to uint8", testStringToNumberZeroOnError[uint8])
    t.Run("to uint16", testStringToNumberZeroOnError[uint16])
    t.Run("to uint32", testStringToNumberZeroOnError[uint32])
    t.Run("to uint64", testStringToNumberZeroOnError[uint64])
    t.Run("to int", testStringToNumberZeroOnError[int])
    t.Run("to int8", testStringToNumberZeroOnError[int8])
    t.Run("to int16", testStringToNumberZeroOnError[int16])
    t.Run("to int32", testStringToNumberZeroOnError[int32])
    t.Run("to int64", testStringToNumberZeroOnError[int64])
    t.Run("to float32", testStringToNumberZeroOnError[float32])
    t.Run("to float64", testStringToNumberZeroOnError[float64])
}
