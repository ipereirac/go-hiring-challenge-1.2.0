package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type CategoriesRepository interface {
	GetAll(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, c *Category) error
}

type sqlCategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) CategoriesRepository {
	return &sqlCategoriesRepository{
		db: db,
	}
}

func (r *sqlCategoriesRepository) GetAll(ctx context.Context) ([]Category, error) {
	var categories []Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *sqlCategoriesRepository) Create(ctx context.Context, c *Category) error {
	err := r.db.WithContext(ctx).Create(c).Error
	if err != nil {
		if isDuplicateKeyError(err) {
			return errors.New("category code already exists")
		}
	}

	return err
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
