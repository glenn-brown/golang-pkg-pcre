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

func pcregroups(ptr *C.pcre) (count C.int) {
	C.pcre_fullinfo(ptr, nil,
		C.PCRE_INFO_CAPTURECOUNT, unsafe.Pointer(&count))
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

func (p PCRE) Groups() int {
	return int(pcregroups((*C.pcre)(unsafe.Pointer(&p.ptr[0]))))
}

type Matcher struct {
	pcre PCRE
	groups int
	ovector []C.int
	matches bool
	subject []byte
}

func (p PCRE) Matcher(subject []byte) (m *Matcher) {
	m = new(Matcher)
	m.Reset(p, subject)
	return
}

func (m *Matcher) Reset(p PCRE, subject []byte) {
	if p.ptr == nil {
		panic("PCRE.Matcher: uninitialized")
	}
	m.init(p)
	m.Match(subject)
}

func (m *Matcher) init(p PCRE) {
	m.matches = false
	if m.pcre.ptr != nil && &m.pcre.ptr[0] == &p.ptr[0] {
		// Skip group count extraction if the matcher has
		// already been initialized with the same regular
		// expression.
		return
	}
	m.pcre = p
	m.groups = p.Groups()
	m.ovector = make([]C.int, 3 * (1 + m.groups))
}

var nullbyte = []byte{0}

func (m *Matcher) Match(subject []byte) bool {
	if m.pcre.ptr == nil {
		panic("Matcher.Match: uninitialized")
	}
	length := len(subject)
	m.subject = subject
	if length == 0 {
		subject = nullbyte // make first character adressable
	}
	subjectptr := (*C.char)(unsafe.Pointer(&subject[0]))
	ovectorptr := &m.ovector[0]
	rc := C.pcre_exec((*C.pcre)(unsafe.Pointer(&m.pcre.ptr[0])), nil,
		subjectptr, C.int(length),
		0, 0, ovectorptr, C.int(len(m.ovector)))
	switch{
	case rc >= 0:
		m.matches = true
		return true
	case rc == C.PCRE_ERROR_NOMATCH:
		m.matches = false
		return false
	}
	panic("unexepcted return code from pcre_exec: " +
		strconv.Itoa(int(rc)))
}

func (m *Matcher) Matches() bool {
	return m.matches
}

func (m *Matcher) Groups() int {
	return m.groups
}

// Returns true if the numbered capture group is present.  Group
// numbers start at 1.  A capture group can be present and match the
// empty string.
func (m *Matcher) Present(group int) bool {
	return m.ovector[2 * group] >= 0
}

// Returns the numbered capture group.  Group 0 is the part of the
// subject which matches the whole pattern; the first actual capture
// group is numbered 1.  Capture groups which are not present return a
// nil slice.
func (m *Matcher) Group(group int) []byte {
	start := m.ovector[2 * group]
	end := m.ovector[2 * group + 1]
	if start >= 0 {
		return m.subject[start:end]
	}
	return nil
}

// Returns the numbered capture group as a string.  Group 0 is the
// part of the subject which matches the whole pattern; the first
// actual capture group is numbered 1.  Capture groups which are not
// present return a nil slice.
func (m *Matcher) GroupString(group int) string {
	return string(m.Group(group))
}

type CompileError struct {
	Pattern string
	Message string
	Offset int
}

func (e *CompileError) String() string {
	return e.Pattern + " (" + strconv.Itoa(e.Offset) + "): " + e.Message
}
