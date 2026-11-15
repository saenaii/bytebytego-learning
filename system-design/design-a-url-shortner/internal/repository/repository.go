package repository

import "context"

type Record struct {
	ID            int64  `json:"id"`
	ShortCode     string `json:"short_code"`
	OriginalURL   string `json:"original_url"`
	NormalizedURL string `json:"normalized_url"`
	Mode          string `json:"mode"`
}

type Store interface {
	FindByKey(context.Context, string, string) (Record, bool, error)
	CreateOrGet(context.Context, Record) (Record, bool, error)
	FindByCode(context.Context, string) (Record, bool, error)
	Ping(context.Context) error
}

type Cache interface {
	Get(context.Context, string) (string, bool, error)
	Set(context.Context, string, string) error
	Ping(context.Context) error
}
