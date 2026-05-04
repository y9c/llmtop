package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

// Pre-created styles — created once at init, not per-frame.
var (
	styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00d4ff"))
	styleGray     = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	styleGrayPipe = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	styleTag      = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	styleFooter   = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	styleUtilChart = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80"))
	styleKVChart   = lipgloss.NewStyle().Foreground(lipgloss.Color("#c084fc"))
	styleDecChart  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00d4ff"))
	styleMemChart  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00"))

	styleEmpty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	styleLowVal   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Bold(true)
	styleMidVal   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00")).Bold(true)
	styleHighVal  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00dd66"))

	stylePctRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff3333")).Bold(true)
	stylePctOrange = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00")).Bold(true)
	stylePctYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffdd00")).Bold(true)
	stylePctGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00cc44")).Bold(true)

	styleValCyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("#38bdf8")).Bold(true)
	styleValOrange = lipgloss.NewStyle().Foreground(lipgloss.Color("#fb923c")).Bold(true)
	styleValGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true)
	styleValPurple = lipgloss.NewStyle().Foreground(lipgloss.Color("#c084fc")).Bold(true)
	styleValYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbbf24")).Bold(true)
	styleValRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f87171")).Bold(true)
	styleValTeal   = lipgloss.NewStyle().Foreground(lipgloss.Color("#2dd4bf")).Bold(true)
	styleValPink   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f472b6")).Bold(true)

	styleTH      = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ade80")).Bold(true)  // Throughput header
	styleSpec    = lipgloss.NewStyle().Foreground(lipgloss.Color("#c084fc")).Bold(true)  // Speculative header

	// Pre-rendered fixed strings — lipgloss.Render() allocates, so render once at init
	grayPipe    = styleGrayPipe.Render("│")
	grayCornerTL = styleGray.Render("┌")
	grayCornerBL = styleGray.Render("└")
	grayCornerTR = styleGray.Render("┐")
	grayCornerBR = styleGray.Render("┘")
	grayTeeT     = styleGray.Render("┬")
	grayTeeB     = styleGray.Render("┴")
	graySep      = styleGray.Render(" | ")
	grayPrefix   = styleGray.Render("─ ")

)

// Reusable buffer to avoid time.Now().Format() allocation per tick
var timeBuf = make([]byte, 0, 8)

// Pre-rendered tag strings — avoids per-frame styleTag.Render() allocation
var (
	tagRun   = styleTag.Render("run")
	tagWait  = styleTag.Render("wait")
	tagDec   = styleTag.Render("dec")
	tagPre   = styleTag.Render("pre")
	tagGen   = styleTag.Render("gen")
	tagPrm   = styleTag.Render("prm")
	tagUp    = styleTag.Render("up")
	tagTTFT  = styleTag.Render("ttft")
	tagTPOT  = styleTag.Render("tpot")
	tagAccept = styleTag.Render("accept")
	tagTD    = styleTag.Render("t/d")
	tagDraft = styleTag.Render("draft")
	tagRej   = styleTag.Render("rej")
	tagHit   = styleTag.Render("hit")
	tagQ     = styleTag.Render("q")
	tagCache = styleTag.Render("cache")
	tagCmp   = styleTag.Render("cmp")
	tagAcc   = styleTag.Render("acc")
)

func fmtNum(v float64) string {
	// Avoid fmt.Sprintf for the common case — use strconv on a reusable buffer
	buf := fmtBuf[:0]
	switch {
	case v >= 1e12:
		buf = strconv.AppendFloat(buf, v/1e12, 'f', 2, 64)
		buf = append(buf, 'T')
	case v >= 1e9:
		buf = strconv.AppendFloat(buf, v/1e9, 'f', 2, 64)
		buf = append(buf, 'B')
	case v >= 1e6:
		buf = strconv.AppendFloat(buf, v/1e6, 'f', 2, 64)
		buf = append(buf, 'M')
	case v >= 1e3:
		buf = strconv.AppendFloat(buf, v/1e3, 'f', 1, 64)
		buf = append(buf, 'K')
	default:
		buf = strconv.AppendFloat(buf, v, 'f', 0, 64)
	}
	return string(buf)
}

var fmtBuf = make([]byte, 0, 32)

