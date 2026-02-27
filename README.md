# k8s-recovery-visualizer

Enterprise Kubernetes Disaster Recovery Assessment Toolkit (PowerShell Edition)

---

## Overview

k8s-recovery-visualizer is a modular PowerShell-based assessment tool designed to:

- Inventory Kubernetes workloads
- Evaluate storage resilience
- Detect backup readiness
- Assess network DR viability
- Evaluate portability constraints
- Generate a unified DR viability report
- Track historical scoring trends

The tool produces JSON, Markdown, and HTML outputs suitable for internal engineering review or customer-facing deliverables.

---

## Architecture

Pipeline Flow:

Collect → Score → Merge → Render → (Optional History Update)

Core Modules:

- Collect-Workloads.ps1
- Score-Storage.ps1
- Detect-BackupReadiness.ps1
- Assess-NetworkReadiness.ps1
- Assess-Portability.ps1
- Build-DrReport.ps1
- Render-DrReport.ps1
- Invoke-DrScan.ps1

Optional:
- Update-History.ps1
- Build-Dist.ps1
- Build-Exe.ps1

---

## Requirements

- PowerShell 7+
- kubectl installed and configured
- Active kubeconfig context
- Read access to cluster resources

---

## Quick Start

Run full scan:

```powershell
pwsh -ExecutionPolicy Bypass -File .\Invoke-DrScan.ps1 -OutDir .\out