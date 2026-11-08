package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Resume parsing: turns plain resume text into work-experience entries using
// rules only (no LLM). Standalone: nothing else in the backend depends on it.
//
// Not yet done: extracting text from the PDF itself. Feed this the text of a
// resume (or the "other experience" box) via parseResumeText.

// Experience is one work-experience entry.
type Experience struct {
	Title     string  `json:"title"`
	Company   string  `json:"company"`
	Location  *string `json:"location"`
	Type      string  `json:"type"`      // internship | job | research | teaching | other
	StartDate *string `json:"startDate"` // YYYY-MM
	EndDate   *string `json:"endDate"`   // YYYY-MM, nil when current
	IsCurrent bool    `json:"isCurrent"`
	Summary   *string `json:"summary"`
}

const (
	monthPat = `(?:jan(?:uary)?|feb(?:ruary)?|mar(?:ch)?|apr(?:il)?|may|june?|july?|aug(?:ust)?|sep(?:t(?:ember)?)?|oct(?:ober)?|nov(?:ember)?|dec(?:ember)?|spring|summer|fall|autumn|winter)`
	datePat  = `(?:\b` + monthPat + `\b\.?,?\s*\d{4}|\b\d{1,2}/\d{4}|\b\d{4})`
	endPat   = `(?:` + datePat + `|\b(?:present|current|now|ongoing)\b)`
	sepPat   = `(?:\s*[-–—]\s*|\s+to\s+)`
)

var (
	rangeRe  = regexp.MustCompile(`(?i)(` + datePat + `)` + sepPat + `(` + endPat + `)`)
	singleRe = regexp.MustCompile(`(?i)\b` + monthPat + `\b\.?,?\s*\d{4}`)
	slashRe  = regexp.MustCompile(`^(\d{1,2})/(\d{4})$`)
	wordRe   = regexp.MustCompile(`(?i)^([a-z]+)\.?,?\s*(\d{4})$`)
	locRe    = regexp.MustCompile(`^[A-Z][A-Za-z.\- ]+,\s*[A-Z]{2}$|^(?i:remote)$`)
	splitRe  = regexp.MustCompile(`\s*\|\s*|\s+[–—-]\s+|\s+at\s+|\s{3,}`)
	bulletRe = regexp.MustCompile(`^\s*[•·▪●○◦*\-–]\s*`)
	titleKw  = regexp.MustCompile(`(?i)\b(intern|engineer|developer|analyst|assistant|associate|researcher|research|manager|director|consultant|specialist|coordinator|tutor|instructor|lecturer|trader|technician|scientist|architect|co-?op|fellow|ta|swe)\b`)
)

var months = map[string]string{
	"jan": "01", "feb": "02", "mar": "03", "apr": "04", "may": "05", "jun": "06",
	"jul": "07", "aug": "08", "sep": "09", "oct": "10", "nov": "11", "dec": "12",
	// seasons, per the extraction rules
	"spr": "01", "sum": "06", "fal": "08", "aut": "08", "win": "12",
}

// Section headers that end the Experience section.
var stopHeaders = map[string]bool{
	"education": true, "projects": true, "project": true, "skills": true, "technical skills": true,
	"leadership": true, "activities": true, "honors": true, "awards": true, "certifications": true,
	"publications": true, "involvement": true, "extracurriculars": true, "summary": true,
	"leadership experience": true, "campus involvement": true, "volunteer experience": true,
}

var expHeaders = map[string]bool{
	"experience": true, "work experience": true, "professional experience": true,
	"relevant experience": true, "employment": true, "work history": true,
	"research experience": true, "internships": true,
}

func headerKey(line string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimRight(strings.TrimSpace(line), ":")))
}

// parseDate converts one date token to YYYY-MM. yearOnly is true when only a
// year was present (month defaults to 01).
func parseDate(tok string) (ym string, yearOnly, ok bool) {
	tok = strings.TrimSpace(tok)
	if m := slashRe.FindStringSubmatch(tok); m != nil {
		return fmt.Sprintf("%s-%02s", m[2], m[1]), false, true
	}
	if regexp.MustCompile(`^\d{4}$`).MatchString(tok) {
		return tok + "-01", true, true
	}
	if m := wordRe.FindStringSubmatch(tok); m != nil {
		w := strings.ToLower(m[1])
		if len(w) >= 3 {
			if mm, found := months[w[:3]]; found {
				return m[2] + "-" + mm, false, true
			}
		}
	}
	return "", false, false
}

