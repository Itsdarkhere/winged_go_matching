package registration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"wingedapp/pgtester/internal/util/errutil"
	"wingedapp/pgtester/internal/util/numrange"

	"github.com/aarondl/null/v8"
	"github.com/alitto/pond"
	"github.com/google/uuid"
)

func (b *Business) DeleteUserPhoto(ctx context.Context, user *User, id string) error {
	filter := &UserPhotoQueryFilter{
		ID:     null.StringFrom(id),
		UserID: null.StringFrom(user.ID),
	}
	userPhoto, err := b.storer.UserPhoto(ctx, b.transBE.DB(), filter)
	if err != nil {
		return fmt.Errorf("get user photo: %w", err)
	}
	if err = b.storer.DeletePhoto(ctx, b.transBE.DB(), userPhoto.ID); err != nil {
		return fmt.Errorf("delete user photo: %w", err)
	}

	return nil
}

func (b *Business) UserPhoto(ctx context.Context, userId string) (*UserPhoto, error) {
	filter := UserPhotoQueryFilter{
		UserID: null.StringFrom(userId),
	}

	userPhoto, err := b.storer.UserPhoto(ctx, b.transBE.DB(), &filter)
	if err != nil {
		return nil, fmt.Errorf("get user photo: %w", err)
	}

	return userPhoto, nil
}

func (b *Business) UserPhotos(ctx context.Context, userId string, enrichPhotos bool) ([]UserPhoto, error) {
	filter := UserPhotoQueryFilter{
		UserID: null.StringFrom(userId),
	}

	userPhotos, err := b.storer.UserPhotos(ctx, b.transBE.DB(), &filter)
	if err != nil {
		return nil, fmt.Errorf("get user photo: %w", err)
	}

	if enrichPhotos {
		if err = b.enrichUserPhotosWithURLS(ctx, userPhotos); err != nil {
			return nil, fmt.Errorf("enrich user photos with urls: %w", err)
		}
	}

	return userPhotos, nil
}

// enrichUserPhotosWithURLS adds public URLs to user photos concurrently.
func (b *Business) enrichUserPhotosWithURLS(ctx context.Context, userPhotos []UserPhoto) error {
	mUserPhoto := make(map[string]int)
	for i, userPhoto := range userPhotos {
		mUserPhoto[userPhoto.Key] = i
	}

	var m sync.Mutex
	var errList errutil.List

	pondLimit := b.cfg.UserMaxPhotos
	p := pond.New(pondLimit, pondLimit)
	fnEnrich := func(key string) func() {
		return func() {
			url, err := b.beUploader.PublicURL(ctx, key)
			if err != nil {
				errList.AddErr(fmt.Errorf("getting public url for key %s: %w", key, err))
				return
			}
			m.Lock()
			defer m.Unlock()
			userPhotos[mUserPhoto[key]].URL = url // add public url
		}
	}
	for key := range mUserPhoto {
		p.Submit(fnEnrich(key))
	}
	p.StopAndWait()
	return errList.Error()
}

// AddUserPhoto updates photo of user, and uploads it to a storage service.
func (b *Business) AddUserPhoto(ctx context.Context, user *User, addUserPhoto *UserPhoto, replace bool) (*UserPhoto, error) {
	// validate photo counts
	userPhotoCount := len(user.Photos)
	if (userPhotoCount >= b.cfg.UserMaxPhotos) && !replace {
		return nil, ErrUserMaxPhotoCountExceeded
	}

	orderNo := addUserPhoto.OrderNo
	if orderNo == 0 {
		orderNo = 1 // default to 1
	}
	beginVal := 1
	maxVal := b.cfg.UserMaxPhotos
	if numrange.NotBetween(orderNo, beginVal, maxVal) {
		return nil, errors.Join(
			fmt.Errorf("must be between %d and %d, got: %d", beginVal, maxVal, orderNo),
			ErrUserPhotoInvalidOrder,
		)
	}

	tx, err := b.transBE.TX()
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer b.transBE.Rollback(tx)

	// handle if already existing
	uploadKey := b.BucketKeyTemplate(user.ID, addUserPhoto.Filename)
	userPhoto, err := b.storer.UserPhoto(ctx, tx, &UserPhotoQueryFilter{
		UserID: null.StringFrom(user.ID),
		Order:  null.IntFrom(addUserPhoto.OrderNo),
	})
	if err != nil && !errors.Is(err, ErrUserPhotoNotFound) {
		return nil, fmt.Errorf("check existing user photo: %w", err)
	}

	alreadyExists := userPhoto != nil
	if alreadyExists {
		if replace {
			if err = b.DeleteUserPhoto(ctx, user, userPhoto.ID); err != nil {
				return nil, fmt.Errorf("delete existing user photo: %w", err)
			}
		} else {
			return nil, errors.Join(
				ErrUserPhotoAlreadyExists,
				fmt.Errorf("overlap order: %v", addUserPhoto.OrderNo),
			)
		}
	}

	// handle colliding keys
	userPhotoCollidingKey, err := b.storer.UserPhoto(ctx, tx, &UserPhotoQueryFilter{
		Key: null.StringFrom(uploadKey),
	})
	if err != nil && !errors.Is(err, ErrUserPhotoNotFound) {
		return nil, fmt.Errorf("check colliding user photo key: %w", err)
	}
	collidingKey := userPhotoCollidingKey != nil
	if collidingKey {
		uploadKey += "-" + uuid.NewString() // mutate key to be unique
	}

	// persist on storage, and db
	uploadedUserPhoto, err := b.beUploader.Upload(ctx, uploadKey, addUserPhoto.Bytes)
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}
	insertedUserPhoto, err := b.storer.InsertUserPhoto(ctx, tx, user.ID, uploadedUserPhoto.Bucket, uploadKey, addUserPhoto.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("insert user photo: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return insertedUserPhoto, nil
}

// BucketKeyTemplate generates a bucket key template for storing user photos.
// ensure it's unique by appending a uuid to the filename.
func (b *Business) BucketKeyTemplate(userID, filename string) string {
	return fmt.Sprintf("%s/%s", userID, filename)
}
