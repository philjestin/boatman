package triage

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

// PDFReport generates a PDF report from triage result data.
type PDFReport struct {
	pdf *fpdf.Fpdf
}

// triageResultData is the shape of the triage result stored in ModeConfig.
// We use map[string]interface{} since that's what JSON unmarshaling produces.
type triageResultData struct {
	Tickets         []map[string]interface{} `json:"tickets"`
	Classifications []map[string]interface{} `json:"classifications"`
	Clusters        []map[string]interface{} `json:"clusters"`
	ContextDocs     []map[string]interface{} `json:"contextDocs"`
	Stats           map[string]interface{}   `json:"stats"`
}

// GeneratePDF creates a PDF report from a triage result map and writes it to outputPath.
func GeneratePDF(result map[string]interface{}, outputPath string) error {
	r := &PDFReport{
		pdf: fpdf.New("P", "mm", "A4", ""),
	}

	r.pdf.SetAutoPageBreak(true, 15)

	// Parse the data
	stats := asMap(result["stats"])
	classifications := asList(result["classifications"])
	clusters := asList(result["clusters"])
	tickets := asList(result["tickets"])
	contextDocs := asList(result["contextDocs"])

	// Build ticket lookup
	ticketMap := make(map[string]map[string]interface{})
	for _, t := range tickets {
		if id, ok := t["ticketId"].(string); ok {
			ticketMap[id] = t
		}
	}

	// Build context doc lookup
	docMap := make(map[string]map[string]interface{})
	for _, d := range contextDocs {
		if id, ok := d["clusterId"].(string); ok {
			docMap[id] = d
		}
	}

	// Parse plan data (optional — only present if plans were generated).
	plans := asList(result["plans"])
	planStats := asMap(result["planStats"])

	r.titlePage(stats, planStats)
	r.statsPage(stats)
	r.classificationsTable(classifications, ticketMap)
	r.clusterPages(clusters, docMap)
	r.ticketDetailPages(classifications, ticketMap)

	if len(plans) > 0 {
		r.planPages(plans, ticketMap, planStats)
	}

	return r.pdf.OutputFileAndClose(outputPath)
}

func (r *PDFReport) titlePage(stats map[string]interface{}, planStats map[string]interface{}) {
	r.pdf.AddPage()

	// Title
	r.pdf.SetFont("Helvetica", "B", 28)
	r.pdf.Ln(40)
	r.pdf.CellFormat(0, 15, "Triage Report", "", 1, "C", false, 0, "")

	// Date
	r.pdf.SetFont("Helvetica", "", 14)
	r.pdf.SetTextColor(120, 120, 120)
	r.pdf.CellFormat(0, 10, time.Now().Format("January 2, 2006"), "", 1, "C", false, 0, "")

	// Summary stats
	r.pdf.Ln(20)
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.SetFont("Helvetica", "", 12)

	total := intVal(stats["totalTickets"])
	aiDef := intVal(stats["aiDefiniteCount"])
	aiLikely := intVal(stats["aiLikelyCount"])
	humanReview := intVal(stats["humanReviewCount"])
	humanOnly := intVal(stats["humanOnlyCount"])
	clusterCount := intVal(stats["clusterCount"])

	lines := []string{
		fmt.Sprintf("%d tickets analyzed", total),
		fmt.Sprintf("%d clusters identified", clusterCount),
		"",
		fmt.Sprintf("AI Definite: %d", aiDef),
		fmt.Sprintf("AI Likely: %d", aiLikely),
		fmt.Sprintf("Human Review: %d", humanReview),
		fmt.Sprintf("Human Only: %d", humanOnly),
	}
	for _, line := range lines {
		r.pdf.CellFormat(0, 8, line, "", 1, "C", false, 0, "")
	}

	// Plan stats (if available).
	planTotal := intVal(planStats["total"])
	if planTotal > 0 {
		planPassed := intVal(planStats["passed"])
		planFailed := intVal(planStats["failed"])
		r.pdf.Ln(5)
		r.pdf.CellFormat(0, 8, "", "", 1, "C", false, 0, "")
		r.pdf.CellFormat(0, 8, fmt.Sprintf("Plans: %d generated (%d passed, %d failed)", planTotal, planPassed, planFailed), "", 1, "C", false, 0, "")
	}

	// Cost
	tokens := intVal(stats["totalTokensUsed"])
	cost := floatVal(stats["totalCostUsd"])
	if tokens > 0 {
		r.pdf.Ln(10)
		r.pdf.SetFont("Helvetica", "", 10)
		r.pdf.SetTextColor(120, 120, 120)
		r.pdf.CellFormat(0, 8, fmt.Sprintf("%.1fK tokens | $%.4f", float64(tokens)/1000, cost), "", 1, "C", false, 0, "")
	}
}

