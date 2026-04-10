package scanapp

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"
)

type Options struct {
	Kubeconfig  string
	OutDir      string
	DryRun      bool
	CI          bool
	MinScore    int
	Timeout     time.Duration
	CustomerID  string
	Site        string
	Cluster     string
	Environment string
	Target      string
	CSVExport   bool
	Namespaces  []string
	CompareTo   string
	Summary     bool
	RedactOut   bool
	ProfileName string
	Runbook     bool
	Insecure    bool
}

func ParseArgs(args []string, stderr io.Writer) (Options, error) {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		namespaceArg string
		timeoutSec   int
		opts         Options
	)

	fs.StringVar(&opts.Kubeconfig, "kubeconfig", "", "Path to kubeconfig")
	fs.StringVar(&opts.OutDir, "out", "./out", "Output directory")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "Run without Kubernetes")
	fs.BoolVar(&opts.CI, "ci", false, "CI mode (machine-readable output)")
	fs.IntVar(&opts.MinScore, "min-score", 90, "Minimum acceptable DR score")
	fs.IntVar(&timeoutSec, "timeout", 60, "Timeout in seconds for Kubernetes API calls")
	fs.StringVar(&opts.CustomerID, "customer", "", "Customer identifier (optional)")
	fs.StringVar(&opts.Site, "site", "", "Site/region name (optional)")
	fs.StringVar(&opts.Cluster, "cluster", "", "Cluster name (optional)")
	fs.StringVar(&opts.Environment, "env", "", "Environment (prod/dev/test) (optional)")
	fs.StringVar(&opts.Target, "target", "vm", "Recovery target type: baremetal or vm")
	fs.BoolVar(&opts.CSVExport, "csv", false, "Also write CSV exports alongside HTML report")
	fs.StringVar(&namespaceArg, "namespace", "", "Comma-separated namespaces to scan (empty = all namespaces)")
	fs.StringVar(&opts.CompareTo, "compare", "", "Path to a previous recovery-scan.json to diff against")
	fs.BoolVar(&opts.Summary, "summary", false, "Also write a print-optimised executive summary HTML")
	fs.BoolVar(&opts.RedactOut, "redact", false, "Also write redacted JSON and HTML with masked identifiers")
	fs.StringVar(&opts.ProfileName, "profile", "standard", "Scoring profile: standard|enterprise|dev|airgap")
	fs.BoolVar(&opts.Runbook, "runbook", false, "Also write a customer-facing DR runbook HTML")
	fs.BoolVar(&opts.Insecure, "insecure", false, "Skip TLS certificate verification (use for self-signed certs, e.g. RKE2/k3s)")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if opts.Target != "baremetal" && opts.Target != "vm" {
		return Options{}, fmt.Errorf("--target must be 'baremetal' or 'vm', got %q", opts.Target)
	}

	opts.Timeout = time.Duration(timeoutSec) * time.Second
	if strings.TrimSpace(namespaceArg) != "" {
		for _, ns := range strings.Split(namespaceArg, ",") {
			ns = strings.TrimSpace(ns)
			if ns != "" {
				opts.Namespaces = append(opts.Namespaces, ns)
			}
		}
	}

	return opts, nil
}
