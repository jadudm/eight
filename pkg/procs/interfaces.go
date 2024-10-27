package procs

type Storage interface {
	Store(string, map[string]string) error
}
