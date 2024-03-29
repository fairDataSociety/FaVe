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

package hnsw

import (
	"fmt"
	"github.com/fairDataSociety/FaVe/pkg/hnsw/distancer"
	"github.com/fairDataSociety/FaVe/pkg/hnsw/priorityqueue"
	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/adapters/repos/db/helpers"
)

func (h *hnsw) KnnSearchByVectorMaxDist(searchVec []float32, dist float32,
	ef int, allowList helpers.AllowList,
) ([]uint64, []float32, error) {
	if h.distancerProvider.Type() == "cosine-dot" {
		// cosine-dot requires normalized vectors, as the dot product and cosine
		// similarity are only identical if the vector is normalized
		searchVec = distancer.Normalize(searchVec)
	}
	entryPointID := h.entryPointID
	entryPointDistance, ok, err := h.distBetweenNodeAndVec(entryPointID, searchVec)
	if err != nil {
		return nil, nil, errors.Wrap(err, "knn search: distance between entrypoint and query node")
	}
	if !ok {
		return nil, nil, fmt.Errorf("entrypoint was deleted in the object store, " +
			"it has been flagged for cleanup and should be fixed in the next cleanup cycle")
	}

	// stop at layer 1, not 0!
	for level := h.currentMaximumLayer; level >= 1; level-- {
		eps := priorityqueue.NewMin(1)
		eps.Insert(entryPointID, entryPointDistance)
		// ignore allowList on layers > 0
		res, err := h.searchLayerByVector(searchVec, eps, 1, level, nil)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "knn search: search layer at Level %d", level)
		}
		if res.Len() > 0 {
			best := res.Pop()
			entryPointID = best.ID
			entryPointDistance = best.Dist
		}

		h.pools.pqResults.Put(res)
	}
	eps := priorityqueue.NewMin(1)
	eps.Insert(entryPointID, entryPointDistance)
	res, err := h.searchLayerByVector(searchVec, eps, ef, 0, allowList)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "knn search: search layer at Level %d", 0)
	}

	all := make([]priorityqueue.Item, res.Len())
	i := res.Len() - 1
	for res.Len() > 0 {
		all[i] = res.Pop()
		i--
	}
	out := make([]uint64, len(all))
	dists := make([]float32, len(all))
	i = 0
	for _, elem := range all {
		if elem.Dist < 0 {
			continue
		}
		if elem.Dist > dist {
			break
		}
		out[i] = elem.ID
		dists[i] = elem.Dist
		i++
	}

	h.pools.pqResults.Put(res)
	return out[:i], dists, nil
}
