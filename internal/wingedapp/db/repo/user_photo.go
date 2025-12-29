package repo

import (
	"context"
	"errors"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/jackc/pgx/v4"
)

// InsertUserPhoto inserts a photo for a user.
func (s *Store) InsertUserPhoto(ctx context.Context,
	db boil.ContextExecutor,
	userID,
	bucket,
	key string,
	orderNo int,
) (*pgmodel.UserPhoto, error) {
	userPhoto := &pgmodel.UserPhoto{
		UserID:  userID,
		Bucket:  bucket,
		Key:     key,
		OrderNo: orderNo,
	}

	if err := userPhoto.Insert(ctx, db, boil.Infer()); err != nil {
		return nil, fmt.Errorf("insert userPhoto with key (%s): %w", key, err)
	}

	return userPhoto, nil
}

type UserPhotoQueryFilter struct {
	ID     null.String `json:"id"`
	UserID null.String `json:"user_id"`
	Bucket null.String `json:"bucket"`
	Key    null.String `json:"key"`
	Order  null.Int    `json:"order"`
}

func userPhotoFilter(filter *UserPhotoQueryFilter) []qm.QueryMod {
	filters := make([]qm.QueryMod, 0)

	if filter.ID.Valid {
		filters = append(filters, pgmodel.UserPhotoWhere.ID.EQ(filter.ID.String))
	}
	if filter.UserID.Valid {
		filters = append(filters, pgmodel.UserPhotoWhere.UserID.EQ(filter.UserID.String))
	}
	if filter.Bucket.Valid {
		filters = append(filters, pgmodel.UserPhotoWhere.Bucket.EQ(filter.Bucket.String))
	}
	if filter.Key.Valid {
		filters = append(filters, pgmodel.UserPhotoWhere.Key.EQ(filter.Key.String))
	}
	if filter.Order.Valid {
		filters = append(filters, pgmodel.UserPhotoWhere.OrderNo.EQ(filter.Order.Int))
	}

	return filters
}

// UserPhotos lists all the userPhotos based on the filter.
func (s *Store) UserPhotos(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserPhotoQueryFilter,
) (pgmodel.UserPhotoSlice, error) {
	userPhotos, err := pgmodel.UserPhotos(userPhotoFilter(filter)...).All(ctx, exec)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgmodel.UserPhotoSlice{}, nil // no results found
		}
		return nil, fmt.Errorf("query userPhotos: %w", err)
	}
	return userPhotos, nil
}

// UserPhoto returns a single userPhoto based on the filter.
func (s *Store) UserPhoto(ctx context.Context,
	exec boil.ContextExecutor,
	filter *UserPhotoQueryFilter,
) (*pgmodel.UserPhoto, error) {
	userPhotos, err := s.UserPhotos(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("user photo: %w", err)
	}

	if len(userPhotos) == 0 {
		return &pgmodel.UserPhoto{}, nil
	}

	if len(userPhotos) != 1 {
		return nil, fmt.Errorf("user photo count mismatch, have %d, want 1", len(userPhotos))
	}

	return userPhotos[0], nil
}

// UpsertUserPhoto upserts a photo for a user.
func (s *Store) UpsertUserPhoto(
	ctx context.Context,
	db boil.ContextExecutor,
	userID,
	bucket,
	key string,
) error {
	userPhoto := pgmodel.UserPhoto{
		UserID: userID,
		Bucket: bucket,
		Key:    key,
	}

	conflictCols := []string{
		pgmodel.UserPhotoColumns.UserID,
		pgmodel.UserPhotoColumns.Bucket,
		pgmodel.UserPhotoColumns.Key,
	}

	if err := userPhoto.Upsert(ctx, db, true, conflictCols, boil.Infer(), boil.Infer()); err != nil {
		return fmt.Errorf("upsert user photo: %w", err)
	}

	return nil
}

// DeleteUserPhoto deletes a photo for a user.
func (s *Store) DeleteUserPhoto(
	ctx context.Context,
	db boil.ContextExecutor,
	id string,
) error {
	userPhoto := pgmodel.UserPhoto{ID: id}
	count, err := userPhoto.Delete(ctx, db)
	if err != nil {
		return fmt.Errorf("delete blocked user: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("delete blocked user: no rows affected")
	}

	return nil
}
