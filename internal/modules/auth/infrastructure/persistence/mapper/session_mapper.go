package mapper

import (
	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/model"
)

func SessionToModel(s *entity.UserSession) *model.SessionModel {
	var closeReason *string
	if s.CloseReason != nil {
		v := string(*s.CloseReason)
		closeReason = &v
	}

	return &model.SessionModel{
		ID:          s.ID,
		UserID:      s.UserID,
		TokenHash:   s.TokenHash,
		IPAddress:   s.IPAddress,
		Device:      s.Device,
		StartedAt:   s.StartedAt,
		ExpiresAt:   s.ExpiresAt,
		EndedAt:     s.EndedAt,
		CloseReason: closeReason,
	}
}

func SessionToDomain(m *model.SessionModel) *entity.UserSession {
	var closeReason *entity.SessionCloseReason
	if m.CloseReason != nil {
		v := entity.SessionCloseReason(*m.CloseReason)
		closeReason = &v
	}

	return &entity.UserSession{
		ID:          m.ID,
		UserID:      m.UserID,
		TokenHash:   m.TokenHash,
		IPAddress:   m.IPAddress,
		Device:      m.Device,
		StartedAt:   m.StartedAt,
		ExpiresAt:   m.ExpiresAt,
		EndedAt:     m.EndedAt,
		CloseReason: closeReason,
	}
}

func LoginAttemptToModel(a *entity.LoginAttempt) *model.LoginAttemptModel {
	return &model.LoginAttemptModel{
		ID:            a.ID,
		Email:         a.Email,
		CompanyID:     a.CompanyID,
		IPAddress:     a.IPAddress,
		Successful:    a.Successful,
		FailureReason: a.FailureReason,
		CreatedAt:     a.CreatedAt,
	}
}
