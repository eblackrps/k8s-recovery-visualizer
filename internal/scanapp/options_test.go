package scanapp

import (
	"bytes"
	"testing"
)

func TestParseArgsParsesNamespaceScope(t *testing.T) {
	var stderr bytes.Buffer
	opts, err := ParseArgs([]string{"--namespace=prod, staging", "--target=baremetal"}, &stderr)
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}
	if len(opts.Namespaces) != 2 || opts.Namespaces[0] != "prod" || opts.Namespaces[1] != "staging" {
		t.Fatalf("Namespaces = %v, want [prod staging]", opts.Namespaces)
	}
	if opts.Target != "baremetal" {
		t.Fatalf("Target = %q, want baremetal", opts.Target)
	}
}

func TestParseArgsRejectsUnknownTarget(t *testing.T) {
	var stderr bytes.Buffer
	if _, err := ParseArgs([]string{"--target=moon"}, &stderr); err == nil {
		t.Fatal("ParseArgs() error = nil, want invalid target error")
	}
}

func TestParseArgsEnablesSecretMetadataCollection(t *testing.T) {
	var stderr bytes.Buffer
	opts, err := ParseArgs([]string{"--include-secret-metadata"}, &stderr)
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}
	if !opts.IncludeSecretMetadata {
		t.Fatal("IncludeSecretMetadata = false, want true")
	}
}
