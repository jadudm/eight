package procs

type Storage interface {
	Store(string, any) error
}
