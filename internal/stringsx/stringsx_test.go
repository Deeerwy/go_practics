package stringsx

import "testing"

func TestClip(t *testing.T) {
	cases := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"empty_string", "", 4, ""},
		{"max_zero", "hello", 0, ""},
		{"max_negative", "hello", -15, ""},
		{"max_equals_len", "hello", 25, "hello"},
		{"max_greater_len", "hello", 15, "hello"},
		{"clip_normal", "hello", 3, "hel"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			got := Clip(c.s, c.max)
			if got != c.want {
				t.Fatalf("Clip(%q,%d)=%q; want %q", c.s, c.max, got, c.want)
			}
		})
	}
}