// colorValInline formats v with the given width/precision, then colors it.
func colorValInline(v float64, width, dec int) string {
	var st lipgloss.Style
	switch {
	case v <= 0:
		st = styleEmpty
	case v < 15:
		st = styleLowVal
	case v < 35:
		st = styleMidVal
	default:
		st = styleHighVal
	}
	b := cvBuf[:0]
	if dec > 0 {
		b = strconv.AppendFloat(b, v, 'f', 1, 64)
	} else {
		b = strconv.AppendFloat(b, v, 'f', 0, 64)
	}
	for len(b) < width {
		// Prepend space for right-alignment (like fmt.Sprintf("%*d", width, v))
		b = append(b, 0)
		copy(b[1:], b)
		b[0] = ' '
	}
	return st.Render(string(b))
}

// colorPctInline renders a percentage value with color.
func colorPctInline(v float64) string {
	var st lipgloss.Style
	switch {
	case v >= 90:
		st = stylePctRed
	case v >= 70:
		st = stylePctOrange
	case v >= 40:
		st = stylePctYellow
	default:
		st = stylePctGreen
	}
	b := cvBuf[:0]
	b = strconv.AppendFloat(b, v, 'f', 1, 64)
	b = append(b, '%')
	return st.Render(string(b))
}

var cvBuf = make([]byte, 0, 16)

// Reusable buffers for string building in buildView
var titleBuf = make([]byte, 0, 128)
var footBuf = make([]byte, 0, 128)

type chartDef struct {
	name  string
	vals  []float64
	width int // total width including y-axis labels (~9 chars)
	style lipgloss.Style
}

// chartLines returns body lines of an asciigraph chart.
// Each line is padded/truncated to exactly def.width display characters.
func chartLines(def chartDef) []string {
	if len(def.vals) < 2 {
		return []string{"--", spaceStr(def.width), spaceStr(def.width), spaceStr(def.width), spaceStr(def.width)}
	}
	// Reserve 10 chars for y-axis labels (e.g. "  100.00 ┤")
	plotW := def.width - 10
	if plotW < 10 { plotW = 10 }
	g := asciigraph.Plot(def.vals, asciigraph.Height(5), asciigraph.Width(plotW))

	var out []string
	start := 0
	for i := 0; i < len(g); i++ {
		if g[i] == '\n' {
			out = append(out, renderChartRow(g[start:i], def))
			start = i + 1
		}
	}
	// Handle trailing data or empty rows
	if start < len(g) {
		out = append(out, renderChartRow(g[start:], def))
	}
	// Pad to fixed 5 body rows
	for len(out) < 5 {
		out = append(out, spaceStr(def.width))
	}
	return out[:5]
}

func renderChartRow(row string, def chartDef) string {
	rendered := def.style.Render(row)
	if w := lipgloss.Width(rendered); w < def.width {
		rendered += spaceStr(def.width - w)
	} else if w > def.width {
		rendered = truncateWidth(rendered, def.width)
	}
	return rendered
}

// emitChartBlock outputs one or more charts stacked vertically.
// If more than one chart, they are rendered side-by-side in a row.
func emitChartBlock(p func(string), defs []chartDef, innerW int) {
	if len(defs) == 0 {
		return
	}

	if len(defs) == 1 {
		for _, line := range chartLines(defs[0]) {
			p(iline(line, innerW))
		}
		return
	}

	// Multi-column: gather body lines for each chart
	allLines := make([][]string, len(defs))
	for i, def := range defs {
		allLines[i] = chartLines(def)
	}

	gap := "  "
	for r := 0; r < 5; r++ {
		var b strings.Builder
		for c, lines := range allLines {
			if c > 0 {
				b.WriteString(gap)
			}
			b.WriteString(lines[r])
		}
		p(iline(b.String(), innerW))
	}
}

