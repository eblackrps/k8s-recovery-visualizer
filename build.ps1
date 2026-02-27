<#
Build outputs:
  dist/
    scan.ps1            # single-file script, profile embedded
    scan.exe            # optional, if -Exe and ps2exe available
    SHA256SUMS.txt
    build-info.json
#>

[CmdletBinding()]
param(
  [string] $ProfilePath = ".\profiles\default.json",
  [string] $EntryScript = ".\scan.ps1",
  [string] $Src