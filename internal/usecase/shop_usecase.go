package usecase

import "context"

import (
	"errors"
	"fmt"
	"github.com/tudemaha/marketplace-be/pkg/apperror"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
)

type ShopUseCase interface {
	Create(ctx context.Context, ownerID uuid.UUID, req *dto.CreateShopRequest) (*dto.ShopResponse, error)
	GetByID(id uuid.UUID) (*dto.ShopResponse, error)
	Update(ownerID uuid.UUID, shopID uuid.UUID, req *dto.UpdateShopRequest) (*dto.ShopResponse, error)
}

type shopUseCase struct {
	txManager repository.TxManager
	shopRepo repository.ShopRepository
	userRepo repository.UserRepository
}

func NewShopUseCase(txManager repository.TxManager, shopRepo repository.ShopRepository, userRepo repository.UserRepository) ShopUseCase {
	return &shopUseCase{
		txManager: txManager,
		shopRepo: shopRepo,
		userRepo: userRepo,
	}
}

func (u *shopUseCase) Create(ctx context.Context, ownerID uuid.UUID, req *dto.CreateShopRequest) (*dto.ShopResponse, error) {
	if _, err := u.shopRepo.FindByOwnerID(ownerID); err == nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrConflict, "user already has a shop")
	}

	shop := &entity.Shop{
		Name:           req.Name,
		Address:        req.Address,
		ProfilePicture: req.ProfilePicture,
		OwnerID:        ownerID,
	}

	err := u.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := u.shopRepo.Create(txCtx, shop); err != nil {
			return errors.New("failed to create shop")
		}

		user, err := u.userRepo.FindByID(ownerID)
		if err == nil && user.Role != entity.RoleAdmin {
			user.Role = entity.RoleSeller
			if err := u.userRepo.Update(txCtx, user); err != nil {
				return errors.New("failed to update user role")
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &dto.ShopResponse{
		ID:             shop.ID,
		Name:           shop.Name,
		Address:        shop.Address,
		ProfilePicture: shop.ProfilePicture,
		OwnerID:        shop.OwnerID,
	}, nil
}

func (u *shopUseCase) GetByID(id uuid.UUID) (*dto.ShopResponse, error) {
	shop, err := u.shopRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "shop not found")
	}

	return &dto.ShopResponse{
		ID:             shop.ID,
		Name:           shop.Name,
		Address:        shop.Address,
		ProfilePicture: shop.ProfilePicture,
		OwnerID:        shop.OwnerID,
	}, nil
}

func (u *shopUseCase) Update(ownerID uuid.UUID, shopID uuid.UUID, req *dto.UpdateShopRequest) (*dto.ShopResponse, error) {
	shop, err := u.shopRepo.FindByID(shopID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "shop not found")
	}

	if shop.OwnerID != ownerID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this shop")
	}

	shop.Name = req.Name
	shop.Address = req.Address
	shop.ProfilePicture = req.ProfilePicture

	if err := u.shopRepo.Update(shop); err != nil {
		return nil, errors.New("failed to update shop")
	}

	return &dto.ShopResponse{
		ID:             shop.ID,
		Name:           shop.Name,
		Address:        shop.Address,
		ProfilePicture: shop.ProfilePicture,
		OwnerID:        shop.OwnerID,
	}, nil
}
