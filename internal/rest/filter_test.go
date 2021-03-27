package rest

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestCreateFilter(t *testing.T) {
	t.Run("valid-serviceId", func(t *testing.T) {
		q, err := url.Parse("/foo?serviceId=123")
		if err != nil {
			t.Fatal(err)
		}

		f := createFilter(q.Query(), []string{"serviceId"})

		assert.True(t, f.Has("serviceId"), "no serviceId in filter")
		assert.Equal(t, "123", f.StringOrDefault("serviceId", ""))
	})

	t.Run("not-allowed-filters-are-ignored", func(t *testing.T) {
		q, err := url.Parse("/foo?bar=123")
		if err != nil {
			t.Fatal(err)
		}

		f := createFilter(q.Query(), []string{"foo"})

		assert.False(t, f.Has("bar"))
		assert.Equal(t, "", f.StringOrDefault("bar", ""))
	})
}