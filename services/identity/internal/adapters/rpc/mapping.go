package rpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	identityv1 "github.com/zeroaxiis/halomail/services/shared/gen/halomail/identity/v1"

	"github.com/zeroaxiis/halomail/services/identity/internal/app"
	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
)

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func toProtoUser(u *domain.User) *identityv1.User {
	if u == nil {
		return nil
	}
	var deletionScheduledAt, scheduledDeleteAt *timestamppb.Timestamp
	if u.DeletionScheduledAt != nil {
		deletionScheduledAt = timestamppb.New(*u.DeletionScheduledAt)
	}
	if u.ScheduledDeleteAt != nil {
		scheduledDeleteAt = timestamppb.New(*u.ScheduledDeleteAt)
	}
	return &identityv1.User{
		Id:                  u.ID,
		Email:               u.Email,
		Name:                u.Name,
		Handle:              u.Handle,
		AvatarUrl:           u.AvatarURL,
		Timezone:            u.Timezone,
		OrgId:               u.OrgID,
		Role:                u.Role,
		CreatedAt:           ts(u.CreatedAt),
		UpdatedAt:           ts(u.UpdatedAt),
		DeletionScheduledAt: deletionScheduledAt,
		ScheduledDeleteAt:   scheduledDeleteAt,
	}
}

func toProtoTeamMember(u domain.User) *identityv1.TeamMember {
	return &identityv1.TeamMember{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Handle:    u.Handle,
		AvatarUrl: u.AvatarURL,
		Role:      u.Role,
		CreatedAt: ts(u.CreatedAt),
	}
}

func toProtoInvitation(i domain.Invitation) *identityv1.Invitation {
	return &identityv1.Invitation{
		Id:            i.ID,
		Email:         i.Email,
		Role:          i.Role,
		InvitedByName: i.InvitedByName,
		Status:        i.Status,
		ExpiresAt:     ts(i.ExpiresAt),
		CreatedAt:     ts(i.CreatedAt),
		Token:         i.Token,
	}
}

func toProtoSession(s app.IssuedSession) *identityv1.Session {
	return &identityv1.Session{
		AccessToken:     s.AccessToken,
		RefreshToken:    s.RefreshToken,
		AccessExpiresAt: ts(s.AccessExpiresAt),
	}
}

func toProtoAPIKey(k domain.APIKey) *identityv1.ApiKey {
	var lastUsed *timestamppb.Timestamp
	if k.LastUsedAt != nil {
		lastUsed = timestamppb.New(*k.LastUsedAt)
	}
	return &identityv1.ApiKey{
		Id:         k.ID,
		Name:       k.Name,
		Prefix:     k.Prefix,
		LastFour:   k.LastFour,
		Scopes:     k.Scopes,
		CreatedAt:  ts(k.CreatedAt),
		LastUsedAt: lastUsed,
		Revoked:    k.Revoked,
	}
}

func toProtoAudit(l domain.AuditLog) *identityv1.AuditLog {
	return &identityv1.AuditLog{
		Id:         l.ID,
		ActorId:    l.ActorID,
		Action:     l.Action,
		TargetType: l.TargetType,
		TargetId:   l.TargetID,
		Metadata:   l.Metadata,
		Ip:         l.IP,
		CreatedAt:  ts(l.CreatedAt),
	}
}
