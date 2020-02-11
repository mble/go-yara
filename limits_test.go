package yara

import "testing"

func TestLimits(t *testing.T) {
	l, err := FetchLimits()
	if err != nil {
		t.Errorf("error fetching limits: %v", err)
	}
	t.Logf("Limits: %+v", l)
}
