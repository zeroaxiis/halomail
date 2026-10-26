package rpc

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"github.com/zeroaxiis/halomail/services/shared/errs"
	identityv1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/identity/v1"
	"github.com/zeroaxiis/halomail/services/shared/httpx"
	"google.golang.org/protobuf/encoding/protojson"
)

// MountGoogleAuth registers the internal POST /v1/auth/google endpoint called by web BFF.
func (h *Handlers) MountGoogleAuth(mux *http.ServeMux, deliverySecret string) {
	mux.HandleFunc("POST /v1/auth/google", func(w http.ResponseWriter, r *http.Request) {
		if len(deliverySecret) < 32 || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-OTP-Delivery-Secret")), []byte(deliverySecret)) != 1 {
			httpx.Error(w, errs.Unauthorized("internal authentication required"))
			return
		}
		var body struct {
			Email     string `json:"email"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatarUrl"`
			GoogleID  string `json:"googleId"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			httpx.Error(w, errs.Invalid("invalid request"))
			return
		}
		if decoder.Decode(&struct{}{}) != io.EOF {
			httpx.Error(w, errs.Invalid("invalid request"))
			return
		}

		result, err := h.app.GoogleLogin(r.Context(), body.Email, body.Name, body.AvatarURL, body.GoogleID)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		payload, err := protojson.Marshal(&identityv1.LoginResponse{
			User:    toProtoUser(result.User),
			Session: toProtoSession(result.Session),
		})
		if err != nil {
			httpx.Error(w, errs.Internal("encode session"))
			return
		}
		httpx.JSON(w, http.StatusOK, json.RawMessage(payload))
	})
}
