package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
	log       *logger.Logger
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		log:       logger.Get(),
	}
}

func (s *AuthService) Register(ctx context.Context, fullName, email, password, role string) (*db.User, error) {
	s.log.Infow("attempting to register new user", "email", email, "role", role)

	existing, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Errorw("database error checking existing user", "error", err, "email", email)
		return nil, domain.ErrInternal("something went wrong")
	}
	if existing != nil {
		s.log.Warnw("registration failed: email already in use", "email", email)
		return nil, domain.ErrBadRequest("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		s.log.Errorw("failed to hash password", "error", err)
		return nil, domain.ErrInternal("could not hash password")
	}

	// Default role to candidate if empty or invalid
	userRole := role
	if userRole != "candidate" && userRole != "company" {
		userRole = "candidate"
	}

	userParams := db.CreateUserParams{
		FullName:     fullName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         userRole,
	}

	createdUser, err := s.userRepo.CreateUser(ctx, userParams)
	if err != nil {
		s.log.Errorw("failed to create user in database", "error", err, "email", email)
		return nil, domain.ErrInternal("something went wrong")
	}

	s.log.Infow("user registered successfully", "email", email, "userID", createdUser.ID, "role", createdUser.Role)
	return createdUser, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (string, error) {
	s.log.Infow("attempting login", "email", email)

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		s.log.Errorw("database error during login", "error", err, "email", email)
		return "", domain.ErrInternal("something went wrong")
	}
	if user == nil {
		s.log.Warnw("login failed: user not found", "email", email)
		return "", domain.ErrBadRequest("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		s.log.Warnw("login failed: invalid password", "email", email)
		return "", domain.ErrBadRequest("invalid email or password")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.log.Errorw("failed to sign jwt token", "error", err, "userID", user.ID)
		return "", domain.ErrInternal("could not create token")
	}

	s.log.Infow("user logged in successfully", "email", email, "userID", user.ID)
	return signed, nil
}
