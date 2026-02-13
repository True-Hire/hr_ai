package application

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/ruziba3vich/hr-ai/internal/infrastructure/gemini"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/qdrant"
)

type SearchService struct {
	qdrantClient *qdrant.Client
	geminiClient *gemini.Client
	userSvc      *UserService
}

func NewSearchService(
	qdrantClient *qdrant.Client,
	geminiClient *gemini.Client,
	userSvc *UserService,
) *SearchService {
	return &SearchService{
		qdrantClient: qdrantClient,
		geminiClient: geminiClient,
		userSvc:      userSvc,
	}
}

type SearchResult struct {
	UserID uuid.UUID
	Score  float64
}

func (s *SearchService) SearchUsers(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Translate query to English
	englishQuery, err := s.geminiClient.TranslateToEnglish(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("translate query: %w", err)
	}

	// Embed the English query
	vector, err := s.geminiClient.EmbedText(ctx, englishQuery)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	// Search Qdrant — request more results to account for multiple vectors per user
	qdrantLimit := max(limit*4, 20)

	scored, err := s.qdrantClient.Search(ctx, collectionName, vector, qdrantLimit)
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}

	// Group by user_id and aggregate weighted scores
	userScores := make(map[string]float64)
	for _, sp := range scored {
		userIDStr, _ := sp.Payload["user_id"].(string)
		if userIDStr == "" {
			continue
		}
		weight := 1.0
		if w, ok := sp.Payload["weight"].(float64); ok {
			weight = w
		}
		userScores[userIDStr] += sp.Score * weight
	}

	// Sort by score descending
	type userScore struct {
		id    string
		score float64
	}
	sorted := make([]userScore, 0, len(userScores))
	for id, score := range userScores {
		sorted = append(sorted, userScore{id: id, score: score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	// Take top N
	if len(sorted) > limit {
		sorted = sorted[:limit]
	}

	results := make([]SearchResult, 0, len(sorted))
	for _, us := range sorted {
		uid, err := uuid.Parse(us.id)
		if err != nil {
			continue
		}
		results = append(results, SearchResult{
			UserID: uid,
			Score:  us.score,
		})
	}

	return results, nil
}
