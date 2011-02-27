package pcre

/*
#cgo LDFLAGS: -lpcre
#include <pcre.h>
#include <string.h>

static int
go_pcre_exec(const pcre *code, const pcre_extra *extra,
             _GoString_ subject, int startoffset,
             int options, int *ovector, int ovecsize)
{
  return pcre_exec(code, extra,
                   subject.p, subject.n, startoffset,
                   options, ovector, ovecsize);
}
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

func (p PCRE) Match(subject string, groups []string) bool {
	length := len(subject)
	if length == 0 {
		subject = "\000" // make first character adressable
	}
	ovector := make([]int, 3 * len(groups))
	rc := C.go_pcre_exec((*C.pcre)(unsafe.Pointer(&p.ptr[0])), nil,
		subject, 0, 0,
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

func creategroups(groups []string, subject string, ovector []int) {
	for i := range groups {
		start := ovector[2 * i]
		end := ovector[2 * i + 1]
		if start < end {
			groups[i] = subject[start:end]
		} else {
			groups[i] = ""
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
