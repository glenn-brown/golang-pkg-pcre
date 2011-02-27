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

func MustCompile(pattern string) (p PCRE) {
	p, err := Compile(pattern)
	if err != nil {
		panic(err)
	}
	return
}

var empty = []byte{}
var nullbyte = []byte{0}

func (p PCRE) Match(subject []byte, groups [][]byte) bool {
	length := len(subject)
	if length == 0 {
		subject = nullbyte // make first character adressable
	}
	var ovector []C.int
	var ovectorptr *C.int
	if len(groups) > 0 {
		ovector = make([]C.int, 3 * len(groups))
		ovectorptr = &ovector[0]
	}
	rc := C.pcre_exec((*C.pcre)(unsafe.Pointer(&p.ptr[0])), nil,
		(*C.char)(unsafe.Pointer(&subject[0])),
		C.int(length), 0, 0,
		ovectorptr, C.int(len(ovector)))
	switch{
	case rc >= 0:
		creategroups(groups, subject, ovector, int(rc))
		return true
	case rc == C.PCRE_ERROR_NOMATCH:
		return false
	}
	panic("unexepcted return code from pcre_exec: " +
		strconv.Itoa(int(rc)))
}

func creategroups(groups [][]byte, subject []byte, ovector []C.int, count int) {
	for i := 0; i < count; i++ {
		start := ovector[2 * i]
		end := ovector[2 * i + 1]
		if start == -1 {
			groups[i] = nil
		} else {
			groups[i] = subject[start:end]
		}
	}
	for i := count; i < len(groups); i++{
		groups[i] = nil
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
