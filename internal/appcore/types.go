package appcore

import (
	"strings"
	"time"

	"k8s-recovery-visualizer/internal/model"
	"k8s-recovery-visualizer/internal/theme"
)

const (
	ConnectionMethodCurrent          = "current"
	ConnectionMethodKubeconfigFile   = "kubeconfig_file"
	ConnectionMethodKubeconfigInline = "kubeconfig_inline"
	ConnectionMethodAPIEndpoint      = "api_endpoint"
)

type ScanRequest struct {
	RunID                 string   `json:"runId,omitempty"`
	ConnectionMethod      string   `json:"connectionMethod,omitempty"`
	KubeconfigPath        string   `json:"kubeconfigPath,omitempty"`
	KubeconfigContent     string   `json:"kubeconfigContent,omitempty"`
	ContextName           string   `json:"contextName,omitempty"`
	APIServerEndpoint     string   `json:"apiServerEndpoint,omitempty"`
	BearerToken           string   `json:"bearerToken,omitempty"`
	CACertPath            string   `json:"caCertPath,omitempty"`
	CACertContent         string   `json:"caCertContent,omitempty"`
	OutputDir             string   `json:"outputDir,omitempty"`
	DryRun                bool     `json:"dryRun,omitempty"`
	CI                    bool     `json:"ci,omitempty"`
	MinScore              int      `json:"minScore,omitempty"`
	TimeoutSeconds        int      `json:"timeoutSeconds,omitempty"`
	CustomerID            string   `json:"customerId,omitempty"`
	Site                  string   `json:"site,omitempty"`
	ClusterName           string   `json:"clusterName,omitempty"`
	Environment           string   `json:"environment,omitempty"`
	Target                string   `json:"target,omitempty"`
	CSVExport             bool     `json:"csvExport,omitempty"`
	Namespaces            []string `json:"namespaces,omitempty"`
	CompareTo             string   `json:"compareTo,omitempty"`
	Summary               bool     `json:"summary,omitempty"`
	Redact                bool     `json:"redact,omitempty"`
	ProfileName           string   `json:"profileName,omitempty"`
	Runbook               bool     `json:"runbook,omitempty"`
	Insecure              bool     `json:"insecure,omitempty"`
	IncludeSecretMetadata bool     `json:"includeSecretMetadata,omitempty"`
}

func (r ScanRequest) Normalized() ScanRequest {
	out := r
	out.ConnectionMethod = normalizeConnectionMethod(out)
	if out.OutputDir == "" {
		out.OutputDir = "./out"
	}
	if out.TimeoutSeconds <= 0 {
		out.TimeoutSeconds = 60
	}
	if out.Target == "" {
		out.Target = "vm"
	}
	if out.ProfileName == "" {
		out.ProfileName = "standard"
	}
	return out
}

func (r ScanRequest) Timeout() time.Duration {
	return time.Duration(r.Normalized().TimeoutSeconds) * time.Second
}

func normalizeConnectionMethod(r ScanRequest) string {
	switch strings.TrimSpace(r.ConnectionMethod) {
	case ConnectionMethodCurrent, ConnectionMethodKubeconfigFile, ConnectionMethodKubeconfigInline, ConnectionMethodAPIEndpoint:
		return strings.TrimSpace(r.ConnectionMethod)
	}
	switch {
	case strings.TrimSpace(r.APIServerEndpoint) != "":
		return ConnectionMethodAPIEndpoint
	case strings.TrimSpace(r.KubeconfigContent) != "":
		return ConnectionMethodKubeconfigInline
	case strings.TrimSpace(r.KubeconfigPath) != "":
		return ConnectionMethodKubeconfigFile
	default:
		return ConnectionMethodCurrent
	}
}

type ExportRequest struct {
	OutputDir  string `json:"outputDir"`
	Report     bool   `json:"report,omitempty"`
	BundleJSON bool   `json:"bundleJson,omitempty"`
	CSVExport  bool   `json:"csvExport,omitempty"`
	Summary    bool   `json:"summary,omitempty"`
	Runbook    bool   `json:"runbook,omitempty"`
	Redact     bool   `json:"redact,omitempty"`
}

