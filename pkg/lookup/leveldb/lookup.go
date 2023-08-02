package leveldb

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/pkg/errors"
	"math"
	"strings"
	"unicode"

	lkupr "github.com/fairDataSociety/FaVe/pkg/lookup"
	"github.com/syndtr/goleveldb/leveldb"
)

// Lookup vector embedding from kv store
type Lookup struct {
	db        *leveldb.DB
	stopWords map[string]int
}

func New(path string, stopWords []string) (lkupr.Lookuper, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	lookupMap := map[string]int{}
	for _, word := range stopWords {
		lookupMap[word] = 1
	}
	return &Lookup{db: db, stopWords: lookupMap}, nil
}

func split(corpus string) []string {
	return strings.FieldsFunc(corpus, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
}

func (l *Lookup) Corpi(corpi []string) (*lkupr.Vector, error) {
	var (
		corpusVectors []lkupr.Vector
		err           error
	)
	for i, corpus := range corpi {
		parts := split(corpus)
		if len(parts) == 0 {
			continue
		}

		corpusVectors, err = l.vectors(parts)
		if err != nil {
			return nil, fmt.Errorf("at corpus %d: %v", i, err)
		}
	}
	if len(corpusVectors) == 0 {
		return nil, fmt.Errorf("no vectors found for corpus")
	}

	vector, err := computeCentroid(corpusVectors)
	if err != nil {
		return nil, err
	}

	return vector, nil
}

func computeCentroid(vectors []lkupr.Vector) (*lkupr.Vector, error) {
	var occr = make([]uint64, len(vectors))

	for i := 0; i < len(vectors); i++ {
		occr[i] = uint64(102)
	}
	weights, err := occurrencesToWeight(occr)
	if err != nil {
		return nil, err
	}

	return ComputeWeightedCentroid(vectors, weights)
}

func occurrencesToWeight(occs []uint64) ([]float32, error) {
	max, min := maxMin(occs)

	weigher := makeLogWeigher(min, max)
	weights := make([]float32, len(occs))
	for i, occ := range occs {
		res := weigher(occ)
		weights[i] = res
	}

	return weights, nil
}

func maxMin(input []uint64) (max uint64, min uint64) {
	if len(input) >= 1 {
		min = input[0]
	}

	for _, curr := range input {
		if curr < min {
			min = curr
		} else if curr > max {
			max = curr
		}
	}

	return
}

func makeLogWeigher(min, max uint64) func(uint64) float32 {
	return func(occ uint64) float32 {
		// Note the 1.05 that's 1 + minimal weight of 0.05. This way, the most common
		// word is not removed entirely, but still weighted somewhat
		return float32(2 * (1.05 - (math.Log(float64(occ)) / math.Log(float64(max)))))
	}
}

func ComputeWeightedCentroid(vectors []lkupr.Vector, weights []float32) (*lkupr.Vector, error) {

	if len(vectors) == 0 {
		return nil, fmt.Errorf("can not compute centroid of empty slice")
	} else if len(vectors) != len(weights) {
		return nil, fmt.Errorf("can not compute weighted centroid if len(vectors) != len(weights)")
	} else if len(vectors) == 1 {
		return &vectors[0], nil
	} else {
		vectorLen := vectors[0].Len()

		var newVector = make([]float32, vectorLen)
		var weightSum float32 = 0.0

		for vectorI, v := range vectors {
			if v.Len() != vectorLen {
				return nil, fmt.Errorf("vectors have different lengths", v.Len(), vectorLen)
			}

			weightSum += weights[vectorI]
			vector := v.ToArray()
			for i := 0; i < vectorLen; i++ {
				newVector[i] += vector[i] * weights[vectorI]
			}
		}

		for i := 0; i < vectorLen; i++ {
			newVector[i] /= weightSum
		}

		result := lkupr.NewVector(newVector)
		return &result, nil
	}
}

func (l *Lookup) getVectorForWord(word string) (*lkupr.Vector, error) {
	if _, ok := l.stopWords[strings.ToLower(word)]; ok {
		return nil, nil
	}
	var value []byte
	value, err := l.db.Get([]byte(word), nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		value, err = l.db.Get([]byte(strings.ToLower(word)), nil)
		if err != nil {
			return nil, nil
		}
	}

	vector := make([]float32, 300)
	err = gob.NewDecoder(bytes.NewBuffer(value)).Decode(&vector)
	if err != nil {
		return nil, err
	}
	v := lkupr.NewVector(vector)

	return &v, nil
}

func (l *Lookup) vectors(words []string) ([]lkupr.Vector, error) {
	vectors := make([]lkupr.Vector, len(words))
	for wordPos := 0; wordPos < len(words); wordPos++ {
		vector, err := l.getVectorForWord(words[wordPos])
		if err != nil {
			return nil, err
		}
		if vector != nil {
			// this compound word exists, use its vector and occurrence
			vectors[wordPos] = *vector
		}
	}

	finalVectors := []lkupr.Vector{}
	for _, v := range vectors {
		if v.Len() > 0 {
			finalVectors = append(finalVectors, v)
		}
	}
	return finalVectors, nil
}
