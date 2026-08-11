package models

import (
	"fmt"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID uint `json:"id" gorm:"primaryKey"`
	// Email is the user's login identifier.
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"`
	// NotificationEmail is where intervention report emails are sent. It
	// defaults to Email but can be changed independently from the Settings
	// page, and never follows Email automatically once set.
	//
	// Not DB-enforced NOT NULL: the column was added after users already
	// existed in production, and with only a couple of rows it's simpler to
	// backfill them by hand (see internal/database/gorm.go) than to script a
	// nullable->backfill->tighten migration in code. Enforced at the app
	// level instead, via BeforeCreate below and SetNotificationEmail.
	NotificationEmail string         `json:"notification_email"`
	FirstName         string         `json:"first_name" gorm:"not null"`
	LastName          string         `json:"last_name" gorm:"not null"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate defaults NotificationEmail to Email when it hasn't been set
// explicitly, so every new user starts with notifications going to their
// login address without callers having to set it themselves.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.NotificationEmail == "" {
		u.NotificationEmail = u.Email
	}
	return nil
}

// SetNotificationEmail validates and sets the address intervention report
// emails are sent to. It is independent from Email: changing the login
// email does not change this, and vice versa.
func (u *User) SetNotificationEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid notification email address: %w", err)
	}
	u.NotificationEmail = email
	return nil
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}
