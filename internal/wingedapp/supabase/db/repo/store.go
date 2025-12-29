package repo

// Store implements CRUD operations for supabase auth database.
type Store struct{}

func NewStore() *Store {
	return &Store{}
}
