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

func TestMatcher(t *testing.T) {
	check := func (pattern, subject string, args ...interface{}) {
		re := MustCompile(pattern)
		m := re.Matcher([]byte(subject))
		if len(args) == 0 {
			if m.Matches() {
				t.Error(pattern, subject, "!Matches")
			}
		} else {
			if !m.Matches() {
				t.Error(pattern, subject, "Matches")
				return
			}
			if m.Groups() != len(args) - 1 {
				t.Error(pattern, subject, "Groups", m.Groups())
				return
			}
			for i, arg := range args {
				if s, ok := arg.(string); ok {
					if !m.Present(i) {
						t.Error(pattern, subject,
							"Present", i)

					}
					if s != string(m.Group(i)) {
						t.Error(pattern, subject,
							"Group", i, s)
					}
					if s != m.GroupString(i) {
						t.Error(pattern, subject,
							"GroupString", i, s)
					}
				} else {
					if m.Present(i) {
						t.Error(pattern, subject,
							"!Present", i)
					}
				}
			}
		}
	}

	check(`^$`, "", "")
	check(`^abc$`, "abc", "abc")
	check(`^(X)*ab()c$`, "abc", "abc", nil, "")
}
