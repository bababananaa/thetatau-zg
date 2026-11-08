package main

import "testing"

func TestParseResumeText(t *testing.T) {
	text := `Jane Doe
Education
University of Florida
Aug 2022 – May 2026
Experience
Software Engineering Intern | Google | Mountain View, CA   Jun 2025 – Aug 2025
• Built things
  that wrapped lines
Teaching Assistant
University of Florida, Gainesville, FL
Jan 2024 - Present
• Graded
Research Assistant – UF Lab  Summer 2023
Projects
Cool App   Jan 2024 – Mar 2024
`
	got, warns := parseResumeText(text)
	if len(got) != 3 {
		t.Fatalf("want 3 entries, got %d: %+v (warnings %v)", len(got), got, warns)
	}
	ta := got[0] // current role first
	if !ta.IsCurrent || ta.EndDate != nil || *ta.StartDate != "2024-01" || ta.Type != "teaching" {
		t.Errorf("TA wrong: %+v", ta)
	}
	intern := got[1]
	if intern.Title != "Software Engineering Intern" || intern.Company != "Google" ||
		*intern.Location != "Mountain View, CA" || *intern.StartDate != "2025-06" ||
		*intern.EndDate != "2025-08" || intern.Type != "internship" {
		t.Errorf("intern wrong: %+v", intern)
	}
	ra := got[2]
	if *ra.StartDate != "2023-06" || ra.Type != "research" {
		t.Errorf("RA wrong: %+v", ra)
	}
}

func TestParseDate(t *testing.T) {
	for in, want := range map[string]string{
		"Jun 2025": "2025-06", "September 2020": "2020-09", "Fall 2022": "2022-08",
		"Spring 2021": "2021-01", "Winter 2020": "2020-12", "03/2021": "2021-03", "2019": "2019-01",
	} {
		if got, _, ok := parseDate(in); !ok || got != want {
			t.Errorf("parseDate(%q) = %q,%v want %q", in, got, ok, want)
		}
	}
}

func TestNoSection(t *testing.T) {
	if got, w := parseResumeText("just some text"); got != nil || len(w) == 0 {
		t.Errorf("want nil + warning, got %v %v", got, w)
	}
}
