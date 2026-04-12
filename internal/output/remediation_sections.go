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
		buf.WriteString(`<section class="rem-section">`)
		buf.WriteString(`<div class="rem-section-title">`)
		buf.WriteString(e(title))
		buf.WriteString(`</div><ul class="rem-list">`)
		for _, item := range items {
			buf.WriteString(`<li>`)
			buf.WriteString(e(item))
			buf.WriteString(`</li>`)
		}
		buf.WriteString(`</ul></section>`)
	}

	buf.WriteString(`<div class="rem-body">`)
	if step.Detail != "" {
		buf.WriteString(`<p class="rem-detail">`)
		buf.WriteString(e(step.Detail))
		buf.WriteString(`</p>`)
	}

	if step.WhyItMatters != "" {
		buf.WriteString(`<section class="rem-section"><div class="rem-section-title">Why It Matters</div><p>`)
		buf.WriteString(e(step.WhyItMatters))
		buf.WriteString(`</p></section>`)
	}
	if step.DRImpact != "" {
		buf.WriteString(`<section class="rem-section"><div class="rem-section-title">DR Impact</div><p>`)
		buf.WriteString(e(step.DRImpact))
		buf.WriteString(`</p></section>`)
	}

	writeSection("Validate", step.Validation)
	writeSection("Fix Steps", step.FixSteps)

	if step.TargetNotes != "" {
		noteClass := `note rem-note`
		if !dark {
			noteClass = `note rem-note`
		}
		buf.WriteString(fmt.Sprintf(`<div class="%s">%s</div>`, noteClass, e(step.TargetNotes)))
	}
	if len(step.Commands) > 0 {
		buf.WriteString(`<section class="rem-section"><div class="rem-section-title">Example Commands</div><pre>`)
		buf.WriteString(e(strings.Join(step.Commands, "\n")))
		buf.WriteString(`</pre></section>`)
	}

	writeSection("Caveats", step.Caveats)

	buf.WriteString(`</div>`)
	return buf.String()
}
