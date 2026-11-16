//go:build ignore

// cast_gen.go generates the Cast function (cast_switch.go) from one list of basic types.
// Run with: go generate ./...
package main

import (
    "bytes"
    "fmt"
    "go/format"
    "os"
    "strings"
    "text/template"
)

// Basic types supported by Cast, in the order of the switch branches.
var types = []string{
    "bool",
    "uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
    "int", "int8", "int16", "int32", "int64",
    "float32", "float64",
    "string",
}

func category(t string) string {
    switch t {
    case "bool", "string":
        return t
    }
    return "number"
}

// convert returns the right-hand side of the assignment converting typed_value (src) to dst.
// An empty second result means a plain assignment without error.
func convert(src, dst string) (expr string, withErr bool) {
    switch category(src) + "->" + category(dst) {
    case "bool->bool":
        return "typed_value", false
    case "bool->number":
        return fmt.Sprintf("BoolToNumber[%s](typed_value)", dst), true
    case "bool->string":
        return "BoolToString(typed_value)", true
    case "number->bool":
        return "NumberToBool(typed_value)", true
    case "number->string":
        return numberToString(src), false
    case "string->string":
        return "string(typed_value)", false
    }
    panic("unknown conversion " + src + " -> " + dst)
}

// Limits of the integer types: typed constants from cast.go (MinInt8, MaxUInt64, ...).
// Only 64 bit platforms are supported: int, uint and uintptr have 64 bits.
var integerLimits = map[string][2]string{
    "int":     {"MinInt", "MaxInt"},
    "int8":    {"MinInt8", "MaxInt8"},
    "int16":   {"MinInt16", "MaxInt16"},
    "int32":   {"MinInt32", "MaxInt32"},
    "int64":   {"MinInt64", "MaxInt64"},
    "uint":    {"MinUInt", "MaxUInt"},
    "uint8":   {"MinUInt8", "MaxUInt8"},
    "uint16":  {"MinUInt16", "MaxUInt16"},
    "uint32":  {"MinUInt32", "MaxUInt32"},
    "uint64":  {"MinUInt64", "MaxUInt64"},
    "uintptr": {"MinUIntPtr", "MaxUIntPtr"},
}

func isInteger(t string) bool  { _, ok := integerLimits[t]; return ok }
func isUnsigned(t string) bool { return strings.HasPrefix(t, "uint") }

// integerBits returns the size of the integer type in bits (int, uint and uintptr have 64 bits).
func integerBits(t string) int {
    switch strings.TrimPrefix(strings.TrimPrefix(t, "u"), "int") {
    case "8":
        return 8
    case "16":
        return 16
    case "32":
        return 32
    }
    return 64 // int64, uint64, int, uint, uintptr
}

// integerCondition returns only the checks that are really needed for integer src -> dst:
// the lower limit only if src can be smaller than the minimum of dst, the upper limit only
// if src can be greater than the maximum of dst. An empty result means "always fits".
// The limits are typed constants of dst; they are converted to src, which is always possible
// when the check is needed (the limit lies inside the range of src).
func integerCondition(src, dst string) string {
    src_bits, dst_bits := integerBits(src), integerBits(dst)
    var lower, upper bool
    switch {
    case isUnsigned(src) && isUnsigned(dst):
        upper = src_bits > dst_bits
    case isUnsigned(src): // unsigned -> signed: the signed maximum is 2^(bits-1)-1
        upper = src_bits >= dst_bits
    case isUnsigned(dst): // signed -> unsigned: negative values never fit
        lower = true
        upper = src_bits > dst_bits
    default: // signed -> signed
        lower = src_bits > dst_bits
        upper = src_bits > dst_bits
    }
    var checks []string
    if lower {
        checks = append(checks, fmt.Sprintf("typed_value >= %s(%s)", src, integerLimits[dst][0]))
    }
    if upper {
        checks = append(checks, fmt.Sprintf("typed_value <= %s(%s)", src, integerLimits[dst][1]))
    }
    return strings.Join(checks, " && ")
}

