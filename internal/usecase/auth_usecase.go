package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tudemaha/marketplace-be/internal/delivery/http/dto"
	"github.com/tudemaha/marketplace-be/internal/domain/entity"
	"github.com/tudemaha/marketplace-be/internal/domain/repository"
	"github.com/tudemaha/marketplace-be/pkg/jwt"
	"github.com/tudemaha/marketplace-be/pkg/password"
)

type AuthUseCase interface {
	Register(req *dto.RegisterRequest) (*dto.TokenResponse, error)
	Login(req *dto.LoginRequest) (*dto.TokenResponse, error)
	Refresh(req *dto.RefreshRequest) (*dto.TokenResponse, error)
	Logout(tokenStr string) error
	GetProfile(userID uuid.UUID) (*dto.UserProfileResponse, error)
}

type authUseCase struct {
	userRepo  repository.UserRepository
	rtRepo    repository.RefreshTokenRepository
	jwtSecret string
}

func NewAuthUseCase(userRepo repository.UserRepository, rtRepo repository.RefreshTokenRepository, jwtSecret string) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		rtRepo:    rtRepo,
		jwtSecret: jwtSecret,
	}
}

func (u *authUseCase) Register(req *dto.RegisterRequest) (*dto.TokenResponse, error) {
	// 1. Check if email exists
	if _, err := u.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already in use")
	}

	// 2. Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. Create user entity
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashedPassword,
		Role:     entity.RoleBuyer, // Default role
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// 4. Generate tokens
	return u.generateAndSaveTokens(user.ID, string(user.Role))
}

func (u *authUseCase) Login(req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// 1. Find user by email
	user, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 2. Verify password
	if !password.Check(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	// 3. Generate tokens
	return u.generateAndSaveTokens(user.ID, string(user.Role))
}

func (u *authUseCase) Refresh(req *dto.RefreshRequest) (*dto.TokenResponse, error) {
	// 1. Find token in DB
	rt, err := u.rtRepo.FindByToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 2. Check expiry
	if time.Now().After(rt.ExpiresAt) {
		_ = u.rtRepo.DeleteByToken(req.RefreshToken)
		return nil, errors.New("refresh token expired")
	}

	// 3. Get User to know the role
	user, err := u.userRepo.FindByID(rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 4. Delete old token
	_ = u.rtRepo.DeleteByToken(req.RefreshToken)

	// 5. Issue new pair
	return u.generateAndSaveTokens(user.ID, string(user.Role))
}

func (u *authUseCase) Logout(tokenStr string) error {
	return u.rtRepo.DeleteByToken(tokenStr)
}

func (u *authUseCase) GetProfile(userID uuid.UUID) (*dto.UserProfileResponse, error) {
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &dto.UserProfileResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Role:  string(user.Role),
	}, nil
}

// Helper to generate access + refresh pair and save refresh token to DB
func (u *authUseCase) generateAndSaveTokens(userID uuid.UUID, role string) (*dto.TokenResponse, error) {
	access, refresh, err := jwt.GenerateTokenPair(userID, role, u.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Save refresh token to DB
	rtEntity := &entity.RefreshToken{
		UserID:    userID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := u.rtRepo.Create(rtEntity); err != nil {
		return nil, errors.New("failed to save refresh token")
	}

	return &dto.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
