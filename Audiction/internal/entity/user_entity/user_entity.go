package user_entity

import (
	"context"
	"fullcycle-auction_go/internal/internal_error"

	"github.com/google/uuid"
)

type User struct {
	Id   string
	Name string
}

func CreateUser(name string) (*User, *internal_error.InternalError) {
	user := &User{
		Id:   uuid.New().String(),
		Name: name,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) Validate() *internal_error.InternalError {
	if len(u.Name) <= 1 {
		return internal_error.NewBadRequestError("Name must have at least 2 characters")
	}
	return nil
}

type UserRepositoryInterface interface {
	FindUserById(
		ctx context.Context, userId string) (*User, *internal_error.InternalError)

	CreateUser(
		ctx context.Context, user *User) *internal_error.InternalError
}
