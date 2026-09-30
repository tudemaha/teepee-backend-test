package usecase

import (
	"errors"
	"fmt"

	"github.com/tudemaha/marketplace-be/pkg/apperror"

	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"github.com/tudemaha/marketplace-be/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type ProductUseCase interface {
	Create(sellerID uuid.UUID, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetByID(id uuid.UUID) (*dto.ProductResponse, error)
	GetAll(filter dto.ProductListFilter) ([]dto.ProductListResponse, error)
	Update(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	UpdateStock(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductStockRequest) (*dto.ProductResponse, error)
	UpdateAvailability(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductAvailabilityRequest) (*dto.ProductResponse, error)
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
	var shop *entity.Shop
	var categories []entity.Category
	var mu sync.Mutex

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		var err error
		shop, err = u.shopRepo.FindByOwnerID(sellerID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrNotFound, "shop not found for this seller")
		}
		return nil
	})

	for _, catID := range req.CategoryIDs {
		g.Go(func() error {
			cat, err := u.categoryRepo.FindByID(catID)
			if err != nil {
				return fmt.Errorf("%w: %s", apperror.ErrBadRequest, "one or more categories not found")
			}

			mu.Lock()
			categories = append(categories, *cat)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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

func (u *productUseCase) GetAll(filter dto.ProductListFilter) ([]dto.ProductListResponse, error) {
	products, err := u.productRepo.FindAll(filter)
	if err != nil {
		return nil, errors.New("failed to fetch products")
	}

	var res []dto.ProductListResponse
	for _, p := range products {
		res = append(res, *utils.MapToProductListResponse(&p))
	}

	if res == nil {
		res = []dto.ProductListResponse{}
	}
	return res, nil
}

func (u *productUseCase) Update(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	var product *entity.Product
	var shop *entity.Shop

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		var err error
		product, err = u.productRepo.FindByID(productID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
		}
		return nil
	})

	g.Go(func() error {
		var err error
		shop, err = u.shopRepo.FindByOwnerID(sellerID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this product")
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if shop.ID != product.ShopID {
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

func (u *productUseCase) UpdateStock(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductStockRequest) (*dto.ProductResponse, error) {
	var product *entity.Product
	var shop *entity.Shop

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		var err error
		product, err = u.productRepo.FindByID(productID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
		}
		return nil
	})

	g.Go(func() error {
		var err error
		shop, err = u.shopRepo.FindByOwnerID(sellerID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this product")
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if shop.ID != product.ShopID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this product")
	}

	product.Stock += req.Stock
	if err := u.productRepo.Update(product, nil); err != nil {
		return nil, errors.New("failed to update product stock")
	}

	updatedProduct, _ := u.productRepo.FindByID(productID)
	return utils.MapToProductResponse(updatedProduct), nil
}

func (u *productUseCase) UpdateAvailability(sellerID uuid.UUID, productID uuid.UUID, req *dto.UpdateProductAvailabilityRequest) (*dto.ProductResponse, error) {
	var product *entity.Product
	var shop *entity.Shop

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		var err error
		product, err = u.productRepo.FindByID(productID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrNotFound, "product not found")
		}
		return nil
	})

	g.Go(func() error {
		var err error
		shop, err = u.shopRepo.FindByOwnerID(sellerID)
		if err != nil {
			return fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this product")
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if shop.ID != product.ShopID {
		return nil, fmt.Errorf("%w: %s", apperror.ErrForbidden, "unauthorized to update this product")
	}

	product.Available = *req.Available
	if err := u.productRepo.Update(product, nil); err != nil {
		return nil, errors.New("failed to update product availability")
	}

	updatedProduct, _ := u.productRepo.FindByID(productID)
	return utils.MapToProductResponse(updatedProduct), nil
}
