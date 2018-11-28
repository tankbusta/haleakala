package pingrelay

import "testing"

func Test_checkMessageForSpam(t *testing.T) {
	words := []string{"incursion", "twitch", "stream"}

	ok := checkMessageForSpam("check out my twitch stream!", words)
	if !ok {
		t.Fatal("expected to find a forbidden word")
	}

	ok = checkMessageForSpam("Hey guys ! This weekend is burn Jita so keep that in mind ! DONT FEED PLS <3<3", words)
	if ok {
		t.Fatal("expected to find a forbidden word")
	}
}
