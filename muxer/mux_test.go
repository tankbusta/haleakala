package muxer

import (
	"testing"
)

func TestFuzzyMatch(t *testing.T) {
	muxer := New()
	muxer.Route("penis", "Testing", func(_ *Context) {})
	muxer.Route("ping_stats", "Testing", func(_ *Context) {})

	route, _ := muxer.FuzzyMatch("p")
	if route == nil || route != nil && route.Pattern != "penis" {
		t.Errorf("Expected route to be returned as penis")
	}
}

func TestMatch(t *testing.T) {
	muxer := New()
	muxer.Route("penis", "Testing", func(_ *Context) {})
	muxer.Route("ping_stats", "Testing", func(_ *Context) {})

	route, _ := muxer.Match("p")
	if route != nil {
		t.Errorf("Expected route to be nil")
	}
}