// floatToIntegerLimits returns the condition "typed_value (float src) fits into dst".
// 8 and 16 bit limits are exact floats. For 32 and 64 bit the largest float not above the
// integer maximum is used (MaxFloat64ForInt64 etc. in cast.go), because e.g. float64(MaxInt64)
// rounds up to 2^63. int, uint and uintptr have 64 bits. NaN fails every comparison, so it is
// out of range too.
func floatToIntegerCondition(src, dst string) string {
    bits := map[string]string{"float32": "Float32", "float64": "Float64"}[src]
    sized := func(size string) string {
        name := "Int" + size // MinFloat64ForInt64 ...
        if isUnsigned(dst) {
            name = "UInt" + size // MaxFloat32ForUInt32 ...
        }
        return fmt.Sprintf("typed_value >= Min%sFor%s && typed_value <= Max%sFor%s", bits, name, bits, name)
    }
    switch dst {
    case "int8", "int16", "uint8", "uint16":
        limits := integerLimits[dst]
        return fmt.Sprintf("typed_value >= %s(%s) && typed_value <= %s(%s)", src, limits[0], src, limits[1])
    case "int32", "uint32":
        return sized("32")
    }
    // int64, uint64 and the 64 bit int, uint, uintptr
    return sized("64")
}

// numberConversion writes number -> number conversions inline: a range check, the conversion
// and, only for values out of range, a call of castError, which builds the error.
// Only the checks needed for the given types are written; without any check there is no
// error branch at all.
func numberConversion(target, src, dst string) string {
    convert := fmt.Sprintf("%s = %s(typed_value)", target, dst)
    var condition, min_value, max_value string
    switch {
    case src == dst:
        return target + " = typed_value"
    case isInteger(src) && !isInteger(dst):
        // integer -> float: always valid (may round)
        return convert
    case !isInteger(src) && !isInteger(dst):
        if dst == "float64" {
            // float32 -> float64: always exact
            return convert
        }
        // float64 -> float32: NaN and infinities are kept, finite values must fit.
        // (Not inlinable: cost about 120. Variants without math.IsInf were tried, they still
        // exceed the budget 80 and are harder to read.)
        condition = "(typed_value >= float64(MinFloat32) && typed_value <= float64(MaxFloat32)) || typed_value != typed_value || math.IsInf(typed_value, 0)"
        min_value, max_value = "MinFloat32", "MaxFloat32"
    case !isInteger(src):
        condition = floatToIntegerCondition(src, dst)
        min_value, max_value = integerLimits[dst][0], integerLimits[dst][1]
    default:
        condition = integerCondition(src, dst)
        if condition == "" {
            // e.g. int8 -> int64, uint8 -> int16: every value fits
            return convert
        }
        min_value, max_value = integerLimits[dst][0], integerLimits[dst][1]
    }
    return fmt.Sprintf("if %s {\n%s\n} else {\nerr = castError(typed_value, %s, %s)\n}",
        condition, convert, min_value, max_value)
}

// boolConversion writes bool <-> number and bool -> string conversions inline (no call).
// The target is always zero before, so only the "true" case needs an assignment.
func boolConversion(target, src, dst string) string {
    switch category(src) + "->" + category(dst) {
    case "bool->number":
        return fmt.Sprintf("if typed_value {\n%s = 1\n}", target)
    case "number->bool":
        return fmt.Sprintf("%s = typed_value != 0", target)
    case "bool->string":
        return fmt.Sprintf("%s = \"false\"\nif typed_value {\n%s = \"true\"\n}", target, target)
    }
    return ""
}

// numberToString returns the strconv call formatting typed_value (src) as a decimal string,
// the same format as NumberToString: integers in base 10, floats with 'g' and the shortest
// representation that reads back to the same value.
func numberToString(src string) string {
    switch {
    case src == "float32":
        return "strconv.FormatFloat(float64(typed_value), 'g', -1, 32)"
    case src == "float64":
        return "strconv.FormatFloat(typed_value, 'g', -1, 64)"
    case src == "uint64":
        return "strconv.FormatUint(typed_value, 10)"
    case isUnsigned(src):
        return "strconv.FormatUint(uint64(typed_value), 10)"
    case src == "int64":
        return "strconv.FormatInt(typed_value, 10)"
    }
    return "strconv.FormatInt(int64(typed_value), 10)"
}

