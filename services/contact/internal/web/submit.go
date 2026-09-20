package web

import (
	"encoding/json"
	"fmt"
	"github.com/aashishrajdev/halomail/services/contact/internal/app"
	"github.com/aashishrajdev/halomail/services/shared/accesskey"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"mime"
	"net/http"
	"strings"
)

func SubmitHandler(service *app.Service, pool *pgxpool.Pool) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(writer, request.Body, 64<<10)
		fields, err := decodeFields(request)
		if err != nil {
			httpx.Error(writer, errs.Invalid("provide a JSON or HTML form with text fields, up to 64 KB"))
			return
		}
		secret := fields["access_key"]
		delete(fields, "access_key")
		owner, err := accesskey.Verify(request.Context(), pool, secret, "forms")
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		honeypot := fields["_hl_hp"]
		delete(fields, "_hl_hp")
		result, err := service.SubmitMessage(request.Context(), owner.ID, app.SubmitInput{SenderName: fields["name"], SenderEmail: fields["email"], Data: fields, Honeypot: honeypot, IP: httpx.Peer(request.RemoteAddr), UserAgent: request.UserAgent()})
		if err != nil {
			httpx.Error(writer, err)
			return
		}
		httpx.JSON(writer, 200, map[string]any{"success": true, "id": result.Message.ID, "message": "Submission received"})
	}
}

func decodeFields(request *http.Request) (map[string]string, error) {
	fields := map[string]string{}
	media, _, _ := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if media == "application/json" {
		var raw map[string]any
		decoder := json.NewDecoder(request.Body)
		decoder.UseNumber()
		if err := decoder.Decode(&raw); err != nil || raw == nil {
			return nil, fmt.Errorf("a JSON object is required")
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("provide one JSON object")
		}
		for key, value := range raw {
			switch typed := value.(type) {
			case string:
				fields[key] = typed
			case json.Number, bool:
				fields[key] = fmt.Sprint(typed)
			case nil:
				fields[key] = ""
			default:
				return nil, fmt.Errorf("text fields required")
			}
		}
	} else {
		if media == "multipart/form-data" {
			if err := request.ParseMultipartForm(64 << 10); err != nil {
				return nil, err
			}
			defer request.MultipartForm.RemoveAll()
			if len(request.MultipartForm.File) > 0 {
				return nil, fmt.Errorf("file uploads are not supported")
			}
		} else if media == "application/x-www-form-urlencoded" {
			if err := request.ParseForm(); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unsupported content type")
		}
		for key, values := range request.PostForm {
			fields[key] = strings.Join(values, ", ")
		}
	}
	return fields, nil
}
