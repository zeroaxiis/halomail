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