// stringConversion writes string -> number and string -> bool conversions inline: the strconv
// parser for the destination type and size, and wrapParseError for the errors (ErrInvalidSyntax,
// ErrOutOfRange). On error the target keeps its zero value, the same as StringToNumber.
func stringConversion(target, src, dst string) string {
    if src != "string" {
        return ""
    }
    var parse, value string
    switch {
    case dst == "bool":
        parse = "strconv.ParseBool(typed_value)"
    case dst == "float32" || dst == "float64":
        parse = fmt.Sprintf("strconv.ParseFloat(typed_value, %s)", strings.TrimPrefix(dst, "float"))
    case isUnsigned(dst):
        parse = fmt.Sprintf("strconv.ParseUint(typed_value, 10, %d)", integerBits(dst))
    case isInteger(dst):
        parse = fmt.Sprintf("strconv.ParseInt(typed_value, 10, %d)", integerBits(dst))
    default:
        return "" // string -> string
    }
    // the parser returns bool, float64, uint64 or int64; convert only when the type differs
    switch dst {
    case "bool", "float64", "uint64", "int64":
        value = "parsed"
    default:
        value = dst + "(parsed)"
    }
    return fmt.Sprintf("if parsed, parse_err := %s; parse_err == nil {\n%s = %s\n} else {\nerr = wrapParseError(parse_err)\n}",
        parse, target, value)
}

// assign builds the whole assignment statement for the given target expression.
func assign(target, src, dst string) string {
    if inline := stringConversion(target, src, dst); inline != "" {
        return inline
    }
    if category(src) == "number" && category(dst) == "number" {
        return numberConversion(target, src, dst)
    }
    if inline := boolConversion(target, src, dst); inline != "" {
        return inline
    }
    expr, withErr := convert(src, dst)
    if withErr {
        return target + ", err = " + expr
    }
    return target + " = " + expr
}

// typeTitle returns the type name as used in the exported function names, in the style of
// the existing names (MaxUInt64, HexStringToUInt): Int8, UInt64, UIntPtr, Float32, String.
func typeTitle(t string) string {
    switch {
    case t == "uintptr":
        return "UIntPtr"
    case strings.HasPrefix(t, "uint"):
        return "UInt" + strings.TrimPrefix(t, "uint")
    }
    return strings.ToUpper(t[:1]) + t[1:]
}

// existing are the conversion functions written by hand in cast.go with the same name and
// signature; they are used as they are and not generated.
var existing = map[string]bool{"BoolToString": true, "StringToBool": true}

// helper returns the name of the exported non-generic conversion function, e.g. Int8ToUInt64.
func helper(src, dst string) string {
    return typeTitle(src) + "To" + typeTitle(dst)
}

// generated reports whether the conversion function src -> dst is generated.
func generated(src, dst string) bool { return !existing[helper(src, dst)] }

// describe returns the doc comment sentence for the conversion function src -> dst.
func describe(src, dst string) string {
    switch {
    case src == dst:
        return "returns the value unchanged."
    case category(src) == "string":
        return "parses the string; invalid syntax returns ErrInvalidSyntax, a value out of range ErrOutOfRange."
    case category(dst) == "string":
        return "formats the value as a string."
    case category(src) == "bool":
        return "converts true to 1 and false to 0."
    case category(dst) == "bool":
        return "returns true for a non-zero value."
    case strings.Contains(numberConversion("result", src, dst), "castError"):
        return "converts the value; a value out of range returns ErrOutOfRange and 0."
    }
    return "converts the value, every value fits."
}

// safe reports whether src -> dst is a numeric conversion where every value fits (e.g. uint8 ->
// uint16, int32 -> float64): a plain Go conversion without a check and without an error, so Cast
// and castKind write it inline instead of calling the conversion function.
func safe(src, dst string) bool {
    return src != dst && category(src) == "number" && category(dst) == "number" &&
        !strings.Contains(numberConversion("result", src, dst), "castError")
}

func kind(t string) string { return "reflect." + strings.ToUpper(t[:1]) + t[1:] }

