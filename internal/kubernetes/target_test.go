package kubernetes

import "testing"

func TestParseTarget(t *testing.T) {
	tests := []struct {
		raw, ns string
		want    Target
	}{
		{"deployment/payment-api", "payments", Target{Kind: "Deployment", Name: "payment-api", Namespace: "payments"}},
		{"deploy/payment-api", "", Target{Kind: "Deployment", Name: "payment-api", Namespace: "default"}},
		{"pod/web-0", "prod", Target{Kind: "Pod", Name: "web-0", Namespace: "prod"}},
		{"prod/svc/api", "ignored", Target{Kind: "Service", Name: "api", Namespace: "prod"}},
		{"payment-api", "shop", Target{Kind: "Deployment", Name: "payment-api", Namespace: "shop"}},
	}
	for _, tt := range tests {
		got, err := ParseTarget(tt.raw, tt.ns)
		if err != nil {
			t.Fatalf("ParseTarget(%q): %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("ParseTarget(%q) = %+v, want %+v", tt.raw, got, tt.want)
		}
	}
}

func TestParseTargetErrors(t *testing.T) {
	if _, err := ParseTarget("", "default"); err == nil {
		t.Fatal("expected error for empty resource")
	}
	if _, err := ParseTarget("secret/db", "default"); err == nil {
		t.Fatal("expected error for unsupported kind")
	}
}
