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
