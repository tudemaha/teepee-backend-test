package usecase

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
	Create(ownerID uuid.UUID, req *dto.CreateShopRequest) (*dto.ShopResponse, error)
	GetByID(id uuid.UUID) (*dto.ShopResponse, error)
	Update(ownerID uuid.UUID, shopID uuid.UUID, req *dto.UpdateShopRequest) (*dto.ShopResponse, error)
}

type shopUseCase struct {
	shopRepo repository.ShopRepository
	userRepo repository.UserRepository
}

func NewShopUseCase(shopRepo repository.ShopRepository, userRepo repository.UserRepository) ShopUseCase {
	return &shopUseCase{
		shopRepo: shopRepo,
		userRepo: userRepo,
	}
}

func (u *shopUseCase) Create(ownerID uuid.UUID, req *dto.CreateShopRequest) (*dto.ShopResponse, error) {
	if _, err := u.shopRepo.FindByOwnerID(ownerID); err == nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrConflict, "user already has a shop")
	}

	shop := &entity.Shop{
		Name:           req.Name,
		Address:        req.Address,
		ProfilePicture: req.ProfilePicture,
		OwnerID:        ownerID,
	}

	if err := u.shopRepo.Create(shop); err != nil {
		return nil, errors.New("failed to create shop")
	}

	user, err := u.userRepo.FindByID(ownerID)
	if err == nil && user.Role != entity.RoleAdmin {
		user.Role = entity.RoleSeller
		_ = u.userRepo.Update(user) // Ignore error, not strictly fatal if this fails, but ideally wrapped in a Tx
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