func (r *PDFReport) statsPage(stats map[string]interface{}) {
	r.pdf.AddPage()
	r.sectionHeader("Summary")

	aiDef := intVal(stats["aiDefiniteCount"])
	aiLikely := intVal(stats["aiLikelyCount"])
	humanReview := intVal(stats["humanReviewCount"])
	humanOnly := intVal(stats["humanOnlyCount"])
	total := aiDef + aiLikely + humanReview + humanOnly

	if total == 0 {
		r.pdf.SetFont("Helvetica", "", 11)
		r.pdf.CellFormat(0, 8, "No classifications produced.", "", 1, "", false, 0, "")
		return
	}

	categories := []struct {
		name  string
		count int
		r, g, b int
	}{
		{"AI Definite", aiDef, 34, 197, 94},
		{"AI Likely", aiLikely, 59, 130, 246},
		{"Human Review", humanReview, 234, 179, 8},
		{"Human Only", humanOnly, 239, 68, 68},
	}

	barWidth := 140.0
	for _, cat := range categories {
		r.pdf.SetFont("Helvetica", "", 11)
		r.pdf.CellFormat(40, 8, cat.name, "", 0, "", false, 0, "")

		// Bar
		pct := 0.0
		if total > 0 {
			pct = float64(cat.count) / float64(total)
		}
		x, y := r.pdf.GetXY()
		r.pdf.SetFillColor(230, 230, 230)
		r.pdf.Rect(x, y+1, barWidth, 6, "F")
		r.pdf.SetFillColor(cat.r, cat.g, cat.b)
		r.pdf.Rect(x, y+1, barWidth*pct, 6, "F")
		r.pdf.SetX(x + barWidth + 5)
		r.pdf.CellFormat(0, 8, fmt.Sprintf("%d (%.0f%%)", cat.count, pct*100), "", 1, "", false, 0, "")
	}
}

