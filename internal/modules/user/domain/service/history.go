package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/sigif/sigif-go/internal/modules/user/domain/entity"
	"github.com/sigif/sigif-go/internal/modules/user/domain/repository"
	"github.com/sigif/sigif-go/internal/shared/clock"
	sharedErrors "github.com/sigif/sigif-go/internal/shared/errors"
)

type HistoryService struct {
	historyRepo repository.UserHistoryRepository
	userRepo    repository.UserRepository
	clock       clock.Clock
}

func NewHistoryService(
	historyRepo repository.UserHistoryRepository,
	userRepo repository.UserRepository,
	clock clock.Clock,
) *HistoryService {
	return &HistoryService{
		historyRepo: historyRepo,
		userRepo:    userRepo,
		clock:       clock,
	}
}

type RecordHistoryInput struct {
	UserID  uuid.UUID
	ActorID *uuid.UUID
	Action  entity.HistoryAction
	Changes entity.HistoryChanges
}

func (s *HistoryService) Record(ctx context.Context, in RecordHistoryInput) error {
	actor, err := s.resolveActor(ctx, in.ActorID)
	if err != nil {
		return err
	}
	return s.historyRepo.Create(ctx, entity.NewUserHistory(s.clock, in.UserID, in.Action, in.Changes, actor))
}

func (s *HistoryService) List(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.UserHistory, int64, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	if user == nil {
		return nil, 0, sharedErrors.New(sharedErrors.CodeNotFound, "user not found", 404)
	}
	return s.historyRepo.ListByUserID(ctx, userID, offset, limit)
}

func (s *HistoryService) resolveActor(ctx context.Context, actorID *uuid.UUID) (*entity.HistoryActor, error) {
	if actorID == nil || *actorID == uuid.Nil {
		return nil, nil
	}
	actor := &entity.HistoryActor{ID: *actorID}
	user, err := s.userRepo.GetByID(ctx, *actorID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		actor.FullName = strings.TrimSpace(user.FirstName + " " + user.LastName)
		actor.Email = user.Email
	}
	return actor, nil
}