// parseResumeText extracts experience entries from plain resume text.
// Warnings flag anything the rules had to guess.
func parseResumeText(text string) ([]Experience, []string) {
	var warnings []string
	lines := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")

	// 1. Slice out the Experience section.
	var section []string
	in := false
	for _, l := range lines {
		k := headerKey(l)
		switch {
		case expHeaders[k]:
			in = true
		case in && stopHeaders[k]:
			in = false
		case in:
			section = append(section, l)
		}
	}
	if len(section) == 0 {
		return nil, []string{"no Experience section found"}
	}

	// 2. Anchor an entry on each line containing a date range.
	var out []Experience
	var pending []string // non-bullet header lines waiting for a date line
	afterBullet := false
	for _, raw := range section {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if bulletRe.MatchString(line) {
			afterBullet = true
			continue
		}
		loc := rangeRe.FindStringSubmatchIndex(line)
		single := singleRe.FindStringIndex(line)
		if loc == nil && single == nil {
			// Lowercase start after a bullet = wrapped bullet text, not a header.
			if afterBullet && line[0] >= 'a' && line[0] <= 'z' {
				continue
			}
			pending = append(pending, line)
			if len(pending) > 2 {
				pending = pending[len(pending)-2:]
			}
			afterBullet = false
			continue
		}

		var e Experience
		var rest string
		var startTok, endTok string
		if loc != nil {
			startTok, endTok = line[loc[2]:loc[3]], line[loc[4]:loc[5]]
			rest = strings.TrimSpace(line[:loc[0]] + " " + line[loc[1]:])
		} else {
			startTok, endTok = line[single[0]:single[1]], line[single[0]:single[1]]
			rest = strings.TrimSpace(line[:single[0]] + " " + line[single[1]:])
		}

		s, sYearOnly, _ := parseDate(startTok)
		e.StartDate = &s
		if sYearOnly {
			warnings = append(warnings, fmt.Sprintf("start date %q is year only, defaulted to %s", startTok, s))
		}
		if regexp.MustCompile(`(?i)present|current|now|ongoing`).MatchString(endTok) {
			e.IsCurrent = true
		} else if end, eYearOnly, ok := parseDate(endTok); ok {
			e.EndDate = &end
			if eYearOnly {
				warnings = append(warnings, fmt.Sprintf("end date %q is year only, defaulted to %s", endTok, end))
			}
		}

		header := append([]string{}, pending...)
		if t := strings.Trim(rest, " ,|–—-"); t != "" {
			header = append(header, t)
		}
		pending, afterBullet = nil, false

		w := assignFields(&e, header)
		warnings = append(warnings, w...)
		if e.Title == "" && e.Company == "" {
			warnings = append(warnings, fmt.Sprintf("dates %s–%s found but no title/company, skipped", startTok, endTok))
			continue
		}
		e.Type = guessType(e.Title + " " + e.Company)
		out = append(out, e)
	}

	// 3. Most recent first (current roles on top).
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsCurrent != out[j].IsCurrent {
			return out[i].IsCurrent
		}
		return deref(out[i].EndDate) > deref(out[j].EndDate)
	})
	return out, warnings
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func guessType(s string) string {
	l := strings.ToLower(s)
	switch {
	case strings.Contains(l, "intern") || strings.Contains(l, "co-op") || strings.Contains(l, "coop"):
		return "internship"
	case strings.Contains(l, "research"):
		return "research"
	case regexp.MustCompile(`teaching|\bta\b|tutor|instructor|lecturer`).MatchString(l):
		return "teaching"
	}
	return "job"
}

// assignFields splits the header lines into title / company / location.
func assignFields(e *Experience, header []string) []string {
	var pieces []string
	for _, h := range header {
		for _, p := range splitRe.Split(h, -1) {
			if p = strings.TrimSpace(p); p != "" {
				pieces = append(pieces, p)
			}
		}
	}
	var rest []string
	for _, p := range pieces {
		if e.Location == nil && locRe.MatchString(p) {
			loc := p
			e.Location = &loc
		} else {
			rest = append(rest, p)
		}
	}
	// "Software Intern, Google" in a single piece.
	if len(rest) == 1 && strings.Contains(rest[0], ", ") {
		parts := strings.SplitN(rest[0], ", ", 2)
		rest = []string{parts[0], parts[1]}
	}

	var warns []string
	switch len(rest) {
	case 0:
	case 1:
		if titleKw.MatchString(rest[0]) {
			e.Title = rest[0]
		} else {
			e.Company = rest[0]
		}
		warns = append(warns, fmt.Sprintf("only one of title/company found (%q)", rest[0]))
	default:
		ti := 0
		for i, p := range rest {
			if titleKw.MatchString(p) {
				ti = i
				break
			}
		}
		e.Title = rest[ti]
		for i, p := range rest {
			if i != ti {
				e.Company = p
				break
			}
		}
		if !titleKw.MatchString(e.Title) {
			warns = append(warns, fmt.Sprintf("title/company order guessed for %q / %q", e.Title, e.Company))
		}
		if len(rest) > 2 {
			warns = append(warns, fmt.Sprintf("extra header text ignored for %q", e.Company))
		}
	}
	return warns
}
