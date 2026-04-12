package theme

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Palette struct {
	Background string `json:"background"`
	Surface    string `json:"surface"`
	Border     string `json:"border"`
	Text       string `json:"text"`
	Muted      string `json:"muted"`
	Accent     string `json:"accent"`
	Success    string `json:"success"`
	Critical   string `json:"critical"`
	High       string `json:"high"`
	Medium     string `json:"medium"`
}

type MaturityPalette struct {
	Platinum string `json:"platinum"`
	Gold     string `json:"gold"`
	Silver   string `json:"silver"`
	Bronze   string `json:"bronze"`
}

type Typography struct {
	Body  string `json:"body"`
	Title string `json:"title"`
	Mono  string `json:"mono"`
}

type Radius struct {
	XL string `json:"xl"`
	LG string `json:"lg"`
	MD string `json:"md"`
	SM string `json:"sm"`
}

type Tokens struct {
	Palette    Palette         `json:"palette"`
	Maturity   MaturityPalette `json:"maturity"`
	Typography Typography      `json:"typography"`
	Radius     Radius          `json:"radius"`
}

var defaultTokens = Tokens{
	Palette: Palette{
		Background: "#0d1117",
		Surface:    "#161b22",
		Border:     "#30363d",
		Text:       "#c9d1d9",
		Muted:      "#8b949e",
		Accent:     "#58a6ff",
		Success:    "#7ee787",
		Critical:   "#f85149",
		High:       "#ffa657",
		Medium:     "#f2cc60",
	},
	Maturity: MaturityPalette{
		Platinum: "#79c0ff",
		Gold:     "#f2cc60",
		Silver:   "#c9d1d9",
		Bronze:   "#ffa657",
	},
	Typography: Typography{
		Body:  `"Segoe UI Variable Text","Segoe UI","Trebuchet MS",sans-serif`,
		Title: `"Segoe UI Variable Display","Segoe UI","Trebuchet MS",sans-serif`,
		Mono:  `"Consolas","SFMono-Regular","Liberation Mono",monospace`,
	},
	Radius: Radius{
		XL: "30px",
		LG: "24px",
		MD: "18px",
		SM: "14px",
	},
}

func Default() Tokens {
	return defaultTokens
}

func (t Tokens) Marshal() []byte {
	raw, _ := json.Marshal(t)
	return raw
}

func (t Tokens) CSSVariables() string {
	return fmt.Sprintf(`
:root{
  color-scheme:dark;
  --bg:%s;
  --surface:%s;
  --surface-raised:%s;
  --panel:%s;
  --panel-strong:%s;
  --line:%s;
  --line-soft:%s;
  --text:%s;
  --muted:%s;
  --accent:%s;
  --accent-soft:%s;
  --accent-strong:%s;
  --success:%s;
  --danger:%s;
  --warning-high:%s;
  --warning-medium:%s;
  --maturity-platinum:%s;
  --maturity-gold:%s;
  --maturity-silver:%s;
  --maturity-bronze:%s;
  --radius-xl:%s;
  --radius-lg:%s;
  --radius-md:%s;
  --radius-sm:%s;
  --shadow:0 28px 60px rgba(1,4,9,0.45);
  --shadow-soft:0 18px 38px rgba(1,4,9,0.3);
  --mono:%s;
  --body:%s;
  --title:%s;
}
`,
		t.Palette.Background,
		t.Palette.Surface,
		WithAlpha(t.Palette.Surface, 0.94),
		WithAlpha(t.Palette.Surface, 0.88),
		WithAlpha(t.Palette.Surface, 0.96),
		t.Palette.Border,
		WithAlpha(t.Palette.Border, 0.45),
		t.Palette.Text,
		t.Palette.Muted,
		t.Palette.Accent,
		WithAlpha(t.Palette.Accent, 0.82),
		WithAlpha(t.Palette.Accent, 0.55),
		t.Palette.Success,
		t.Palette.Critical,
		t.Palette.High,
		t.Palette.Medium,
		t.Maturity.Platinum,
		t.Maturity.Gold,
		t.Maturity.Silver,
		t.Maturity.Bronze,
		t.Radius.XL,
		t.Radius.LG,
		t.Radius.MD,
		t.Radius.SM,
		t.Typography.Mono,
		t.Typography.Body,
		t.Typography.Title,
	)
}

func MaturityColor(maturity string) string {
	switch strings.ToUpper(strings.TrimSpace(maturity)) {
	case "PLATINUM":
		return defaultTokens.Maturity.Platinum
	case "GOLD":
		return defaultTokens.Maturity.Gold
	case "SILVER":
		return defaultTokens.Maturity.Silver
	case "BRONZE":
		return defaultTokens.Maturity.Bronze
	default:
		return defaultTokens.Maturity.Silver
	}
}

func ScoreColor(score int) string {
	switch {
	case score < 50:
		return defaultTokens.Palette.Critical
	case score < 75:
		return defaultTokens.Palette.High
	default:
		return defaultTokens.Palette.Success
	}
}

func SeverityColor(severity string) string {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case "CRITICAL":
		return defaultTokens.Palette.Critical
	case "HIGH":
		return defaultTokens.Palette.High
	case "MEDIUM":
		return defaultTokens.Palette.Medium
	case "LOW", "INFO":
		return defaultTokens.Palette.Muted
	default:
		return defaultTokens.Palette.Muted
	}
}

func StatusColor(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "PASS", "PASSED", "SUCCESS", "CONFIRMED", "VERIFIED":
		return defaultTokens.Palette.Success
	case "WARN", "WARNING", "UNVERIFIED", "PARTIAL", "DEGRADED":
		return defaultTokens.Palette.Medium
	case "FAIL", "FAILED", "ERROR", "CRITICAL", "HIGH":
		return defaultTokens.Palette.Critical
	default:
		return defaultTokens.Palette.Muted
	}
}

func WithAlpha(hex string, alpha float64) string {
	trimmed := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(trimmed) != 6 {
		return hex
	}
	r, errR := strconv.ParseInt(trimmed[0:2], 16, 64)
	g, errG := strconv.ParseInt(trimmed[2:4], 16, 64)
	b, errB := strconv.ParseInt(trimmed[4:6], 16, 64)
	if errR != nil || errG != nil || errB != nil {
		return hex
	}
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.2f)", r, g, b, alpha)
}
