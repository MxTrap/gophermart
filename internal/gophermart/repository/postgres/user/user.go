package posgress

import (
	"context"
	"errors"

	"github.com/MxTrap/gophermart/internal/gophermart/common"
	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	storage "github.com/MxTrap/gophermart/internal/gophermart/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: pool,
	}
}

const repoName = "postgres.UserRepo."

func (s UserRepository) SaveUser(
	ctx context.Context,
	user entity.User,
) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, insertStmt, user.Login, user.Password).Scan(&id)

	if err != nil {
		return 0, storage.NewRepositoryError(repoName+".SaveUser", err)
	}

	return id, nil
}

func (s UserRepository) collectUser(rows pgx.Rows) (entity.User, error) {
	const op = repoName + "collectUser"
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[entity.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, storage.NewRepositoryError(op, common.ErrUserNotFound)
		}
		return user, storage.NewRepositoryError(op, err)
	}
	return user, nil
}

func (s UserRepository) FindUserByID(ctx context.Context, userID int64) (entity.User, error) {
	var user entity.User

	row, err := s.db.Query(ctx, findByIdStmt, userID)
	if err != nil {
		return user, storage.NewRepositoryError(repoName+"FindUserByID", err)
	}

	return s.collectUser(row)
}

func (s UserRepository) FindUserByUsername(ctx context.Context, username string) (entity.User, error) {
	var user entity.User

	row, err := s.db.Query(ctx, findByUsernameStmt, username)
	if err != nil {
		return user, storage.NewRepositoryError(repoName+"FindUserByUsername", err)
	}

	return s.collectUser(row)
}
