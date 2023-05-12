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

// Dump to stdout for debugging purposes
func (h *hnsw) Dump(labels ...string) {

}

// DumpJSON to stdout for debugging purposes
func (h *hnsw) DumpJSON(labels ...string) {

}

type JSONDump struct {
	Labels              []string            `json:"labels"`
	ID                  string              `json:"ID"`
	Entrypoint          uint64              `json:"entrypoint"`
	CurrentMaximumLayer int                 `json:"currentMaximumLayer"`
	Tombstones          map[uint64]struct{} `json:"tombstones"`
	Nodes               []JSONDumpNode      `json:"nodes"`
}

type JSONDumpNode struct {
	ID          uint64     `json:"ID"`
	Level       int        `json:"Level"`
	Connections [][]uint64 `json:"Connections"`
}

type JSONDumpMap struct {
	Labels              []string            `json:"labels"`
	ID                  string              `json:"ID"`
	Entrypoint          uint64              `json:"entrypoint"`
	CurrentMaximumLayer int                 `json:"currentMaximumLayer"`
	Tombstones          map[uint64]struct{} `json:"tombstones"`
	Nodes               []JSONDumpNodeMap   `json:"nodes"`
}

type JSONDumpNodeMap struct {
	ID          uint64           `json:"ID"`
	Level       int              `json:"Level"`
	Connections map[int][]uint64 `json:"Connections"`
}

func NewFromJSONDump(dumpBytes []byte, vecForID VectorForID) (*hnsw, error) {
	return nil, nil
}

func NewFromJSONDumpMap(dumpBytes []byte, vecForID VectorForID) (*hnsw, error) {
	return nil, nil

}

// was added as part of
// https://github.com/weaviate/weaviate/issues/1868 for debugging. It
// is not currently in use anywhere as it is somewhat costly, it would lock the
// entire graph and iterate over every node which would lead to disruptions in
// production. However, keeping this method around may be valuable for future
// investigations where the amount of links may be a problem.
func (h *hnsw) ValidateLinkIntegrity() {

}