const tmpl = `// Code generated by cast_gen.go; DO NOT EDIT.

package cast

import (
    "encoding"
    "fmt"
    "math"
    "reflect"
    "strconv"
    "time"
    "unsafe"
)

// Cast converts value of type T to type TT. Both types are bool, a number type, string or a type
// derived from them (Castable).
//
// Numbers: a value out of range of TT returns ErrOutOfRange and 0 (NaN and infinities are kept
// only in float32/float64). A number converted to bool is true when non-zero, a bool converted
// to a number is 1 or 0. Strings are parsed and formatted in decimal; invalid syntax returns
// ErrInvalidSyntax and the zero value. A derived type with MarshalText or String is formatted
// by it, a derived type with UnmarshalText (and time.Duration) is parsed by it.
func Cast[T Castable, TT Castable](value T) (result TT, err error) {
    // Fast path: basic types are handled by the type switch.
    // Both switches use values (not pointers), so for every GC shape the compiler keeps
    // only the single matching branch of each switch. The branch calls the exported
    // conversion function (e.g. StringToFloat32); the same type and a conversion where every
    // value fits (e.g. uint8 -> uint16) are written without a call. The result is written through
    // unsafe.Pointer, TT is exactly the basic type of the matching branch.
    // Every converted branch returns; a type without a branch (a derived type) falls through
    // both switches to castDerived.
    switch typed_value := any(value).(type) {
{{- range $src := .Types}}
    case {{$src}}:
        switch any(result).(type) {
{{- range $dst := $.Types}}
        case {{$dst}}:
{{- if eq $src $dst}}
            return *(*TT)(unsafe.Pointer(&typed_value)), nil
{{- else if safe $src $dst}}
            temp := {{$dst}}(typed_value)
            return *(*TT)(unsafe.Pointer(&temp)), nil
{{- else}}
            temp, temp_err := {{helper $src $dst}}(typed_value)
            return *(*TT)(unsafe.Pointer(&temp)), temp_err
{{- end}}
{{- end}}
        }
{{- end}}
    }
    return castDerived[T, TT](value)
}

// castDerived is the slow path of Cast: at least one of the types is derived from a basic type
// (e.g. type MyInt int). It is a separate function, so its code and stack frame do not burden
// the fast path of Cast.
//
//go:noinline
func castDerived[T Castable, TT Castable](value T) (result TT, err error) {
    // Slow path: at least one of the types is derived from a basic type (e.g. type MyInt int).
    // The checks below use temporary variables, so value and result do not escape to the heap
    // and the fast path above stays without allocations.

    // Same type: no conversion needed.
    if same, ok := any(value).(TT); ok {
        return same, nil
    }

    src_type, dst_type := reflect.TypeFor[T](), reflect.TypeFor[TT]()
    src_kind, dst_kind := src_type.Kind(), dst_type.Kind()

    // Derived type with its own text representation -> string.
    // The interface is first checked on a nil pointer (no allocation), so types without
    // these methods are not slowed down. Pointer method sets include value methods.
    if dst_kind == reflect.String {
        switch any((*T)(nil)).(type) {
        case encoding.TextMarshaler:
            value_copy := value
            text, text_err := any(&value_copy).(encoding.TextMarshaler).MarshalText()
            if text_err != nil {
                return result, text_err
            }
            temp := string(text)
            return *(*TT)(unsafe.Pointer(&temp)), nil
        case fmt.Stringer:
            value_copy := value
            temp := any(&value_copy).(fmt.Stringer).String()
            return *(*TT)(unsafe.Pointer(&temp)), nil
        }
    }

    // String -> derived type with its own text representation.
    if src_kind == reflect.String {
        text := *(*string)(unsafe.Pointer(&value))
        if dst_type == reflect.TypeFor[time.Duration]() {
            duration, parse_err := time.ParseDuration(text)
            if parse_err != nil {
                return result, &parseError{kind: ErrInvalidSyntax, err: parse_err}
            }
            return *(*TT)(unsafe.Pointer(&duration)), nil
        }
        if _, ok := any((*TT)(nil)).(encoding.TextUnmarshaler); ok {
            temp := new(TT)
            if parse_err := any(temp).(encoding.TextUnmarshaler).UnmarshalText([]byte(text)); parse_err != nil {
                return result, &parseError{kind: ErrInvalidSyntax, err: parse_err}
            }
            return *temp, nil
        }
    }

    // Kind switch: not generic (castKind), so its code exists only once in the binary.
    // The kinds are passed, so castKind does not ask reflect again; the types are only
    // for the error messages.
    err = castKind(unsafe.Pointer(&value), src_kind, src_type, unsafe.Pointer(&result), dst_kind, dst_type)
    return
}

// castKind converts *src to *dst by the kinds of the types. A derived type has the same memory
// layout as its underlying basic type, so the value is read and the result written through
// unsafe.Pointer as the basic type of that kind. The conversion is done by the exported
// functions (Int8ToUInt64, StringToFloat32 ...); the same kind and a conversion where every
// value fits (e.g. uint8 -> uint16) are written without a call.
// castKind only reads *src and writes *dst, the
// pointers do not escape, so value and result of the caller stay on the stack.
func castKind(src unsafe.Pointer, src_kind reflect.Kind, src_type reflect.Type, dst unsafe.Pointer, dst_kind reflect.Kind, dst_type reflect.Type) (err error) {
    switch src_kind {
{{- range $src := .Types}}
    case {{kind $src}}:
        typed_value := *(*{{$src}})(src)
        switch dst_kind {
{{- range $dst := $.Types}}
        case {{kind $dst}}:
{{- if eq $src $dst}}
            *(*{{$dst}})(dst) = typed_value
{{- else if safe $src $dst}}
            *(*{{$dst}})(dst) = {{$dst}}(typed_value)
{{- else}}
            *(*{{$dst}})(dst), err = {{helper $src $dst}}(typed_value)
{{- end}}
{{- end}}
        default:
            err = fmt.Errorf("%w: destination type %v", ErrUnsupportedType, dst_type)
        }
{{- end}}
    default:
        err = fmt.Errorf("%w: source type %v", ErrUnsupportedType, src_type)
    }
    return
}

// Conversions between the basic types. They are not generic, so they can be called directly
// without the overhead of Cast (type dictionary, type switches) when the types are known.
// castDerived uses them as well.
{{- range $src := .Types}}
{{- range $dst := $.Types}}
{{- if generated $src $dst}}

// {{helper $src $dst}} {{describe $src $dst}}
func {{helper $src $dst}}(typed_value {{$src}}) (result {{$dst}}, err error) {
    {{assign "result" $src $dst}}
    return
}
{{- end}}
{{- end}}
{{- end}}
`

