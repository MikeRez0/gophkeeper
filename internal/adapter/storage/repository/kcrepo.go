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
			Columns("id", "owner_id", "name", "create_at", "update_at").
			Values(kcdata.ID, kcdata.OwnerID, kcdata.Name, kcdata.Created, kcdata.Changed).
			Suffix("returning id")

		sql, args, err := stat.ToSql()
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, sql, args...).Scan(&(kcdata.ID))
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
	return r.selectKeyChainList(ctx, r.db.Pool, domain.KeychainID(uuid.Nil), user)
}

func (r *KeychainRepository) KeyChainItemUpsert(ctx context.Context,
	item *domain.KCItemData) (*domain.KCItemData, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		itemSt := r.db.QueryBuilder.
			Insert("keychain_item").
			Columns("id", "keychain_id", "item_type", "label", "enc_key", "enc_value", "create_at", "update_at").
			Values(item.ID, item.KeyChainID, item.ItemType, item.Label, item.Key, item.Value, time.Now(), time.Now()).
			Suffix("returning id")

		sql, args, err := itemSt.ToSql()
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, sql, args...).Scan(&(item.ID))
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

	return item, nil
}

func (r *KeychainRepository) KeyChainItemSelect(ctx context.Context, keyChainID domain.KeychainID,
	itemID domain.KeychainItemID) (*domain.KCItemData, error) {
	items, err := r.selectKeyChainItems(ctx, r.db.Pool, keyChainID, itemID, time.Time{})
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

func (r *KeychainRepository) KeyChainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
	since time.Time) ([]*domain.KCItemData, error) {
	return r.selectKeyChainItems(ctx, r.db, keyChainID, domain.KeychainItemID(uuid.Nil), since)
}

func (r *KeychainRepository) selectKeyChainItems(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID, itemID domain.KeychainItemID, since time.Time) ([]*domain.KCItemData, error) {
	statement := r.db.QueryBuilder.
		Select("id", "keychain_id", "item_type", "label", "enc_key", "enc_value", "create_at", "update_at").
		From("keychain_item").
		Where(sq.Eq{"keychain_id": keyChainID})

	if itemID != domain.KeychainItemID(uuid.Nil) {
		statement = statement.Where(sq.Eq{"id": itemID})
	}

	if !since.IsZero() {
		statement = statement.Where(sq.GtOrEq{"update_at": since})
	}

	statement = statement.OrderBy("update_at desc")

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
	for rows.Next() {
		item := domain.KCItemData{}
		err := rows.Scan(
			&item.ID,
			&item.KeyChainID,
			&item.ItemType,
			&item.Label,
			&item.Key,
			&item.Value,
			&item.Created,
			&item.Changed,
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

func (r *KeychainRepository) selectKeyChainList(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID, userID domain.UserID) ([]*domain.KCData, error) {
	statement := r.db.QueryBuilder.
		Select("id", "owner_id", "create_at", "update_at").
		From("keychain")

	if keyChainID != domain.KeychainID(uuid.Nil) {
		statement = statement.Where(sq.Eq{"keychain_id": keyChainID})
	}

	if userID != 0 {
		statement = statement.Where(sq.Eq{"owner": userID})
	}

	statement = statement.OrderBy("update_at desc")

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
			&item.Created,
			&item.Changed,
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
