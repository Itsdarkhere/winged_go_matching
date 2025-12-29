package pg

import (
	"context"
	"fmt"
	"wingedapp/pgtester/internal/wingedapp/business/domain/setting"
	"wingedapp/pgtester/internal/wingedapp/db/pgmodel"
	"wingedapp/pgtester/internal/wingedapp/db/repo"

	"github.com/aarondl/sqlboiler/v4/boil"
)

func toRepoQueryFilterAnonymizedContact(f *setting.QueryFilterAnonymizedContact) *repo.QueryFilterAnonymizedContact {
	return &repo.QueryFilterAnonymizedContact{
		ContactHashes: f.ContactHashes,
		Pagination:    f.Pagination,
		OrderBy:       f.OrderBy,
		Sort:          f.Sort,
	}
}

// toBizAnonymizedContact converts a repo.AnonymizedContact to a setting.AnonymizedContact,
// with relevant translation logic of its `status` field.
func toBizAnonymizedContact(repoAC *repo.AnonymizedContact) *setting.AnonymizedContact {
	bizAC := &setting.AnonymizedContact{
		ContactHash:     repoAC.ContactHash,
		FriendsOnWinged: repoAC.Count,
	}

	switch true {
	case repoAC.IsMember:
		bizAC.Status = setting.AnonymizedContactStatusMember
		break
	case repoAC.IsInvited:
		bizAC.Status = setting.AnonymizedContactStatusInvited
		break
	default:
		bizAC.Status = setting.AnonymizedContactStatusNonMember
	}

	return bizAC
}

func toBizAnonymizedContacts(acs []repo.AnonymizedContact) []setting.AnonymizedContact {
	bizACs := make([]setting.AnonymizedContact, 0, len(acs))
	for _, contactHash := range acs {
		bizACs = append(bizACs, *toBizAnonymizedContact(&contactHash))
	}
	return bizACs
}

func toRepoInsertAnonymizedContact(b *setting.UpsertAnonymizedContact) *repo.InsertAnonymizedContact {
	return &repo.InsertAnonymizedContact{
		OwnerHash:   b.OwnerHash,
		ContactHash: b.ContactHash,
	}
}

func toBizInsertAnonymizedContact(p *pgmodel.AnonymizedContact) *setting.UpsertAnonymizedContact {
	return &setting.UpsertAnonymizedContact{
		OwnerHash:   p.OwnerHash,
		ContactHash: p.ContactHash,
	}
}

func (s *Store) UpsertAnonymizedContacts(ctx context.Context,
	exec boil.ContextExecutor,
	inserter []setting.UpsertAnonymizedContact,
) ([]setting.UpsertAnonymizedContact, error) {
	var inserts []setting.UpsertAnonymizedContact
	for _, ac := range inserter {
		inserted, err := s.repoBackendApp.UpsertAnonymizedContact(ctx, exec,
			toRepoInsertAnonymizedContact(&ac))
		if err != nil {
			return nil, fmt.Errorf("anonymized contact store: %w", err)
		}
		inserts = append(inserts, *toBizInsertAnonymizedContact(inserted))
	}
	return inserts, nil
}

func (s *Store) AnonymizedContacts(ctx context.Context,
	exec boil.ContextExecutor,
	f *setting.QueryFilterAnonymizedContact,
) ([]setting.AnonymizedContact, error) {
	fRepo := toRepoQueryFilterAnonymizedContact(f)
	bizACs, err := s.repoBackendApp.AnonymizedContactsByCount(ctx, exec, fRepo)
	if err != nil {
		return nil, fmt.Errorf("repo sysparams: %w", err)
	}
	return toBizAnonymizedContacts(bizACs), nil
}
