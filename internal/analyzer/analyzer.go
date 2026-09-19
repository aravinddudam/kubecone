package analyzer

import "github.com/aravinddudam/kubecone/internal/evidence"

// Analyzer is a deterministic rule. It reads a snapshot and emits findings.
type Analyzer interface {
	Name() string
	Analyze(snap evidence.Snapshot) []evidence.Finding
}

func Default() []Analyzer {
	return []Analyzer{
		ImagePull{},
		Scheduling{},
		OOM{},
		Probe{},
		CrashLoop{},
	}
}

func Run(snap evidence.Snapshot, analyzers []Analyzer) []evidence.Finding {
	if analyzers == nil {
		analyzers = Default()
	}
	var findings []evidence.Finding
	for _, a := range analyzers {
		for _, f := range a.Analyze(snap) {
			findings = append(findings, evidence.Correlate(snap, f))
		}
	}
	return Dedupe(Rank(findings))
}

func Dedupe(findings []evidence.Finding) []evidence.Finding {
	if len(findings) < 2 {
		return findings
	}
	seen := map[string]struct{}{}
	out := make([]evidence.Finding, 0, len(findings))
	for _, f := range findings {
		key := f.Code + "\x00" + f.Resource + "\x00" + f.Title
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, f)
	}
	return out
}

var severityRank = map[string]int{
	evidence.SevCritical: 5,
	evidence.SevHigh:     4,
	evidence.SevMedium:   3,
	evidence.SevLow:      2,
	evidence.SevInfo:     1,
}

var codeRank = map[string]int{
	evidence.CodeImagePull:        50,
	evidence.CodeFailedScheduling: 45,
	evidence.CodeOOMKilled:        40,
	evidence.CodeProbeFailed:      30,
	evidence.CodeCrashLoop:        20,
	evidence.CodeHealthy:          0,
}

func Rank(findings []evidence.Finding) []evidence.Finding {
	if len(findings) < 2 {
		return findings
	}
	sorted := append([]evidence.Finding(nil), findings...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if less(sorted[j], sorted[i]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func less(a, b evidence.Finding) bool {
	if severityRank[a.Severity] != severityRank[b.Severity] {
		return severityRank[a.Severity] > severityRank[b.Severity]
	}
	if codeRank[a.Code] != codeRank[b.Code] {
		return codeRank[a.Code] > codeRank[b.Code]
	}
	return a.Confidence > b.Confidence
}

func Primary(findings []evidence.Finding) *evidence.Finding {
	if len(findings) == 0 {
		return nil
	}
	f := findings[0]
	return &f
}

func resourceName(namespace, kind, name string) string {
	return namespace + "/" + kind + "/" + name
}
