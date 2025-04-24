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
	updateUserQuery = `
	UPDATE "user" SET 
	login = CASE WHEN $2::character varying IS NULL THEN login ELSE $2 END,
	password = CASE WHEN $3::character varying IS NULL THEN password ELSE $3 END,
	email = CASE WHEN $4::character varying IS NULL THEN email ELSE $4 END
	WHERE user_id=$1
	RETURNING user_id`
	removeUserQuery       = `DELETE FROM "user" WHERE user_id=$1 RETURNING user_id`
	getRolesByUserIdQuery = `
	SELECT _r.name FROM role _r
	JOIN user_role _ur
	ON _r.role_id=_ur.role_id
	WHERE _ur.user_id=$1`
	addTokenWithIdQuery                         = `INSERT INTO "token" VALUES ($1,$2,$3,$4,$5,$6,$7)`
	removeTokenByUserIdAndDeviceCodeQuery       = `DELETE FROM "token" WHERE user_id=$1 AND device_code=$2`
	UpdateTokenRevokeByUserIdAndDeviceCodeQuery = `UPDATE "token" SET is_valid=false WHERE user_id=$1 AND device_code=$2`
	UpdateTokenRevokeByUserId                   = `UPDATE "token" SET is_valid=false WHERE user_id=$1`
)

func (s *Store) AddUser(dto *dto.AddUser) (*uuid.UUID, error) {
	const op = "store.AddUser"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), addUserQuery, dto.Login, dto.Password, dto.Email).Scan(userId)

	if err != nil {
		if pgError, ok := err.(*pgconn.PgError); ok && pgError.Code == "23505" {
			s.lg.Error("failed to add user", slog.String("op", op), slog.Any("error", err))
			return nil, servererrors.ErrorLoginAlreadyExists
		}
		s.lg.Error("failed to add user", slog.String("op", op), slog.Any("error", err))
		return nil, servererrors.ErrorInternalServerError
	}
	return userId, nil
}
func (s *Store) GetUserByLogin(login string) (*entity.User, error) {
	const op = "store.GetUser"
	user := new(entity.User)
	err := s.pool.QueryRow(context.Background(), getUserQuery, login).Scan(&user.UserId, &user.Login, &user.Password, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.lg.Error("failed to get user", slog.String("op", op), slog.Any("error", err))
			return nil, servererrors.ErrorRecordNotFound
		}
		s.lg.Error("failed to get user", slog.String("op", op), slog.Any("error", err))
		return nil, servererrors.ErrorInternalServerError
	}
	return user, err
}
func (s *Store) UpdateUser(dto *dto.UpdateUser) error {
	const op = "store.UpdateUser"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), updateUserQuery, dto.UserId, dto.Login, dto.Password, dto.Email).Scan(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.lg.Error("failed to update user", slog.String("op", op), slog.Any("error", err))
			return servererrors.ErrorRecordNotFound
		}
		s.lg.Error("failed to update user", slog.String("op", op), slog.Any("error", err))
		return servererrors.ErrorInternalServerError
	}
	return nil
}

func (s *Store) RemoveUser(userId *uuid.UUID) error {
	const op = "store.RemoveUser"
	err := s.pool.QueryRow(context.Background(), removeUserQuery, userId).Scan(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.lg.Error("failed to remove user", slog.String("op", op), slog.Any("error", err))
			return servererrors.ErrorRecordNotFound
		}
		s.lg.Error("failed to remove user", slog.String("op", op), slog.Any("error", err))
		return servererrors.ErrorInternalServerError
	}
	return nil
}
func (s *Store) GetRolesByUserId(userId *uuid.UUID) ([]string, error) {
	rows, err := s.pool.Query(context.Background(), getRolesByUserIdQuery, userId)
	if err != nil {
		return nil, servererrors.ErrorInternalServerError
	}
	defer rows.Close()
	names := []string{}

	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, servererrors.ErrorInternalServerError
		}
		names = append(names, name)
	}
	return names, nil
}
func (s *Store) AddTokenWithId(dto *dto.AddTokenWithId) error {
	//const op = "store.AddTokenWithId"
	_, err := s.pool.Exec(context.Background(), addTokenWithIdQuery, dto.TokenId, dto.UserId, dto.DeviceCode, dto.Token, dto.TokenTypeCode, dto.ExpirationAt, dto.IsRevoke)
	if err != nil {
		return servererrors.ErrorInternalServerError
	}
	return nil
}
func (s *Store) RemoveTokenByUserIdAndDeviceCode(dto *dto.RemoveTokenByUserIdAndDeviceCode) error {
	//const op = "store.RemoveTokenByUserIdAndDeviceCode"
	_, err := s.pool.Exec(context.Background(), removeTokenByUserIdAndDeviceCodeQuery, dto.UserId, dto.DeviceCode)
	if err != nil {
		return servererrors.ErrorInternalServerError
	}
	return nil
}

func (s *Store) UpdateTokenRevokeByUserIdAndDeviceCode(dto *dto.UpdateTokenRevokeByUserIdAndDeviceCode) error {
	//const op = "store.UpdateTokenRevokeByUserIdAndDeviceCode"
	_, err := s.pool.Exec(context.Background(), UpdateTokenRevokeByUserIdAndDeviceCodeQuery, dto.UserId, dto.DeviceCode)
	if err != nil {
		return servererrors.ErrorInternalServerError
	}
	return nil
}
func (s *Store) UpdateTokenRevokeByUserId(userId *uuid.UUID) error {
	//const op = "store.UpdateTokenRevokeByUserId"
	_, err := s.pool.Exec(context.Background(), UpdateTokenRevokeByUserId, userId)
	if err != nil {
		return servererrors.ErrorInternalServerError
	}
	return nil
}
func (s *Store) GetToken(tokenId *uuid.UUID) (*entity.Token, error) {
	return nil, nil
}
