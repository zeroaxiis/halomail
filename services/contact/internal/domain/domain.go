// Package domain holds contact-form entities. Pure: no I/O.
package domain

import "time"

// Field types.
const (
	FieldText     = "text"
	FieldEmail    = "email"
	FieldTextarea = "textarea"
	FieldSelect   = "select"
	FieldNumber   = "number"
)

// Spam-protection modes.
const (
	SpamNone      = "none"
	SpamHoneypot  = "honeypot"
	SpamRecaptcha = "recaptcha"
)

// FormField describes one input of a contact form.
type FormField struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Placeholder string   `json:"placeholder"`
	Options     []string `json:"options,omitempty"`
}

// Form is an embeddable contact form.
type Form struct {
	ID             string
	OwnerID        string
	Name           string
	Slug           string
	TargetEmail    string
	SpamProtection string
	RedirectURL    string
	Fields         []FormField
	Active         bool
	CreatedAt      time.Time
}

// Message is a single submission.
type Message struct {
	ID          string
	FormID      string
	OwnerID     string
	SenderName  string
	SenderEmail string
	Data        map[string]string
	IP          string
	UserAgent   string
	SpamScore   float64
	IsSpam      bool
	Read        bool
	CreatedAt   time.Time
}

// UsageStats holds aggregate metrics for the dashboard.
type UsageStats struct {
	TotalMessages  int
	UnreadMessages int
	SpamPrevented  int
}
