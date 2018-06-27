package postgres

import "github.com/WasinWatt/slumbot/cache"

func New(c cache.Cacher) *Repository {
	return &Repository{c}
}

type Repository struct {
	memcache cache.Cacher
}
