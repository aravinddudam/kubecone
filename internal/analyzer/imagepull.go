package analyzer

import (
	"fmt"
	"strings"

	"github.com/aravinddudam/kubecone/internal/evidence"
)

type ImagePull struct{}

func (ImagePull) Name() string { return "imagepull" }

func (ImagePull) Analyze(snap evidence.Snapshot) []evidence.Finding {
	var findings []evidence.Finding
	seen := map[string]struct{}{}

	for _, c := range snap.Containers {
		if !looksLikeImagePull(c.WaitingReason, c.WaitingMessage) {
			continue
		}
		if _, ok := seen[c.Pod]; ok {
			continue
		}
		seen[c.Pod] = struct{}{}
		class := classifyImagePull(c.WaitingReason, c.WaitingMessage, c.Image)
		msg := firstNonEmpty(c.WaitingMessage, c.WaitingReason, "image could not be pulled")
		findings = append(findings, evidence.Finding{
			Code:     evidence.CodeImagePull,
			Severity: evidence.SevCritical,
			Title:    "Image cannot be pulled",
			Summary: fmt.Sprintf(
				"%s/%s cannot start because image %q failed to pull (%s). %s",
				c.Pod, c.Name, valueOr(c.Image, "unknown"), valueOr(c.WaitingReason, "ErrImagePull"), class.Reason,
			),
			Resource:       resourceName(c.Namespace, "Pod", c.Pod),
			Recommendation: class.Next,
			Confidence:     0.96,
			Attributes: map[string]string{
				"container":         c.Name,
				"image":             c.Image,
				"reason":            c.WaitingReason,
				"cause":             class.Cause,
				"classifiedReason":  class.Reason,
				"runtimeError":      trim(msg, 300),
			},
		})
	}

	for _, ev := range snap.Events {
		if !looksLikeImagePull(ev.Reason, ev.Message) {
			continue
		}
		key := ev.InvolvedName
		if key == "" {
			key = ev.Name
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		class := classifyImagePull(ev.Reason, ev.Message, "")
		findings = append(findings, evidence.Finding{
			Code:           evidence.CodeImagePull,
			Severity:       evidence.SevCritical,
			Title:          "Image cannot be pulled",
			Summary:        trim(firstNonEmpty(class.Reason, ev.Message), 240),
			Resource:       resourceName(ev.Namespace, ev.InvolvedKind, ev.InvolvedName),
			Recommendation: class.Next,
			Confidence:     0.9,
			Attributes: map[string]string{
				"cause":            class.Cause,
				"classifiedReason": class.Reason,
				"runtimeError":     trim(ev.Message, 300),
			},
		})
	}
	return findings
}

type pullClass struct {
	Cause  string
	Reason string
	Next   string
}

func classifyImagePull(waitingReason, message, image string) pullClass {
	blob := strings.ToLower(waitingReason + " " + message + " " + image)
	ecr := strings.Contains(blob, ".dkr.ecr.") || strings.Contains(blob, "amazonaws.com/")
	repo, tag := splitImageRef(image)

	switch {
	case containsAny(blob, "unauthorized", "authentication", "no basic auth", "401", "403") ||
		(strings.Contains(blob, "denied") && !strings.Contains(blob, "not found")):
		if ecr {
			return pullClass{
				Cause:  "auth",
				Reason: "ECR authentication or authorization failed",
				Next:   "Confirmed registry auth failure. Check the node instance profile or IRSA for ecr:GetAuthorizationToken, ecr:BatchGetImage, and ecr:GetDownloadUrlForLayer. A dockerconfig pull secret is the wrong fix for ECR on AWS.",
			}
		}
		return pullClass{
			Cause:  "auth",
			Reason: "Registry authentication failed",
			Next:   "Create or fix an imagePullSecret and attach it to the service account or pod spec.",
		}
	case containsAny(blob, "not found", "manifest unknown", "failed to resolve", "no such image", "repository does not exist", "name unknown"):
		reason := "Registry returned: repository/tag not found"
		next := "Correct the image name and tag. The registry does not have this reference."
		if ecr {
			next = "Verify the repository and tag exist in ECR before changing credentials."
			if tag != "" && ecrRepoName(repo) != "" {
				next = fmt.Sprintf("Image tag %s does not appear to exist in ECR. Verify with: aws ecr describe-images --repository-name %s --image-ids imageTag=%s", tag, ecrRepoName(repo), tag)
			}
		}
		return pullClass{Cause: "missing", Reason: reason, Next: next}
	case containsAny(blob, "no such host", "lookup ", "dns"):
		return pullClass{
			Cause:  "dns",
			Reason: "Node cannot resolve the registry hostname",
			Next:   "Check DNS from the node (CoreDNS / VPC DNS) and that the registry hostname is correct.",
		}
	case containsAny(blob, "i/o timeout", "connection refused", "tls", "network", "dial tcp", "timeout"):
		return pullClass{
			Cause:  "network",
			Reason: "Node cannot reach the registry",
			Next:   "Check node egress, security groups, and that the registry endpoint is reachable from the node.",
		}
	case containsAny(blob, "no match for platform", "exec format", "image architecture"):
		return pullClass{
			Cause:  "platform",
			Reason: "Image architecture does not match the node",
			Next:   "Pull an image variant that matches the node OS/arch, or set a matching nodeSelector.",
		}
	default:
		reason := firstNonEmpty(trim(message, 160), waitingReason, "container runtime could not pull the image")
		next := "Inspect the full ErrImagePull / container-runtime message before changing credentials. Common causes: missing tag, registry auth, node network/DNS, or IAM for ECR."
		if ecr {
			next = "This is an Amazon ECR image, but the pull error does not confirm authentication yet. Read the full runtime message; if it is a missing tag, fix the reference, and only then check instance profile or IRSA."
		}
		return pullClass{Cause: "unknown", Reason: reason, Next: next}
	}
}

func containsAny(blob string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(blob, p) {
			return true
		}
	}
	return false
}

func splitImageRef(image string) (repo, tag string) {
	image = strings.TrimSpace(image)
	if image == "" {
		return "", ""
	}
	tag = "latest"
	if i := strings.LastIndex(image, ":"); i > 0 && !strings.Contains(image[i:], "/") {
		tag = image[i+1:]
		image = image[:i]
	}
	return image, tag
}

func ecrRepoName(repo string) string {
	// 123.dkr.ecr.region.amazonaws.com/app-dev/customer-service
	_, path, ok := strings.Cut(repo, "/")
	if !ok || path == "" {
		return ""
	}
	return path
}

// imagePullAdvice is kept for tests and older call sites.
func imagePullAdvice(message, image string) string {
	return classifyImagePull("", message, image).Next
}
