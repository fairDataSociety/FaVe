package leveldb

import (
	"fmt"
	"testing"

	dfsLookup "github.com/fairDataSociety/FaVe/pkg/lookup/dfs"
)

func TestLookup(t *testing.T) {
	lukpr, err := New("../../../tools/dev/gloveToLeveldb/embds", dfsLookup.Stopwords["en"])
	if err != nil {
		t.Fatal(err)
	}
	v, err := lukpr.Corpi([]string{"What"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(v)
}
