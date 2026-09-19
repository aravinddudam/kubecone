package analyzer

import (
	"strings"
	"testing"
)

func TestRecommendMemory(t *testing.T) {
	got := RecommendMemory("512Mi", "")
	if !strings.Contains(got, "512Mi") {
		t.Fatalf("recommendation %q should mention current limit", got)
	}
	if !strings.Contains(got, "768") && !strings.Contains(strings.ToLower(got), "raise") {
		t.Fatalf("expected a raised limit, got %q", got)
	}
}
