package adapters

import (
	gormadapter "github.com/isauran/logger/adapters/gorm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ExampleGormLogger() {
	// Initialize GORM logger adapter
	gormLogger := gormadapter.NewLogger("debug")

	// Initialize GORM with our logger
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the schema - this will log the migration SQL
	db.AutoMigrate(&User{})

	// Create a new user - this will generate INSERT SQL logs
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	db.Create(&user)

	// Query operations - these will generate SELECT SQL logs
	var foundUser User
	db.First(&foundUser, "email = ?", "john@example.com")

	// Update operation - this will generate UPDATE SQL logs
	db.Model(&foundUser).Update("name", "John Smith")

	// Delete operation - this will generate DELETE SQL logs
	db.Delete(&foundUser)
}
