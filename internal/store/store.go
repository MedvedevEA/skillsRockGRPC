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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

const (
	addUserQuery = `
		INSERT INTO "user" (login,password,email) VALUES ($1, $2, $3) RETURNING user_id;
	`
	getUserQuery = `
		SELECT * FROM "user" WHERE login=$1;
	`
	updateUserQuery = `
		UPDATE "user" SET 
		login = CASE WHEN $2::character varying IS NULL THEN login ELSE $2 END,
		password = CASE WHEN $3::character varying IS NULL THEN password ELSE $3 END,
		email = CASE WHEN $4::character varying IS NULL THEN email ELSE $4 END
		WHERE user_id=$1
		RETURNING user_id;
	`
	removeUserQuery = `
		DELETE FROM "user" WHERE user_id=$1 RETURNING user_id;
	`
	removeUserByLoginQuery = `
		DELETE FROM "user" WHERE login=$1 RETURNING user_id;
	`
	getRoleIdsByUserIdQuery = `
		SELECT _r.name FROM role _r
		JOIN user_role _ur
		ON _r.role_id=_ur.role_id
		WHERE _ur.user_id=$1;
	`

	addTokenWithIdQuery = `
		INSERT INTO "token" VALUES ($1,$2,$3,$4,$5,$6,$7);
	`
	getTokenQuery = `
		SELECT * FROM "token" WHERE token_id=$1;
	`
	UpdateTokenRevokeByTokenIdQuery = `
		UPDATE "token" SET is_valid=false WHERE token_id=$1 RETURNING token_id;
	`
	UpdateTokenRevokeByUserId = `
		UPDATE "token" SET is_valid=false WHERE user_id=$1;
	`
	removeTokenByUserIdAndDeviceCodeQuery = `
		DELETE FROM "token" WHERE user_id=$1 AND device_code=$2;
	`
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
		log.Fatal(errors.Wrap(err, op))
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatal(errors.Wrap(err, op))
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal(errors.Wrap(err, op))
	}

	return &Store{
		pool,
		lg,
	}
}

func (s *Store) AddUser(dto *dto.AddUser) (*uuid.UUID, error) {
	const op = "store.AddUser"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), addUserQuery, dto.Login, dto.Password, dto.Email).Scan(userId)

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
	err := s.pool.QueryRow(context.Background(), getUserQuery, login).Scan(&user.UserId, &user.Login, &user.Password, &user.Email)
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
	err := s.pool.QueryRow(context.Background(), updateUserQuery, dto.UserId, dto.Login, dto.Password, dto.Email).Scan(userId)
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
func (s *Store) RemoveUserByLogin(login string) error {
	const op = "store.RemoveUserByLogin"
	userId := new(uuid.UUID)
	err := s.pool.QueryRow(context.Background(), removeUserQuery, login).Scan(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) GetRoleIdsByUserId(userId *uuid.UUID) ([]*uuid.UUID, error) {
	const op = "store.GetRoleIdsByUserId"
	rows, err := s.pool.Query(context.Background(), getRoleIdsByUserIdQuery, userId)
	if err != nil {
		return nil, errors.Wrap(repository.ErrInternalServerError, op)
	}
	defer rows.Close()

	roleIds := []*uuid.UUID{}
	for rows.Next() {
		var roleId uuid.UUID
		err := rows.Scan(&roleId)
		if err != nil {
			return nil, errors.Wrap(repository.ErrInternalServerError, op)
		}
		roleIds = append(roleIds, &roleId)
	}
	return roleIds, nil
}

func (s *Store) AddTokenWithId(dto *dto.AddTokenWithId) error {
	const op = "store.AddTokenWithId"
	_, err := s.pool.Exec(context.Background(), addTokenWithIdQuery, dto.TokenId, dto.UserId, dto.DeviceCode, dto.Token, dto.TokenTypeCode, dto.ExpirationAt, dto.IsRevoke)
	if err != nil {
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) GetToken(tokenId *uuid.UUID) (*entity.Token, error) {
	const op = "store.GetToken"
	token := new(entity.Token)
	err := s.pool.QueryRow(context.Background(), getTokenQuery, tokenId).Scan(&token.TokenId, &token.UserId, &token.DeviceCode, &token.Token, &token.TokenTypeCode, &token.ExpirationAt, &token.IsRevoke)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return nil, errors.Wrap(repository.ErrInternalServerError, op)
	}
	return token, err
}
func (s *Store) UpdateTokenRevokeByTokenId(tokenId *uuid.UUID) error {
	const op = "store.UpdateTokenRevokeByTokenId"
	if err := s.pool.QueryRow(context.Background(), UpdateTokenRevokeByTokenIdQuery, tokenId).Scan(tokenId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(repository.ErrRecordNotFound, op)
		}
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) UpdateTokensRevokeByUserId(userId *uuid.UUID) error {
	const op = "store.UpdateTokenRevokeByUserId"
	_, err := s.pool.Exec(context.Background(), UpdateTokenRevokeByUserId, userId)
	if err != nil {
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
func (s *Store) RemoveTokensByUserIdAndDeviceCode(dto *dto.RemoveTokensByUserIdAndDeviceCode) error {
	const op = "store.RemoveTokenByUserIdAndDeviceCode"
	_, err := s.pool.Exec(context.Background(), removeTokenByUserIdAndDeviceCodeQuery, dto.UserId, dto.DeviceCode)
	if err != nil {
		return errors.Wrap(repository.ErrInternalServerError, op)
	}
	return nil
}
