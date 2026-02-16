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

// ResumeEntry represents a piece of resume data collected from the user.
type ResumeEntry struct {
	Type     string `json:"type"`      // "text", "file"
	Text     string `json:"text"`      // for type=text
	Data     string `json:"data"`      // base64-encoded for type=file
	MimeType string `json:"mime_type"` // for type=file
}

func (s *BotStateService) AddResumeEntry(ctx context.Context, telegramID string, entry *ResumeEntry) error {
	key := fmt.Sprintf("bot:resume:%s", telegramID)
	entries, _ := s.GetResumeEntries(ctx, telegramID)
	entries = append(entries, *entry)
	return s.redis.Set(ctx, key, entries, s.ttl)
}

func (s *BotStateService) GetResumeEntries(ctx context.Context, telegramID string) ([]ResumeEntry, error) {
	key := fmt.Sprintf("bot:resume:%s", telegramID)
	var entries []ResumeEntry
	found, err := s.redis.Get(ctx, key, &entries)
	if err != nil {
		return nil, fmt.Errorf("get resume entries: %w", err)
	}
	if !found {
		return nil, nil
	}
	return entries, nil
}

func (s *BotStateService) ClearResumeEntries(ctx context.Context, telegramID string) error {
	key := fmt.Sprintf("bot:resume:%s", telegramID)
	return s.redis.Delete(ctx, key)
}
