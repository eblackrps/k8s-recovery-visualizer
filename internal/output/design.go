package output

import (
	"fmt"

	"k8s-recovery-visualizer/internal/theme"
)

func reportDocumentStart(title, bodyClass, extraCSS string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html lang="en"><head>
<meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>%s</title>
<style>%s%s</style></head><body class="%s">
<div class="page-glow page-glow-left"></div>
<div class="page-glow page-glow-right"></div>
<main class="layout">`, title, sharedOutputCSS(), extraCSS, bodyClass)
}

func reportDocumentEnd() string {
	return `</body></html>`
}

func maturityAccent(maturity string) string {
	return theme.MaturityColor(maturity)
}

func scoreAccent(score int) string {
	return theme.ScoreColor(score)
}

func sharedOutputCSS() string {
	return theme.Default().CSSVariables() + `
*{box-sizing:border-box}

html,body{
  margin:0;
  min-height:100%;
  background:
    radial-gradient(circle at top left, rgba(88,166,255,0.16), transparent 32%),
    radial-gradient(circle at top right, rgba(126,231,135,0.08), transparent 24%),
    linear-gradient(180deg, rgba(13,17,23,0.92) 0%, var(--bg) 32%, var(--bg-deep) 100%);
  color:var(--text);
  font-family:var(--body);
}

body{
  position:relative;
  font-size:14px;
  line-height:1.6;
}

a{
  color:var(--accent-soft);
  text-decoration:none;
}

a:hover,
a:focus-visible{
  text-decoration:underline;
}

button,
input,
select,
textarea{
  font:inherit;
}

button{
  color:inherit;
}

:focus-visible{
  outline:3px solid var(--accent-strong);
  outline-offset:2px;
}

h1,h2,h3{
  margin:0;
  color:var(--text);
  font-family:var(--title);
  letter-spacing:-0.02em;
}

h1{
  font-size:clamp(2rem,3vw,3.35rem);
  line-height:1.03;
}

h2{
  font-size:1.18rem;
  line-height:1.2;
}

h3{
  font-size:1rem;
  line-height:1.28;
}

p{
  margin:0;
}

code{
  font-family:var(--mono);
  font-size:0.92em;
}

pre{
  margin-top:0.8rem;
  padding:0.95rem 1rem;
  border-radius:18px;
  border:1px solid rgba(48,54,61,0.55);
  background:rgba(1,4,9,0.84);
  color:var(--text);
  overflow-x:auto;
  white-space:pre-wrap;
  font-family:var(--mono);
  font-size:0.83rem;
  line-height:1.55;
}

table{
  width:100%;
  margin-top:0.75rem;
  border-collapse:separate;
  border-spacing:0;
  table-layout:auto;
  border:1px solid var(--line);
  border-radius:20px;
  overflow:hidden;
  background:var(--panel);
}

thead{
  background:rgba(255,255,255,0.03);
}

th{
  padding:0.82rem 0.95rem;
  color:var(--muted-strong);
  text-align:left;
  font-size:0.74rem;
  font-weight:700;
  letter-spacing:0.08em;
  text-transform:uppercase;
  border-bottom:1px solid var(--line);
  cursor:pointer;
  user-select:none;
  white-space:nowrap;
}

th:hover,
th:focus-visible{
  color:var(--text);
}

th.asc::after{content:" ↑";color:var(--accent-soft)}
th.desc::after{content:" ↓";color:var(--accent-soft)}

td{
  padding:0.82rem 0.95rem;
  color:var(--text);
  border-bottom:1px solid var(--line-soft);
  vertical-align:top;
  word-break:normal;
  overflow-wrap:anywhere;
}

tbody tr:last-child td{
  border-bottom:none;
}

tbody tr:hover td,
tbody tr:focus-within td{
  background:rgba(255,255,255,0.03);
}

.page-glow{
  position:fixed;
  width:30rem;
  height:30rem;
  border-radius:999px;
  filter:blur(120px);
  opacity:0.42;
  pointer-events:none;
  z-index:0;
}

.page-glow-left{
  top:-10rem;
  left:-6rem;
  background:rgba(88,166,255,0.16);
}

.page-glow-right{
  top:9rem;
  right:-8rem;
  background:rgba(126,231,135,0.12);
}

.layout{
  position:relative;
  z-index:1;
  width:min(1480px, calc(100% - 2rem));
  margin:0 auto;
  padding:1.2rem 0 3rem;
}

.eyebrow,
.section-tag{
  display:inline-flex;
  align-items:center;
  gap:0.4rem;
  color:var(--accent-soft);
  font-size:0.72rem;
  font-weight:700;
  letter-spacing:0.18em;
  text-transform:uppercase;
}

.hero,
.card,
.step,
.toc{
  backdrop-filter:blur(18px);
  -webkit-backdrop-filter:blur(18px);
}

.hero{
  position:relative;
  display:grid;
  grid-template-columns:minmax(0,1.35fr) minmax(300px,0.9fr);
  gap:1.35rem;
  padding:1.6rem;
  border:1px solid var(--line);
  border-radius:var(--radius-xl);
  background:linear-gradient(145deg, var(--surface-raised), var(--panel));
  box-shadow:var(--shadow);
  overflow:hidden;
}

.hero::after{
  content:"";
  position:absolute;
  right:-7rem;
  bottom:-7rem;
  width:20rem;
  height:20rem;
  border-radius:999px;
  background:radial-gradient(circle, var(--glow), transparent 70%);
  pointer-events:none;
}

.hero-copy p{
  max-width:56rem;
  margin-top:0;
  color:var(--muted);
  font-size:0.98rem;
}

.hero-copy{
  display:flex;
  flex-direction:column;
  gap:0.9rem;
}

.hero-meta-grid{
  display:flex;
  flex-wrap:wrap;
  gap:0.65rem;
  margin-top:0;
}

.hero-brief-grid{
  display:grid;
  grid-template-columns:repeat(3, minmax(0, 1fr));
  gap:0.75rem;
}

.hero-brief{
  padding:0.9rem 0.95rem;
  border-radius:20px;
  border:1px solid var(--line-soft);
  background:rgba(255,255,255,0.03);
}

.hero-brief-label{
  color:var(--muted);
  font-size:0.7rem;
  font-weight:700;
  letter-spacing:0.14em;
  text-transform:uppercase;
}

.hero-brief strong{
  display:block;
  margin-top:0.35rem;
  color:var(--text);
  font-size:1rem;
}

.hero-brief p{
  max-width:none;
  margin-top:0.42rem;
  font-size:0.84rem;
}

.meta-chip{
  display:inline-flex;
  align-items:center;
  gap:0.45rem;
  padding:0.48rem 0.82rem;
  border-radius:999px;
  border:1px solid var(--accent-strong);
  background:rgba(255,255,255,0.04);
  color:var(--muted-strong);
  font-size:0.84rem;
}

.meta-chip strong{
  font-weight:700;
}

.hero-panel{
  display:flex;
  flex-direction:column;
  gap:0.95rem;
  padding:1.2rem;
  border-radius:24px;
  border:1px solid var(--accent-strong);
  background:linear-gradient(180deg, var(--surface-raised), var(--panel-strong));
  box-shadow:inset 0 1px 0 rgba(255,255,255,0.04);
}

.hero-pill{
  display:inline-flex;
  align-items:center;
  justify-content:center;
  width:fit-content;
  padding:0.5rem 0.88rem;
  border-radius:999px;
  border:1px solid var(--accent-strong);
  background:var(--accent-faint);
  color:var(--text);
  font-size:0.84rem;
  font-weight:700;
}

.hero-score{
  display:flex;
  flex-wrap:wrap;
  align-items:flex-end;
  gap:0.9rem;
}

.hero-score-value{
  font-size:clamp(2.6rem,5vw,4rem);
  font-weight:800;
  line-height:0.92;
}

.hero-score-label{
  color:var(--muted);
  font-size:0.75rem;
  font-weight:700;
  letter-spacing:0.16em;
  text-transform:uppercase;
}

.hero-stat-grid{
  display:grid;
  grid-template-columns:repeat(2, minmax(0,1fr));
  gap:0.7rem;
}

.hero-stat{
  padding:0.82rem 0.9rem;
  border-radius:18px;
  border:1px solid var(--line-soft);
  background:rgba(255,255,255,0.04);
}

.hero-stat span{
  display:block;
  color:var(--muted);
  font-size:0.72rem;
  font-weight:700;
  letter-spacing:0.1em;
  text-transform:uppercase;
}

.hero-stat strong{
  display:block;
  margin-top:0.34rem;
  font-size:1.08rem;
  color:var(--text);
}

.badge-row{
  display:flex;
  flex-wrap:wrap;
  gap:0.55rem;
}

.badge{
  display:inline-flex;
  align-items:center;
  gap:0.35rem;
  padding:0.4rem 0.86rem;
  border-radius:999px;
  border:1px solid currentColor;
  color:var(--text);
  font-size:0.8rem;
  font-weight:700;
}

.badge-subtle{
  border-color:var(--line-soft);
  background:rgba(255,255,255,0.03);
  color:var(--muted-strong);
}

.card{
  margin-bottom:1rem;
  padding:1.15rem 1.2rem;
  border-radius:var(--radius-lg);
  border:1px solid var(--line);
  background:linear-gradient(180deg, var(--surface-raised), var(--panel));
  box-shadow:var(--shadow-soft);
}

.card > h2:first-child,
.card > h3:first-child{
  margin-top:0;
}

.card > p + table,
.card > div + table,
.card > h3 + table{
  margin-top:0.9rem;
}

.grid{
  display:grid;
  grid-template-columns:repeat(auto-fit, minmax(180px,1fr));
  gap:0.9rem;
  margin:1rem 0;
}

.sbox{
  position:relative;
  min-height:116px;
  padding:1rem;
  border-radius:22px;
  border:1px solid var(--line-soft);
  background:linear-gradient(180deg, var(--surface-raised), var(--panel-strong));
  box-shadow:inset 0 1px 0 rgba(255,255,255,0.04), 0 20px 44px rgba(3,2,10,0.18);
  overflow:hidden;
}

.sbox::before{
  content:"";
  position:absolute;
  inset:0 auto auto 0;
  width:100%;
  height:1px;
  opacity:0.45;
  background:linear-gradient(90deg, rgba(255,255,255,0.24), transparent 68%);
}

.sbox .v{
  color:var(--text);
  font-size:2rem;
  font-weight:800;
  line-height:1;
}

.sbox .l{
  margin-top:0.38rem;
  color:var(--muted);
  font-size:0.74rem;
  font-weight:700;
  letter-spacing:0.08em;
  text-transform:uppercase;
}

.sbox .bar{
  margin-top:0.85rem;
  height:7px;
  border-radius:999px;
  background:rgba(255,255,255,0.06);
  overflow:hidden;
}

.sbox .fill{
  height:100%;
  border-radius:999px;
  background:linear-gradient(90deg, var(--accent-strong), var(--accent-soft));
}

.empty{
  padding:0.45rem 0 0;
  color:var(--muted);
  font-style:italic;
}

.chip{
  display:inline-flex;
  align-items:center;
  gap:0.35rem;
  padding:0.28rem 0.72rem;
  border-radius:999px;
  border:1px solid transparent;
  background:rgba(255,255,255,0.05);
  color:var(--muted-strong);
  font-size:0.76rem;
  font-weight:700;
  letter-spacing:0.02em;
}

.chip.p{
  border-color:var(--success);
  background:var(--success-faint);
  color:var(--success);
}

.chip.f{
  border-color:var(--danger);
  background:var(--danger-faint);
  color:var(--danger);
}

.chip.w{
  border-color:var(--warning);
  background:var(--warning-medium-faint);
  color:var(--warning);
}

.chip.n{
  border-color:var(--line-soft);
  background:rgba(255,255,255,0.04);
  color:var(--muted);
}

.pub{
  border-color:var(--danger);
  background:var(--danger-faint);
  color:var(--danger);
}

.prv{
  border-color:var(--success);
  background:var(--success-faint);
  color:var(--success);
}

.c-CRITICAL,.sev-CRITICAL{
  color:var(--danger);
  font-weight:700;
}

.c-HIGH,.sev-HIGH{
  color:var(--warning-high);
  font-weight:700;
}

.c-MEDIUM,.sev-MEDIUM{
  color:var(--warning);
  font-weight:600;
}

.c-LOW,.c-INFO,.sev-LOW,.sev-INFO{
  color:var(--muted);
}

.ok{color:var(--success)}
.bad{color:var(--danger)}

.note{
  margin-top:0.8rem;
  padding:0.8rem 0.95rem;
  border-radius:18px;
  border:1px solid var(--accent-strong);
  background:var(--accent-faint);
  color:var(--muted-strong);
}

.page-toolbar{
  display:flex;
  justify-content:flex-end;
  gap:0.75rem;
  margin-bottom:1rem;
}

.btn-sm,
.print-btn{
  display:inline-flex;
  align-items:center;
  justify-content:center;
  gap:0.45rem;
  padding:0.72rem 1rem;
  border-radius:999px;
  border:1px solid var(--accent-strong);
  background:rgba(255,255,255,0.04);
  color:var(--text);
  cursor:pointer;
  text-decoration:none;
  transition:transform 160ms ease, border-color 160ms ease, background 160ms ease, color 160ms ease;
}

.btn-sm:hover,
.btn-sm:focus-visible,
.print-btn:hover,
.print-btn:focus-visible{
  transform:translateY(-1px);
  border-color:var(--accent);
  background:var(--accent-faint);
  text-decoration:none;
}

.filter-bar{
  display:flex;
  flex-wrap:wrap;
  align-items:center;
  gap:0.55rem;
}

.filter-bar span{
  color:var(--muted);
  font-size:0.82rem;
  font-weight:700;
}

.fbtn{
  padding:0.42rem 0.9rem;
  border-radius:999px;
  border:1px solid var(--line-soft);
  background:rgba(255,255,255,0.04);
  color:var(--muted);
  cursor:pointer;
  font-size:0.78rem;
  font-weight:700;
}

.fbtn:hover,
.fbtn:focus-visible{
  border-color:var(--accent);
  color:var(--text);
}

.fbtn.active{
  background:var(--accent-faint);
  border-color:var(--accent);
  color:var(--text);
}

.fbtn.fc{border-color:var(--danger);color:var(--danger)}
.fbtn.fc.active{background:var(--danger-faint)}
.fbtn.fh{border-color:var(--warning-high);color:var(--warning-high)}
.fbtn.fh.active{background:var(--warning-high-faint)}
.fbtn.fm{border-color:var(--warning);color:var(--warning)}
.fbtn.fm.active{background:var(--warning-medium-faint)}

.step{
  margin-bottom:0.9rem;
  border:1px solid var(--line);
  border-radius:22px;
  background:linear-gradient(180deg, var(--surface-raised), var(--panel));
  overflow:hidden;
  page-break-inside:avoid;
}

.step-h{
  width:100%;
  padding:1rem 1.1rem;
  border:none;
  border-radius:0;
  background:linear-gradient(180deg, rgba(255,255,255,0.035), rgba(255,255,255,0.01));
  display:flex;
  align-items:center;
  gap:0.75rem;
  text-align:left;
  cursor:pointer;
}

.step-h:hover,
.step-h:focus-visible{
  background:rgba(255,255,255,0.05);
}

.step-b{
  display:none;
  padding:1rem 1.1rem 1.15rem;
  border-top:1px solid var(--line);
}

.step-b.open{
  display:block;
}

.rem-body{
  display:grid;
  gap:0.9rem;
}

.rem-detail{
  color:var(--muted-strong);
}

.rem-section{
  display:grid;
  gap:0.45rem;
  padding:0.85rem 0.95rem;
  border-radius:18px;
  border:1px solid var(--line-soft);
  background:rgba(255,255,255,0.03);
}

.rem-section-title{
  color:var(--accent-soft);
  font-size:0.76rem;
  font-weight:700;
  letter-spacing:0.12em;
  text-transform:uppercase;
}

.rem-list{
  margin:0;
  padding-left:1.15rem;
}

.rem-list li{
  margin:0.28rem 0;
}

.rem-note{
  margin-top:0;
}

.subtle{
  color:var(--muted);
}

.footer{
  margin-top:1.5rem;
  padding-top:0.95rem;
  border-top:1px solid var(--line);
  color:var(--muted);
  font-size:0.82rem;
}

.summary-grid{
  display:grid;
  grid-template-columns:minmax(0,1.15fr) minmax(280px,0.85fr);
  gap:1rem;
  margin-top:1rem;
}

.kv-grid{
  display:grid;
  grid-template-columns:repeat(auto-fit, minmax(220px, 1fr));
  gap:0.85rem;
  margin-top:0.9rem;
}

.kv-card{
  padding:0.95rem 1rem;
  border-radius:20px;
  border:1px solid var(--line-soft);
  background:rgba(255,255,255,0.03);
}

.kv-card h3{
  margin:0 0 0.7rem;
  color:var(--muted-strong);
  font-size:0.78rem;
  font-weight:700;
  letter-spacing:0.12em;
  text-transform:uppercase;
}

.kv-list{
  display:grid;
  gap:0.15rem;
}

.kv-row{
  display:grid;
  grid-template-columns:minmax(112px, 0.72fr) minmax(0, 1fr);
  gap:0.8rem;
  align-items:start;
  padding:0.52rem 0;
  border-top:1px solid var(--line-soft);
}

.kv-row:first-child{
  padding-top:0;
  border-top:none;
}

.kv-row:last-child{
  padding-bottom:0;
}

.kv-key{
  color:var(--muted);
  font-size:0.75rem;
  font-weight:700;
  letter-spacing:0.05em;
  text-transform:uppercase;
}

.kv-value{
  color:var(--text);
  font-size:0.9rem;
  line-height:1.5;
}

.kv-value .chip{
  vertical-align:middle;
}

.stack{
  display:grid;
  gap:1rem;
}

.toc{
  display:inline-block;
  margin:1rem 0 1.25rem;
  padding:1rem 1.1rem;
  border-radius:22px;
  border:1px solid var(--line);
  background:linear-gradient(180deg, var(--surface-raised), var(--panel));
}

.toc ol{
  margin:0.75rem 0 0;
  padding-left:1.2rem;
  color:var(--muted-strong);
  line-height:1.9;
}

.toc h3{
  margin:0;
}

@media (max-width: 1120px){
  .hero,
  .summary-grid{
    grid-template-columns:1fr;
  }

  .hero-brief-grid{
    grid-template-columns:repeat(auto-fit, minmax(180px, 1fr));
  }
}

@media (max-width: 760px){
  .layout{
    width:min(100% - 1rem, 1480px);
    padding-top:0.9rem;
  }

  .hero,
  .card,
  .toc{
    padding:1rem;
  }

  .grid,
  .hero-stat-grid,
  .kv-grid,
  .kv-row{
    grid-template-columns:1fr;
  }

  .page-toolbar{
    justify-content:stretch;
  }

  .page-toolbar > *{
    flex:1 1 auto;
  }

  table{
    display:block;
    overflow-x:auto;
  }
}

@media (prefers-reduced-motion: reduce){
  *,*::before,*::after{
    animation:none !important;
    transition:none !important;
    scroll-behavior:auto !important;
  }
}

@media print{
  :root{
    color-scheme:light;
  }

  html,body{
    background:#fff !important;
    color:#24292f !important;
  }

  body{
    font-size:11pt;
  }

  .page-glow,
  .page-toolbar{
    display:none !important;
  }

  .layout{
    width:100% !important;
    padding:0 !important;
  }

  .hero,
  .hero-panel,
  .card,
  .sbox,
  .step,
  .toc,
  .kv-card,
  table,
  pre{
    background:#fff !important;
    color:#24292f !important;
    border-color:#d0d7de !important;
    box-shadow:none !important;
    backdrop-filter:none !important;
    -webkit-backdrop-filter:none !important;
  }

  .hero-copy p,
  .hero-score-label,
  .meta-chip,
  .subtle,
  .footer,
  .filter-bar span,
  .sbox .l,
  .rem-section-title{
    color:#57606a !important;
  }

  a{
    color:#0969da !important;
  }

  .badge,
  .chip,
  .btn-sm,
  .print-btn,
  .fbtn{
    background:#fff !important;
    color:#24292f !important;
    border-color:#d0d7de !important;
  }

  th,
  td{
    color:#24292f !important;
    border-color:#d8dee4 !important;
  }

  th{
    background:#f6f8fa !important;
  }
}
`
}

func reportPageCSS() string {
	return `
body.report-page .layout{
  width:min(1560px, calc(100% - 1.75rem));
}

body.report-page .tabs{
  position:sticky;
  top:0.65rem;
  z-index:20;
  display:flex;
  gap:0.6rem;
  overflow-x:auto;
  margin-top:1rem;
  padding:0.55rem;
  border:1px solid var(--line);
  border-radius:999px;
  background:var(--panel-strong);
  box-shadow:var(--shadow-soft);
  backdrop-filter:blur(18px);
  -webkit-backdrop-filter:blur(18px);
}

body.report-page .tab{
  appearance:none;
  border:1px solid transparent;
  border-radius:999px;
  background:rgba(255,255,255,0.03);
  color:var(--muted);
  padding:0.78rem 1.02rem;
  cursor:pointer;
  white-space:nowrap;
  font-size:0.84rem;
  font-weight:700;
}

body.report-page .tab:hover,
body.report-page .tab:focus-visible{
  color:var(--text);
  border-color:var(--accent-strong);
}

body.report-page .tab.active{
  color:var(--text);
  background:linear-gradient(135deg, var(--accent-surface), var(--accent-faint));
  border-color:var(--accent);
  box-shadow:inset 0 1px 0 rgba(255,255,255,0.12);
}

body.report-page .pane{
  display:none;
  padding-top:1.35rem;
}

body.report-page .pane.active{
  display:block;
}

body.report-page .pane > h2{
  margin-bottom:1rem;
}

body.report-page .pane > h3{
  margin-top:1.35rem;
  margin-bottom:0.3rem;
}

body.report-page .rem-controls{
  display:flex;
  gap:0.65rem;
  flex-wrap:wrap;
  margin-bottom:1rem;
}

body.report-page .inline-link{
  border:none;
  background:none;
  padding:0;
  color:var(--accent-soft);
  font:inherit;
  font-weight:700;
  cursor:pointer;
}

body.report-page .inline-link:hover,
body.report-page .inline-link:focus-visible{
  text-decoration:underline;
}

@media print{
  body.report-page .tabs{
    position:static;
    display:none;
  }

  body.report-page .pane{
    display:block !important;
    padding-top:1rem;
  }
}
`
}

func summaryPageCSS() string {
	return `
body.summary-page .layout{
  width:min(1280px, calc(100% - 1.75rem));
}

body.summary-page .hero{
  margin-bottom:1rem;
}

body.summary-page .hero-copy p{
  max-width:48rem;
}

body.summary-page .summary-grid{
  grid-template-columns:minmax(0,1.3fr) minmax(320px,0.7fr);
}
`
}

func runbookPageCSS() string {
	return `
body.runbook-page .layout{
  width:min(1180px, calc(100% - 1.75rem));
}

body.runbook-page section.card{
  page-break-inside:avoid;
}

body.runbook-page .section-break{
  page-break-before:always;
}
`
}

func legacyPageCSS() string {
	return `
body.legacy-page .layout{
  width:min(1120px, calc(100% - 1.75rem));
}
`
}
