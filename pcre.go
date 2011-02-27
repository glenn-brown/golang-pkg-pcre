package pcre

/*
#cgo LDFLAGS: -lpcre
#include <pcre.h>
#include <string.h>
*/
import "C"

import (
	"strconv"
	"unsafe"
)

const ANCHORED = C.PCRE_ANCHORED

type PCRE struct {
	ptr []byte
}

func pcresize(ptr *C.pcre) (size C.size_t) {
	C.pcre_fullinfo(ptr, nil, C.PCRE_INFO_SIZE, unsafe.Pointer(&size))
	return
}

func toheap(ptr *C.pcre) (p PCRE) {
	defer C.free(unsafe.Pointer(ptr))
	size := pcresize(ptr)
	p.ptr = make([]byte, size)
	C.memcpy(unsafe.Pointer(&p.ptr[0]), unsafe.Pointer(ptr), size)
	return
}

func Compile(pattern string) (PCRE, *CompileError) {
	pattern1 := C.CString(pattern)
	defer C.free(unsafe.Pointer(pattern1))
	var errptr *C.char
	var erroffset C.int
	ptr := C.pcre_compile(pattern1, 0, &errptr, &erroffset, nil)
	if ptr == nil {
		return PCRE{}, &CompileError{
		        Pattern: pattern,
		        Message: C.GoString(errptr),
		        Offset: int(erroffset),
		}
	}
	return toheap(ptr), nil
}

var empty = []byte{}
var nullbyte = []byte{0}

func (p PCRE) Match(subject []byte, groups [][]byte) bool {
	length := len(subject)
	if length == 0 {
		subject = nullbyte // make first character adressable
	}
	ovector := make([]int, 3 * len(groups))
	rc := C.pcre_exec((*C.pcre)(unsafe.Pointer(&p.ptr[0])), nil,
		(*C.char)(unsafe.Pointer(&subject[0])),
		C.int(len(subject)), 0, 0,
		(*C.int)(unsafe.Pointer(&ovector[0])), C.int(len(ovector)))
	switch rc {
	case 0:
		creategroups(groups, subject, ovector)
		return true
	case C.PCRE_ERROR_NOMATCH:
		return false
	}
	panic("unexepcted return code from pcre_exec: " +
		strconv.Itoa(int(rc)))
}

func creategroups(groups [][]byte, subject []byte, ovector []int) {
	for i := range groups {
		start := ovector[2 * i]
		end := ovector[2 * i + 1]
		switch {
		case start < end:
			groups[i] = subject[start:end]
		case start == end:
			groups[i] = empty
		default:
			groups[i] = nil
		}
	}
}

type CompileError struct {
	Pattern string
	Message string
	Offset int
}

func (e *CompileError) String() string {
	return e.Pattern + " (" + strconv.Itoa(e.Offset) + "): " + e.Message
}
