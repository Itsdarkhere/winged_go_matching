package registration

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/util/errutil"

	"github.com/aarondl/null/v8"
)

// DeployUserAgent toggles a user's agent if the user has photos and a voice selected.
// This is the final phase of the registration process (important domain knowledge).
// Note: Referral bonus is NOT triggered here - per MVP spec it triggers on first paid action
// (connect or schedule), handled in matching business layer.
func (b *Business) DeployUserAgent(ctx context.Context, user *User) error {
	// validations
	var errList errutil.List
	if len(user.Photos) == 0 {
		errList.AddErr(ErrUserNoPhotos)
	}
	if !user.SelectedIntroID.Valid {
		errList.AddErr(ErrUserNoSelectedIntroID)
	}
	if errList.HasErrors() {
		return errList.Error()
	}

	tx, err := b.transBE.TX()
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}
	defer b.transBE.Rollback(tx)

	// Set user agent deployed to true
	if _, err = b.storer.UpdateUser(ctx, tx, b.dbAI(), &UpdateUser{
		ID:            user.ID,
		AgentDeployed: null.BoolFrom(true),
	}); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
