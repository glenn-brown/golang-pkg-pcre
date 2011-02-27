package pcre

import (
	"reflect"
	"testing"
)

func TestCompile(t *testing.T) {
	var check = func (p string) {
		_, err := Compile(p)
		if err != nil {
			t.Error(p, err)
		}
	}
	check("")
	check("^")
	check("^$")
	check("()")
}

func strings(b [][]byte) (r []string) {
	r = make([]string, len(b))
	for i, v := range b {
		r[i] = string(v)
	} 
	return
}

func equal(l, r []string) bool {
	if len(l) != len(r) {
		return false
	}
	for i, lv := range l {
		if lv != r[i] {
			return false
		}
	}
	return true
}

func TestMatch(t *testing.T) {
	if !MustCompile("^$").Match([]byte(""), [][]byte{}) {
		t.Error("empty")
	}
	if !MustCompile("^$").Match([]byte(""), nil) {
		t.Error("empty/nil")
	}
	s1 := make([][]byte, 1)
	s2 := make([][]byte, 2)
	s3 := make([][]byte, 3)
	if !MustCompile("^abc$").Match([]byte("abc"), s1) {
		t.Error("abc")
	}
	if !equal(strings(s1), []string{"abc"}) {
		t.Error("abc", s1)
	}
	s2[1] = []byte{65}
	if !MustCompile("^abc$").Match([]byte("abc"), s2) {
		t.Error("abc")
	}
	if !reflect.DeepEqual(s2, [][]byte{[]byte("abc"), nil}) {
		t.Error("abc", s2)
	}
	if !MustCompile("^(X)*ab(c)$").Match([]byte("abc"), s3) {
		t.Error("^(X)*abc$")
	}
	if !reflect.DeepEqual(s3, [][]byte{[]byte("abc"), nil, []byte("c")}) {
		t.Error("^(X)*abc$", s3)
	}
}
