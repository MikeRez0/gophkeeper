package sqlite

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"go.uber.org/zap"
)

// UserRepository pgsql implementation of user store.
type UserRepository struct {
	db  *DB
	log *zap.Logger
}

// NewUserRepository creates new user sqlite-repositiry
func NewUserRepository(db *DB, log *zap.Logger) (*UserRepository, error) {
	if db == nil {
		return nil, errors.New("nil not allowed as database")
	}
	if log == nil {
		return nil, errors.New("nil not allowed as log")
	}
	return &UserRepository{
		db:  db,
		log: log,
	}, nil
}

// CreateUser creates new user.
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}
	defer func() {
		if err := wrapSQLErr(tx.Rollback()); err != nil && !errors.Is(err, sql.ErrTxDone) {
			r.log.Error("rollback error", zap.Error(err))
		}
	}()

	userSt := r.db.QueryBuilder.
		Insert("users").
		Columns("login", "password").
		Values(user.Login, user.Password)

	sqlstr, args, err := userSt.ToSql()
	if err := wrapStatmentErr(err); err != nil {
		return nil, err
	}

	sqlRes, err := tx.ExecContext(ctx, sqlstr, args...)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}
	id, err := sqlRes.LastInsertId()
	// err = tx.QueryRowContext(ctx, "SELECT last_insert_rowid()").Scan(&(user.ID))
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}
	user.ID = domain.UserID(id)

	err = tx.Commit()
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) selectUser(ctx context.Context, tx queryAble, login string) (*domain.User, error) {
	statement := r.db.QueryBuilder.
		Select("id", "login", "password").
		From("users").
		Where(sq.Eq{"login": login})

	sqlstr, args, err := statement.ToSql()
	if err := wrapStatmentErr(err); err != nil {
		return nil, err
	}

	user := domain.User{}

	err = tx.QueryRowContext(ctx, sqlstr, args...).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
	)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByLogin finds user.
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*domain.User, error) {
	return r.selectUser(ctx, r.db, login)
}
