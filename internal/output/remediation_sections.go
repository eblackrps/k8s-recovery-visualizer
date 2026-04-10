package output

import (
	"bytes"
	"fmt"
	"html"
	"strings"

	"k8s-recovery-visualizer/internal/model"
)

func remediationBodyHTML(step model.RemediationStep) string {
	return renderRemediationBody(step, true)
}

func remediationBodyPrintHTML(step model.RemediationStep) string {
	return renderRemediationBody(step, false)
}

func renderRemediationBody(step model.RemediationStep, dark bool) string {
	var buf bytes.Buffer
	e := html.EscapeString
	writeSection := func(title string, items []string) {
		if len(items) == 0 {
			return
		}
		buf.WriteString(`<div style="margin-top:10px">`)
		buf.WriteString(`<strong style="display:block;margin-bottom:4px">`)
		buf.WriteString(e(title))
		buf.WriteString(`</strong><ul style="margin:0;padding-left:18px">`)
		for _, item := range items {
			buf.WriteString(`<li style="margin:2px 0">`)
			buf.WriteString(e(item))
			buf.WriteString(`</li>`)
		}
		buf.WriteString(`</ul></div>`)
	}

	buf.WriteString(`<p>`)
	buf.WriteString(e(step.Detail))
	buf.WriteString(`</p>`)

	if step.WhyItMatters != "" {
		buf.WriteString(`<div style="margin-top:10px"><strong style="display:block;margin-bottom:4px">Why It Matters</strong><p>`)
		buf.WriteString(e(step.WhyItMatters))
		buf.WriteString(`</p></div>`)
	}
	if step.DRImpact != "" {
		buf.WriteString(`<div style="margin-top:10px"><strong style="display:block;margin-bottom:4px">DR Impact</strong><p>`)
		buf.WriteString(e(step.DRImpact))
		buf.WriteString(`</p></div>`)
	}

	writeSection("Validate", step.Validation)
	writeSection("Fix Steps", step.FixSteps)

	if step.TargetNotes != "" {
		noteClass := `note`
		if !dark {
			noteClass = `note`
		}
		buf.WriteString(fmt.Sprintf(`<div class="%s">%s</div>`, noteClass, e(step.TargetNotes)))
	}
	if len(step.Commands) > 0 {
		buf.WriteString(`<div style="margin-top:10px"><strong style="display:block;margin-bottom:4px">Example Commands</strong><pre>`)
		buf.WriteString(e(strings.Join(step.Commands, "\n")))
		buf.WriteString(`</pre></div>`)
	}

	writeSection("Caveats", step.Caveats)

	return buf.String()
}
