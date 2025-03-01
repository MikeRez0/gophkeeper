package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type KeychainSqliteRepository struct {
	db  *DB
	log *zap.Logger
}

type queryAble interface {
	QueryRowContext(ctx context.Context, sql string, args ...any) *sql.Row
	QueryContext(ctx context.Context, sql string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, sql string, arguments ...any) (sql.Result, error)
}

func NewKeychainSqliteRepository(db *DB, log *zap.Logger) (*KeychainSqliteRepository, error) {
	if db == nil || log == nil {
		return nil, errors.New("nil not allowed as argument")
	}
	return &KeychainSqliteRepository{db: db, log: log}, nil
}

func (r *KeychainSqliteRepository) KeychainUpsert(ctx context.Context, kcdata *domain.KCData) (*domain.KCData, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		rerr := tx.Rollback()
		if rerr != nil && !errors.Is(rerr, sql.ErrTxDone) {
			r.log.Error("rollback error", zap.Error(rerr))
		}
	}()

	klist, err := r.selectKeychainList(ctx, tx, kcdata.ID, domain.UserID(0))
	if err != nil {
		return nil, err
	}

	if len(klist) == 0 {
		stat := r.db.QueryBuilder.
			Insert("keychain").
			Columns("id", "owner_id", "name").
			Values(kcdata.ID, kcdata.OwnerID, kcdata.Name)

		sql, args, err := stat.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, sql, args...)
		if err != nil {
			var sqlErr sqlite3.Error
			if errors.As(err, &sqlErr) {
				if errors.Is(sqlErr, sqlite3.ErrConstraint) {
					return nil, domain.ErrConflictingData
				}
			}
			return nil, err
		}
	} else {
		updateSt := r.db.QueryBuilder.
			Update("keychain").
			Set("name", kcdata.Name).
			Set("owner_id", kcdata.OwnerID).
			Where(sq.Eq{"id": kcdata.ID})

		sql, args, err := updateSt.ToSql()
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return kcdata, nil
}

func (r *KeychainSqliteRepository) KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error) {
	return r.selectKeychainList(ctx, r.db, domain.KeychainID(uuid.Nil), user)
}

func (r *KeychainSqliteRepository) KeychainGet(ctx context.Context,
	keychainID domain.KeychainID) (*domain.KCData, error) {
	list, err := r.selectKeychainList(ctx, r.db, keychainID, 0)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, domain.ErrDataNotFound
	}
	return list[0], nil
}

func (r *KeychainSqliteRepository) KeychainItemUpsert(ctx context.Context,
	item *domain.KCItemData) (*domain.KCItemData, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		rerr := tx.Rollback()
		if rerr != nil && errors.Is(err, sql.ErrConnDone) {
			r.log.Error("rollback error", zap.Error(rerr))
		}
	}()

	var (
		sql  string
		args []interface{}
	)

	upsertSt := r.db.QueryBuilder.
		Insert("keychain_item").
		Columns("id", "keychain_id", "item_type", "label", "enc_key", "enc_value", "client_ts", "server_ts").
		Values(item.ID, item.KeyChainID, item.ItemType, item.Label, item.Key, item.Value,
			item.ClientTime.UTC(), item.ServerTime.UTC()).
		Suffix(`ON CONFLICT (id) DO UPDATE SET 
				item_type = excluded.item_type, 
				label = excluded.label, 
				enc_key = excluded.enc_key, 
				enc_value = excluded.enc_value, 
				client_ts = excluded.client_ts,
				server_ts = excluded.server_ts
			WHERE excluded.client_ts >= keychain_item.client_ts`)

	sql, args, err = upsertSt.ToSql()
	if err != nil {
		return nil, false, err
	}

	sqlRes, err := tx.ExecContext(ctx, sql, args...)
	if err := wrapSQLErr(err); err != nil {
		return nil, false, err
	}
	rows, err := sqlRes.RowsAffected()
	if err := wrapSQLErr(err); err != nil {
		return nil, false, err
	}
	// check: if rows affected is 0, then need to return actual record from database
	if rows == 0 {
		//TODO: fill item from database
		return item, false, nil
	}

	deleteSt := r.db.QueryBuilder.
		Delete("keychain_item_meta").
		Where(sq.Eq{"keychain_item_id": item.ID,
			"keychain_id": item.KeyChainID})
	sql, args, err = deleteSt.ToSql()
	if err != nil {
		return nil, false, err
	}
	_, err = tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, false, err
	}

	insertSt := r.db.QueryBuilder.
		Insert("keychain_item_meta").
		Columns("keychain_item_id", "keychain_id", "k", "v")

	for k, v := range item.MetaData {
		s := insertSt.Values(item.ID, item.KeyChainID, k, v)
		sql, args, err = s.ToSql()
		if err != nil {
			return nil, false, err
		}
		_, err = tx.ExecContext(ctx, sql, args...)
		if err != nil {
			return nil, false, err
		}
	}

	err = tx.Commit()
	if err := wrapSQLErr(err); err != nil {
		return nil, false, err
	}

	return item, true, nil
}

