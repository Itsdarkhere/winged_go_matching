package registration

import (
	"github.com/aarondl/null/v8"
)

type QueryFilterUser struct {
	ID         null.String `json:"id"`
	MobileCode null.String `json:"mobile_code"`
	Email      null.String `json:"email"`
	UserTypeID null.String `json:"user_type_id"`

	EnrichPhotos     bool `json:"enrich_photos"`      // add user photos
	EnrichCallStates bool `json:"enrich_call_states"` // add call states
}

func (u *QueryFilterUser) HasFilters() bool {
	if u == nil {
		return true
	}
	return u.ID.Valid || u.MobileCode.Valid || u.Email.Valid
}

type UserInviteCodeQueryFilter struct {
	Code null.String `json:"code"`
}
