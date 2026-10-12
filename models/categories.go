package models

// Category represents a product categgory in the catalog.
// It includes ID for internal use only, Code  and Name
type Category struct {
	ID   uint   `gorm:"primaryKey" json:"-"`
	Code string `gorm:"uniqueIndex;not null" json:"code"`
	Name string `gorm:"not null" json:"name"`
}

func (v *Category) TableName() string {
	return "categories"
}
