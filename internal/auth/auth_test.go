package auth

import "testing"

func TestFirstNonEmpty(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"", "  ", "x"}, "x"},
		{[]string{"a", "b"}, "a"},
		{[]string{"  a  ", "b"}, "a"},
		{[]string{"", ""}, ""},
	}
	for _, c := range cases {
		if got := firstNonEmpty(c.in...); got != c.want {
			t.Fatalf("firstNonEmpty(%v)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestParseJWTClaims_Invalid(t *testing.T) {
	if _, err := parseJWTClaims("notatoken"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
