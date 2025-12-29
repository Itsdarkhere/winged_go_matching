package setting

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/lib/userhasher"

	"github.com/aarondl/null/v8"
)

// InviteUser sends an invite to a user based on the provided parameters.
func (p *Business) InviteUser(ctx context.Context, params *InviteUser) error {
	//  enclose all in tx for point of consistency
	tx, err := p.trans.TX()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer p.trans.Rollback(tx)

	// check if invite code already exists
	existInviteCode, err := p.storer.UserInviteCodes(ctx, tx, &QueryFilterUserInviteCode{
		InviteCode: null.StringFrom(params.InviteCode),
	})
	if err != nil {
		return fmt.Errorf("user invite codes: %w", err)
	}
	if len(existInviteCode) > 0 {
		return ErrInviteUserCodeAlreadyExists
	}

	// check if already a "member"
	users, err := p.storer.Users(ctx, tx, &QueryFilterUser{
		Number: null.StringFrom(params.ExtNumber),
	})
	if err != nil {
		return fmt.Errorf("users: %w", err)
	}
	if len(users) > 0 {
		return ErrInviteUserAlreadyAMember
	}

	// check if already "invited"
	userInviteCodes, err := p.storer.UserInviteCodes(ctx, tx, &QueryFilterUserInviteCode{
		Number: null.StringFrom(params.ExtNumber),
	})
	if err != nil {
		return fmt.Errorf("user invite codes: %w", err)
	}
	if len(userInviteCodes) > 0 {
		return ErrInviteUserAlreadyInvited
	}

	// insert into invite codes
	if err = p.storer.InsertUserInviteCode(ctx, tx, &InsertUserInviteCode{
		InviteCode:         params.InviteCode,
		ExtNumber:          params.ExtNumber,
		ExtNumberHash:      userhasher.Sha256(params.ExtNumber),
		ReferrerNumberHash: userhasher.Sha256(params.ReferrerNumber),
	}); err != nil {
		return fmt.Errorf("insert user invite code: %w", err)
	}

	// commit
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
