package registration

import (
	"context"
	"fmt"
)

func (b *Business) Profile(ctx context.Context, id string) (*Profile, error) {
	filter := ProfileQueryFilter{
		ID: id,
	}

	userProfile, err := b.storer.Profile(ctx, b.dbAI(), &filter)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	return userProfile, nil
}

func (b *Business) Profiles(ctx context.Context) ([]Profile, error) {
	userProfiles, err := b.storer.Profiles(ctx, b.dbAI())
	if err != nil {
		return nil, fmt.Errorf("get user photo: %w", err)
	}

	return userProfiles, nil
}
