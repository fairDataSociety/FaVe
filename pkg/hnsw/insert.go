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
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/pkg/errors"
	"github.com/weaviate/weaviate/adapters/repos/db/helpers"
	"github.com/weaviate/weaviate/adapters/repos/db/vector/hnsw/distancer"
)

const (
	entrypointKey = "entrypoint"
	countKey      = "count"
)

type entrypoint struct {
	ID uint64 `json:"id"`
}

type count struct {
	Count uint64 `json:"count"`
}

func (h *hnsw) ValidateBeforeInsert(vector []float32) error {
	if h.isEmpty() {
		return nil
	}
	// check if vector length is the same as existing nodes
	existingNodeVector, err := h.cache.get(context.Background(), h.entryPointID)
	if err != nil {
		return err
	}

	if len(existingNodeVector) != len(vector) {
		return fmt.Errorf("new node has a vector with length %v. "+
			"Existing nodes have vectors with length %v", len(vector), len(existingNodeVector))
	}

	return nil
}

func (h *hnsw) Add(id uint64, vector []float32) error {
	before := time.Now()
	if len(vector) == 0 {
		return errors.Errorf("insert called with nil-vector")
	}

	h.metrics.InsertVector()
	defer h.insertMetrics.total(before)

	node := &vertex{
		Id: id,
	}

	if h.distancerProvider.Type() == "cosine-dot" {
		// cosine-dot requires normalized vectors, as the dot product and cosine
		// similarity are only identical if the vector is normalized
		vector = distancer.Normalize(vector)
	}

	h.compressActionLock.RLock()
	defer h.compressActionLock.RUnlock()
	return h.insert(node, vector)
}

func (h *hnsw) insertInitialElement(node *vertex, nodeVec []float32) error {
	h.Lock()
	defer h.Unlock()
	if err := h.commitLog.SetEntryPointWithMaxLayer(node.Id, 0); err != nil {
		return err
	}

	h.entryPointID = node.Id
	h.currentMaximumLayer = 0
	node.Connections = [][]uint64{
		make([]uint64, 0, h.maximumConnectionsLayerZero),
	}
	node.Vector = nodeVec
	node.Level = 0
	if err := h.commitLog.AddNode(node); err != nil {
		return err
	}

	err := h.growIndexToAccomodateNode(node.Id, h.logger)
	if err != nil {
		return errors.Wrapf(err, "grow HNSW index to accommodate node %d", node.Id)
	}

	h.indexCache.Add(node.Id, node)

	if h.compressed.Load() {
		compressed := h.pq.Encode(nodeVec)
		h.storeCompressedVector(node.Id, compressed)
		h.compressedVectorsCache.preload(node.Id, compressed)
	} else {
		h.cache.preload(node.Id, nodeVec)
	}

	// go h.insertHook(node.ID, 0, node.Connections)
	return nil
}

func (h *hnsw) Flush(docCount uint64) error {
	keys := h.indexCache.Keys()
	for _, key := range keys {
		iNode, ok := h.indexCache.Get(key)
		if !ok {
			continue
		}
		node := iNode.(*vertex)
		if node.Committed {
			continue
		}
		nodeBytes, err := json.Marshal(node)
		if err != nil {
			return errors.Wrapf(err, "marshal node %d", node.Id)
		}
		err = h.nodes.KVPut(h.className, fmt.Sprintf("%d", node.Id), nodeBytes)
		if err != nil {
			return errors.Wrapf(err, "put node %d", node.Id)
		}
		node.Committed = true
		h.indexCache.Add(node.Id, node)
	}
	ep := &entrypoint{
		ID: h.entryPointID,
	}
	epBytes, err := json.Marshal(ep)
	if err != nil {
		return errors.Wrapf(err, "marshal entrypoint %d", h.entryPointID)
	}
	err = h.nodes.KVPut(h.className, entrypointKey, epBytes)
	if err != nil {
		return errors.Wrapf(err, "put entrypoint %d", h.entryPointID)
	}

	c := &count{
		Count: docCount,
	}
	cBytes, err := json.Marshal(c)
	if err != nil {
		return errors.Wrapf(err, "marshal count %d", h.entryPointID)
	}
	err = h.nodes.KVPut(h.className, countKey, cBytes)
	if err != nil {
		return errors.Wrapf(err, "put count %d", h.entryPointID)
	}
	return h.commitLog.Flush()
}

