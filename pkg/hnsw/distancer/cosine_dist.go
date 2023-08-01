//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2023 Weaviate B.V. All rights reserved.
//
//  CONTACT: hello@weaviate.io
//

package distancer

import (
	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/adapters/repos/db/vector/hnsw/distancer"
	"math"
)

type CosineDistance struct {
	a []float32
}

func (d *CosineDistance) Distance(b []float32) (float32, bool, error) {
	if len(d.a) != len(b) {
		return 0, false, errors.Errorf("vector lengths don't match: %d vs %d",
			len(d.a), len(b))
	}
	d.a = normalizeVector(d.a)
	b = normalizeVector(b)
	dist := 1 - dotProductImplementation(d.a, b)
	return dist, true, nil
}

func normalizeVector(vector []float32) []float32 {
	norm := float32(0)

	// Calculate the Euclidean norm
	for _, val := range vector {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))

	// Normalize the vector
	for i := range vector {
		vector[i] /= norm
	}

	return vector
}

type CosineDistanceProvider struct{}

func NewCosineDistanceProvider() CosineDistanceProvider {
	return CosineDistanceProvider{}
}

func (d CosineDistanceProvider) SingleDist(a, b []float32) (float32, bool, error) {
	if len(a) != len(b) {
		return 0, false, errors.Errorf("vector lengths don't match: %d vs %d",
			len(a), len(b))
	}
	a = normalizeVector(a)
	b = normalizeVector(b)
	prod := 1 - dotProductImplementation(a, b)

	return prod, true, nil
}

func (d CosineDistanceProvider) Type() string {
	return "cosine-dot"
}

func (d CosineDistanceProvider) New(a []float32) distancer.Distancer {
	return &CosineDistance{a: a}
}

func (d CosineDistanceProvider) Step(x, y []float32) float32 {
	var sum float32
	for i := range x {
		sum += x[i] * y[i]
	}

	return sum
}

func (d CosineDistanceProvider) Wrap(x float32) float32 {
	return 1 - x
}
