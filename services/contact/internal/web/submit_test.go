package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeFields(test *testing.T) {
	cases := []struct {
		name, contentType, body string
		want                    map[string]string
		invalid                 bool
	}{
		{name: "json fields", contentType: "application/json", body: `{"name":"Asha","count":9007199254740993,"agreed":true}`, want: map[string]string{"name": "Asha", "count": "9007199254740993", "agreed": "true"}},
		{name: "html fields", contentType: "application/x-www-form-urlencoded", body: "name=Asha&topic=One&topic=Two", want: map[string]string{"name": "Asha", "topic": "One, Two"}},
		{name: "nested objects", contentType: "application/json", body: `{"data":{"message":"hello"}}`, invalid: true},
		{name: "multiple objects", contentType: "application/json", body: `{} {}`, invalid: true},
		{name: "null", contentType: "application/json", body: `null`, invalid: true},
		{name: "unknown format", contentType: "text/plain", body: "name=Asha", invalid: true},
	}
	for _, example := range cases {
		test.Run(example.name, func(test *testing.T) {
			request := httptest.NewRequest("POST", "/submit?topic=Ignored", strings.NewReader(example.body))
			request.Header.Set("Content-Type", example.contentType)
			fields, err := decodeFields(request)
			if example.invalid {
				if err == nil {
					test.Fatal("accepted invalid input")
				}
				return
			}
			if err != nil {
				test.Fatal(err)
			}
			for key, want := range example.want {
				if fields[key] != want {
					test.Errorf("%s: got %q, want %q", key, fields[key], want)
				}
			}
		})
	}
}
