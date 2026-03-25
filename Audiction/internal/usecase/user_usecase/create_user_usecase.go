package user_usecase

import (
	"context"
	"fullcycle-auction_go/internal/entity/user_entity"
	"fullcycle-auction_go/internal/internal_error"
)

func (u *UserUseCase) CreateUser(
	ctx context.Context,
	userInput UserInputDTO) (*UserOutputDTO, *internal_error.InternalError) {
	userEntity, err := user_entity.CreateUser(userInput.Name)
	if err != nil {
		return nil, err
	}

	if err := u.UserRepository.CreateUser(ctx, userEntity); err != nil {
		return nil, err
	}

	return &UserOutputDTO{
		Id:   userEntity.Id,
		Name: userEntity.Name,
	}, nil
}
