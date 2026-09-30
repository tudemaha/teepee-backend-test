package usecase

import (
	"errors"
	"fmt"
	"github.com/tudemaha/marketplace-be/pkg/apperror"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"github.com/tudemaha/marketplace-be/pkg/utils"
)

type ProductUseCase interface {
	Create(sellerID uuid.UUID, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetByID(id uuid.UUID) (*dto.ProductResponse, error)
	GetAll(filter dto.ProductListFilter) ([]dto.ProductResponse, error)
	Update(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(sellerID uuid.UUID, productID uuid.UUID) error
}

type productUseCase struct {
	productRepo  repository.ProductRepository
	shopRepo     repository.ShopRepository
	categoryRepo repository.CategoryRepository
}

func NewProductUseCase(productRepo repository.ProductRepository, shopRepo repository.ShopRepository, categoryRepo repository.CategoryRepository) ProductUseCase {
	return &productUseCase{
		productRepo:  productRepo,
		shopRepo:     shopRepo,
		categoryRepo: categoryRepo,
	}
}

func (u *productUseCase) Create(sellerID uuid.UUID, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	shop, err := u.shopRepo.FindByOwnerID(sellerID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "shop not found for this seller")
	}

	var categories []entity.Category
	for _, catID := range req.CategoryIDs {
		cat, err := u.categoryRepo.FindByID(catID)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", apperror.ErrBadRequest, "one or more categories not found")
		}
		categories = append(categories, *cat)
	}

	var images []entity.ProductImage
	for i, url := range req.ImageURLs {
		images = append(images, entity.ProductImage{
			URL:       url,
			IsPrimary: i == 0,
		})
	}

	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Available:   true,
		ShopID:      shop.ID,
		Categories:  categories,
		Images:      images,
	}

	if err := u.productRepo.Create(product); err != nil {
		return nil, errors.New("failed to create product")
	}

	// Attach shop to product for the response formatting
	product.Shop = *shop
	return utils.MapToProductResponse(product), nil
}

func (u *productUseCase) GetByID(id uuid.UUID) (*dto.ProductResponse, error) {
	product, err := u.productRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
	}
	return utils.MapToProductResponse(product), nil
}

func (u *productUseCase) GetAll(filter dto.ProductListFilter) ([]dto.ProductResponse, error) {
	products, err := u.productRepo.FindAll(filter)
	if err != nil {
		return nil, errors.New("failed to fetch products")
	}

	var res []dto.ProductResponse
	for _, p := range products {
		res = append(res, *utils.MapToProductResponse(&p))
	}

	if res == nil {
		res = []dto.ProductResponse{}
	}
	return res, nil
}

func (u *productUseCase) Update(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := u.productRepo.FindByID(productID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
	}

	shop, err := u.shopRepo.FindByOwnerID(sellerID)
	if err != nil || shop.ID != product.ShopID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this product")
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock

	if err := u.productRepo.Update(product, req.CategoryIDs); err != nil {
		return nil, errors.New("failed to update product")
	}

	updatedProduct, _ := u.productRepo.FindByID(productID)
	return utils.MapToProductResponse(updatedProduct), nil
}

func (u *productUseCase) Delete(sellerID uuid.UUID, productID uuid.UUID) error {
	product, err := u.productRepo.FindByID(productID)
	if err != nil {
		return fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
	}

	shop, err := u.shopRepo.FindByOwnerID(sellerID)
	if err != nil || shop.ID != product.ShopID {
		return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to delete this product")
	}

	if err := u.productRepo.SoftDelete(productID); err != nil {
		return errors.New("failed to delete product")
	}

	return nil
}
