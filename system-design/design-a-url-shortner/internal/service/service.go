package service

import (
	"context"
	"fmt"

	"github.com/bytebytego-learning/design-a-url-shortner/internal/metrics"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/repository"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/shortcode"
)

type Service struct {
	store       repository.Store
	cache       repository.Cache
	metrics     *metrics.Metrics
	baseURL     string
	defaultMode string
	blacklist   []string
}

type CreateRequest struct {
	URL  string `json:"url"`
	Mode string `json:"mode"`
}

type CreateResponse struct {
	ShortURL      string `json:"short_url"`
	ShortCode     string `json:"short_code"`
	OriginalURL   string `json:"original_url"`
	NormalizedURL string `json:"normalized_url"`
	Mode          string `json:"mode"`
	Existing      bool   `json:"existing"`
}

func New(store repository.Store, cache repository.Cache, collector *metrics.Metrics, baseURL, defaultMode string, blacklist []string) *Service {
	return &Service{store: store, cache: cache, metrics: collector, baseURL: baseURL, defaultMode: defaultMode, blacklist: blacklist}
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (CreateResponse, error) {
	mode := request.Mode
	if mode == "" {
		mode = s.defaultMode
	}
	normalized, err := shortcode.Normalize(request.URL, mode, s.blacklist)
	if err != nil {
		s.metrics.Inc("errors")
		return CreateResponse{}, err
	}
	existing, found, err := s.store.FindByKey(ctx, normalized, mode)
	if err != nil {
		s.metrics.Inc("errors")
		return CreateResponse{}, err
	}
	if found {
		s.metrics.Inc("shorten")
		return toResponse(existing, true, s.baseURL), nil
	}
	record, wasExisting, err := s.store.CreateOrGet(ctx, repository.Record{OriginalURL: request.URL, NormalizedURL: normalized, Mode: mode})
	if err != nil {
		s.metrics.Inc("errors")
		return CreateResponse{}, err
	}
	s.metrics.Inc("shorten")
	return toResponse(record, wasExisting, s.baseURL), nil
}

func (s *Service) Resolve(ctx context.Context, code string) (repository.Record, error) {
	s.metrics.Inc("resolve")
	key := cacheKey(code)
	if value, found, err := s.cache.Get(ctx, key); err == nil && found {
		s.metrics.Inc("cache_hit")
		return repository.Record{ShortCode: code, OriginalURL: value}, nil
	}
	s.metrics.Inc("cache_miss")
	record, found, err := s.store.FindByCode(ctx, code)
	if err != nil {
		s.metrics.Inc("errors")
		return repository.Record{}, err
	}
	if !found {
		return repository.Record{}, fmt.Errorf("short code %q not found", code)
	}
	if err := s.cache.Set(ctx, key, record.OriginalURL); err != nil {
		s.metrics.Inc("errors")
	}
	return record, nil
}

func (s *Service) Health(ctx context.Context) error {
	if err := s.store.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	if err := s.cache.Ping(ctx); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	return nil
}

func (s *Service) Metrics() metrics.Snapshot {
	return s.metrics.Snapshot()
}

func (s *Service) IncMetric(name string) {
	s.metrics.Inc(name)
}

func toResponse(record repository.Record, existing bool, baseURL string) CreateResponse {
	return CreateResponse{ShortURL: baseURL + "/shorturl/" + record.ShortCode, ShortCode: record.ShortCode, OriginalURL: record.OriginalURL, NormalizedURL: record.NormalizedURL, Mode: record.Mode, Existing: existing}
}

func cacheKey(code string) string {
	return "url-shortener:code:" + code
}
