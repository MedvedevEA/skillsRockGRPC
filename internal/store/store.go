package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"
	"skillsRockGRPC/pkg/servererrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Store struct {
	pool *pgxpool.Pool
	lg   *slog.Logger
}

func MustNew(ctx context.Context, lg *slog.Logger, cfg *config.PostgreSQL) *Store {
	const op = "postgresql.MustNew"
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
		log.Fatalf("%s: %s", op, err)
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}

	return &Store{
		pool,
		lg,
	}
}

const (
	addUserQuery    = `INSERT INTO "user" (login,password,email) VALUES ($1, $2, $3) RETURNING user_id`
	getUserQuery    = `SELECT * FROM "user" WHERE login=$1`
	updateUserQuery = `UPDATE "user" SET login=$2, password=$3, email=$4 WHERE user_id=$1 RETURNING user_id`
)

func (s *Store) AddUser(dto *dto.AddUser) (*uuid.UUID, error) {
	const op = "store.AddUser"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), addUserQuery, dto.Login, dto.Password, dto.Email).Scan(userId)

	if err, ok := err.(*pgconn.PgError); ok {
		if err.Code == "23505" && err.ConstraintName == "user_login_unique" {
			s.lg.Error("failed to add user", slog.String("op", op), slog.Any("error", err))
			return nil, servererrors.ErrorUsernameAlreadyExists
		}
	}
	if err != nil {
		s.lg.Error("failed to add user", slog.String("op", op), slog.Any("error", err))
		return nil, servererrors.ErrorInternalServerError
	}
	return userId, err
}
func (s *Store) GetUser(login string) (*entity.User, error) {
	const op = "store.GetUser"
	user := new(entity.User)
	err := s.pool.QueryRow(context.Background(), getUserQuery, login).Scan(&user.UserId, &user.Login, &user.Password, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		s.lg.Error("failed to get user", slog.String("op", op), slog.Any("error", err))
		return nil, servererrors.ErrorRecordNotFound
	}
	if err != nil {
		s.lg.Error("failed to get user", slog.String("op", op), slog.Any("error", err))
		return nil, servererrors.ErrorInternalServerError
	}
	return user, err
}
func (s *Store) UpdateUser(dto *dto.UpdateUser) error {
	const op = "store.UpdateUser"
	user := new(entity.User)
	err := s.pool.QueryRow(context.Background(), getUserQuery, dto.UserId, dto.Login, dto.Password, dto.Email).Scan(&user.UserId, &user.Login, &user.Password, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		s.lg.Error("failed to update user", slog.String("op", op), slog.Any("error", err))
		return servererrors.ErrorRecordNotFound
	}
	if err != nil {
		s.lg.Error("failed to update user", slog.String("op", op), slog.Any("error", err))
		return servererrors.ErrorInternalServerError
	}
	return nil
}
