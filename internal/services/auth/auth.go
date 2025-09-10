package auth

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/exp/slog"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		phone string,
		passHash []byte,
	) (uid int64, err error)
	UpdateUser(
		ctx context.Context,
		user_id int64,
		name string,
		email string,
		phone string,
		passHash []byte,
	) (bool, error)
	DeleteUser(
		ctx context.Context,
		phone string,
	) (bool, error)
}

type UserProvider interface {
	User(ctx context.Context, phone string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// New return a new instance of the Auth service.
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {

	return &Auth{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if user with given credentials exist in the system
//
// if user exists, but password is incorrect, return error,
// if user doesn't exist, return error.
func (a *Auth) Login(
	ctx context.Context,
	phone string,
	password string,
	appID int,
) (string, string, string, int64, error) {
	const op = "auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("phone", phone),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, phone)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			return "", "", "", 0, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", "", "", 0, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("Invalid credentials", sl.Err(err))

		return "", "", "", 0, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", "", "", 0, fmt.Errorf("%s:%w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		return "", "", "", 0, fmt.Errorf("%s, %w", op, err)
	}

	return user.Name, user.Email, token, user.ID, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, phone string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("phone", phone),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate hash", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, phone, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user not found", sl.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

// GetUser receive name of user
func (a *Auth) GetUser(ctx context.Context, phone string) (string, error) {
	const op = "auth.GetUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("phone", phone),
	)

	log.Info("getting information of user")

	user, err := a.usrProvider.User(ctx, phone)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("getted information of user", slog.String("name", user.Name))

	return user.Name, nil
}

// UpdateUser put new information of user
//
// Check token.
// If owner is not user or admin, return error.
// If user doesn't exist, return error.
// Else update information and return success.
func (a *Auth) UpdateUser(
	ctx context.Context,
	user_id int64,
	name string,
	email string,
	phone string,
	password string,
	token string,
) (bool, error) {
	const op = "auth.UpdateUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("phone", phone),
	)
	log.Info("checking token")

	//	TODO: implement logic from godoc

	log.Info("not implemented")

	return false, nil
}

// DeleteUser put new information of user
//
// Check token.
// If owner is not user or admin, return error.
// If user doesn't exist, return error.
// Else delete user and return success.
func (a *Auth) DeleteUser(
	ctx context.Context,
	phone string,
	token string,
) (bool, error) {
	const op = "auth.DeleteUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("phone", phone),
	)
	log.Info("checking token")

	//	TODO: implement logic from godoc

	log.Info("not implemented")

	return false, nil
}
