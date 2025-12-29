package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

// UserPhotos retrieves a list of user photos based on the provided filter.
func (s *Store) UserPhotos(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.UserPhotoQueryFilter,
) ([]registration.UserPhoto, error) {
	pgUserPhotos, err := s.repoBackendApp.UserPhotos(ctx,
		exec,
		newRepoUserPhotoQueryFilter(filter),
	)
	if err != nil {
		return nil, fmt.Errorf("list blocked contacts: %w", err)
	}

	return newUserPhotosFromSlice(pgUserPhotos), nil
}

func newRepoUserPhotoQueryFilter(filter *registration.UserPhotoQueryFilter) *repo.UserPhotoQueryFilter {
	return &repo.UserPhotoQueryFilter{
		ID:     filter.ID,
		UserID: filter.UserID,
		Bucket: filter.Bucket,
		Key:    filter.Key,
		Order:  filter.Order,
	}
}

func newUserPhotosFromSlice(pgUserPhotos pgmodel.UserPhotoSlice) []registration.UserPhoto {
	userPhotos := make([]registration.UserPhoto, 0, len(pgUserPhotos))
	for _, pgUserPhoto := range pgUserPhotos {
		userPhotos = append(userPhotos, registration.UserPhoto{
			ID:      pgUserPhoto.ID,
			UserID:  pgUserPhoto.UserID,
			Bucket:  pgUserPhoto.Bucket,
			Key:     pgUserPhoto.Key,
			OrderNo: pgUserPhoto.OrderNo,
		})
	}

	return userPhotos
}

// UserPhoto gets details of a specific user photo based on the provided filter.
func (s *Store) UserPhoto(ctx context.Context,
	exec boil.ContextExecutor,
	filter *registration.UserPhotoQueryFilter,
) (*registration.UserPhoto, error) {
	userPhotos, err := s.UserPhotos(ctx, exec, filter)
	if err != nil {
		return nil, fmt.Errorf("list blocked contacts: %w", err)
	}
	if len(userPhotos) == 0 {
		return nil, registration.ErrUserPhotoNotFound
	}
	if len(userPhotos) != 1 {
		return nil, fmt.Errorf("user photo count mismatch, have %d, want 1", len(userPhotos))
	}
	return &userPhotos[0], nil
}

// InsertUserPhoto inserts a new user photo record into the database.
func (s *Store) InsertUserPhoto(ctx context.Context, exec boil.ContextExecutor, userID, bucket, key string, orderNo int) (*registration.UserPhoto, error) {
	repodUserPhoto, err := s.repoBackendApp.InsertUserPhoto(ctx, exec, userID, bucket, key, orderNo)
	if err != nil {
		return nil, fmt.Errorf("insert user photo into db: %w", err)
	}
	return newUserPhotoFromPG(repodUserPhoto), nil
}

func newUserPhotoFromPG(userPhoto *pgmodel.UserPhoto) *registration.UserPhoto {
	return &registration.UserPhoto{
		ID:      userPhoto.ID,
		UserID:  userPhoto.UserID,
		Bucket:  userPhoto.Bucket,
		Key:     userPhoto.Key,
		OrderNo: userPhoto.OrderNo,
	}
}

// DeletePhoto deletes a user photo based on the provided UserID and returns the remaining photos.
func (s *Store) DeletePhoto(ctx context.Context,
	exec boil.ContextExecutor,
	id string,
) error {
	return s.repoBackendApp.DeleteUserPhoto(ctx, exec, id)
}
