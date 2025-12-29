package setting

import (
	"context"
	"fmt"
)

// UpsertAnonymizedContacts upserts multiple anon contacts
func (p *Business) UpsertAnonymizedContacts(ctx context.Context, acs []UpsertAnonymizedContact) ([]UpsertAnonymizedContact, error) {
	acs, err := p.storer.UpsertAnonymizedContacts(ctx, p.trans.DB(), acs)
	if err != nil {
		return nil, fmt.Errorf("storer acs: %w", err)
	}
	return acs, nil
}

func (p *Business) AnonymizedContacts(ctx context.Context,
	f *QueryFilterAnonymizedContact,
) (*AnonymizedContacts, error) {
	acs, err := p.storer.AnonymizedContacts(ctx, p.trans.DB(), f)
	if err != nil {
		return nil, fmt.Errorf("storer acs: %w", err)
	}

	return &AnonymizedContacts{
		Data:       acs,
		Pagination: f.Pagination,
	}, nil
}
