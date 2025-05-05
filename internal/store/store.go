package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository"
	"skillsRockGRPC/internal/repository/dto"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

const (
	addUserQuery = `
INSERT INTO "user" (login,password) 
VALUES ($1, $2) RETURNING user_id;`
	getUserByLoginQuery = `
SELECT * FROM "user" 
WHERE login=$1;`
	updateUserQuery = `
UPDATE "user" SET 
login = CASE WHEN $2::character varying IS NULL THEN login ELSE $2 END,
password = CASE WHEN $3::character varying IS NULL THEN password ELSE $3 END
WHERE user_id=$1
RETURNING user_id;`
	removeUserQuery = `
DELETE FROM "user" WHERE user_id=$1 RETURNING user_id;`
	addRefreshTokenWithRefreshTokenIdQuery = `
INSERT INTO refresh_token VALUES ($1,$2,$3,$4,$5);`
	getRefreshTokenQuery = `
SELECT * FROM refresh_token 
WHERE refresh_token_id=$1;`
	revokeRefreshTokensByUserIdAndDeviceCodeQuery = `
UPDATE refresh_token 
SET is_revoke=true
WHERE user_id=$1 AND ($2::character varying IS NULL OR device_code=$2);`
	revokeRefreshTokenByRefreshTokenIdQuery = `
UPDATE refresh_token 
SET is_revoke=true
WHERE refresh_token_id = $1;`
	removeRefreshTokensByExpirationAtQuery = `
DELETE FROM refresh_token
WHERE expiration_at < $1;`
)

type Store struct {
	pool *pgxpool.Pool
	lg   *slog.Logger
}

func MustNew(lg *slog.Logger, cfg *config.Store) *Store {
	const op = "store.MustNew"
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("%s: %v", op, err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("%s: %v", op, err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("%s: %v", op, err)
	}
	return &Store{
		pool,
		lg,
	}
}

func (s *Store) AddUser(dto *dto.AddUser) (*uuid.UUID, error) {
	const op = "store.AddUser"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), addUserQuery, dto.Login, dto.Password).Scan(userId)
	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == "23505" {
			return nil, errors.Wrap(repository.ErrUniqueViolation, op)

		}
		return nil, errors.Wrap(repository.ErrInternalServerError, op)

	}
	return userId, nil
}
func (s *Store) GetUserByLogin(login string) (*entity.User, error) {
	const op = "store.GetUserByLogin"
	user := new(entity.User)
	err := s.pool.QueryRow(context.Background(), getUserByLoginQuery, login).Scan(&user.UserId, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return nil, errors.Wrap(repository.ErrInternalServerError, op)
	}
	return user, err
}
func (s *Store) UpdateUser(dto *dto.UpdateUser) error {
	const op = "store.UpdateUser"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), updateUserQuery, dto.UserId, dto.Login, dto.Password).Scan(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) RemoveUser(userId *uuid.UUID) error {
	const op = "store.RemoveUser"
	err := s.pool.QueryRow(context.Background(), removeUserQuery, userId).Scan(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}

func (s *Store) AddRefreshTokenWithRefreshTokenId(dto *dto.AddRefreshTokenWithRefreshTokenId) error {
	const op = "store.AddRefreshTokenWithRefreshTokenId"
	_, err := s.pool.Exec(context.Background(), addRefreshTokenWithRefreshTokenIdQuery, dto.RefreshTokenId, dto.UserId, dto.DeviceCode, dto.ExpirationAt, dto.IsRevoke)
	if err != nil {
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) GetRefreshToken(refreshTokenId *uuid.UUID) (*entity.RefreshToken, error) {
	const op = "store.GetRefreshToken"
	refreshToken := new(entity.RefreshToken)
	err := s.pool.QueryRow(context.Background(), getRefreshTokenQuery, refreshTokenId).Scan(&refreshToken.RefreshTokenId, &refreshToken.UserId, &refreshToken.DeviceCode, &refreshToken.ExpirationAt, &refreshToken.IsRevoke)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return nil, errors.Wrap(repository.ErrInternalServerError, op)
	}
	return refreshToken, nil
}
func (s *Store) RevokeRefreshTokenByRefreshTokenId(refreshTokenId *uuid.UUID) error {
	const op = "store.RevokeRefreshTokenByRefreshTokenIdAndIsRevoke"
	_, err := s.pool.Exec(context.Background(), revokeRefreshTokenByRefreshTokenIdQuery, refreshTokenId)
	if err != nil {
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) RevokeRefreshTokensByUserIdAndDeviceCode(dto *dto.RevokeRefreshTokensByUserIdAndDeviceCode) error {
	const op = "store.RevokeRefreshTokensByUserIdAndDeviceCode"
	_, err := s.pool.Exec(context.Background(), revokeRefreshTokensByUserIdAndDeviceCodeQuery, dto.UserId, dto.DeviceCode)
	if err != nil {
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) RemoveRefreshTokensByExpirationAt(now time.Time) (int64, error) {
	const op = "store.RemoveRefreshTokensByExpirationAtQuery"
	result, err := s.pool.Exec(context.Background(), removeRefreshTokensByExpirationAtQuery, now)
	if err != nil {
		return -1, errors.Wrap(repository.ErrInternalServerError, op)
	}
	return result.RowsAffected(), nil
}
