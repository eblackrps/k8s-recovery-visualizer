package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func main() {
	schemaPath := flag.String("schema", "", "Path to the JSON schema file")
	inputPath := flag.String("input", "", "Path to the JSON document to validate")
	flag.Parse()

	if strings.TrimSpace(*schemaPath) == "" || strings.TrimSpace(*inputPath) == "" {
		fmt.Fprintln(os.Stderr, "usage: schema-validate -schema <schema.json> -input <document.json>")
		os.Exit(2)
	}

	schemaAbs, err := filepath.Abs(*schemaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema path error: %v\n", err)
		os.Exit(2)
	}
	inputAbs, err := filepath.Abs(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "input path error: %v\n", err)
		os.Exit(2)
	}

	schemaDoc, err := loadJSON(schemaAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema load error: %v\n", err)
		os.Exit(1)
	}
	inputDoc, err := loadJSON(inputAbs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "input load error: %v\n", err)
		os.Exit(1)
	}

	compiler := jsonschema.NewCompiler()
	schemaURI := fileURI(schemaAbs)
	if err := compiler.AddResource(schemaURI, schemaDoc); err != nil {
		fmt.Fprintf(os.Stderr, "schema compile setup error: %v\n", err)
		os.Exit(1)
	}
	schema, err := compiler.Compile(schemaURI)
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema compile error: %v\n", err)
		os.Exit(1)
	}
	if err := schema.Validate(inputDoc); err != nil {
		fmt.Fprintf(os.Stderr, "validation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("schema validation passed: %s <- %s\n", inputAbs, schemaAbs)
}

func loadJSON(path string) (any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return jsonschema.UnmarshalJSON(f)
}

func fileURI(path string) string {
	return "file:///" + strings.TrimPrefix(filepath.ToSlash(path), "/")
}
