package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

func sampleReport() *evidence.Report {
	return &evidence.Report{
		Target:  evidence.Target{Kind: "Deployment", Name: "customer-service", Namespace: "banking-dev"},
		Cluster: "default",
		Primary: &evidence.Finding{
			Code:           evidence.CodeImagePull,
			Severity:       evidence.SevCritical,
			Title:          "Image cannot be pulled",
			Summary:        "customer-service-1/app cannot start",
			Recommendation: "Verify the repository and tag exist in ECR.",
			Attributes: map[string]string{
				"image":            "707405801682.dkr.ecr.us-east-1.amazonaws.com/app-dev/customer-service:d54a1c7",
				"classifiedReason": "Registry returned: repository/tag not found",
				"cause":            "missing",
			},
		},
		Evidence: evidence.Snapshot{
			Containers: []evidence.ContainerFact{{
				Pod:            "customer-service-1",
				Name:           "app",
				Image:          "707405801682.dkr.ecr.us-east-1.amazonaws.com/app-dev/customer-service:d54a1c7",
				WaitingReason:  "ImagePullBackOff",
				WaitingMessage: "failed to resolve reference",
			}},
		},
	}
}

func TestBuildPromptIncludesDiagnosis(t *testing.T) {
	p := BuildPrompt(sampleReport())
	for _, want := range []string{"IMAGE_PULL", "d54a1c7", "repository/tag not found", "customer-service"} {
		if !strings.Contains(p, want) {
			t.Fatalf("prompt missing %q:\n%s", want, p)
		}
	}
}

func TestOpenAIMissingKey(t *testing.T) {
	note, err := (OpenAI{}).Diagnose(context.Background(), sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if !note.Skipped || !strings.Contains(note.Reason, "OPENAI_API_KEY") {
		t.Fatalf("%+v", note)
	}
}

func TestOpenAIComplete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("auth=%s", r.Header.Get("Authorization"))
		}
		raw, _ := io.ReadAll(r.Body)
		var req chatRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			t.Errorf("body: %v", err)
		}
		if req.Model != "gpt-4o-mini" {
			t.Errorf("model=%s", req.Model)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message chatMessage `json:"message"`
			}{{
				Message: chatMessage{Role: "assistant", Content: "Likely cause:\nMissing ECR tag d54a1c7.\n\nWhat to do next:\n1. aws ecr describe-images\n"},
			}},
		})
	}))
	defer srv.Close()

	o := OpenAI{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: srv.URL, Client: srv.Client()}
	note, err := o.Diagnose(context.Background(), sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if note.Skipped {
		t.Fatalf("skipped: %s", note.Reason)
	}
	if !strings.Contains(note.Text, "Missing ECR tag") {
		t.Fatalf("text=%s", note.Text)
	}
}

func TestOpenAIHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"Incorrect API key provided"}}`))
	}))
	defer srv.Close()

	o := OpenAI{APIKey: "bad", BaseURL: srv.URL, Client: srv.Client()}
	note, err := o.Diagnose(context.Background(), sampleReport())
	if err != nil {
		t.Fatal(err)
	}
	if !note.Skipped || !strings.Contains(note.Reason, "Incorrect API key") {
		t.Fatalf("%+v", note)
	}
}
