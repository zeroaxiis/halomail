package themes

import (
	"strings"
	"testing"

	"github.com/aashishrajdev/halomail/services/template/internal/domain"
)

func TestRenderBuiltins(t *testing.T) {
	vars := map[string]string{"heading": "Booked!", "body": "See you Monday."}
	for _, k := range domain.BuiltinThemes {
		subj, htmlOut := Render(k, "Hi {{name}}", map[string]string{"name": "Grace", "heading": vars["heading"], "body": vars["body"]})
		if subj != "Hi Grace" {
			t.Errorf("%s: subject substitution = %q", k, subj)
		}
		if !strings.Contains(htmlOut, "Booked!") {
			t.Errorf("%s: rendered HTML missing heading", k)
		}
		if !strings.HasPrefix(htmlOut, "<!doctype html>") {
			t.Errorf("%s: not a full HTML doc", k)
		}
	}
}

func TestRenderCustomEscapes(t *testing.T) {
	_, out := RenderCustom(`<div>{{name}}</div>`, "s", map[string]string{"name": "<script>"})
	if strings.Contains(out, "<script>") {
		t.Fatalf("custom render did not escape variable: %q", out)
	}
}

func TestThemesGallery(t *testing.T) {
	all := Themes()
	if len(all) != len(domain.BuiltinThemes) {
		t.Fatalf("gallery has %d themes, want %d", len(all), len(domain.BuiltinThemes))
	}
	for _, ti := range all {
		if ti.PreviewHTML == "" || ti.Name == "" {
			t.Errorf("theme %s missing preview/name", ti.Kind)
		}
	}
}

func TestReplaceVars(t *testing.T) {
	cases := []struct {
		in   string
		vars map[string]string
		want string
	}{
		{"Hi {{name}}", map[string]string{"name": "Grace"}, "Hi Grace"},
		{"Hi {{ name }}!", map[string]string{"name": "Grace"}, "Hi Grace!"},
		{"Hi {{unknown}}!", map[string]string{"name": "Grace"}, "Hi !"},
		{"{{name}}", map[string]string{"name": "<b>Grace</b>"}, "&lt;b&gt;Grace&lt;/b&gt;"},
		{"no variables", nil, "no variables"},
	}
	for _, tc := range cases {
		if got := replaceVars(tc.in, tc.vars); got != tc.want {
			t.Errorf("replaceVars(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRenderUnknownThemeFallsBackToMinimal(t *testing.T) {
	vars := map[string]string{"heading": "Booked!", "body": "See you Monday."}
	_, unknown := Render("does-not-exist", "s", vars)
	_, minimal := Render(domain.ThemeMinimal, "s", vars)
	if unknown != minimal {
		t.Fatal("unknown theme did not render as minimal")
	}
	if _, apple := Render(domain.ThemeApple, "s", vars); apple == minimal {
		t.Fatal("apple and minimal rendered identically")
	}
}

func TestSplitParasAndFirstNonEmpty(t *testing.T) {
	if got := strings.Join(splitParas("one\r\n\r\ntwo\n   \nthree"), "|"); got != "one|two|three" {
		t.Errorf("splitParas = %q", got)
	}
	if got := splitParas(""); len(got) != 1 || got[0] != "" {
		t.Errorf("empty body should yield one empty paragraph, got %q", got)
	}
	if got := firstNonEmpty("", "   ", "x", "y"); got != "x" {
		t.Errorf("firstNonEmpty = %q", got)
	}
	if got := firstNonEmpty(); got != "" {
		t.Errorf("firstNonEmpty() = %q", got)
	}
}

func TestRenderButton(t *testing.T) {
	_, without := Render(domain.ThemeMinimal, "s", map[string]string{"body": "No call to action."})
	if strings.Contains(without, "<a href") {
		t.Fatal("button rendered without button_text")
	}

	_, with := Render(domain.ThemeMinimal, "s", map[string]string{
		"button_text": "Open <now>",
		"button_url":  `https://example.com/?a=1&b="x"`,
	})
	if !strings.Contains(with, `href="https://example.com/?a=1&amp;b=&#34;x&#34;"`) {
		t.Fatalf("button url not escaped: %s", with)
	}
	if !strings.Contains(with, "Open &lt;now&gt;</a>") {
		t.Fatalf("button text not escaped: %s", with)
	}
}

func TestSubjectFallbacks(t *testing.T) {
	if got, _ := Render(domain.ThemeMinimal, "", map[string]string{"subject": "From vars"}); got != "From vars" {
		t.Errorf("subject = %q, want the subject variable", got)
	}
	if got, _ := Render(domain.ThemeMinimal, "{{missing}}", nil); got != "Hello from HaloMail" {
		t.Errorf("subject = %q, want the default", got)
	}
	if got, _ := RenderCustom("<p>hi</p>", "   ", nil); got != "Hello from HaloMail" {
		t.Errorf("custom subject = %q, want the default", got)
	}
}
