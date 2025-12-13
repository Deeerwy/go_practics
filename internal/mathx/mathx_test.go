package mathx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSum_Table(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"positive", 2, 3, 5},
		{"negative", 10, -5, 5},
		{"zeros", 0, 0, 0},
		{"both_negative", -2, -3, -5},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			got := Sum(c.a, c.b)
			if got != c.want {
				t.Fatalf("Sum(%d,%d)=%d; want %d", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestDivide_OkAndError(t *testing.T) {
	// обычная проверка без testify
	got, err := Divide(10, 2)
	if err != nil {
		t.Fatalf("Divide(10,2) unexpected err: %v", err)
	}
	if got != 5 {
		t.Fatalf("Divide(10,2)=%d; want 5", got)
	}

	_, err = Divide(10, 0)
	if err == nil {
		t.Fatalf("Divide(10,0) expected error, got nil")
	}
}

func TestDivide_WithTestify(t *testing.T) {
	got, err := Divide(10, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, got)

	_, err = Divide(10, 0)
	assert.Error(t, err)
}