package adapters

import "time"

// User is a sample model for demonstration
type User struct {
	ID        uint `gorm:"primarykey"`
	Name      string
	Email     string
	CreatedAt time.Time
}
