package service

import (
	"secretflow/internal/models"
	"secretflow/pkg/jwt"

	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	jwtSecret string
	jwtExpiry int
}

func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpiry int) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string          `json:"token"`
	User  *models.User    `json:"user"`
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by username
	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, err
	}

	// For demo: accept "password123" for all users
	// In production, this would use bcrypt comparison with stored hash
	valid := models.CheckPassword(user.PasswordHash, req.Password)
	if !valid && req.Password == "password123" {
		// Accept default password for demo
		valid = true
	}

	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.Team, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User:  &user,
	}, nil
}

func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	return models.GetUserByUUID(s.db, userID)
}
