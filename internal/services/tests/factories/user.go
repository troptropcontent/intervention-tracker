package factories

import (
	"github.com/troptropcontent/qr_code_maintenance/internal/models"
	"gorm.io/gorm"
)

type UserBuilder struct {
	user     *models.User
	password string
}

func NewUser() *UserBuilder {
	return &UserBuilder{
		user: &models.User{
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		},
		password: "password123",
	}
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithName(firstName, lastName string) *UserBuilder {
	b.user.FirstName = firstName
	b.user.LastName = lastName
	return b
}

func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.password = password
	return b
}

func (b *UserBuilder) WithIsActive(isActive bool) *UserBuilder {
	b.user.IsActive = isActive
	return b
}

func (b *UserBuilder) Build() *models.User {
	b.user.SetPassword(b.password)
	return b.user
}

func (b *UserBuilder) Create(db *gorm.DB) *models.User {
	b.user.SetPassword(b.password)
	db.Create(b.user)
	return b.user
}
