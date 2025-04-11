package postgres

import (
	"context"

	"github.com/MxTrap/gophermart/internal/gophermart/entity"
	storage "github.com/MxTrap/gophermart/internal/gophermart/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepo struct {
	db *pgxpool.Pool
}

func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{
		db: pool,
	}
}

const repoName = "postgres.TokenRepo."

func (r TokenRepo) GetTokens(ctx context.Context, userId int64) ([]entity.RefreshToken, error) {
	const op = repoName + "GetTokens"
	var tokens []entity.RefreshToken
	rows, err := r.db.Query(ctx, selectTokensStmt, userId)
	if err != nil {
		return tokens, storage.NewRepositoryError(op, err)
	}

	tokens, err = pgx.CollectRows(rows, pgx.RowToStructByName[entity.RefreshToken])
	if err != nil {
		return tokens, storage.NewRepositoryError(op, err)
	}

	return tokens, nil
}

func (r TokenRepo) DeleteToken(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, deleteTokenStmt, token)
	if err != nil {
		return storage.NewRepositoryError(repoName+"DeleteToken", err)
	}
	return nil
}

func (r TokenRepo) ClearTokens(ctx context.Context, userId int64) error {
	_, err := r.db.Exec(ctx, deleteAllTokensStmt, userId)
	if err != nil {
		return storage.NewRepositoryError(repoName+"ClearTokens", err)
	}
	return nil
}

func (r *TokenRepo) SaveToken(ctx context.Context, token entity.RefreshToken) error {
	_, err := r.db.Exec(ctx, insertTokenStmt, token.UserId, token.RefreshToken, token.ExpirationTime)
	if err != nil {
		return storage.NewRepositoryError(repoName+"SaveToken", err)
	}

	return nil
}
