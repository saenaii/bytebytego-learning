package service

import (
	"context"
	"testing"

	"github.com/bytebytego-learning/design-a-url-shortner/internal/metrics"
	"github.com/bytebytego-learning/design-a-url-shortner/internal/repository"
)

type fakeStore struct {
	record   repository.Record
	found    bool
	created  bool
	findCode bool
}

func (f *fakeStore) FindByKey(context.Context, string, string) (repository.Record, bool, error) {
	return f.record, f.found, nil
}
func (f *fakeStore) CreateOrGet(_ context.Context, record repository.Record) (repository.Record, bool, error) {
	if f.created {
		return f.record, true, nil
	}
	record.ID = 1
	record.ShortCode = "1"
	f.record = record
	f.record.ShortCode = "1"
	return f.record, false, nil
}
func (f *fakeStore) FindByCode(context.Context, string) (repository.Record, bool, error) {
	return f.record, f.findCode, nil
}
func (f *fakeStore) Ping(context.Context) error { return nil }

type fakeCache struct {
	value string
	found bool
}

func (f *fakeCache) Get(context.Context, string) (string, bool, error) { return f.value, f.found, nil }
func (f *fakeCache) Set(context.Context, string, string) error         { return nil }
func (f *fakeCache) Ping(context.Context) error                        { return nil }

func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		found    bool
		wantCode string
		wantOld  bool
	}{
		{name: "creates new", wantCode: "1"},
		{name: "returns existing", found: true, wantCode: "existing", wantOld: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{found: tt.found, record: repository.Record{ShortCode: "existing", OriginalURL: "https://example.com", NormalizedURL: "https://example.com", Mode: "full"}}
			svc := New(store, &fakeCache{}, &metrics.Metrics{}, "http://localhost:8080", "full", nil)
			response, err := svc.Create(context.Background(), CreateRequest{URL: "https://example.com"})
			if err != nil {
				t.Fatal(err)
			}
			if response.ShortCode != tt.wantCode || response.Existing != tt.wantOld {
				t.Fatalf("response = %+v", response)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		cacheHit  bool
		storeHit  bool
		wantError bool
	}{
		{name: "cache hit", cacheHit: true},
		{name: "database hit", storeHit: true},
		{name: "missing", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{findCode: tt.storeHit, record: repository.Record{OriginalURL: "https://example.com"}}
			cache := &fakeCache{value: "https://cached.example.com", found: tt.cacheHit}
			svc := New(store, cache, &metrics.Metrics{}, "http://localhost:8080", "full", nil)
			_, err := svc.Resolve(context.Background(), "abc")
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
