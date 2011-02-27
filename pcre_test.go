package pcre

import (
	"testing"
)

func TestCompile(t *testing.T) {
	var check = func (p string, groups int) {
		re, err := Compile(p)
		if err != nil {
			t.Error(p, err)
		}
		if g := re.Groups(); g != groups {
			t.Error(p, g)
		}
	}
	check("",0 )
	check("^", 0)
	check("^$", 0)
	check("()", 1)
	check("(())", 2)
	check("((?:))", 1)
}

func TestCompileFail(t *testing.T) {
	var check = func (p, msg string, off int) {
		_, err := Compile(p)
		switch {
		case err == nil:
			t.Error(p)
		case err.Message != msg:
			t.Error(p, "Message", err.Message)
		case err.Offset != off:
			t.Error(p, "Offset", err.Offset)
		}
	}
	check("(", "missing )", 1)
	check("\\", "\\ at end of pattern", 1)
	check("abc\\", "\\ at end of pattern", 4)
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

func checkmatch1(t *testing.T, dostring bool, pattern, subject string,
	args ...interface{}) {
	re := MustCompile(pattern)
	var m *Matcher
	var prefix string
	if dostring {
		m = re.MatcherString(subject)
		prefix = "string"
	} else {
		m = re.Matcher([]byte(subject))
		prefix = "[]byte"
	}
	if len(args) == 0 {
		if m.Matches() {
			t.Error(prefix, pattern, subject, "!Matches")
		}
	} else {
		if !m.Matches() {
			t.Error(prefix, pattern, subject, "Matches")
			return
		}
		if m.Groups() != len(args) - 1 {
			t.Error(prefix, pattern, subject, "Groups", m.Groups())
			return
		}
		for i, arg := range args {
			if s, ok := arg.(string); ok {
				if !m.Present(i) {
					t.Error(prefix, pattern, subject,
						"Present", i)

				}
				if g := string(m.Group(i)); g != s {
					t.Error(prefix, pattern, subject,
						"Group", i, g, "!=", s)
				}
				if g := m.GroupString(i); g != s {
					t.Error(prefix, pattern, subject,
						"GroupString", i, g, "!=", s)
				}
			} else {
				if m.Present(i) {
					t.Error(prefix, pattern, subject,
						"!Present", i)
				}
			}
		}
	}
}

func TestMatcher(t *testing.T) {
	check := func(pattern, subject string, args ...interface{}) {
		checkmatch1(t, false, pattern, subject, args...)
		checkmatch1(t, true, pattern, subject, args...)
	}

	check(`^$`, "", "")
	check(`^abc$`, "abc", "abc")
	check(`^(X)*ab(c)$`, "abc", "abc", nil, "c")
	check(`^(X)*ab()c$`, "abc", "abc", nil, "")
	check(`^.*$`, "abc", "abc")
	check(`^.*$`, "a\000c", "a\000c")
	check(`^(.*)$`, "a\000c", "a\000c", "a\000c")
}
