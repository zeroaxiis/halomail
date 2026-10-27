package outbox

import (
	"strings"
	"testing"
)

func TestFieldsEscapeUserHTML(test *testing.T) {
	body := Fields("<script>", map[string]string{"<label>": `<img src=x onerror="alert(1)">`})
	if strings.Contains(body, "<script>") || strings.Contains(body, "<img src=x") || !strings.Contains(body, "&lt;img") {
		test.Fatal("unescaped user HTML")
	}
}

func TestFieldsAreSortedByLabel(test *testing.T) {
	body := Fields("Contact", map[string]string{"name": "Grace", "email": "grace@example.com", "message": "hi"})
	email := strings.Index(body, "email</td>")
	message := strings.Index(body, "message</td>")
	name := strings.Index(body, "name</td>")
	if email < 0 || !(email < message && message < name) {
		test.Fatalf("fields not in label order: %s", body)
	}
	if !strings.Contains(body, "Contact</h2>") || !strings.Contains(body, "ZeroAxiis") {
		test.Fatalf("unexpected layout: %s", body)
	}
}
