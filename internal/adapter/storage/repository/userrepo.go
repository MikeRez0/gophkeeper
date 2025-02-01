package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/MikeRez0/gophkeeper/internal/adapter/storage"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *storage.DB
}

type queryAble interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func NewUserRepository(db *storage.DB) (*UserRepository, error) {
	return &UserRepository{db: db}, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		userSt := r.db.QueryBuilder.
			Insert("users").
			Columns("login", "password").
			Values(user.Login, user.Password).
			Suffix("returning id")

		sql, args, err := userSt.ToSql()
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, sql, args...).Scan(&(user.ID))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, domain.ErrConflictingData
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) selectUser(ctx context.Context, tx queryAble, login string, forUpdate bool) (*domain.User, error) {
	statement := r.db.QueryBuilder.
		Select("id", "login", "password").
		From("users").
		Where(sq.Eq{"login": login})

	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	user := domain.User{}

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return r.selectUser(ctx, r.db.Pool, login, false)
}
