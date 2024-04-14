package handler

import (
	"avito-testovoe/internal/cache"
	sqlite "avito-testovoe/internal/storage"
	"context"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"sync"
)

type StorageI interface {
	GetBannerFromStorage(query sqlite.Query, ctx context.Context) (content string, active bool, err error)
	GetAllBannersFromStorage(query sqlite.Query, ctx context.Context) (banners []sqlite.Banner, err error)
	PostBannerToStorage(banner sqlite.Banner, ctx context.Context) (id int, err error)
	UpdateBannerInStorage(banner sqlite.BannerUpdate, ctx context.Context) (err error)
	DeleteBannerFromStorage(id int, ctx context.Context) (keys []string, err error)
	DeleteBannerFromStorageByFeature(featureId int, ctx context.Context) (keys []string, err error)
	DeleteBannerFromStorageByTag(tag int, ctx context.Context) (keys []string, err error)
	GetBannerVersionsFromStorage(id int, ctx context.Context) (banners []sqlite.Banner, err error)
	CheckToken(token string, ctx context.Context) (role string, err error)
}

type Handler struct {
	rwMu sync.RWMutex
	S    StorageI
	Log  *slog.Logger
	C    *cache.Cache
	Ctx  context.Context
}

func NewServer(log *slog.Logger, storage *sqlite.Storage, c *cache.Cache, ctx context.Context) http.Handler {
	h := Handler{
		S:   storage,
		Log: log,
		C:   c,
		Ctx: ctx,
	}

	r := chi.NewRouter()

	r.Get("/user_banner", h.GetBanner)
	r.Get("/banner", h.GetAllBanners)
	r.Post("/banner", h.PostBanner)
	r.Patch("/banner/{id}", h.PatchBanner)
	r.Delete("/banner/{id}", h.DeleteBanner)
	r.Delete("/banner", h.DeleteBannerByTagOrFeature)
	r.Get("/banner/{id}", h.GetBannerVersions)

	return r
}
