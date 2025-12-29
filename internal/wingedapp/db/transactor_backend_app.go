package db

type TransactorBackendApp struct {
	*Transactor
}

// NewTransactorBackendApp creates a new TransactorBackendApp instance,
// embedding the provided Transactor.
// This is just so we can differentiate our struct types in DI containers.
func NewTransactorBackendApp(t *Transactor) *TransactorBackendApp {
	if t == nil {
		panic("TransactorBackendApp: nil Transactor provided")
	}
	return &TransactorBackendApp{Transactor: t}
}
