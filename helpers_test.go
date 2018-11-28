package haleakala

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrSliceContainsStr(t *testing.T) {
	testArray := []string{"hello", "world"}

	assert.True(t, StrSliceContainsStr("hello", testArray))
	assert.False(t, StrSliceContainsStr("apples", testArray))
}
