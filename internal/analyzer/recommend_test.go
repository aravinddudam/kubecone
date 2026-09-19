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

func TestImagePullAdviceECR(t *testing.T) {
	got := imagePullAdvice("failed to resolve reference", "707405801682.dkr.ecr.us-east-1.amazonaws.com/app-dev/customer-service:d54a1c7")
	if !strings.Contains(strings.ToLower(got), "ecr") {
		t.Fatalf("ECR pull should mention ECR IAM/IRSA, got %q", got)
	}
}