type ArtifactPaths struct {
	OutputDir           string `json:"outputDir"`
	BundleJSON          string `json:"bundleJson,omitempty"`
	EnrichedJSON        string `json:"enrichedJson,omitempty"`
	HTMLReport          string `json:"htmlReport,omitempty"`
	MarkdownReport      string `json:"markdownReport,omitempty"`
	SummaryHTML         string `json:"summaryHtml,omitempty"`
	RunbookHTML         string `json:"runbookHtml,omitempty"`
	RedactedJSON        string `json:"redactedJson,omitempty"`
	RedactedHTML        string `json:"redactedHtml,omitempty"`
	CSVDir              string `json:"csvDir,omitempty"`
	HistoryIndex        string `json:"historyIndex,omitempty"`
	HistoryLatestHTML   string `json:"historyLatestHtml,omitempty"`
	LoadedBundlePath    string `json:"loadedBundlePath,omitempty"`
	LoadedBundleDirHint string `json:"loadedBundleDirHint,omitempty"`
}

type HistoryEntry struct {
	TimestampUTC string `json:"timestampUtc"`
	ClusterName  string `json:"clusterName,omitempty"`
	Environment  string `json:"environment,omitempty"`
	Overall      int    `json:"overall"`
	Storage      int    `json:"storage,omitempty"`
	Workload     int    `json:"workload,omitempty"`
	Config       int    `json:"config,omitempty"`
	Backup       int    `json:"backup,omitempty"`
	Findings     int    `json:"findings,omitempty"`
	Maturity     string `json:"maturity"`
	JSONPath     string `json:"jsonPath,omitempty"`
	HTMLPath     string `json:"htmlPath,omitempty"`
}

type HistoryDomainTrend struct {
	Name      string `json:"name"`
	Current   int    `json:"current"`
	Delta     int    `json:"delta"`
	Direction string `json:"direction"`
}

type HistoryDashboard struct {
	Entries      []HistoryEntry       `json:"entries"`
	TrendLabel   string               `json:"trendLabel,omitempty"`
	TrendDelta   int                  `json:"trendDelta,omitempty"`
	AverageScore int                  `json:"averageScore,omitempty"`
	BestScore    int                  `json:"bestScore,omitempty"`
	WorstScore   int                  `json:"worstScore,omitempty"`
	RunCount     int                  `json:"runCount,omitempty"`
	DomainTrends []HistoryDomainTrend `json:"domainTrends,omitempty"`
}

type Workspace struct {
	Bundle    model.Bundle     `json:"bundle"`
	Artifacts ArtifactPaths    `json:"artifacts"`
	History   HistoryDashboard `json:"history"`
	Source    string           `json:"source"`
	LoadedAt  string           `json:"loadedAt"`
}

type ProjectSummary struct {
	Name         string `json:"name"`
	ClusterName  string `json:"clusterName,omitempty"`
	Environment  string `json:"environment,omitempty"`
	OutputDir    string `json:"outputDir"`
	LastScanPath string `json:"lastScanPath"`
	ReportPath   string `json:"reportPath,omitempty"`
	Score        int    `json:"score"`
	Maturity     string `json:"maturity"`
	TimestampUTC string `json:"timestampUtc,omitempty"`
}

type PreflightReport struct {
	CanRun      bool              `json:"canRun"`
	Degraded    bool              `json:"degraded"`
	Server      string            `json:"server,omitempty"`
	ContextName string            `json:"contextName,omitempty"`
	Scope       string            `json:"scope"`
	Diagnosis   *FailureDiagnosis `json:"diagnosis,omitempty"`
	Checks      []PreflightCheck  `json:"checks"`
	Warnings    []string          `json:"warnings,omitempty"`
}

type ContextCatalog struct {
	Contexts       []string `json:"contexts,omitempty"`
	CurrentContext string   `json:"currentContext,omitempty"`
	Source         string   `json:"source,omitempty"`
}

type ConnectionAdvisor struct {
	RecommendedMethod               string `json:"recommendedMethod,omitempty"`
	RecommendedReason               string `json:"recommendedReason,omitempty"`
	KubectlAvailable                bool   `json:"kubectlAvailable,omitempty"`
	KubectlPath                     string `json:"kubectlPath,omitempty"`
	CurrentLoginAvailable           bool   `json:"currentLoginAvailable,omitempty"`
	CurrentContext                  string `json:"currentContext,omitempty"`
	CurrentLoginDetail              string `json:"currentLoginDetail,omitempty"`
	CurrentLoginWarning             string `json:"currentLoginWarning,omitempty"`
	DefaultKubeconfigAvailable      bool   `json:"defaultKubeconfigAvailable,omitempty"`
	DefaultKubeconfigPath           string `json:"defaultKubeconfigPath,omitempty"`
	DefaultKubeconfigCurrentContext string `json:"defaultKubeconfigCurrentContext,omitempty"`
	DefaultKubeconfigDetail         string `json:"defaultKubeconfigDetail,omitempty"`
	DefaultKubeconfigPortable       bool   `json:"defaultKubeconfigPortable,omitempty"`
	DefaultKubeconfigWarning        string `json:"defaultKubeconfigWarning,omitempty"`
}

