package db

import (
	"context"

	"github.com/milkymilky0116/jellyfish/internal/repository"
)

type IRepository interface {
	GetEmailById(context.Context, int64) (repository.Email, error)
	CreateEmail(context.Context, repository.CreateEmailParams) (repository.Email, error)
	CreateCategory(context.Context, repository.CreateCategoryParams) (repository.Category, error)
	GetCategory(context.Context, string) (repository.Category, error)
	RegisterEmailAndCategory(context.Context, repository.RegisterEmailAndCategoryParams) error
}
