package repository

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/MikeRez0/gophkeeper/internal/adapter/storage"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type KeychainRepository struct {
	db *storage.DB
}

func NewKeychainRepository(db *storage.DB) (*KeychainRepository, error) {
	return &KeychainRepository{db: db}, nil
}

func (r *KeychainRepository) KeyChainInsert(ctx context.Context, kcdata *domain.KCData) (*domain.KCData, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		stat := r.db.QueryBuilder.
			Insert("keychain").
			Columns("id", "owner_id", "name").
			Values(kcdata.ID, kcdata.OwnerID, kcdata.Name)

		sql, args, err := stat.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, sql, args...)
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

	return kcdata, nil
}

func (r *KeychainRepository) KeyChainUpdate(ctx context.Context, kcdata *domain.KCData) (*domain.KCData, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		updateSt := r.db.QueryBuilder.
			Update("keychain").
			Set("name", kcdata.Name).
			Set("owner_id", kcdata.OwnerID).
			Where(sq.Eq{"id": kcdata.ID})

		sql, args, err := updateSt.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, sql, args...)
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

	return kcdata, nil
}

func (r *KeychainRepository) KeyChainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error) {
	return r.selectKeychainList(ctx, r.db.Pool, domain.KeychainID(uuid.Nil), user, false)
}

func (r *KeychainRepository) KeyChainGet(ctx context.Context, keychainID domain.KeychainID) (*domain.KCData, error) {
	list, err := r.selectKeychainList(ctx, r.db.Pool, keychainID, 0, false)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, domain.ErrDataNotFound
	}
	return list[0], nil
}

func (r *KeychainRepository) KeyChainItemUpsert(ctx context.Context,
	item *domain.KCItemData) (*domain.KCItemData, bool, error) {

	var (
		updated bool
		result  *domain.KCItemData
	)

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var oldItem *domain.KCItemData
		if items, err := r.selectKeyChainItems(ctx, tx, item.KeyChainID, item.ID, time.Time{}, true); err == nil {
			if len(items) > 0 {
				oldItem = items[0]
			}
		} else if !errors.Is(err, domain.ErrDataNotFound) {
			return err
		}

		var (
			sql  string
			args []interface{}
			err  error
		)

		if oldItem == nil {
			insertSt := r.db.QueryBuilder.
				Insert("keychain_item").
				Columns("id", "keychain_id", "item_type", "label", "enc_key", "enc_value", "client_ts", "server_ts").
				Values(item.ID, item.KeyChainID, item.ItemType, item.Label, item.Key, item.Value, item.ClientTime, item.ServerTime)

			sql, args, err = insertSt.ToSql()
			if err != nil {
				return err
			}
		} else {

			if oldItem.ClientTime.After(item.ClientTime) {
				result = oldItem
				updated = false
				return nil
			}

			updateSt := r.db.QueryBuilder.
				Update("keychain_item").
				Set("item_type", item.ItemType).
				Set("label", item.Label).
				Set("enc_key", item.Key).
				Set("enc_value", item.Value).
				Set("client_ts", item.ClientTime).
				Set("server_ts", item.ServerTime).
				Where(sq.Eq{"id": item.ID,
					"keychain_id": item.KeyChainID})

			sql, args, err = updateSt.ToSql()
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}

		deleteSt := r.db.QueryBuilder.
			Delete("keychain_item_meta").
			Where(sq.Eq{"keychain_item_id": item.ID,
				"keychain_id": item.KeyChainID})
		sql, args, err = deleteSt.ToSql()
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}

		insertSt := r.db.QueryBuilder.
			Insert("keychain_item_meta").
			Columns("keychain_item_id", "keychain_id", "k", "v")

		for k, v := range item.MetaData {
			s := insertSt.Values(item.ID, item.KeyChainID, k, v)
			sql, args, err = s.ToSql()
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, sql, args...)
			if err != nil {
				return err
			}
		}
		result = item
		updated = true
		return nil
	})

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, false, domain.ErrConflictingData
		}
		return nil, false, err
	}

	return result, updated, nil
}

func (r *KeychainRepository) KeyChainItemSelect(ctx context.Context, keyChainID domain.KeychainID,
	itemID domain.KeychainItemID) (*domain.KCItemData, error) {
	items, err := r.selectKeyChainItems(ctx, r.db.Pool, keyChainID, itemID, time.Time{}, false)
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

func (r *KeychainRepository) KeyChainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	return r.selectKeyChainItems(ctx, r.db, keyChainID, domain.KeychainItemID(uuid.Nil), since, false)
}

func (r *KeychainRepository) selectKeyChainItems(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID,
	itemID domain.KeychainItemID,
	since time.Time,
	forUpdate bool) ([]*domain.KCItemData, error) {
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

	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

	list := make([]*domain.KCItemData, 0)
	item_ids := make([]domain.KeychainItemID, 0)
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
		item_ids = append(item_ids, item.ID)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	statement = r.db.QueryBuilder.
		Select("keychain_item_id", "k", "v").
		From("keychain_item_meta").
		Where(sq.Eq{"keychain_item_id": item_ids, "keychain_id": keyChainID})
	sql, args, err = statement.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err = tx.Query(ctx, sql, args...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

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

func (r *KeychainRepository) selectKeychainList(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID, userID domain.UserID, forUpdate bool) ([]*domain.KCData, error) {
	statement := r.db.QueryBuilder.
		Select("id", "owner_id", "name").
		From("keychain")

	if keyChainID != domain.KeychainID(uuid.Nil) {
		statement = statement.Where(sq.Eq{"id": &keyChainID})
	}

	if userID != 0 {
		statement = statement.Where(sq.Eq{"owner_id": userID})
	}

	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDataNotFound
		}
		return nil, err
	}

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
	if err != nil {
		return nil, err
	}

	return list, nil
}
