package models

import (
	"context"

	"gorm.io/gorm"
)

type ProductsRepository interface {
	//GetAll(ctx context.Context) ([]Product, error)
	GetPaginated(ctx context.Context, pageSize, offset int, category string, priceLessThan *float64) (int64, []Product, error)
	GetByCode(ctx context.Context, code string) (*Product, error)
}

type sqlProductRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) ProductsRepository {
	return &sqlProductRepository{
		db: db,
	}
}

func (r *sqlProductRepository) GetAll(ctx context.Context) ([]Product, error) {
	var products []Product

	err := r.db.WithContext(ctx).Preload("Variants").Preload("Category").Find(&products).Error
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *sqlProductRepository) GetPaginated(ctx context.Context, pageSize, offset int, category string, priceLessThan *float64) (int64, []Product, error) {
	var products []Product
	var count int64

	query := r.db.WithContext(ctx).Model(&Product{})

	if category != "" {
		query = query.Joins("Join categories ON categories.id = products.category_id").
			Where("categories.name = ?", category)
	}

	if priceLessThan != nil {
		query = query.Where("products.price < ?", *priceLessThan)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, nil, err
	}

	err := query.Limit(pageSize).Offset(offset).Preload("Variants").Preload("Category").Find(&products).Error
	if err != nil {
		return 0, nil, err
	}

	return count, products, nil
}

func (r *sqlProductRepository) GetByCode(ctx context.Context, code string) (*Product, error) {
	var p Product

	err := r.db.WithContext(ctx).Preload("Variants").Preload("Category").Where("code = ?", code).First(&p).Error
	if err != nil {
		return nil, err
	}

	return &p, nil
}
