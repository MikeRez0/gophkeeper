package pgsql

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
)

type KeychainPgRepository struct {
	db  *DB
	log *zap.Logger
}

func NewKeychainPgRepository(db *DB, log *zap.Logger) (*KeychainPgRepository, error) {
	if db == nil || log == nil {
		return nil, errors.New("nil not allowed as argument")
	}
	return &KeychainPgRepository{db: db, log: log}, nil
}

func (r *KeychainPgRepository) KeychainUpsert(ctx context.Context, kcdata *domain.KCData) (*domain.KCData, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		klist, err := r.selectKeychainList(ctx, tx, kcdata.ID, domain.UserID(0), true)
		if err != nil {
			return err
		}
		if len(klist) == 0 {
			stat := r.db.QueryBuilder.
				Insert("keychain").
				Columns("id", "owner_id", "name").
				Values(kcdata.ID, kcdata.OwnerID, kcdata.Name)

			sql, args, err := stat.ToSql()
			if err := wrapStatmentErr(err); err != nil {
				return err
			}

			_, err = tx.Exec(ctx, sql, args...)
			if err := wrapSQLErr(err); err != nil {
				return err
			}
		} else {
			updateSt := r.db.QueryBuilder.
				Update("keychain").
				Set("name", kcdata.Name).
				Set("owner_id", kcdata.OwnerID).
				Where(sq.Eq{"id": kcdata.ID})

			sql, args, err := updateSt.ToSql()
			if err := wrapStatmentErr(err); err != nil {
				return err
			}

			_, err = tx.Exec(ctx, sql, args...)
			if err := wrapSQLErr(err); err != nil {
				return err
			}
		}

		return nil
	})

	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return kcdata, nil
}

func (r *KeychainPgRepository) KeychainList(ctx context.Context, user domain.UserID) ([]*domain.KCData, error) {
	return r.selectKeychainList(ctx, r.db.Pool, domain.KeychainID(uuid.Nil), user, false)
}

func (r *KeychainPgRepository) KeychainGet(ctx context.Context, keychainID domain.KeychainID) (*domain.KCData, error) {
	list, err := r.selectKeychainList(ctx, r.db.Pool, keychainID, 0, false)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, domain.ErrDataNotFound
	}
	return list[0], nil
}

func (r *KeychainPgRepository) KeychainItemUpsert(ctx context.Context,
	item *domain.KCItemData) (*domain.KCItemData, bool, error) {
	var (
		updated bool
		result  *domain.KCItemData
	)

	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var oldItem *domain.KCItemData
		if items, err := r.selectKeychainItems(ctx, tx,
			item.KeyChainID, item.ID, time.Time{}, time.Time{}, true); err == nil {
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
				Values(item.ID, item.KeyChainID, item.ItemType, item.Label, item.Key, item.Value,
					item.ClientTime.UTC(), item.ServerTime.UTC())

			sql, args, err = insertSt.ToSql()
			if err := wrapStatmentErr(err); err != nil {
				return err
			}
		} else {
			if oldItem.ClientTime.UTC().After(item.ClientTime.UTC()) {
				r.log.Debug("old item saved, because it is after",
					zap.Time("CT_old", oldItem.ClientTime.UTC()),
					zap.Time("CT_new", item.ClientTime.UTC()))
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
				Set("client_ts", item.ClientTime.UTC()).
				Set("server_ts", item.ServerTime.UTC()).
				Where(sq.Eq{"id": item.ID,
					"keychain_id": item.KeyChainID})

			sql, args, err = updateSt.ToSql()
			if err := wrapStatmentErr(err); err != nil {
				return err
			}
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err := wrapSQLErr(err); err != nil {
			return err
		}

		deleteSt := r.db.QueryBuilder.
			Delete("keychain_item_meta").
			Where(sq.Eq{"keychain_item_id": item.ID,
				"keychain_id": item.KeyChainID})
		sql, args, err = deleteSt.ToSql()
		if err := wrapStatmentErr(err); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, sql, args...)
		if err := wrapSQLErr(err); err != nil {
			return err
		}

		insertSt := r.db.QueryBuilder.
			Insert("keychain_item_meta").
			Columns("keychain_item_id", "keychain_id", "k", "v")

		for k, v := range item.MetaData {
			s := insertSt.Values(item.ID, item.KeyChainID, k, v)
			sql, args, err = s.ToSql()
			if err := wrapStatmentErr(err); err != nil {
				return err
			}
			_, err = tx.Exec(ctx, sql, args...)
			if err := wrapSQLErr(err); err != nil {
				return err
			}
		}
		result = item
		updated = true
		return nil
	})

	if err := wrapSQLErr(err); err != nil {
		return nil, false, err
	}

	return result, updated, nil
}