func (m Model) buildView() string {
	s := m.Snap
	d := m.Delta
	w := m.Width
	if w <= 0 { w = 80 }
	if w > 88 { w = 88 }

	chr := 0.0
	if s.PrefixCacheQueries > 0 { chr = s.PrefixCacheHits / s.PrefixCacheQueries * 100 }
	draftAcceptPct := d.AcceptRate * 100
	if draftAcceptPct == 0 && s.SpecDraftToksTotal > 0 {
		draftAcceptPct = s.SpecAcceptedTotal / s.SpecDraftToksTotal * 100
	}
	rej := s.SpecDraftToksTotal - s.SpecAcceptedTotal
	accPerDraftBatch := 0.0
	if s.SpecDraftsTotal > 0 { accPerDraftBatch = s.SpecAcceptedTotal / s.SpecDraftsTotal }

	uptime := formatDuration(m.Uptime)
	innerW := w - 4

	var out strings.Builder
	p := func(s string) { out.WriteString(s); out.WriteString("\n") }

	// Render timestamp once, reuse buffer
	nowBuf := timeBuf[:0]
	nowBuf = time.Now().AppendFormat(nowBuf, "15:04:05")
	now := string(nowBuf)

	// Title: timestamp, backend, GPU name, temp/power
	tb := titleBuf[:0]
	tb = append(tb, now...)
	tb = append(tb, " llmtop ┃ "...)
	tb = append(tb, m.Backend...)
	tb = append(tb, " ┃ "...)
	if cnt := s.GPUCount(); cnt > 1 {
		tb = strconv.AppendInt(tb, int64(cnt), 10)
		tb = append(tb, []byte("×")...)
	}
	tb = append(tb, s.GPUName...)
	if s.GPUTempC > 0 {
		tb = append(tb, " ┃ "...)
		tb = strconv.AppendFloat(tb, s.GPUTempC, 'f', 0, 64)
		tb = append(tb, []byte("°C")...)
		if s.GPUPowerW > 0 {
			tb = append(tb, ' ')
			tb = strconv.AppendFloat(tb, s.GPUPowerW, 'f', 0, 64)
			tb = append(tb, 'W')
		}
	}
	p(styleTitle.Render(string(tb)))
	p(sepLine(w))

	// Charts box: 4 mini timelines (Util, KV, Dec, Mem)
	if w >= 80 {
		// 2×2 grid
		half := (innerW - 2) / 2 // gap=2 between columns
		defs1 := []chartDef{
			{"Util", m.UtilHist, half, styleUtilChart},
			{"KV", m.KVHist, half, styleKVChart},
		}
		defs2 := []chartDef{
			{"Dec", m.DecHist, half, styleDecChart},
			{"Mem", m.MemHist, half, styleMemChart},
		}
		var names []string
		for _, d := range defs1 {
			names = append(names, d.style.Render(d.name))
		}
		for _, d := range defs2 {
			names = append(names, d.style.Render(d.name))
		}
		p(hline(w, "", names...))
		emitChartBlock(p, defs1, innerW)
		p(iline("", innerW)) // blank line between chart rows
		emitChartBlock(p, defs2, innerW)
	} else {
		// Vertical stack: each chart on its own with a blank line in between
		defs := []chartDef{
			{"Util", m.UtilHist, innerW, styleUtilChart},
			{"KV", m.KVHist, innerW, styleKVChart},
			{"Dec", m.DecHist, innerW, styleDecChart},
			{"Mem", m.MemHist, innerW, styleMemChart},
		}
		var names []string
		for _, d := range defs {
			names = append(names, d.style.Render(d.name))
		}
		p(hline(w, "", names...))
		for i, def := range defs {
			if i > 0 {
				p(iline("", innerW)) // blank line between charts
			}
			emitChartBlock(p, []chartDef{def}, innerW)
		}
	}
	p(footerLine(w))

	// Two-column: Throughput + Speculative/Prefetch
	colW := (innerW - 3) * 2 / 5
	lW2 := colW
	rW2 := innerW - 3 - colW

	// Tags are package-level vars, pre-rendered at init — no per-frame Render calls

	var tpRows, spRows []string

	// --- Left column rows ---

	// Row 1: run/wait + uptime
	rr := colorValInline(s.RunningReqs, 3, 0)
	wr := colorValInline(s.WaitingReqs, 3, 0)
	tpRows = append(tpRows, tagRun+" "+rr+"  "+tagWait+" "+wr+"  "+tagUp+" "+uptime)

	// Row 2: dec/pre (bright cyan) — instant(rolling_avg)
	dt := styleValCyan.Render(fmt.Sprintf("%.1f(%.0f)", d.DecodeTokS, avgRecent(m.DecHist, 5)))
	pt := styleValCyan.Render(fmt.Sprintf("%.1f(%.0f)", d.PrefillTokS, avgRecent(m.PreHist, 5)))
	tpRows = append(tpRows, tagDec+" "+dt+"  "+tagPre+" "+pt)

	// Row 3: gen/prm totals
	tpRows = append(tpRows, tagGen+" "+fmtNum(s.GenTokensTotal)+"  "+
		tagPrm+" "+fmtNum(s.PromptTokensTotal))

	// Row 4: ttft/tpot + session latency (if available)
	if s.TTFTCount > 0 {
		row4 := ""
		avgTTFT := s.TTFTTotalS / s.TTFTCount * 1000
		avgTPOT := s.TPOTTotalS / s.TPOTCount * 1000
		ttft := styleValOrange.Render(fmt.Sprintf("%.0fms", avgTTFT))
		tpot := styleValGreen.Render(fmt.Sprintf("%.0fms", avgTPOT))
		row4 = tagTTFT + " " + ttft + "  " + tagTPOT + " " + tpot
		if m.Latency.TTFTAvgMs > 0 {
			ttftStr := styleValOrange.Render(fmt.Sprintf("%.0f/%.0f/%.0f", m.Latency.TTFTMinMs, m.Latency.TTFTAvgMs, m.Latency.TTFTMaxMs))
			tpotStr := styleValGreen.Render(fmt.Sprintf("%.0f/%.0f/%.0f", m.Latency.TPOTMinMs, m.Latency.TPOTAvgMs, m.Latency.TPOTMaxMs))
			row4 += "  ttft " + ttftStr + "  tpot " + tpotStr
		}
		tpRows = append(tpRows, row4)
	} else if m.Latency.TTFTAvgMs > 0 {
		ttftStr := styleValOrange.Render(fmt.Sprintf("%.0f/%.0f/%.0f", m.Latency.TTFTMinMs, m.Latency.TTFTAvgMs, m.Latency.TTFTMaxMs))
		tpotStr := styleValGreen.Render(fmt.Sprintf("%.0f/%.0f/%.0f", m.Latency.TPOTMinMs, m.Latency.TPOTAvgMs, m.Latency.TPOTMaxMs))
		tpRows = append(tpRows, "ttft "+ttftStr+"  tpot "+tpotStr)
	}

	// --- Right column rows ---

	// Right column: speculative decoding + prefix cache, packed 3 per row
	if s.SpecDraftsTotal > 0 {
		// Row 1: accept t/d draft
		spRows = append(spRows,
			tagAccept+" "+colorPctInline(draftAcceptPct)+"  "+
				tagTD+" "+styleValPurple.Render(fmt.Sprintf("%.2f", accPerDraftBatch))+"  "+
				tagDraft+" "+styleValYellow.Render(fmtNum(s.SpecDraftsTotal)))
		// Row 2: rej hit q
		if s.PrefixCacheQueries > 0 || s.PromptCachedTotal > 0 {
			spRows = append(spRows,
				tagRej+" "+styleValRed.Render(fmtNum(rej))+"  "+
					tagHit+" "+colorPctInline(chr)+"  "+
					tagQ+" "+styleValTeal.Render(fmtNum(s.PrefixCacheQueries)))
		} else {
			spRows = append(spRows,
				tagRej+" "+styleValRed.Render(fmtNum(rej)))
		}
		// Row 3+: acc positions (may be long, keep separate)
		if len(s.SpecAcceptedPos) > 0 {
			var posParts []string
			for _, v := range s.SpecAcceptedPos {
				if v > 0 {
					posParts = append(posParts, styleValPink.Render(fmtNum(v)))
				}
			}
			if len(posParts) > 0 {
				spRows = append(spRows, tagAcc+" "+strings.Join(posParts, " → "))
			}
		}
		// Row 4: cache cmp
		if s.PromptCachedTotal > 0 {
			spRows = append(spRows,
				tagCache+" "+styleValTeal.Render(fmtNum(s.PromptCachedTotal))+"  "+
					tagCmp+" "+styleValTeal.Render(fmtNum(s.PromptLocalTotal)))
		}
	} else {
		// No spec data — show prefix cache only
		if s.PrefixCacheQueries > 0 || s.PromptCachedTotal > 0 {
			spRows = append(spRows,
				tagHit+" "+colorPctInline(chr)+"  "+
					tagQ+" "+styleValTeal.Render(fmtNum(s.PrefixCacheQueries)))
			spRows = append(spRows,
				tagCache+" "+styleValTeal.Render(fmtNum(s.PromptCachedTotal))+"  "+
					tagCmp+" "+styleValTeal.Render(fmtNum(s.PromptLocalTotal)))
		}
	}

	nr := len(tpRows)
	if len(spRows) > nr { nr = len(spRows) }

	p(twoColTop("Throughput", "Speculative", lW2, rW2))
	for i := 0; i < nr; i++ {
		lt := ""; if i < len(tpRows) { lt = tpRows[i] }
		rt := ""; if i < len(spRows) { rt = spRows[i] }
		p(twoColLine(lt, rt, lW2, rW2))
	}
	p(twoColBot(lW2, rW2))

	return out.String()
}

