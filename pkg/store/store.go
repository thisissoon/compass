package store

// The Store is the mian Store inerface incoperating
// all the store interface into a single interface
type Store interface {
	ServiceStore
	DentryStore
}
