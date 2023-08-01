package contextionary

import (
	"context"
	"fmt"

	lkupr "github.com/fairDataSociety/FaVe/pkg/lookup"
	"github.com/weaviate/weaviate/modules/text2vec-contextionary/client"
)

type lookup struct {
	client *client.Client
}

func New(url string) (lkupr.Lookuper, error) {
	c, err := client.NewClient(url, nil)
	if err != nil {
		return nil, err
	}
	return &lookup{client: c}, nil
}

func (l *lookup) Corpi(corpi []string) (*lkupr.Vector, error) {
	fmt.Println("contextionary corpi: ", corpi)
	vec, _, err := l.client.VectorForCorpi(context.Background(), corpi, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("contextionary vec: ", vec)
	vector := lkupr.NewVector(vec)
	return &vector, nil
}
