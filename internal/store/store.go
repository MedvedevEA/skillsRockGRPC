package store

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"skillsRockGRPC/internal/config"
	"skillsRockGRPC/internal/entity"
	"skillsRockGRPC/internal/repository/dto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	addUserQuery    = ``
	getUserQuery    = ``
	updateUserQuery = ``
)

func (s *Store) AddUser(dto *dto.AddUser) (*entity.User, error) {
	const op = "store.AddUser"
	return nil, nil
}
func (s *Store) GetUser(userId *uuid.UUID) (*entity.User, error) {
	const op = "store.GetUser"
	return nil, nil

}
func (s *Store) UpdateUser(dto *dto.UpdateUser) error {
	const op = "store.UpdateUser"
	return nil
}

/*
func (p *PostgreSql) Login(userName string) (string, error) {
	const op = "postgresql.Login"
	var password string
	err := p.pool.QueryRow(context.Background(), LoginQuery, userName).Scan(&password)
	if errors.Is(err, sql.ErrNoRows) {
		return "", servererrors.RecordNotFound
	}
	if err != nil {
		return "", servererrors.InternalServerError
	}
	return password, nil

}
*/