func (r *KeychainSqliteRepository) KeychainItemSelect(ctx context.Context, keyChainID domain.KeychainID,
	itemID domain.KeychainItemID) (*domain.KCItemData, error) {
	items, err := r.selectKeychainItems(ctx, r.db, keyChainID, itemID, time.Time{})
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

func (r *KeychainSqliteRepository) KeychainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	return r.selectKeychainItems(ctx, r.db, keyChainID, domain.KeychainItemID(uuid.Nil), since)
}

func (r *KeychainSqliteRepository) selectKeychainItems(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID,
	itemID domain.KeychainItemID,
	since time.Time) ([]*domain.KCItemData, error) {
	statement := r.db.QueryBuilder.
		Select("id", "keychain_id", "item_type", "label", "enc_key", "enc_value", "client_ts", "server_ts").
		From("keychain_item").
		Where(sq.Eq{"keychain_id": keyChainID})

	if itemID != domain.KeychainItemID(uuid.Nil) {
		statement = statement.Where(sq.Eq{"id": itemID})
	}

	if !since.IsZero() {
		statement = statement.Where(sq.GtOrEq{"server_ts": since})
	}

	statement = statement.OrderBy("client_ts desc")

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*domain.KCItemData, 0)
	itemIDs := make([]domain.KeychainItemID, 0)
	for rows.Next() {
		item := domain.KCItemData{}
		err := rows.Scan(
			&item.ID,
			&item.KeyChainID,
			&item.ItemType,
			&item.Label,
			&item.Key,
			&item.Value,
			&item.ClientTime,
			&item.ServerTime,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &item)
		itemIDs = append(itemIDs, item.ID)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, domain.ErrDataNotFound
	}

	statement = r.db.QueryBuilder.
		Select("keychain_item_id", "k", "v").
		From("keychain_item_meta").
		Where(sq.Eq{"keychain_item_id": itemIDs, "keychain_id": keyChainID})
	sql, args, err = statement.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err = tx.QueryContext(ctx, sql, args...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			i    domain.KeychainItemID
			k, v string
		)
		err := rows.Scan(&i, &k, &v)
		if err != nil {
			return nil, err
		}
		for _, j := range list {
			if j.ID == i {
				if j.MetaData == nil {
					j.MetaData = make(domain.KeychainItemMeta, 0)
				}
				j.MetaData[k] = v
				break
			}
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (r *KeychainSqliteRepository) selectKeychainList(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID, userID domain.UserID) ([]*domain.KCData, error) {
	statement := r.db.QueryBuilder.
		Select("id", "owner_id", "name").
		From("keychain")

	if keyChainID != domain.KeychainID(uuid.Nil) {
		statement = statement.Where(sq.Eq{"id": &keyChainID})
	}

	if userID != 0 {
		statement = statement.Where(sq.Eq{"owner_id": userID})
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, sql, args...)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*domain.KCData, 0)
	for rows.Next() {
		item := domain.KCData{}
		err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.Name,
		)
		list = append(list, &item)
		if err != nil {
			return nil, err
		}
	}

	err = rows.Err()
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return list, nil
}

func wrapSQLErr(err error) error {
	if err != nil {
		var sqlErr sqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.Code == sqlite3.ErrConstraint {
				return domain.ErrConflictingData
			}
		}
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrDataNotFound
		}
		return err
	}
	return nil
}