func (r *PDFReport) classificationsTable(classifications []map[string]interface{}, ticketMap map[string]map[string]interface{}) {
	r.pdf.AddPage()
	r.sectionHeader("Classifications")

	if len(classifications) == 0 {
		r.pdf.SetFont("Helvetica", "", 11)
		r.pdf.CellFormat(0, 8, "No classifications available.", "", 1, "", false, 0, "")
		return
	}

	// Table header
	r.pdf.SetFont("Helvetica", "B", 9)
	r.pdf.SetFillColor(60, 60, 70)
	r.pdf.SetTextColor(255, 255, 255)
	r.pdf.CellFormat(25, 7, "Ticket", "1", 0, "", true, 0, "")
	r.pdf.CellFormat(65, 7, "Title", "1", 0, "", true, 0, "")
	r.pdf.CellFormat(35, 7, "Category", "1", 0, "", true, 0, "")
	r.pdf.CellFormat(18, 7, "Clarity", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(18, 7, "Locality", "1", 0, "C", true, 0, "")
	r.pdf.CellFormat(18, 7, "Blast", "1", 0, "C", true, 0, "")
	r.pdf.Ln(-1)

	// Table rows
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.SetFont("Helvetica", "", 9)
	for i, c := range classifications {
		ticketId := strVal(c["ticketId"])
		category := strVal(c["category"])
		rubric := asMap(c["rubric"])
		title := ""
		if t, ok := ticketMap[ticketId]; ok {
			title = strVal(t["title"])
		}
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		if i%2 == 0 {
			r.pdf.SetFillColor(245, 245, 250)
		} else {
			r.pdf.SetFillColor(255, 255, 255)
		}

		r.pdf.CellFormat(25, 7, ticketId, "1", 0, "", true, 0, "")
		r.pdf.CellFormat(65, 7, title, "1", 0, "", true, 0, "")
		r.pdf.CellFormat(35, 7, formatCategory(category), "1", 0, "", true, 0, "")
		r.pdf.CellFormat(18, 7, fmt.Sprintf("%d", intVal(rubric["clarity"])), "1", 0, "C", true, 0, "")
		r.pdf.CellFormat(18, 7, fmt.Sprintf("%d", intVal(rubric["codeLocality"])), "1", 0, "C", true, 0, "")
		r.pdf.CellFormat(18, 7, fmt.Sprintf("%d", intVal(rubric["blastRadius"])), "1", 0, "C", true, 0, "")
		r.pdf.Ln(-1)
	}
}

func (r *PDFReport) clusterPages(clusters []map[string]interface{}, docMap map[string]map[string]interface{}) {
	r.pdf.AddPage()
	r.sectionHeader("Clusters")

	if len(clusters) == 0 {
		r.pdf.SetFont("Helvetica", "", 11)
		r.pdf.CellFormat(0, 8, "No clusters formed.", "", 1, "", false, 0, "")
		return
	}

	for _, cluster := range clusters {
		clusterId := strVal(cluster["clusterId"])
		rationale := strVal(cluster["rationale"])
		tickets := asStringList(cluster["tickets"])

		r.pdf.SetFont("Helvetica", "B", 11)
		r.pdf.CellFormat(0, 8, fmt.Sprintf("%s  (%d tickets)", clusterId, len(tickets)), "", 1, "", false, 0, "")

		r.pdf.SetFont("Helvetica", "", 10)
		r.pdf.SetTextColor(80, 80, 80)
		r.writeWrapped(rationale)
		r.pdf.Ln(2)

		r.pdf.SetTextColor(0, 0, 0)
		r.pdf.SetFont("Helvetica", "", 9)
		r.pdf.CellFormat(0, 6, "Tickets: "+strings.Join(tickets, ", "), "", 1, "", false, 0, "")

		// Repo areas from context doc
		if doc, ok := docMap[clusterId]; ok {
			areas := asStringList(doc["repoAreas"])
			// Filter to file paths only
			var filePaths []string
			for _, a := range areas {
				if !strings.Contains(a, "//") && (strings.Contains(a, "/") || strings.Contains(a, ".")) {
					filePaths = append(filePaths, a)
				}
			}
			if len(filePaths) > 0 {
				r.pdf.SetFont("Helvetica", "", 8)
				r.pdf.SetTextColor(80, 80, 80)
				r.writeWrapped("Repo Areas: " + strings.Join(filePaths, ", "))
				r.pdf.SetTextColor(0, 0, 0)
			}

			risks := asStringList(doc["risks"])
			if len(risks) > 0 {
				r.pdf.SetFont("Helvetica", "I", 9)
				r.pdf.SetTextColor(180, 120, 0)
				r.pdf.CellFormat(0, 6, "Risks: "+strings.Join(risks, ", "), "", 1, "", false, 0, "")
				r.pdf.SetTextColor(0, 0, 0)
			}
		}

		r.pdf.Ln(5)
	}
}

func (r *PDFReport) ticketDetailPages(classifications []map[string]interface{}, ticketMap map[string]map[string]interface{}) {
	if len(classifications) == 0 {
		return
	}

	r.pdf.AddPage()
	r.sectionHeader("Ticket Details")

	rubricLabels := []struct {
		key      string
		label    string
		positive bool
	}{
		{"clarity", "Clarity", true},
		{"codeLocality", "Code Locality", true},
		{"patternMatch", "Pattern Match", true},
		{"validationStrength", "Validation Strength", true},
		{"dependencyRisk", "Dependency Risk", false},
		{"productAmbiguity", "Product Ambiguity", false},
		{"blastRadius", "Blast Radius", false},
	}

	for i, c := range classifications {
		if i > 0 {
			r.pdf.Ln(3)
			// Check if we need a new page
			if r.pdf.GetY() > 220 {
				r.pdf.AddPage()
			}
			r.pdf.SetDrawColor(200, 200, 200)
			r.pdf.Line(10, r.pdf.GetY(), 200, r.pdf.GetY())
			r.pdf.Ln(5)
		}

		ticketId := strVal(c["ticketId"])
		category := strVal(c["category"])
		rubric := asMap(c["rubric"])
		reasons := asStringList(c["reasons"])
		hardStops := asStringList(c["hardStops"])
		gateResults := asList(c["gateResults"])

		title := ""
		if t, ok := ticketMap[ticketId]; ok {
			title = strVal(t["title"])
		}

		// Ticket header
		r.pdf.SetFont("Helvetica", "B", 12)
		r.pdf.CellFormat(30, 8, ticketId, "", 0, "", false, 0, "")
		r.pdf.SetFont("Helvetica", "", 10)
		setCategoryColor(r.pdf, category)
		r.pdf.CellFormat(35, 8, formatCategory(category), "", 0, "", false, 0, "")
		r.pdf.SetTextColor(0, 0, 0)
		r.pdf.Ln(-1)

		if title != "" {
			r.pdf.SetFont("Helvetica", "", 10)
			r.pdf.SetTextColor(80, 80, 80)
			r.writeWrapped(title)
			r.pdf.SetTextColor(0, 0, 0)
		}
		r.pdf.Ln(3)

		// Rubric scores as horizontal bars
		barWidth := 80.0
		for _, rl := range rubricLabels {
			score := intVal(rubric[rl.key])
			r.pdf.SetFont("Helvetica", "", 8)
			r.pdf.CellFormat(35, 5, rl.label, "", 0, "R", false, 0, "")
			r.pdf.Cell(2, 5, " ")

			x, y := r.pdf.GetXY()
			// Background bar
			r.pdf.SetFillColor(230, 230, 230)
			r.pdf.Rect(x, y+0.5, barWidth, 4, "F")
			// Score bar
			if rl.positive {
				if score >= 4 {
					r.pdf.SetFillColor(34, 197, 94)
				} else if score >= 3 {
					r.pdf.SetFillColor(59, 130, 246)
				} else {
					r.pdf.SetFillColor(234, 179, 8)
				}
			} else {
				if score <= 1 {
					r.pdf.SetFillColor(34, 197, 94)
				} else if score <= 2 {
					r.pdf.SetFillColor(59, 130, 246)
				} else {
					r.pdf.SetFillColor(239, 68, 68)
				}
			}
			pct := float64(score) / 5.0
			r.pdf.Rect(x, y+0.5, barWidth*pct, 4, "F")

			r.pdf.SetX(x + barWidth + 3)
			r.pdf.CellFormat(0, 5, fmt.Sprintf("%d", score), "", 1, "", false, 0, "")
		}

		// Hard stops
		if len(hardStops) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.SetTextColor(220, 50, 50)
			r.pdf.CellFormat(0, 6, "Hard Stops:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 9)
			for _, hs := range hardStops {
				r.pdf.CellFormat(0, 5, "  - "+hs, "", 1, "", false, 0, "")
			}
			r.pdf.SetTextColor(0, 0, 0)
		}

		// Gate results
		if len(gateResults) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Gate Results:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 9)
			for _, g := range gateResults {
				gate := strVal(g["gate"])
				passed := boolVal(g["passed"])
				mark := "PASS"
				if !passed {
					mark = "FAIL"
					r.pdf.SetTextColor(220, 50, 50)
				} else {
					r.pdf.SetTextColor(34, 150, 60)
				}
				r.pdf.CellFormat(12, 5, mark, "", 0, "", false, 0, "")
				r.pdf.SetTextColor(0, 0, 0)
				r.pdf.CellFormat(0, 5, gate, "", 1, "", false, 0, "")
			}
		}

		// Reasons (compact)
		if len(reasons) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Reasons:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 8)
			r.pdf.SetTextColor(60, 60, 60)
			for _, reason := range reasons {
				r.pdf.Cell(5, 5, "")
				r.writeWrapped("- " + reason)
			}
			r.pdf.SetTextColor(0, 0, 0)
		}
	}
}

