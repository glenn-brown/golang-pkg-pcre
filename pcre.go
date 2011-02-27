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

// Flags for Compile and Match functions
const ANCHORED = C.PCRE_ANCHORED
const BSR_ANYCRLF = C.PCRE_BSR_ANYCRLF
const BSR_UNICODE = C.PCRE_BSR_UNICODE
const NEWLINE_ANY = C.PCRE_NEWLINE_ANY
const NEWLINE_ANYCRLF = C.PCRE_NEWLINE_ANYCRLF
const NEWLINE_CR = C.PCRE_NEWLINE_CR
const NEWLINE_CRLF = C.PCRE_NEWLINE_CRLF
const NEWLINE_LF = C.PCRE_NEWLINE_LF
const NO_UTF8_CHECK = C.PCRE_NO_UTF8_CHECK

// Flags for Compile functions
const CASELESS = C.PCRE_CASELESS
const DOLLAR_ENDONLY = C.PCRE_DOLLAR_ENDONLY
const DOTALL = C.PCRE_DOTALL
const DUPNAMES = C.PCRE_DUPNAMES
const EXTENDED = C.PCRE_EXTENDED
const EXTRA = C.PCRE_EXTRA
const FIRSTLINE = C.PCRE_FIRSTLINE
const JAVASCRIPT_COMPAT = C.PCRE_JAVASCRIPT_COMPAT
const MULTILINE = C.PCRE_MULTILINE
const NO_AUTO_CAPTURE = C.PCRE_NO_AUTO_CAPTURE
const UNGREEDY = C.PCRE_UNGREEDY
const UTF8 = C.PCRE_UTF8

// Flags for Match functions
const NOTBOL = C.PCRE_NOTBOL 
const NOTEOL = C.PCRE_NOTEOL
const NOTEMPTY = C.PCRE_NOTEMPTY
const NOTEMPTY_ATSTART = C.PCRE_NOTEMPTY_ATSTART
const NO_START_OPTIMIZE = C.PCRE_NO_START_OPTIMIZE
const PARTIAL_HARD = C.PCRE_PARTIAL_HARD
const PARTIAL_SOFT = C.PCRE_PARTIAL_SOFT

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

func Compile(pattern string, flags int) (PCRE, *CompileError) {
	pattern1 := C.CString(pattern)
	defer C.free(unsafe.Pointer(pattern1))
	var errptr *C.char
	var erroffset C.int
	ptr := C.pcre_compile(pattern1, C.int(flags), &errptr, &erroffset, nil)
	if ptr == nil {
		return PCRE{}, &CompileError{
		        Pattern: pattern,
		        Message: C.GoString(errptr),
		        Offset: int(erroffset),
		}
	}
	return toheap(ptr), nil
}

func MustCompile(pattern string, flags int) (p PCRE) {
	p, err := Compile(pattern, flags)
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
	subjects string
	subjectb []byte
}

func (p PCRE) Matcher(subject []byte, flags int) (m *Matcher) {
	m = new(Matcher)
	m.Reset(p, subject, flags)
	return
}

func (p PCRE) MatcherString(subject string, flags int) (m *Matcher) {
	m = new(Matcher)
	m.ResetString(p, subject, flags)
	return
}

func (m *Matcher) Reset(p PCRE, subject []byte, flags int) {
	if p.ptr == nil {
		panic("PCRE.Matcher: uninitialized")
	}
	m.init(p)
	m.Match(subject, flags)
}

func (m *Matcher) ResetString(p PCRE, subject string, flags int) {
	if p.ptr == nil {
		panic("PCRE.Matcher: uninitialized")
	}
	m.init(p)
	m.MatchString(subject, flags)
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

func (m *Matcher) Match(subject []byte, flags int) bool {
	if m.pcre.ptr == nil {
		panic("Matcher.Match: uninitialized")
	}
	length := len(subject)
	m.subjects = ""
	m.subjectb = subject
	if length == 0 {
		subject = nullbyte // make first character adressable
	}
	subjectptr := (*C.char)(unsafe.Pointer(&subject[0]))
	ovectorptr := &m.ovector[0]
	rc := C.pcre_exec((*C.pcre)(unsafe.Pointer(&m.pcre.ptr[0])), nil,
		subjectptr, C.int(length),
		0, C.int(flags), ovectorptr, C.int(len(m.ovector)))
	return m.match(rc)
}

func (m *Matcher) MatchString(subject string, flags int) bool {
	if m.pcre.ptr == nil {
		panic("Matcher.Match: uninitialized")
	}
	length := len(subject)
	m.subjects = subject
	m.subjectb = nil
	subjectptr := C.CString(subject)
	if subjectptr == nil {
		panic("pcre.MatchString: malloc")
	}
	defer C.free(unsafe.Pointer(subjectptr))
	ovectorptr := &m.ovector[0]
	rc := C.pcre_exec((*C.pcre)(unsafe.Pointer(&m.pcre.ptr[0])), nil,
		subjectptr, C.int(length),
		0, C.int(flags), ovectorptr, C.int(len(m.ovector)))
	return m.match(rc)
}

func (m *Matcher) match(rc C.int) bool {
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
		if m.subjectb != nil {
			return m.subjectb[start:end]
		}
		return []byte(m.subjects[start:end])
	}
	return nil
}

// Returns the numbered capture group as a string.  Group 0 is the
// part of the subject which matches the whole pattern; the first
// actual capture group is numbered 1.  Capture groups which are not
// present return a nil slice.
func (m *Matcher) GroupString(group int) string {
	start := m.ovector[2 * group]
	end := m.ovector[2 * group + 1]
	if start >= 0 {
		if m.subjectb != nil {
			return string(m.subjectb[start:end])
		}
		return m.subjects[start:end]
	}
	return ""
}

type CompileError struct {
	Pattern string
	Message string
	Offset int
}

func (e *CompileError) String() string {
	return e.Pattern + " (" + strconv.Itoa(e.Offset) + "): " + e.Message
}
