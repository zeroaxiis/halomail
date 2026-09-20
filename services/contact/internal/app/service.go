package app

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/aashishrajdev/halomail/services/contact/internal/domain"
	"github.com/aashishrajdev/halomail/services/contact/internal/spam"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/idgen"
	"github.com/aashishrajdev/halomail/services/shared/ratelimit"
)

type Service struct {
	forms     FormRepo
	messages  MessageRepo
	limiter   ratelimit.Limiter
	forwarder Forwarder
	now       func() time.Time
}

func New(r Repos, limiter ratelimit.Limiter, forwarder Forwarder) *Service {
	return &Service{
		forms:     r.Forms,
		messages:  r.Messages,
		limiter:   limiter,
		forwarder: forwarder,
		now:       time.Now,
	}
}

// ---- Forms ---------------------------------------------------------------

type FormInput struct {
	Name           string
	Slug           string
	TargetEmail    string
	SpamProtection string
	RedirectURL    string
	Fields         []domain.FormField
}

func (s *Service) CreateForm(ctx context.Context, ownerID string, in FormInput) (*domain.Form, error) {
	if strings.HasPrefix(in.Slug, "inbox_") {
		return nil, errs.Invalid("this slug is reserved")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.Invalid("form name is required")
	}
	slug := in.Slug
	if slug == "" {
		slug = slugify(in.Name)
	}
	sp := in.SpamProtection
	if sp == "" {
		sp = domain.SpamHoneypot
	}
	f := &domain.Form{
		ID:             idgen.Prefixed("frm_"),
		OwnerID:        ownerID,
		Name:           in.Name,
		Slug:           slug,
		TargetEmail:    in.TargetEmail,
		SpamProtection: sp,
		RedirectURL:    in.RedirectURL,
		Fields:         in.Fields,
		Active:         true,
		CreatedAt:      s.now(),
	}
	if err := s.forms.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) GetForm(ctx context.Context, id, slug string) (*domain.Form, error) {
	if id != "" {
		return s.forms.GetByID(ctx, id)
	}
	if slug != "" {
		return s.forms.GetBySlug(ctx, slug)
	}
	return nil, errs.Invalid("id or slug is required")
}

func (s *Service) ListForms(ctx context.Context, ownerID string) ([]domain.Form, error) {
	return s.forms.ListByOwner(ctx, ownerID)
}

func (s *Service) UpdateForm(ctx context.Context, ownerID, id string, in FormInput, active bool) (*domain.Form, error) {
	f, err := s.forms.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if f.OwnerID != ownerID {
		return nil, errs.Forbidden("not your form")
	}
	if in.Name != "" {
		f.Name = in.Name
	}
	if in.TargetEmail != "" {
		f.TargetEmail = in.TargetEmail
	}
	if in.SpamProtection != "" {
		f.SpamProtection = in.SpamProtection
	}
	if in.RedirectURL != "" {
		f.RedirectURL = in.RedirectURL
	}
	if in.Fields != nil {
		f.Fields = in.Fields
	}
	f.Active = active
	if err := s.forms.Update(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) DeleteForm(ctx context.Context, ownerID, id string) error {
	return s.forms.Delete(ctx, id, ownerID)
}

// ---- Messages ------------------------------------------------------------

type SubmitInput struct {
	SenderName  string
	SenderEmail string
	Data        map[string]string
	Honeypot    string
	IP          string
	UserAgent   string
}

type SubmitResult struct {
	Message     *domain.Message
	RedirectURL string
}

// SubmitMessage is the public submit path.
func (s *Service) SubmitMessage(ctx context.Context, ownerID string, in SubmitInput) (*SubmitResult, error) {
	if len(in.Data) > 50 || len(in.SenderName) > 200 {
		return nil, errs.Invalid("submission is too large")
	}
	for key, value := range in.Data {
		if len(key) > 100 || len(value) > 10000 {
			return nil, errs.Invalid("form field is too large")
		}
	}
	if in.SenderEmail != "" {
		parsed, parseErr := mail.ParseAddress(in.SenderEmail)
		if parseErr != nil || parsed.Address != in.SenderEmail || len(in.SenderEmail) > 254 || strings.ContainsAny(in.SenderEmail, "\r\n") {
			return nil, errs.Invalid("invalid sender email")
		}
	}
	form, err := s.forms.Inbox(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !form.Active {
		return nil, errs.Invalid("this form is not accepting submissions")
	}

	if s.limiter != nil {
		key := "submit:" + form.ID + ":" + in.IP
		if ok, _ := s.limiter.Allow(ctx, key); !ok {
			return nil, errs.RateLimited("too many submissions, please slow down")
		}
	}

	score := 0.0
	isSpam := false
	if strings.TrimSpace(in.Honeypot) != "" {
		// A filled honeypot means a bot — flag without scoring.
		score, isSpam = 1, true
	} else {
		score = spam.Score(in.SenderName, in.SenderEmail, values(in.Data))
		isSpam = spam.IsSpam(score)
	}

	msg := &domain.Message{
		ID:          idgen.Prefixed("msg_"),
		FormID:      form.ID,
		OwnerID:     form.OwnerID,
		SenderName:  in.SenderName,
		SenderEmail: in.SenderEmail,
		Data:        in.Data,
		IP:          in.IP,
		UserAgent:   in.UserAgent,
		SpamScore:   score,
		IsSpam:      isSpam,
		CreatedAt:   s.now(),
	}
	if err := s.messages.Create(ctx, msg); err != nil {
		return nil, err
	}

	if !isSpam && s.forwarder != nil {
		// Best-effort: a forwarding failure must not fail the submission.
		_ = s.forwarder.Forward(ctx, form, msg)
	}

	return &SubmitResult{Message: msg, RedirectURL: form.RedirectURL}, nil
}

func (s *Service) ListMessages(ctx context.Context, ownerID, formID string, unreadOnly bool, limit, offset int) ([]domain.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.messages.List(ctx, ownerID, formID, unreadOnly, limit, offset)
}

func (s *Service) GetMessage(ctx context.Context, ownerID, id string) (*domain.Message, error) {
	m, err := s.messages.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.OwnerID != ownerID {
		return nil, errs.NotFound("message not found")
	}
	return m, nil
}

func (s *Service) MarkRead(ctx context.Context, ownerID, id string, read bool) error {
	return s.messages.MarkRead(ctx, id, ownerID, read)
}

func (s *Service) DeleteMessage(ctx context.Context, ownerID, id string) error {
	return s.messages.Delete(ctx, id, ownerID)
}

func (s *Service) GetUsageStats(ctx context.Context, ownerID string) (*domain.UsageStats, error) {
	return s.messages.GetUsageStats(ctx, ownerID)
}

// ---- helpers -------------------------------------------------------------

func values(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func slugify(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r == ' ' || r == '-' || r == '_':
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "form"
	}
	return out + "-" + idgen.New()[:6]
}