func (r *KeychainPgRepository) KeychainItemSelect(ctx context.Context, keyChainID domain.KeychainID,
	itemID domain.KeychainItemID) (*domain.KCItemData, error) {
	items, err := r.selectKeychainItems(ctx, r.db.Pool, keyChainID, itemID, time.Time{}, time.Time{}, false)
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

func (r *KeychainPgRepository) KeychainGetItemsSince(ctx context.Context, keyChainID domain.KeychainID,
	sinceClient time.Time, sinceServer time.Time) ([]*domain.KCItemData, error) {
	return r.selectKeychainItems(ctx, r.db, keyChainID, domain.KeychainItemID(uuid.Nil), sinceClient, sinceServer, false)
}

func (r *KeychainPgRepository) selectKeychainItems(ctx context.Context, tx queryAble,
	keyChainID domain.KeychainID,
	itemID domain.KeychainItemID,
	sinceClient time.Time, sinceServer time.Time,
	forUpdate bool) ([]*domain.KCItemData, error) {
	statement := r.db.QueryBuilder.
		Select("id", "keychain_id", "item_type", "label", "enc_key", "enc_value", "client_ts", "server_ts").
		From("keychain_item").
		Where(sq.Eq{"keychain_id": keyChainID})

	if itemID != domain.KeychainItemID(uuid.Nil) {
		statement = statement.Where(sq.Eq{"id": itemID})
	}

	if !sinceClient.IsZero() {
		statement = statement.Where(sq.Gt{"client_ts": sinceClient})
	}
	if !sinceServer.IsZero() {
		statement = statement.Where(sq.Gt{"server_ts": sinceServer})
	}

	statement = statement.OrderBy("client_ts desc")

	if forUpdate {
		statement = statement.Suffix("FOR UPDATE")
	}

	sql, args, err := statement.ToSql()
	if err := wrapStatmentErr(err); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	list := make([]*domain.KCItemData, 0)
	itemIDs := make([]domain.KeychainItemID, 0)
	defer rows.Close()
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
		if err := wrapSQLErr(err); err != nil {
			return nil, err
		}
		list = append(list, &item)
		itemIDs = append(itemIDs, item.ID)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("read row error: %w", err)
	}

	if len(list) == 0 {
		return nil, domain.ErrDataNotFound
	}
	statement = r.db.QueryBuilder.
		Select("keychain_item_id", "k", "v").
		From("keychain_item_meta").
		Where(sq.Eq{"keychain_item_id": itemIDs, "keychain_id": keyChainID})
	sql, args, err = statement.ToSql()
	if err := wrapStatmentErr(err); err != nil {
		return nil, err
	}
	rowsMeta, err := tx.Query(ctx, sql, args...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, wrapSQLErr(err)
	}
	defer rowsMeta.Close()

	for rowsMeta.Next() {
		var (
			i    domain.KeychainItemID
			k, v string
		)
		err := rowsMeta.Scan(&i, &k, &v)
		if err := wrapSQLErr(err); err != nil {
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
	err = rowsMeta.Err()
	if err := wrapSQLErr(err); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *KeychainPgRepository) selectKeychainList(ctx context.Context, tx queryAble,
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
	if err := wrapStatmentErr(err); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
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
		if err := wrapSQLErr(err); err != nil {
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return domain.ErrConflictingData
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrDataNotFound
		}
		return err
	}
	return nil
}

func wrapStatmentErr(err error) error {
	if err != nil {
		return fmt.Errorf("statemt build error: %w", err)
	}
	return nil
}
