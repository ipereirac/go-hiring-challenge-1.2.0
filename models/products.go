package models

import (
	"github.com/shopspring/decimal"
)

// Product represents a product in the catalog.
// It includes a unique code and a price.
type Product struct {
	ID         uint            `gorm:"primaryKey"`
	Code       string          `gorm:"uniqueIndex;not null"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2);not null"`
	Variants   []Variant       `gorm:"foreignKey:ProductID"`
	CategoryID uint            `gorm:"not null"`              // Foreign key
	Category   Category        `gorm:"foreignKey:CategoryID"` // Relation
}

func (p *Product) TableName() string {
	return "products"
}