func iline(content string, innerW int) string {
	vis := lipgloss.Width(content)
	pad := innerW - vis
	if pad < 0 { pad = 0 }
	return grayPipe + " " + content + spaceStr(pad) + " " + grayPipe
}

// Fixed-size cache for padding strings (max 256, covers all terminal widths).
// No sync needed — single-threaded render path.
var spaceCache [256]string

func spaceStr(n int) string {
	if n <= 0 {
		return ""
	}
	if n >= len(spaceCache) {
		return strings.Repeat(" ", n)
	}
	if spaceCache[n] == "" {
		spaceCache[n] = strings.Repeat(" ", n)
	}
	return spaceCache[n]
}

// Pre-rendered gray dash lines for common box widths — avoids per-frame styleGray.Render + strings.Repeat.
var grayDashCache [256]string

func grayDash(w int) string {
	if w <= 0 {
		return ""
	}
	if w < len(grayDashCache) {
		if grayDashCache[w] == "" {
			grayDashCache[w] = styleGray.Render(strings.Repeat("─", w))
		}
		return grayDashCache[w]
	}
	return styleGray.Render(strings.Repeat("─", w))
}

func twoColLine(left, right string, lW, rW int) string {
	l := truncateWidth(left, lW)
	r := truncateWidth(right, rW)
	return grayPipe + " " + l + " " + grayPipe + " " + r + " " + grayPipe
}

