package service

import (
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"taskflow/internal/config"
	"taskflow/internal/dto"
	"taskflow/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAuthInvalidInput       = errors.New("invalid auth input")
	ErrAuthEmailAlreadyExists = errors.New("email already exists")
	ErrAuthInvalidCredentials = errors.New("invalid email or password")
)
type AuthRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type AuthService struct {
	Repo AuthRepository
	JWTConfig *config.JWTConfig
}

func NewAuthService(
	repo AuthRepository,
	jwtConfig *config.JWTConfig,
) *AuthService {
	return &AuthService{
		Repo: repo,
		JWTConfig: jwtConfig,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req dto.RegisterRequest,
) (*dto.UserResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if len(name) < 2 || len(name) > 100 {
		return nil, ErrInvalidInput
	}

	if len(email) > 150 {
		return nil, ErrInvalidInput
	}

	parsedEmail, err := mail.ParseAddress(email) 
	if err != nil || parsedEmail.Address != email {
		return nil, ErrInvalidInput
	}

	if len(password) < 8 || len(password) > 72 {
		return nil, ErrInvalidInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name: name,
		Email: email,
		PasswordHash: string(passwordHash),
	}

	err = s.Repo.Create(ctx, user)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrAuthEmailAlreadyExists
		}

		return nil, err
	}

	return &dto.UserResponse{
		ID: user.ID,
		Name: user.Name,
		Email: user.Email,
	}, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	req dto.LoginRequest,
) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if len(email) == 0 || len(email) > 150 {
		return nil, ErrAuthInvalidCredentials
	}

	if len(req.Password) == 0 || len(req.Password) > 72 {
		return nil, ErrAuthInvalidCredentials
	}

	user, err := s.Repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAuthInvalidCredentials
		}

		return nil, err
	}

	if user.PasswordHash == "" {
		return nil, ErrAuthInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)
	if err != nil {
		return nil, ErrAuthInvalidCredentials
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer: "taskflow",
		Subject: strconv.FormatInt(user.ID, 10),
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.JWTConfig.ExpiresIn)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.JWTConfig.Secret)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken: signedToken,
		TokenType: "Bearer",
		ExpiresIn: int64(s.JWTConfig.ExpiresIn.Seconds()),
		User: dto.UserResponse{
			ID: user.ID,
			Name: user.Name,
			Email: user.Email,
		},
	}, nil
}