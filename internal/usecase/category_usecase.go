package usecase

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
)

type CategoryUseCase interface {
	Create(req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetAll() ([]dto.CategoryResponse, error)
	Delete(id uuid.UUID) error
}

type categoryUseCase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUseCase(categoryRepo repository.CategoryRepository) CategoryUseCase {
	return &categoryUseCase{
		categoryRepo: categoryRepo,
	}
}

func generateSlug(name string) string {
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(name, "-")
	return strings.Trim(slug, "-")
}

func (u *categoryUseCase) Create(req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	slug := generateSlug(req.Name)

	if _, err := u.categoryRepo.FindBySlug(slug); err == nil {
		return nil, errors.New("category with similar name already exists")
	}

	cat := &entity.Category{
		Name: req.Name,
		Slug: slug,
	}

	if err := u.categoryRepo.Create(cat); err != nil {
		return nil, errors.New("failed to create category")
	}

	return &dto.CategoryResponse{
		ID:   cat.ID,
		Name: cat.Name,
		Slug: cat.Slug,
	}, nil
}

func (u *categoryUseCase) GetAll() ([]dto.CategoryResponse, error) {
	categories, err := u.categoryRepo.FindAll()
	if err != nil {
		return nil, errors.New("failed to fetch categories")
	}

	var res []dto.CategoryResponse
	for _, c := range categories {
		res = append(res, dto.CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
		})
	}

	if res == nil {
		res = []dto.CategoryResponse{}
	}
	return res, nil
}

func (u *categoryUseCase) Delete(id uuid.UUID) error {
	if _, err := u.categoryRepo.FindByID(id); err != nil {
		return errors.New("category not found")
	}

	if err := u.categoryRepo.Delete(id); err != nil {
		return errors.New("failed to delete category")
	}
	return nil
}
