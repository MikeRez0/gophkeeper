package pgsql

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *DB
}

type queryAble interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func NewUserRepository(db *DB) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("nil not allowed as argument")
	}
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
		if err := wrapStatmentErr(err); err != nil {
			return err
		}

		err = tx.QueryRow(ctx, sql, args...).Scan(&(user.ID))
		if err := wrapSQLErr(err); err != nil {
			return err
		}

		return nil
	})

	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) selectUser(ctx context.Context, tx queryAble,
	login string, forUpdate bool) (*domain.User, error) {
	statement := r.db.QueryBuilder.
		Select("id", "login", "password").
		From("users").
		Where(sq.Eq{"login": login})

	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err := wrapStatmentErr(err); err != nil {
		return nil, err
	}

	user := domain.User{}

	err = tx.QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return r.selectUser(ctx, r.db.Pool, login, false)
}
