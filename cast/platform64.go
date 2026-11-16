package cast

import "unsafe"

// require64Bit stops the compilation on a platform where int, uint or uintptr is not 64-bit.
// The generated conversions (e.g. uint64 -> uint, int64 -> int, uint64 -> uintptr) assume
// 64-bit types and have no range check, so on a 32-bit platform they would silently truncate.
//
// The check is done by the compiler, the function does not need to be called: each line indexes
// a one-element array by the constant Sizeof - 8. On a 64-bit platform the index is 0, on a 32-bit
// platform it is negative and the compilation fails with
// "unsafe.Sizeof(int(0)) - 8 (constant -4 of type uintptr) overflows uintptr".
// The file name starts with a capital letter, so the compiler reports this error first.
func require64Bit() {
    _ = [1]struct{}{}[unsafe.Sizeof(int(0))-8]     // int must be 64-bit
    _ = [1]struct{}{}[unsafe.Sizeof(uint(0))-8]    // uint must be 64-bit
    _ = [1]struct{}{}[unsafe.Sizeof(uintptr(0))-8] // uintptr must be 64-bit
}