func (h *hnsw) GetDocCount() (uint64, error) {
	_, cBytes, err := h.nodes.KVGet(h.className, countKey)
	if err != nil {
		return 0, errors.Wrapf(err, "put count %d", h.entryPointID)
	}

	c := &count{}
	err = json.Unmarshal(cBytes, c)
	if err != nil {
		return 0, err
	}

	return c.Count, nil
}

func (h *hnsw) LoadEntrypoint() error {
	h.RLock()
	defer h.RUnlock()

	_, v, err := h.nodes.KVGet(h.className, entrypointKey)
	if err != nil {
		h.entryPointID = 0
		return nil
	}
	ep := &entrypoint{}
	err = json.Unmarshal(v, ep)
	if err != nil {
		h.entryPointID = 0
		return nil
	}

	h.entryPointID = ep.ID
	return nil
}

func (h *hnsw) insert(node *vertex, nodeVec []float32) error {
	h.deleteVsInsertLock.RLock()
	defer h.deleteVsInsertLock.RUnlock()

	before := time.Now()

	wasFirst := false
	var firstInsertError error
	h.initialInsertOnce.Do(func() {
		if h.isEmpty() {
			wasFirst = true
			firstInsertError = h.insertInitialElement(node, nodeVec)
		}
	})
	if wasFirst {
		return firstInsertError
	}

	node.markAsMaintenance()

	h.RLock()
	// initially use the "global" entrypoint which is guaranteed to be on the
	// currently highest layer
	entryPointID := h.entryPointID
	// initially use the Level of the entrypoint which is the highest Level of
	// the h-graph in the first iteration
	currentMaximumLayer := h.currentMaximumLayer
	h.RUnlock()

	targetLevel := int(math.Floor(-math.Log(h.randFunc()) * h.levelNormalizer))
	// before = time.Now()
	// m.addBuildingItemLocking(before)
	node.Level = targetLevel
	node.Connections = make([][]uint64, targetLevel+1)
	node.Vector = nodeVec
	for i := targetLevel; i >= 0; i-- {
		capacity := h.maximumConnections
		if i == 0 {
			capacity = h.maximumConnectionsLayerZero
		}

		node.Connections[i] = make([]uint64, 0, capacity)
	}

	if err := h.commitLog.AddNode(node); err != nil {
		return err
	}

	nodeId := node.Id

	// before = time.Now()
	h.Lock()
	// m.addBuildingLocking(before)
	err := h.growIndexToAccomodateNode(node.Id, h.logger)
	if err != nil {
		h.Unlock()
		return errors.Wrapf(err, "grow HNSW index to accommodate node %d", node.Id)
	}
	h.Unlock()

	// // make sure this new vec is immediately present in the cache, so we don't
	// // have to read it from disk again
	if h.compressed.Load() {
		compressed := h.pq.Encode(nodeVec)
		h.storeCompressedVector(node.Id, compressed)
		h.compressedVectorsCache.preload(node.Id, compressed)
	} else {
		h.cache.preload(node.Id, nodeVec)
	}

	h.indexCache.Add(node.Id, node)

	h.insertMetrics.prepareAndInsertNode(before)
	before = time.Now()
	entryPointID, err = h.findBestEntrypointForNode(currentMaximumLayer, targetLevel,
		entryPointID, nodeVec)
	if err != nil {
		return errors.Wrap(err, "find best entrypoint")
	}

	h.insertMetrics.findEntrypoint(before)
	before = time.Now()

	if err := h.findAndConnectNeighbors(node, entryPointID, nodeVec,
		targetLevel, currentMaximumLayer, helpers.NewAllowList()); err != nil {
		return errors.Wrap(err, "find and connect neighbors")
	}

	h.insertMetrics.findAndConnectTotal(before)
	before = time.Now()
	defer h.insertMetrics.updateGlobalEntrypoint(before)

	// go h.insertHook(nodeId, targetLevel, neighborsAtLevel)
	node.unmarkAsMaintenance()

	h.Lock()
	if targetLevel > h.currentMaximumLayer {
		// before = time.Now()
		// m.addBuildingLocking(before)
		if err := h.commitLog.SetEntryPointWithMaxLayer(nodeId, targetLevel); err != nil {
			h.Unlock()
			return err
		}

		h.entryPointID = nodeId
		h.currentMaximumLayer = targetLevel
	}
	h.Unlock()

	return nil
}
