package spam

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		name      string
		nameField string
		email     string
		values    []string
		wantSpam  bool
	}{
		{
			name:      "legit message",
			nameField: "Grace Hopper",
			email:     "grace@example.com",
			values:    []string{"Hi, I'd love to chat about a project next week."},
			wantSpam:  false,
		},
		{
			name:      "link spam",
			nameField: "x",
			email:     "x@x.io",
			values:    []string{"cheap loans http://spam.example https://spam2.example www.more.example"},
			wantSpam:  true,
		},
		{
			name:      "keyword spam",
			nameField: "bot",
			email:     "b@b.io",
			values:    []string{"buy viagra and crypto bitcoin now"},
			wantSpam:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSpam(Score(tc.nameField, tc.email, tc.values))
			if got != tc.wantSpam {
				t.Fatalf("IsSpam=%v want %v (score=%.2f)", got, tc.wantSpam, Score(tc.nameField, tc.email, tc.values))
			}
		})
	}
}

func TestUpperRatio(t *testing.T) {
	cases := map[string]float64{
		"":        0,
		"1234 !!": 0,
		"ABCD":    1,
		"ABcd":    0.5,
		"hello":   0,
	}
	for in, want := range cases {
		if got := upperRatio(in); got != want {
			t.Errorf("upperRatio(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestScoreIsClampedToOne(t *testing.T) {
	got := Score("bot", "b@b.io", []string{"http://a.example http://b.example http://c.example http://d.example casino lottery"})
	if got != 1 {
		t.Fatalf("score = %v, want 1", got)
	}
}

func TestScoreSignals(t *testing.T) {
	cases := []struct {
		name   string
		values []string
		want   float64
	}{
		{"ordinary message", []string{"Hi, I'd like to book a call next week."}, 0},
		{"very short body", []string{"hi"}, 0.2},
		{"shouting", []string{"PLEASE CALL ME BACK TODAY"}, 0.2},
		{"single link", []string{"My portfolio is at https://example.com/portfolio if useful."}, 0.35},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Score("Grace Hopper", "grace@example.com", tc.values); got != tc.want {
				t.Fatalf("score = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsSpamThreshold(t *testing.T) {
	if !IsSpam(Threshold) {
		t.Fatal("a score equal to the threshold should be spam")
	}
	if IsSpam(Threshold - 0.01) {
		t.Fatal("a score just under the threshold should not be spam")
	}
	// One link is fine on its own; a link plus a spammy keyword crosses the line.
	if IsSpam(Score("Grace", "grace@example.com", []string{"My portfolio is at https://example.com/portfolio if useful."})) {
		t.Fatal("a single link was flagged")
	}
	if !IsSpam(Score("x", "x@x.io", []string{"great casino offer at https://spam.example today"})) {
		t.Fatal("link plus keyword was not flagged")
	}
}
