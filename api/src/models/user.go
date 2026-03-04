package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the relational identity of a bank client in PostgreSQL.
// It maps the "users" table created in the init.sql script.
type Users struct {
	UUIDUser      uuid.UUID      `gorm:"column:uuid_user;primaryKey"`
	UserFullname  string         `gorm:"column:user_fullname"`
	UserEmail     string         `gorm:"column:user_email"`
	UserPassword  string         `gorm:"column:user_password"`
	TBAccountID   string         `gorm:"column:tb_account_id;type:jsonb"`
	UserCreatedAt time.Time      `gorm:"column:user_created_at"`
	UserUpdatedAt time.Time      `gorm:"column:user_updated_at"`
	UserDeletedAt gorm.DeletedAt `gorm:"column:user_deleted_at"`
}

// BeforeCreate is a GORM Hook that ensures a UUID is generated before insertion.
// This provides a fallback if the database driver doesn't handle the default.
func (u *Users) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UUIDUser == uuid.Nil {
		u.UUIDUser = uuid.New()
	}
	return
}
