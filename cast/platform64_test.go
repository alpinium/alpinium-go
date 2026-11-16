package cast

import (
    "math/bits"
    "testing"
    "unsafe"
)

// TestRequire64Bit calls require64Bit (the real check is done by the compiler) and checks
// the sizes at run time as well.
func TestRequire64Bit(t *testing.T) {
    require64Bit()
    if bits.UintSize != 64 || unsafe.Sizeof(int(0)) != 8 || unsafe.Sizeof(uintptr(0)) != 8 {
        t.Fatalf("Cast requires a 64-bit platform, uint is %d-bit", bits.UintSize)
    }
}