// derivedName returns the name of the test type derived from t, e.g. genInt8.
func derivedName(t string) string { return "gen" + typeTitle(t) }

// testTmpl generates cast_switch_test.go: one derived type per basic type and a check of every
// pair against the reference implementation (cast_reference_test.go).
const testTmpl = `// Code generated by cast_gen.go; DO NOT EDIT.

package cast

import "testing"

// Types derived from the basic types (same memory layout, no methods).
{{- range $src := .Types}}
type {{derived $src}} {{$src}}
{{- end}}

// TestGeneratedPairs checks all conversion paths of every pair of basic types against the
// reference implementation: Cast, Cast with derived types (castDerived, castKind) and the
// exported non-generic function.
func TestGeneratedPairs(t *testing.T) {
{{- range $src := .Types}}
{{- range $dst := $.Types}}
    t.Run("{{$src}} to {{$dst}}", func(t *testing.T) { checkPair[{{$src}}, {{$dst}}, {{derived $src}}, {{derived $dst}}](t, {{helper $src $dst}}) })
{{- end}}
{{- end}}
}
`

// indentWithSpaces replaces the leading tabs of every line (go/format indents with tabs) with
// 4 spaces each, the indentation used in the whole package.
func indentWithSpaces(src []byte) []byte {
    lines := strings.Split(string(src), "\n")
    for i, line := range lines {
        trimmed := strings.TrimLeft(line, "\t")
        lines[i] = strings.Repeat("    ", len(line)-len(trimmed)) + trimmed
    }
    return []byte(strings.Join(lines, "\n"))
}

// generate executes the template and writes the formatted result to the file.
func generate(name, text string, funcs template.FuncMap) {
    t := template.Must(template.New(name).Funcs(funcs).Parse(text))
    var buf bytes.Buffer
    if err := t.Execute(&buf, struct{ Types []string }{types}); err != nil {
        panic(err)
    }
    src, err := format.Source(buf.Bytes())
    if err != nil {
        panic(err)
    }
    if err := os.WriteFile(name, indentWithSpaces(src), 0o644); err != nil {
        panic(err)
    }
}

func main() {
    funcs := template.FuncMap{"assign": assign, "helper": helper, "generated": generated, "describe": describe, "kind": kind, "safe": safe, "derived": derivedName}
    generate("cast_switch.go", tmpl, funcs)
    generate("cast_switch_test.go", testTmpl, funcs)
}
