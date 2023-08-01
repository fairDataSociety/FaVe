package lookup

type Lookuper interface {
	Corpi(corpi []string) (*Vector, error)
}
