package dfs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"math"
	"strings"
	"sync"
	"unicode"

	lkupr "github.com/fairDataSociety/FaVe/pkg/vectorizer"
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
	cache     *lru.Cache
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

	lkup := &Lookup{
		store:     store,
		storeName: storeName,
		stopWords: lookupMap,
	}
	cache, err := lru.New(10000)
	if err == nil {
		lkup.cache = cache
	}
	return lkup, nil
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

func (lookup *Lookup) Corpi(corpi []string) (*lkupr.Vector, error) {
	var (
		corpusVectors []lkupr.Vector
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

func (lookup *Lookup) getVectorForWord(word string) (*lkupr.Vector, error) {
	if _, ok := lookup.stopWords[word]; ok {
		return nil, nil
	}
	// check if available in cache
	if vector, ok := lookup.cache.Get(word); ok {
		v := lkupr.NewVector(vector.([]float32))
		return &v, nil
	}

	_, b, err := lookup.store.KVGet(lookup.storeName, word)
	if err != nil {
		// TODO add logger
		return nil, nil
	}

	vector := make([]float32, 300)
	err = gob.NewDecoder(bytes.NewBuffer(b)).Decode(&vector)
	if err != nil {
		return nil, err
	}
	v := lkupr.NewVector(vector)

	lookup.cache.Add(word, vector)
	return &v, nil
}

func (lookup *Lookup) vectors(words []string) ([]lkupr.Vector, error) {
	vectors := make([]lkupr.Vector, len(words))
	wg := sync.WaitGroup{}
	for wordPos := 0; wordPos < len(words); wordPos++ {
		wg.Add(1)
		go func(wordPos int) {
			defer wg.Done()
			vector, err := lookup.getVectorForWord(words[wordPos])
			if err != nil {
				return
			}
			if vector != nil {
				// this compound word exists, use its vector and occurrence
				vectors[wordPos] = *vector
			}
		}(wordPos)
	}
	wg.Wait()

	finalVectors := []lkupr.Vector{}
	for _, v := range vectors {
		if v.Len() > 0 {
			finalVectors = append(finalVectors, v)
		}
	}
	return finalVectors, nil
}

func nextWords(words []string, startPos int, additional int) []string {
	endPos := startPos + 1 + additional
	return words[startPos:endPos]
}

func compound(words ...string) string {
	return strings.Join(words, "_")
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

		result := lkupr.NewVector(newVector)
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
