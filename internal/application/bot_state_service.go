package application

import (
	"context"
	"fmt"
	"time"

	"github.com/ruziba3vich/hr-ai/internal/domain"
	redisclient "github.com/ruziba3vich/hr-ai/internal/infrastructure/redis"
)

type BotStateService struct {
	redis *redisclient.Client
	ttl   time.Duration
}

func NewBotStateService(redis *redisclient.Client) *BotStateService {
	return &BotStateService{
		redis: redis,
		ttl:   24 * time.Hour,
	}
}

func (s *BotStateService) SetState(ctx context.Context, telegramID, state string) error {
	key := fmt.Sprintf("bot:state:%s", telegramID)
	bs := &domain.BotState{
		State: state,
		Data:  make(map[string]string),
	}
	return s.redis.Set(ctx, key, bs, s.ttl)
}

func (s *BotStateService) SetStateWithData(ctx context.Context, telegramID, state string, data map[string]string) error {
	key := fmt.Sprintf("bot:state:%s", telegramID)
	bs := &domain.BotState{
		State: state,
		Data:  data,
	}
	return s.redis.Set(ctx, key, bs, s.ttl)
}

func (s *BotStateService) GetState(ctx context.Context, telegramID string) (*domain.BotState, error) {
	key := fmt.Sprintf("bot:state:%s", telegramID)
	var state domain.BotState
	found, err := s.redis.Get(ctx, key, &state)
	if err != nil {
		return nil, fmt.Errorf("get bot state: %w", err)
	}
	if !found {
		return nil, nil
	}
	return &state, nil
}

func (s *BotStateService) ClearState(ctx context.Context, telegramID string) error {
	key := fmt.Sprintf("bot:state:%s", telegramID)
	return s.redis.Delete(ctx, key)
}
