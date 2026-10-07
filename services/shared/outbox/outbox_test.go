package outbox

import (
	"strings"
	"testing"
)

func TestFieldsEscapeUserHTML(test *testing.T) {
	body := Fields("<script>", map[string]string{"<label>": `<img src=x onerror="alert(1)">`})
	if strings.Contains(body, "<script>") || strings.Contains(body, "<img") || !strings.Contains(body, "&lt;img") {
		test.Fatal("unescaped user HTML")
	}
}

func TestFieldsAreSortedByLabel(test *testing.T) {
	body := Fields("Contact", map[string]string{"name": "Grace", "email": "grace@example.com", "message": "hi"})
	email := strings.Index(body, "<strong>email</strong>")
	message := strings.Index(body, "<strong>message</strong>")
	name := strings.Index(body, "<strong>name</strong>")
	if email < 0 || !(email < message && message < name) {
		test.Fatalf("fields not in label order: %s", body)
	}
	if !strings.HasPrefix(body, "<h1>Contact</h1><dl>") || !strings.HasSuffix(body, "</dl>") {
		test.Fatalf("unexpected layout: %s", body)
	}
}