func (r *PDFReport) planPages(plans []map[string]interface{}, ticketMap map[string]map[string]interface{}, planStats map[string]interface{}) {
	r.pdf.AddPage()
	r.sectionHeader("Execution Plans")

	// Summary bar
	total := intVal(planStats["total"])
	passed := intVal(planStats["passed"])
	failed := intVal(planStats["failed"])
	planTokens := intVal(planStats["totalTokensUsed"])
	planCost := floatVal(planStats["totalCostUsd"])

	r.pdf.SetFont("Helvetica", "", 11)
	summary := fmt.Sprintf("%d plans generated: %d passed gates, %d failed", total, passed, failed)
	if planTokens > 0 {
		summary += fmt.Sprintf(" (%.1fK tokens, $%.4f)", float64(planTokens)/1000, planCost)
	}
	r.pdf.CellFormat(0, 8, summary, "", 1, "", false, 0, "")
	r.pdf.Ln(5)

	for i, p := range plans {
		if i > 0 {
			if r.pdf.GetY() > 220 {
				r.pdf.AddPage()
			}
			r.pdf.SetDrawColor(200, 200, 200)
			r.pdf.Line(10, r.pdf.GetY(), 200, r.pdf.GetY())
			r.pdf.Ln(5)
		}

		ticketId := strVal(p["ticketId"])
		errMsg := strVal(p["error"])
		plan := asMap(p["plan"])
		validation := asMap(p["validation"])

		title := ""
		if t, ok := ticketMap[ticketId]; ok {
			title = strVal(t["title"])
		}

		// Header: ticket ID + pass/fail badge
		r.pdf.SetFont("Helvetica", "B", 12)
		r.pdf.CellFormat(30, 8, ticketId, "", 0, "", false, 0, "")

		if errMsg != "" {
			r.pdf.SetTextColor(239, 68, 68)
			r.pdf.SetFont("Helvetica", "B", 10)
			r.pdf.CellFormat(20, 8, "ERROR", "", 0, "", false, 0, "")
			r.pdf.SetTextColor(0, 0, 0)
		} else if boolVal(validation["passed"]) {
			r.pdf.SetTextColor(34, 197, 94)
			r.pdf.SetFont("Helvetica", "B", 10)
			r.pdf.CellFormat(20, 8, "PASS", "", 0, "", false, 0, "")
			r.pdf.SetTextColor(0, 0, 0)
		} else {
			r.pdf.SetTextColor(234, 179, 8)
			r.pdf.SetFont("Helvetica", "B", 10)
			r.pdf.CellFormat(20, 8, "FAIL", "", 0, "", false, 0, "")
			r.pdf.SetTextColor(0, 0, 0)
		}
		r.pdf.Ln(-1)

		if title != "" {
			r.pdf.SetFont("Helvetica", "", 10)
			r.pdf.SetTextColor(80, 80, 80)
			r.writeWrapped(title)
			r.pdf.SetTextColor(0, 0, 0)
		}

		if errMsg != "" {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "", 9)
			r.pdf.SetTextColor(220, 50, 50)
			r.writeWrapped("Error: " + errMsg)
			r.pdf.SetTextColor(0, 0, 0)
			continue
		}

		// Approach
		approach := strVal(plan["approach"])
		if approach != "" {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Approach:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 9)
			r.pdf.SetTextColor(60, 60, 60)
			if len(approach) > 500 {
				approach = approach[:497] + "..."
			}
			r.writeWrapped(approach)
			r.pdf.SetTextColor(0, 0, 0)
		}

		// Candidate files
		candidateFiles := asStringList(plan["candidateFiles"])
		if len(candidateFiles) > 0 {
			validatedFiles := asStringList(validation["validatedFiles"])
			missingFiles := asStringList(validation["missingFiles"])
			validatedSet := toSet(validatedFiles)
			missingSet := toSet(missingFiles)

			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, fmt.Sprintf("Candidate Files (%d):", len(candidateFiles)), "", 1, "", false, 0, "")
			r.pdf.SetFont("Courier", "", 8)
			for _, f := range candidateFiles {
				mark := " "
				if missingSet[f] {
					r.pdf.SetTextColor(220, 50, 50)
					mark = "X"
				} else if validatedSet[f] {
					r.pdf.SetTextColor(34, 150, 60)
					mark = "+"
				}
				r.pdf.CellFormat(0, 5, fmt.Sprintf("  %s %s", mark, f), "", 1, "", false, 0, "")
				r.pdf.SetTextColor(0, 0, 0)
			}
		}

		// Validation commands
		validationCmds := asStringList(plan["validation"])
		if len(validationCmds) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Validation Commands:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Courier", "", 8)
			for _, cmd := range validationCmds {
				r.pdf.CellFormat(0, 5, "  $ "+cmd, "", 1, "", false, 0, "")
			}
		}

		// Stop conditions
		stopConditions := asStringList(plan["stopConditions"])
		if len(stopConditions) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Stop Conditions:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 8)
			for _, sc := range stopConditions {
				r.pdf.Cell(5, 5, "")
				r.writeWrapped("- " + sc)
			}
		}

		// Uncertainties
		uncertainties := asStringList(plan["uncertainties"])
		if len(uncertainties) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Uncertainties:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 8)
			r.pdf.SetTextColor(180, 120, 0)
			for _, u := range uncertainties {
				r.pdf.Cell(5, 5, "")
				r.writeWrapped("? " + u)
			}
			r.pdf.SetTextColor(0, 0, 0)
		}

		// Gate results
		gateResults := asList(validation["gateResults"])
		if len(gateResults) > 0 {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Validation Gates:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 9)
			for _, g := range gateResults {
				gate := strVal(g["gate"])
				gPassed := boolVal(g["passed"])
				reason := strVal(g["reason"])
				mark := "PASS"
				if !gPassed {
					mark = "FAIL"
					r.pdf.SetTextColor(220, 50, 50)
				} else {
					r.pdf.SetTextColor(34, 150, 60)
				}
				r.pdf.CellFormat(12, 5, mark, "", 0, "", false, 0, "")
				r.pdf.SetTextColor(0, 0, 0)
				label := strings.ReplaceAll(gate, "_", " ")
				if reason != "" {
					label += " - " + reason
				}
				r.pdf.CellFormat(0, 5, label, "", 1, "", false, 0, "")
			}
		}

		// Rollback
		rollback := strVal(plan["rollback"])
		if rollback != "" {
			r.pdf.Ln(2)
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.CellFormat(0, 6, "Rollback:", "", 1, "", false, 0, "")
			r.pdf.SetFont("Helvetica", "", 9)
			r.pdf.SetTextColor(60, 60, 60)
			r.writeWrapped(rollback)
			r.pdf.SetTextColor(0, 0, 0)
		}
	}
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