type KubeconfigInspection struct {
	Source                       string   `json:"source,omitempty"`
	Path                         string   `json:"path,omitempty"`
	CurrentContext               string   `json:"currentContext,omitempty"`
	Contexts                     []string `json:"contexts,omitempty"`
	ClusterCount                 int      `json:"clusterCount"`
	UserCount                    int      `json:"userCount"`
	UsesExecAuth                 bool     `json:"usesExecAuth,omitempty"`
	UsesClientCertificate        bool     `json:"usesClientCertificate,omitempty"`
	UsesCertificateAuthorityFile bool     `json:"usesCertificateAuthorityFile,omitempty"`
	UsesCertificateAuthorityData bool     `json:"usesCertificateAuthorityData,omitempty"`
	Servers                      []string `json:"servers,omitempty"`
	LoopbackServers              []string `json:"loopbackServers,omitempty"`
	ReferencedFiles              []string `json:"referencedFiles,omitempty"`
	MissingReferencedFiles       []string `json:"missingReferencedFiles,omitempty"`
	Summary                      string   `json:"summary,omitempty"`
	NextAction                   string   `json:"nextAction,omitempty"`
}

type ConnectionTestCheck struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Detail string `json:"detail"`
	Hint   string `json:"hint,omitempty"`
}

type FailureDiagnosis struct {
	Code       string `json:"code,omitempty"`
	Label      string `json:"label,omitempty"`
	Summary    string `json:"summary,omitempty"`
	Detail     string `json:"detail,omitempty"`
	NextAction string `json:"nextAction,omitempty"`
}

type ConnectionTestReport struct {
	CanConnect    bool                  `json:"canConnect"`
	Source        string                `json:"source,omitempty"`
	Server        string                `json:"server,omitempty"`
	ContextName   string                `json:"contextName,omitempty"`
	Summary       string                `json:"summary,omitempty"`
	NextAction    string                `json:"nextAction,omitempty"`
	Diagnosis     *FailureDiagnosis     `json:"diagnosis,omitempty"`
	FieldErrors   map[string]string     `json:"fieldErrors,omitempty"`
	FieldWarnings map[string]string     `json:"fieldWarnings,omitempty"`
	Checks        []ConnectionTestCheck `json:"checks,omitempty"`
}

type PreflightCheck struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Required bool     `json:"required"`
	Scope    string   `json:"scope,omitempty"`
	Resource string   `json:"resource,omitempty"`
	Detail   string   `json:"detail"`
	Hint     string   `json:"hint,omitempty"`
	Manifest string   `json:"manifest,omitempty"`
	Commands []string `json:"commands,omitempty"`
}

type RunEvent struct {
	Type      string               `json:"type"`
	RunID     string               `json:"runId"`
	Timestamp string               `json:"timestamp"`
	Step      string               `json:"step,omitempty"`
	Level     string               `json:"level,omitempty"`
	Message   string               `json:"message"`
	Percent   float64              `json:"percent,omitempty"`
	Artifact  string               `json:"artifact,omitempty"`
	Warning   string               `json:"warning,omitempty"`
	Skip      *model.CollectorSkip `json:"skip,omitempty"`
}

type RunResult struct {
	RunID      string          `json:"runId"`
	ExitCode   int             `json:"exitCode"`
	TrendLabel string          `json:"trendLabel,omitempty"`
	TrendDelta int             `json:"trendDelta,omitempty"`
	Artifacts  ArtifactPaths   `json:"artifacts"`
	Workspace  Workspace       `json:"workspace"`
	Preflight  PreflightReport `json:"preflight"`
}

type EventSink interface {
	Emit(RunEvent)
}

type Bootstrap struct {
	Theme theme.Tokens `json:"theme"`
}
