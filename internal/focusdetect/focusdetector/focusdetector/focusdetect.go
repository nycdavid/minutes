//go:build darwin

package focusdetector

import (
	"fmt"
	"log"
	"unsafe"
)

/*
#cgo CFLAGS: -I${SRCDIR}/focusdetector/focusdetector
#cgo darwin CFLAGS: -x objective-c
#cgo darwin LDFLAGS: -framework AppKit
#include "focusdetector.h"
*/
import "C"

func Foo() {
	buf := make([]byte, 256)
	p := (*C.char)(unsafe.Pointer(&buf[0]))
	rc := C.FrontmostAppName(p, C.int(len(buf)))
	if rc != 0 {
		log.Fatalf("error calling Frontmostapp: %d", rc)
	}

	fmt.Println(cString(buf))
}

func cString(b []byte) string {
	for i, v := range b {
		if v == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
