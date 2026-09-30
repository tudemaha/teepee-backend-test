package utils

import (
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
)

func MapToProductResponse(p *entity.Product) *dto.ProductResponse {
	var catRes []dto.CategoryResponse
	for _, c := range p.Categories {
		catRes = append(catRes, dto.CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
		})
	}

	var imgRes []dto.ProductImageResponse
	for _, i := range p.Images {
		imgRes = append(imgRes, dto.ProductImageResponse{
			ID:        i.ID,
			URL:       i.URL,
			IsPrimary: i.IsPrimary,
		})
	}

	return &dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      string(p.Status),
		ShopID:      p.ShopID,
		ShopName:    p.Shop.Name,
		Categories:  catRes,
		Images:      imgRes,
	}
}

func MapToProductListResponse(p *entity.Product) *dto.ProductListResponse {
	var primaryImage string
	for _, img := range p.Images {
		if img.IsPrimary {
			primaryImage = img.URL
			break
		}
	}
	
	// Fallback to first image if no primary is explicitly set
	if primaryImage == "" && len(p.Images) > 0 {
		primaryImage = p.Images[0].URL
	}

	return &dto.ProductListResponse{
		ID:       p.ID,
		Name:     p.Name,
		Price:    p.Price,
		Stock:    p.Stock,
		Status:   string(p.Status),
		ShopName: p.Shop.Name,
		Image:    primaryImage,
	}
}
