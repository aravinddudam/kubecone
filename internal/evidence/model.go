package evidence

import "time"

const (
	KindPod         = "Pod"
	KindDeployment  = "Deployment"
	KindReplicaSet  = "ReplicaSet"
	KindService     = "Service"
	KindNode        = "Node"
	KindEvent       = "Event"
	KindContainer   = "Container"
	KindLog         = "Log"
	KindWorkload    = "Workload"
)

const (
	RelOwns      = "owns"
	RelCreates   = "creates"
	RelRunsOn    = "runs-on"
	RelReports   = "reports"
	RelContains  = "contains"
	RelSelects   = "selects"
	RelCausedBy  = "caused-by"
)

type Target struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (t Target) ID() string {
	return t.Namespace + "/" + t.Kind + "/" + t.Name
}

type Condition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

type WorkloadFact struct {
	Kind            string      `json:"kind"`
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	Replicas        int32       `json:"replicas"`
	ReadyReplicas   int32       `json:"readyReplicas"`
	Unavailable     int32       `json:"unavailableReplicas"`
	Conditions      []Condition `json:"conditions,omitempty"`
	Selector        string      `json:"selector,omitempty"`
	Generation      int64       `json:"generation,omitempty"`
	Observed        int64       `json:"observedGeneration,omitempty"`
}

type PodFact struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Phase      string      `json:"phase"`
	Reason     string      `json:"reason,omitempty"`
	Message    string      `json:"message,omitempty"`
	Node       string      `json:"node,omitempty"`
	OwnerKind  string      `json:"ownerKind,omitempty"`
	OwnerName  string      `json:"ownerName,omitempty"`
	QOS        string      `json:"qos,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type ContainerFact struct {
	Pod                    string `json:"pod"`
	Namespace              string `json:"namespace"`
	Name                   string `json:"name"`
	Image                  string `json:"image,omitempty"`
	Ready                  bool   `json:"ready"`
	RestartCount           int32  `json:"restartCount"`
	WaitingReason          string `json:"waitingReason,omitempty"`
	WaitingMessage         string `json:"waitingMessage,omitempty"`
	TerminatedReason       string `json:"terminatedReason,omitempty"`
	TerminatedMessage      string `json:"terminatedMessage,omitempty"`
	TerminatedExitCode     *int32 `json:"terminatedExitCode,omitempty"`
	LastTerminatedReason   string `json:"lastTerminatedReason,omitempty"`
	LastTerminatedExitCode *int32 `json:"lastTerminatedExitCode,omitempty"`
	LastTerminatedMessage  string `json:"lastTerminatedMessage,omitempty"`
	MemoryRequest          string `json:"memoryRequest,omitempty"`
	MemoryLimit            string `json:"memoryLimit,omitempty"`
	CPURequest             string `json:"cpuRequest,omitempty"`
	CPULimit               string `json:"cpuLimit,omitempty"`
	Started                bool   `json:"started"`
}

type EventFact struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Type         string    `json:"type"`
	Reason       string    `json:"reason"`
	Message      string    `json:"message"`
	Count        int32     `json:"count"`
	InvolvedKind string    `json:"involvedKind"`
	InvolvedName string    `json:"involvedName"`
	LastSeen     time.Time `json:"lastSeen,omitempty"`
}

type LogFact struct {
	Pod       string `json:"pod"`
	Container string `json:"container"`
	Namespace string `json:"namespace"`
	Tail      string `json:"tail"`
}

type NodeFact struct {
	Name       string      `json:"name"`
	Ready      bool        `json:"ready"`
	Conditions []Condition `json:"conditions,omitempty"`
}

type ServiceFact struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	Selector  map[string]string `json:"selector,omitempty"`
}

// Snapshot is the collected picture of a workload. Analyzers only see this.
type Snapshot struct {
	Target      Target          `json:"target"`
	Cluster     string          `json:"cluster,omitempty"`
	CollectedAt time.Time       `json:"collectedAt"`
	Workload    *WorkloadFact   `json:"workload,omitempty"`
	Pods        []PodFact       `json:"pods,omitempty"`
	Containers  []ContainerFact `json:"containers,omitempty"`
	Events      []EventFact     `json:"events,omitempty"`
	Logs        []LogFact       `json:"logs,omitempty"`
	Nodes       []NodeFact      `json:"nodes,omitempty"`
	Services    []ServiceFact   `json:"services,omitempty"`
	Warnings    []string        `json:"warnings,omitempty"`
}

const (
	SevCritical = "critical"
	SevHigh     = "high"
	SevMedium   = "medium"
	SevLow      = "low"
	SevInfo     = "info"
)

const (
	CodeOOMKilled        = "OOM_KILLED"
	CodeCrashLoop        = "CRASH_LOOP"
	CodeImagePull        = "IMAGE_PULL"
	CodeFailedScheduling = "FAILED_SCHEDULING"
	CodeProbeFailed      = "PROBE_FAILED"
	CodeHealthy          = "HEALTHY"
)

type Finding struct {
	Code           string            `json:"code"`
	Severity       string            `json:"severity"`
	Title          string            `json:"title"`
	Summary        string            `json:"summary"`
	Resource       string            `json:"resource"`
	EvidenceIDs    []string          `json:"evidenceIds,omitempty"`
	Recommendation string            `json:"recommendation,omitempty"`
	Confidence     float64           `json:"confidence"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

type Node struct {
	ID         string            `json:"id"`
	Kind       string            `json:"kind"`
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace,omitempty"`
	Status     string            `json:"status,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Edge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

type Report struct {
	Target      Target    `json:"target"`
	Cluster     string    `json:"cluster,omitempty"`
	CollectedAt time.Time `json:"collectedAt"`
	Primary     *Finding  `json:"primary,omitempty"`
	Findings    []Finding `json:"findings"`
	Graph       Graph     `json:"graph"`
	Evidence    Snapshot  `json:"evidence"`
	AI          *AINote   `json:"ai,omitempty"`
}

type AINote struct {
	Provider string `json:"provider"`
	Skipped  bool   `json:"skipped"`
	Reason   string `json:"reason,omitempty"`
	Text     string `json:"text,omitempty"`
}