// --- helpers ---

func (r *PDFReport) sectionHeader(title string) {
	r.pdf.SetFont("Helvetica", "B", 16)
	r.pdf.CellFormat(0, 12, title, "", 1, "", false, 0, "")
	r.pdf.SetDrawColor(200, 160, 50)
	r.pdf.Line(10, r.pdf.GetY(), 200, r.pdf.GetY())
	r.pdf.Ln(5)
}

func (r *PDFReport) writeWrapped(text string) {
	r.pdf.MultiCell(0, 5, text, "", "", false)
}

func formatCategory(cat string) string {
	switch cat {
	case "AI_DEFINITE":
		return "AI Definite"
	case "AI_LIKELY":
		return "AI Likely"
	case "HUMAN_REVIEW_REQUIRED":
		return "Human Review"
	case "HUMAN_ONLY":
		return "Human Only"
	default:
		return cat
	}
}

func setCategoryColor(pdf *fpdf.Fpdf, cat string) {
	switch cat {
	case "AI_DEFINITE":
		pdf.SetTextColor(34, 197, 94)
	case "AI_LIKELY":
		pdf.SetTextColor(59, 130, 246)
	case "HUMAN_REVIEW_REQUIRED":
		pdf.SetTextColor(234, 179, 8)
	case "HUMAN_ONLY":
		pdf.SetTextColor(239, 68, 68)
	default:
		pdf.SetTextColor(0, 0, 0)
	}
}

func asMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func asList(v interface{}) []map[string]interface{} {
	if arr, ok := v.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	}
	return nil
}

func asStringList(v interface{}) []string {
	if arr, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func strVal(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func intVal(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}

func floatVal(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}

func boolVal(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
