package mapper

import (
	"github.com/sigif/sigif-go/internal/modules/auth/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/auth/infrastructure/persistence/model"
)

func ToModel(token *entity.RefreshToken) *model.RefreshTokenModel {
	return &model.RefreshTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		UserAgent: token.UserAgent,
		IPAddress: token.IPAddress,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: token.RevokedAt,
		CreatedAt: token.CreatedAt,
	}
}

func ToDomain(m *model.RefreshTokenModel) *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		UserAgent: m.UserAgent,
		IPAddress: m.IPAddress,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
}

func ToDomainList(models []model.RefreshTokenModel) []*entity.RefreshToken {
	result := make([]*entity.RefreshToken, len(models))
	for i := range models {
		result[i] = ToDomain(&models[i])
	}
	return result
}
