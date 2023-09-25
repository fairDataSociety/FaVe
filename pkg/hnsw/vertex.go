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
	"sync"
)

type vertex struct {
	sync.Mutex `json:"-"`

	Id          uint64     `json:"_id"`
	Level       int        `json:"level"`
	Connections [][]uint64 `json:"connections"`
	Maintenance bool       `json:"maintenance"`
	Committed   bool       `json:"-"`
}

func (v *vertex) markAsMaintenance() {
	v.Lock()
	v.Maintenance = true
	v.Unlock()
}

func (v *vertex) unmarkAsMaintenance() {
	v.Lock()
	v.Maintenance = false
	v.Unlock()
}

func (v *vertex) isUnderMaintenance() bool {
	v.Lock()
	m := v.Maintenance
	v.Unlock()
	return m
}

func (v *vertex) connectionsAtLevelNoLock(level int) []uint64 {
	return v.Connections[level]
}

func (v *vertex) upgradeToLevelNoLock(level int) {
	newConnections := make([][]uint64, level+1)
	copy(newConnections, v.Connections)
	v.Level = level
	v.Connections = newConnections
}

func (v *vertex) setConnectionsAtLevel(level int, connections []uint64) {
	v.Lock()
	defer v.Unlock()
	// before we simply copy the Connections let's check how much smaller the new
	// list is. If it's considerably smaller, we might want to downsize the
	// current allocation
	oldCap := cap(v.Connections[level])
	newLen := len(connections)
	ratio := float64(1) - float64(newLen)/float64(oldCap)
	if ratio > 0.33 || oldCap < newLen {
		// the replaced slice is over 33% smaller than the last one, this makes it
		// worth to replace it entirely. This has a small performance cost, it
		// means that if we need append to this node again, we need to re-allocate,
		// but we gain at least a 33% memory saving on this particular node right
		// away.
		v.Connections[level] = connections
		return
	}

	v.Connections[level] = v.Connections[level][:newLen]
	copy(v.Connections[level], connections)
}

func (v *vertex) appendConnectionAtLevelNoLock(level int, connection uint64, maxConns int) {
	if len(v.Connections[level]) == cap(v.Connections[level]) {
		// if the len is the capacity, this  means a new array needs to be
		// allocated to back this slice. The go runtime would do this
		// automatically, if we just use 'append', but it wouldn't do it very
		// efficiently. It would always double the existing capacity. Since we have
		// a hard limit (maxConns), we don't ever want to grow it beyond that. In
		// the worst case, the current len & cap could be maxConns-1, which would
		// mean we would double to 2*(maxConns-1) which would be way too large.
		//
		// Instead let's grow in 4 steps: 25%, 50%, 75% or full capacity
		ratio := float64(len(v.Connections[level])) / float64(maxConns)

		target := 0
		switch {
		case ratio < 0.25:
			target = int(float64(0.25) * float64(maxConns))
		case ratio < 0.50:
			target = int(float64(0.50) * float64(maxConns))
		case ratio < 0.75:
			target = int(float64(0.75) * float64(maxConns))
		default:
			target = maxConns
		}

		// handle rounding errors on maxConns not cleanly divisble by 4
		if target < len(v.Connections[level])+1 {
			target = len(v.Connections[level]) + 1
		}

		newConns := make([]uint64, len(v.Connections[level]), target)
		copy(newConns, v.Connections[level])
		v.Connections[level] = newConns
	}
	v.Connections[level] = append(v.Connections[level], connection)
}

func (v *vertex) resetConnectionsAtLevelNoLock(level int) {
	v.Connections[level] = v.Connections[level][:0]
}
