package database

import (
	"log"

	"github.com/tudemaha/marketplace-be/config"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedData(db *gorm.DB, cfg config.AppConfig) error {
	log.Println("Running database seeders...")

	if err := seedUsers(db, cfg); err != nil {
		return err
	}

	if err := seedCategories(db); err != nil {
		return err
	}

	log.Println("Database seeding completed.")
	return nil
}

func seedUsers(db *gorm.DB, cfg config.AppConfig) error {
	users := []struct {
		Name     string
		Email    string
		Password string
		Role     entity.UserRole
	}{
		{"Platform Admin", cfg.AdminEmail, cfg.AdminPassword, entity.RoleAdmin},
		{"Shop Seller", cfg.SellerEmail, cfg.SellerPassword, entity.RoleSeller},
		{"Happy Buyer", cfg.BuyerEmail, cfg.BuyerPassword, entity.RoleBuyer},
	}

	for _, u := range users {
		var count int64
		db.Model(&entity.User{}).Where("email = ?", u.Email).Count(&count)
		if count == 0 {
			hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			user := entity.User{
				Name:     u.Name,
				Email:    u.Email,
				Password: string(hashed),
				Role:     u.Role,
			}
			if err := db.Create(&user).Error; err != nil {
				return err
			}
			log.Printf("Seeded user: %s (%s)", u.Email, u.Role)

			if u.Role == entity.RoleSeller {
				shop := entity.Shop{
					Name:    "Gadget Hub",
					Address: "123 Tech Avenue, Silicon Valley",
					OwnerID: user.ID,
				}
				if err := db.Create(&shop).Error; err == nil {
					log.Printf("Seeded shop for seller: %s", u.Email)
				}
			}
		}
	}
	return nil
}

func seedCategories(db *gorm.DB) error {
	categories := []struct {
		Name string
		Slug string
	}{
		{"Electronics", "electronics"},
		{"Fashion", "fashion"},
		{"Home & Living", "home-living"},
		{"Beauty", "beauty"},
		{"Sports", "sports"},
	}

	for _, c := range categories {
		var count int64
		db.Model(&entity.Category{}).Where("slug = ?", c.Slug).Count(&count)
		if count == 0 {
			cat := entity.Category{
				Name: c.Name,
				Slug: c.Slug,
			}
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
			log.Printf("Seeded category: %s", c.Name)
		}
	}
	return nil
}
