package pingrelay

import (
	"fmt"
	"testing"

	"github.com/agext/levenshtein"
	"github.com/stretchr/testify/assert"
)

func TestNCParser(t *testing.T) {
	for i, test := range []struct {
		Ping          string
		ExpectedOut   *NCPing
		ExpectedError error
	}{
		{
			Ping: "FUUUFA [nc]: MAX DUDES FLEET 3 NOW IN GEHI GOGOGOGOGOG",
			ExpectedOut: &NCPing{
				Ping:    "MAX DUDES FLEET 3 NOW IN GEHI GOGOGOGOGOG",
				ToGroup: "nc",
				From:    "FUUUFA",
			},
			ExpectedError: nil,
		},
		{
			Ping:          "FUUUFA: MAX DUDES FLEET 3 NOW IN GEHI GOGOGOGOGOG",
			ExpectedOut:   nil,
			ExpectedError: errNotPing,
		},
	} {
		presult, err := ParseNCMessage(test.Ping)
		if err != nil && test.ExpectedError != nil {
			assert.Equal(t, test.ExpectedError, err, "errors do not match in test %d", i)
			continue
		}

		if test.ExpectedError == nil && err != nil {
			assert.Fail(t, "unexpected error in test %d: %s", i, err.Error())
			continue
		}

		assert.Equal(t, test.ExpectedOut, presult)
	}
}

func TestStringCommonality(t *testing.T) {
	fmt.Println(levenshtein.Similarity("testing", "testing hello world", nil))
}
