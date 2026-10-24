package app

import (
	"context"

	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
)

func (s *Service) ProcessRefunds(ctx context.Context, userID, orgID, orderID, paymentID, currency string, paymentAmount int64, refunds []domain.BillingRefund) (*domain.User, domain.RefundResult, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.RefundResult{}, err
	}
	if orgID != "" && orgID != user.OrgID {
		return nil, domain.RefundResult{}, errs.Forbidden("refund account mismatch")
	}
	result, err := s.users.ApplyRefunds(ctx, user.OrgID, orderID, paymentID, currency, paymentAmount, refunds)
	return user, result, err
}
