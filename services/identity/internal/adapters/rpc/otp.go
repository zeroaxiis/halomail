package rpc

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aashishrajdev/halomail/services/shared/errs"
	identityv1 "github.com/aashishrajdev/halomail/services/shared/gen/halomail/identity/v1"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
	"google.golang.org/protobuf/encoding/protojson"
)

// MountOTP exposes verification and an authenticated, server-only delivery API.
func (h *Handlers) MountOTP(mux *http.ServeMux, deliverySecret string) {
	for _, action := range []string{"request", "verify", "cancel"} {
		mux.HandleFunc("POST /v1/auth/otp/"+action, func(w http.ResponseWriter, r *http.Request) {
			if action != "verify" && (len(deliverySecret) < 32 || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-OTP-Delivery-Secret")), []byte(deliverySecret)) != 1) {
				httpx.Error(w, errs.Unauthorized("mail delivery authentication required"))
				return
			}
			var body struct {
				Email       string `json:"email"`
				Password    string `json:"password"`
				ChallengeID string `json:"challengeId"`
				Code        string `json:"code"`
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
			switch action {
			case "request":
				result, err := h.app.RequestLoginOTP(r.Context(), body.Email, body.Password)
				if err != nil {
					httpx.Error(w, err)
					return
				}
				httpx.JSON(w, http.StatusOK, result)
			case "cancel":
				if err := h.app.CancelLoginOTP(r.Context(), body.ChallengeID); err != nil {
					httpx.Error(w, err)
					return
				}
				httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
			case "verify":
				result, err := h.app.VerifyLoginOTP(r.Context(), body.ChallengeID, body.Code)
				if err != nil {
					httpx.Error(w, err)
					return
				}
				payload, err := protojson.Marshal(&identityv1.LoginResponse{User: toProtoUser(result.User), Session: toProtoSession(result.Session)})
				if err != nil {
					httpx.Error(w, errs.Internal("encode session"))
					return
				}
				httpx.JSON(w, http.StatusOK, json.RawMessage(payload))
			}
		})
	}
}
