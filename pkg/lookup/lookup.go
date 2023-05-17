package lookup

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/utils"
)

const (
	maxCompoundWordLength = 1

	GloveStore = "glove"
)

// Lookup vector embedding from kv store
type Lookup struct {
	store     dfs.KVGetter
	storeName string
	stopWords map[string]int
}

// New takes *dfs.API, a pod sharingReference and a kv store name
func New(api *dfs.API, sharingRefString, storeName string, stopWords []string) (*Lookup, error) {

	// parse the sharing reference
	ref, err := utils.ParseHexReference(sharingRefString)
	if err != nil {
		return nil, err
	}

	// get the share info of the pod
	shareInfo, err := api.PublicPodReceiveInfo(ref)
	if err != nil {
		return nil, err
	}

	lookupMap := map[string]int{}
	for _, word := range stopWords {
		lookupMap[word] = 1
	}

	store := api.PublicPodKVGetter(shareInfo)
	if err := store.OpenKVTable(storeName, shareInfo.Password); err != nil {
		return nil, err
	}
	return &Lookup{
		store:     store,
		storeName: storeName,
		stopWords: lookupMap,
	}, nil
}

func (lookup *Lookup) Get(key string) []float32 {
	// TODO
	return nil
}

func split(corpus string) []string {
	return strings.FieldsFunc(corpus, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
}

func (lookup *Lookup) Corpi(corpi []string) (*Vector, error) {
	var (
		corpusVectors []Vector
		err           error
	)
	for i, corpus := range corpi {
		parts := split(corpus)
		if len(parts) == 0 {
			continue
		}

		corpusVectors, err = lookup.vectors(parts)
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

func (lookup *Lookup) getVectorForWord(word string) (*Vector, error) {
	if _, ok := lookup.stopWords[word]; ok {
		return nil, nil
	}
	_, b, err := lookup.store.KVGet(lookup.storeName, word)
	if err != nil {
		// TODO add logger
		fmt.Println("lookup failed", err)
		return nil, nil
	}

	vector := make([]float32, 300)
	err = gob.NewDecoder(bytes.NewBuffer(b)).Decode(&vector)
	if err != nil {
		return nil, err
	}
	v := NewVector(vector)
	return &v, nil
}

func (lookup *Lookup) vectors(words []string) ([]Vector, error) {
	var vectors []Vector

	for wordPos := 0; wordPos < len(words); wordPos++ {
	additionalWordLoop:
		for additionalWords := maxCompoundWordLength - 1; additionalWords >= 0; additionalWords-- {
			if (wordPos + additionalWords) < len(words) {
				// we haven't reached the end of the corpus yet, so this words plus the
				// next n additional words could still form a compound word, we need to
				// check.
				// Note that n goes all the way down to zero, so once we didn't find
				// any compound words, we're checking the individual word.
				// TODO: check if this can be done concurrently
				compound := compound(nextWords(words, wordPos, additionalWords)...)
				vector, err := lookup.getVectorForWord(compound)
				if err != nil {
					return nil, err
				}
				if vector != nil {
					// this compound word exists, use its vector and occurrence
					vectors = append(vectors, *vector)

					// however, now we must make sure to skip the additionalWords
					wordPos += additionalWords
					break additionalWordLoop
				}
			}
		}
	}
	return vectors, nil
}

func nextWords(words []string, startPos int, additional int) []string {
	endPos := startPos + 1 + additional
	return words[startPos:endPos]
}

func compound(words ...string) string {
	return strings.Join(words, "_")
}

func computeCentroid(vectors []Vector) (*Vector, error) {
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

func ComputeWeightedCentroid(vectors []Vector, weights []float32) (*Vector, error) {

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
				return nil, fmt.Errorf("vectors have different lengths")
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

		result := NewVector(newVector)
		return &result, nil
	}
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
