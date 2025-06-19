package message

import (
	"context"
)

type Repository interface {
	GetNextUnsent(ctx context.Context) (*Message, error)
	GetAllUnsent(ctx context.Context) ([]*Message, error)
	GetAllSent(ctx context.Context) ([]*SentMessage, error)
	Save(ctx context.Context, msg *Message) error
}

type RepositoryMiddleware func(Repository) Repository

func RepositoryWithMiddleware(repo Repository, mws ...RepositoryMiddleware) Repository {
	r := repo

	// middleware is applied in reverse
	for i := len(mws) - 1; i >= 0; i-- {
		r = mws[i](r)
	}
	return r
}
