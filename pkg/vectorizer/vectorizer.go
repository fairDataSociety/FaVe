package vectorizer

type Vectorizer interface {
	Corpi(corpi []string) (*Vector, error)
}
