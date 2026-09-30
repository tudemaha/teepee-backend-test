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
