package model

import (
	"errors"
	"github.com/phankanp/csv-to-json/auth"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/pborman/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"primary_key" json:"id"`
	AuthKey   string    `gorm:"not null;" json:"auth_key"`
	Username  string    `gorm:"size:255;not null;unique" json:"username"`
	Email     string    `gorm:"size:100;not null;unique" json:"email"`
	Password  string    `gorm:"not null;" json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (u *User) Prepare() {
	u.ID = uuid.NewRandom()
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(u.Email)
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

func (u *User) ValidateInput(db *gorm.DB) error {
	var err error

	if u.Username == "" {
		return errors.New("username is required")
	}

	if u.Password == "" {
		return errors.New("password is required")
	}

	if u.Email == "" {
		return errors.New("email is required")
	}

	if err = checkmail.ValidateFormat(u.Email); err != nil {
		return errors.New("invalid email format")
	}

	checkEmailInUse := &User{}
	err = db.Model(&User{}).Where("email = ?", u.Email).Take(checkEmailInUse).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.New("connection error")
	}

	if checkEmailInUse.Email != "" {
		return errors.New("email already in use")
	}

	checkUserNameInUse := &User{}
	err = db.Model(&User{}).Where("username = ?", u.Username).Take(checkUserNameInUse).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.New("connection error")
	}

	if checkUserNameInUse.Email != "" {
		return errors.New("email already in use")
	}

	return nil
}

func (u *User) CreateUser(db *gorm.DB) (string, error) {
	hashedPassword, err := auth.HashPassword(u.Password)

	if err != nil {
		return "", err
	}

	u.Password = hashedPassword

	AuthKey := auth.GenerateAPIKey(32)

	hashedAuthKey, err := auth.HashPassword(AuthKey)

	u.AuthKey = hashedAuthKey

	if err != nil {
		return "", err
	}

	err = db.Create(&u).Error

	if err != nil {
		return "", err
	}

	return AuthKey, nil
}

func (u *User) CheckCredentials(db *gorm.DB, email string, password string) (string, error) {
	err := db.Model(&User{}).Where("email = ?", email).Take(&u).Error

	if err != nil {
		return "", err
	}

	ok := auth.CheckPasswordHash(u.Password, password)

	if !ok {
		return "", errors.New("invalid password")
	}

	return u.Email, nil
}

func (u *User) AuthenticateUser(db *gorm.DB, username string) (*User, error) {
	err := db.Model(&User{}).Where("username = ?", username).Take(&u).Error

	if err != nil {
		return &User{}, err
	}

	return u, nil
}

func (u *User) GetUserByEmail(db *gorm.DB, email string) (*User, error) {
	err := db.Model(&User{}).Where("email = ?", email).Take(&u).Error

	if err != nil {
		return &User{}, err
	}

	return u, nil
}
