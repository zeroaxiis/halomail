package rpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	contactv1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/contact/v1"

	"github.com/zeroaxiis/halomail/services/contact/internal/domain"
)

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func fieldTypeToProto(s string) contactv1.FieldType {
	switch s {
	case domain.FieldText:
		return contactv1.FieldType_FIELD_TYPE_TEXT
	case domain.FieldEmail:
		return contactv1.FieldType_FIELD_TYPE_EMAIL
	case domain.FieldTextarea:
		return contactv1.FieldType_FIELD_TYPE_TEXTAREA
	case domain.FieldSelect:
		return contactv1.FieldType_FIELD_TYPE_SELECT
	case domain.FieldNumber:
		return contactv1.FieldType_FIELD_TYPE_NUMBER
	default:
		return contactv1.FieldType_FIELD_TYPE_UNSPECIFIED
	}
}

func fieldTypeFromProto(t contactv1.FieldType) string {
	switch t {
	case contactv1.FieldType_FIELD_TYPE_TEXT:
		return domain.FieldText
	case contactv1.FieldType_FIELD_TYPE_EMAIL:
		return domain.FieldEmail
	case contactv1.FieldType_FIELD_TYPE_TEXTAREA:
		return domain.FieldTextarea
	case contactv1.FieldType_FIELD_TYPE_SELECT:
		return domain.FieldSelect
	case contactv1.FieldType_FIELD_TYPE_NUMBER:
		return domain.FieldNumber
	default:
		return domain.FieldText
	}
}

func spamProtToProto(s string) contactv1.SpamProtection {
	switch s {
	case domain.SpamNone:
		return contactv1.SpamProtection_SPAM_PROTECTION_NONE
	case domain.SpamHoneypot:
		return contactv1.SpamProtection_SPAM_PROTECTION_HONEYPOT
	case domain.SpamRecaptcha:
		return contactv1.SpamProtection_SPAM_PROTECTION_RECAPTCHA
	default:
		return contactv1.SpamProtection_SPAM_PROTECTION_UNSPECIFIED
	}
}

func spamProtFromProto(s contactv1.SpamProtection) string {
	switch s {
	case contactv1.SpamProtection_SPAM_PROTECTION_NONE:
		return domain.SpamNone
	case contactv1.SpamProtection_SPAM_PROTECTION_HONEYPOT:
		return domain.SpamHoneypot
	case contactv1.SpamProtection_SPAM_PROTECTION_RECAPTCHA:
		return domain.SpamRecaptcha
	default:
		return ""
	}
}

func toProtoForm(f *domain.Form) *contactv1.Form {
	if f == nil {
		return nil
	}
	fields := make([]*contactv1.FormField, len(f.Fields))
	for i, ff := range f.Fields {
		fields[i] = &contactv1.FormField{
			Name:        ff.Name,
			Label:       ff.Label,
			Type:        fieldTypeToProto(ff.Type),
			Required:    ff.Required,
			Placeholder: ff.Placeholder,
			Options:     ff.Options,
		}
	}
	return &contactv1.Form{
		Id:             f.ID,
		OwnerId:        f.OwnerID,
		Name:           f.Name,
		Slug:           f.Slug,
		TargetEmail:    f.TargetEmail,
		SpamProtection: spamProtToProto(f.SpamProtection),
		RedirectUrl:    f.RedirectURL,
		Fields:         fields,
		Active:         f.Active,
		CreatedAt:      ts(f.CreatedAt),
	}
}

func fieldsFromProto(in []*contactv1.FormField) []domain.FormField {
	out := make([]domain.FormField, len(in))
	for i, ff := range in {
		out[i] = domain.FormField{
			Name:        ff.GetName(),
			Label:       ff.GetLabel(),
			Type:        fieldTypeFromProto(ff.GetType()),
			Required:    ff.GetRequired(),
			Placeholder: ff.GetPlaceholder(),
			Options:     ff.GetOptions(),
		}
	}
	return out
}

func toProtoMessage(m *domain.Message) *contactv1.Message {
	if m == nil {
		return nil
	}
	return &contactv1.Message{
		Id:          m.ID,
		FormId:      m.FormID,
		OwnerId:     m.OwnerID,
		SenderName:  m.SenderName,
		SenderEmail: m.SenderEmail,
		Data:        m.Data,
		Ip:          m.IP,
		UserAgent:   m.UserAgent,
		SpamScore:   m.SpamScore,
		IsSpam:      m.IsSpam,
		Read:        m.Read,
		CreatedAt:   ts(m.CreatedAt),
	}
}
