package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Run("allow", func(t *testing.T) {
		f := NewFilter([]string{"foo", "bar"})

		assert.True(t, f.Allows("foo"))
		assert.True(t, f.Allows("bar"))
		assert.False(t, f.Allows("baz"))
		assert.False(t, f.Allows(""))
	})

	t.Run("has-and-get", func(t *testing.T) {
		f := NewFilter([]string{"foo", "bar"})
		f.Add("foo", "1234").Add("bar", "baz")

		assert.True(t, f.Has("foo"))
		assert.Equal(t, "1234", f.StringOrDefault("foo", "boo"))
		assert.True(t, f.Has("bar"))
		assert.Equal(t, "baz", f.StringOrDefault("bar", ""))
		assert.False(t, f.Has("baz"))
		assert.False(t, f.Has(""))
	})
}
