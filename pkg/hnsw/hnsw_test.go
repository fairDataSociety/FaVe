package hnsw

import (
	"context"
	"fmt"
	"github.com/fairdatasociety/fairOS-dfs/pkg/account"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/feed"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/weaviate/weaviate/adapters/repos/db/vector/hnsw/distancer"
	"log"
	"os"
	"testing"
)

// roughly grouped into three clusters of three
var testVectors = [][]float32{
	{0.1, 0.9},
	{0.15, 0.8},
	{0.13, 0.65},

	{0.6, 0.1},
	{0.63, 0.2},
	{0.65, 0.08},

	{0.8, 0.8},
	{0.9, 0.75},
	{0.8, 0.7},
}

func testVectorForID(ctx context.Context, id uint64) ([]float32, error) {
	return testVectors[int(id)], nil
}

func TestHnsw(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	logger := logging.New(os.Stdout, 1)
	acc := account.New(logger)
	ai := acc.GetUserAccountInfo()
	_, _, err := acc.CreateUserAccount("")
	if err != nil {
		t.Fatal(err)
	}
	fd := feed.New(acc.GetUserAccountInfo(), mockClient, logger)
	user := acc.GetAddress(account.UserAccountIndex)
	kvStore := collection.NewKeyValueStore("pod1", fd, ai, user, mockClient, logger)
	podPassword := "podPassword"
	tableName := "testTable"
	err = kvStore.CreateKVTable(tableName, podPassword, collection.StringIndex)
	if err != nil {
		t.Fatal(err)
	}

	err = kvStore.OpenKVTable(tableName, podPassword)
	if err != nil {
		t.Fatal(err)
	}

	makeCL := MakeNoopCommitLogger
	index, err := New(Config{
		RootPath:              "doesnt-matter-as-committlogger-is-mocked-out",
		ID:                    "unittest",
		MakeCommitLoggerThunk: makeCL,
		DistanceProvider:      distancer.NewCosineDistanceProvider(),
		VectorForIDThunk:      testVectorForID,
		ClassName:             tableName,
	}, UserConfig{
		MaxConnections: 30,
		EFConstruction: 60,
	}, kvStore)

	if err != nil {
		log.Fatal(err)
	}

	for i, vec := range testVectors {
		err := index.Add(uint64(i), vec)
		if err != nil {
			log.Fatal(err)
		}
	}
	for i, _ := range testVectors {
		s, v, err := kvStore.KVGet(tableName, fmt.Sprintf("%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(s, string(v))
	}

	position := 0
	ids, _, err := index.KnnSearchByVectorMaxDist(testVectors[position], 0.2, 36, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ids)
}
