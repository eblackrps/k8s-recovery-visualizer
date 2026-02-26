package output

import (
	"html/template"
	"os"

	"k8s-recovery-visualizer/internal/model"
)

func WriteHTML(path string, b *model.Bundle) error {

	tpl := `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>DR Assessment Report</title>
<style>
body { font-family: Arial; margin: 40px; }
h1 { color: #333; }
.score { font-size: 24px; font-weight: bold; }
.domain { margin: 10px 0; }
table { border-collapse: collapse; width: 100%; margin-top: 20px; }
th, td { border: 1px solid #ddd; padding: 8px; }
th { background-color: #f2f2f2; }
</style>
</head>
<body>

<h1>DR Assessment Report</h1>

<div class="score">
Overall Score: {{.Score.Overall.Final}} / {{.Score.Overall.Max}}
</div>

<h2>Maturity Level: {{.Score.Maturity}}</h2>

<h3>Domain Scores</h3>
<div class="domain">Storage: {{.Score.Storage.Final}} / {{.Score.Storage.Max}}</div>
<div class="domain">Workload: {{.Score.Workload.Final}} / {{.Score.Workload.Max}}</div>
<div class="domain">Config: {{.Score.Config.Final}} / {{.Score.Config.Max}}</div>

<h3>Findings</h3>

{{if .Inventory.Findings}}
<table>
<tr>
<th>Severity</th>
<th>Resource</th>
<th>Issue</th>
<th>Recommendation</th>
</tr>
{{range .Inventory.Findings}}
<tr>
<td>{{.Severity}}</td>
<td>{{.ResourceID}}</td>
<td>{{.Message}}</td>
<td>{{.Recommendation}}</td>
</tr>
{{end}}
</table>
{{else}}
<p>No critical DR findings detected.</p>
{{end}}

</body>
</html>
`

	t, err := template.New("report").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, b)
}
