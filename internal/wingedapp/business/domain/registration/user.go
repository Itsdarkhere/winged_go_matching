package registration

import (
	"context"
	"fmt"
	"sync"
	"wingedapp/pgtester/internal/util/errutil"

	"github.com/aarondl/null/v8"
	"github.com/alitto/pond"
)

func (b *Business) enrichAudioFiles(ctx context.Context, u *User) error {
	// Should be of the format `folder/subfolder/filename.png`
	mUserAudio := make(map[string]int)
	for i, userAudio := range u.AudioFiles {
		if userAudio.StoragePath == "" {
			continue
		}
		mUserAudio[userAudio.StoragePath] = i
	}

	var m sync.Mutex
	var errList errutil.List

	pondLimit := b.cfg.UserMaxPhotos
	p := pond.New(pondLimit, pondLimit)
	fnEnrichPublicURL := func(key string) func() {
		return func() {
			publicURL, err := b.aiUploader.PublicURL(ctx, key)
			if err != nil {
				errList.AddErr(fmt.Errorf("getting public url for key %s: %w", key, err))
				return
			}
			m.Lock()
			defer m.Unlock()
			u.AudioFiles[mUserAudio[key]].URL = publicURL
		}
	}
	for key := range mUserAudio {
		p.Submit(fnEnrichPublicURL(key))
	}
	p.StopAndWait()
	return errList.Error()
}

func (b *Business) DeleteUser(ctx context.Context, userID string) error {
	if err := b.deleter.DeleteUserData(ctx, b.dbBE(), b.dbAI(), b.dbSupa(), userID); err != nil {
		return fmt.Errorf("deleting user data: %w", err)
	}
	return nil
}

// UserDetails retrieves detailed information about a user, including optional filters.
// This is more formatted on the registration portion.
func (b *Business) UserDetails(ctx context.Context, userID string, filters *QueryFilterUser) (*User, error) {
	filters.ID = null.StringFrom(userID)
	filters.EnrichCallStates = true
	user, err := b.storer.User(ctx, b.dbBE(), b.dbAI(), filters)
	if err != nil {
		return user, fmt.Errorf("user details: %w", err)
	}

	// Enrich audio files with public URLs if requested
	if err = b.enrichAudioFiles(ctx, user); err != nil {
		return nil, fmt.Errorf("enrich audio files: %w", err)
	}

	return user, nil
}
