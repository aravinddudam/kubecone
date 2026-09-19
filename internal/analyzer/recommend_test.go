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
	if strings.Contains(strings.ToLower(got), "imagepullsecret") {
		t.Fatalf("missing tag must not prescribe imagePullSecret, got %q", got)
	}
	if !strings.Contains(got, "d54a1c7") || !strings.Contains(strings.ToLower(got), "ecr") {
		t.Fatalf("ECR missing tag should mention tag and ECR, got %q", got)
	}
}

func TestImagePullAdviceAuth(t *testing.T) {
	got := imagePullAdvice("unauthorized: authentication required", "example.invalid/app:v1")
	if !strings.Contains(strings.ToLower(got), "imagepullsecret") {
		t.Fatalf("generic auth should mention imagePullSecret, got %q", got)
	}
}

func TestImagePullAdviceECRAuth(t *testing.T) {
	got := imagePullAdvice("unauthorized", "707405801682.dkr.ecr.us-east-1.amazonaws.com/app-dev/customer-service:d54a1c7")
	if strings.Contains(strings.ToLower(got), "imagepullsecret") {
		t.Fatalf("ECR auth should not recommend imagePullSecret, got %q", got)
	}
	if !strings.Contains(strings.ToLower(got), "irsa") && !strings.Contains(strings.ToLower(got), "iam") {
		t.Fatalf("ECR auth should mention IAM/IRSA, got %q", got)
	}
}