func hline(w int, prefix string, names ...string) string {
	var inner strings.Builder
	inner.WriteString(grayPrefix)
	inner.WriteString(prefix)
	for i, name := range names {
		if i > 0 {
			inner.WriteString(graySep)
		}
		inner.WriteString(name)
	}
	inner.WriteString(" ")
	innerStr := inner.String()
	innerVis := lipgloss.Width(innerStr)
	pad := w - 2 - innerVis
	if pad < 0 {
		pad = 0
	}
	return grayCornerTL + innerStr + grayDash(pad) + grayCornerTR
}

func footerLine(w int) string {
	return grayCornerBL + grayDash(w-2) + grayCornerBR
}

func sepLine(w int) string {
	return grayDash(w)
}

func twoColTop(lTitle, rTitle string, lW, rW int) string {
	// Leading ─ in gray, title in its own color, trailing ─ in gray
	lStyled := styleGray.Render("─ ") + styleTH.Render(lTitle) + " "
	rStyled := styleGray.Render("─ ") + styleSpec.Render(rTitle) + " "
	lp := lW + 2 - lipgloss.Width(lStyled)
	rp := rW + 2 - lipgloss.Width(rStyled)
	if lp < 0 { lp = 0 }
	if rp < 0 { rp = 0 }
	return grayCornerTL + lStyled + styleGray.Render(strings.Repeat("─", lp)) + grayTeeT + rStyled + styleGray.Render(strings.Repeat("─", rp)) + grayCornerTR
}

func twoColBot(lW, rW int) string {
	return grayCornerBL + grayDash(lW+2) + grayTeeB + grayDash(rW+2) + grayCornerBR
}

func truncateWidth(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis <= w { return s + strings.Repeat(" ", w-vis) }
	var out strings.Builder
	plain := 0
	inANSI := false
	for _, r := range s {
		if r == '\x1b' { inANSI = true }
		if inANSI {
			out.WriteRune(r)
			if r == 'm' { inANSI = false }
			continue
		}
		if plain >= w {
			break
		}
		out.WriteRune(r)
		plain++
	}
	out.WriteString("\x1b[0m")
	return out.String()
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	b := durBuf[:0]
	if h > 0 {
		b = strconv.AppendInt(b, int64(h), 10)
		b = append(b, "h "[:2]...)
		if m < 10 {
			b = append(b, '0')
		}
		b = strconv.AppendInt(b, int64(m), 10)
		b = append(b, 'm')
	} else {
		b = strconv.AppendInt(b, int64(m), 10)
		b = append(b, "m "[:2]...)
		s := int(d.Seconds()) % 60
		if s < 10 {
			b = append(b, '0')
		}
		b = strconv.AppendInt(b, int64(s), 10)
		b = append(b, 's')
	}
	return string(b)
}

var durBuf = make([]byte, 0, 16)
var latBuf = make([]byte, 0, 32)

func avgRecent(hist []float64, n int) float64 {
	sum := 0.0
	count := 0
	for i := len(hist) - 1; i >= 0 && count < n; i-- {
		sum += hist[i]
		count++
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func formatLatency(ms float64) string {
	b := latBuf[:0]
	b = strconv.AppendFloat(b, ms, 'f', 0, 64)
	b = append(b, "ms"...)
	return styleTag.Render(string(b))
}

