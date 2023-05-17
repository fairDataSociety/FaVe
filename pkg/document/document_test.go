package document

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/fairDataSociety/FaVe/pkg/lookup"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/sirupsen/logrus"
)

const (
	username = "testuser"
	password = "testpasswordtestpassword"
)

func TestFave(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()
	logger := logging.New(os.Stdout, logrus.InfoLevel)

	users := user.NewUsers(mockClient, ens, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger)
	defer dfsApi.Close()

	_, err := dfsApi.CreateUserV2(username, password, "", "")
	if err != nil {
		t.Fatal(err)
	}

	ref := lookupPrep(t, dfsApi)
	cfg := Config{
		Verbose:     false,
		GlovePodRef: ref,
	}
	client, err := New(cfg, dfsApi)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Login(username, password)
	if err != nil {
		t.Fatal(err)
	}

	err = client.OpenPod("Fave")
	if err != nil {
		t.Fatal(err)
	}

	col := &Collection{
		Name: "Question",
		Indexes: map[string]collection.IndexType{
			"category": collection.StringIndex,
			"question": collection.StringIndex,
			"answer":   collection.StringIndex,
		},
	}

	err = client.CreateCollection(col)
	if err != nil {
		t.Fatal(err)
	}

	documents := []*Document{
		{
			ID: "36ddd591-2dee-4e7e-a3cc-eb86d30a4303",
			Properties: map[string]interface{}{
				"category": "SCIENCE",
				"question": "This is an organ that filters blood",
				"answer":   "Liver",
			},
		},
		{
			ID: "36ddd591-2dee-4e7e-a3cc-eb86d30a4304",
			Properties: map[string]interface{}{
				"category": "ANIMALS",
				"question": "It's the only living mammal in the order Proboseidea",
				"answer":   "Elephant",
			},
		},
		{
			ID: "36ddd591-2dee-4e7e-a3cc-eb86d30a4305",
			Properties: map[string]interface{}{
				"category": "SCIENCE",
				"question": "Changes in the tropospheric layer of this are what gives us weather",
				"answer":   "the atmosphere",
			},
		},
	}

	err = client.AddDocuments(col.Name, documents...)
	if err != nil {
		t.Fatal(err)
	}
	for i, _ := range documents {
		s, v, err := dfsApi.KVGet(client.sessionId, client.pod, col.Name, fmt.Sprintf("%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(s, string(v))
	}

	// Test search
	docs, err := client.GetNearDocuments(col.Name, "atmosphere", .1)
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range docs {
		fmt.Println(string(doc))
	}
}

func lookupPrep(t *testing.T, api *dfs.API) string {
	resp, err := api.LoginUserV2(username, password, "")
	if err != nil {
		t.Fatal(err)
	}
	sessionId := resp.UserInfo.GetSessionId()
	pod := "glove"
	table := lookup.GloveStore

	_, err = api.CreatePod(pod, sessionId)
	if err != nil {
		t.Fatal(err)
	}
	err = api.KVCreate(sessionId, pod, table, collection.BytesIndex)
	if err != nil {
		t.Fatal(err)
	}
	err = api.KVOpen(sessionId, pod, table)
	if err != nil {
		t.Fatal(err)
	}

	batch, err := api.KVBatch(sessionId, pod, table, []string{})
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open("../../tools/dev/en_test-vectors-small.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var vectorLength int = -1
	var nrWords int = 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		nrWords += 1
		parts := strings.Split(scanner.Text(), " ")

		word := parts[0]
		if vectorLength == -1 {
			vectorLength = len(parts) - 1
		}

		if vectorLength != len(parts)-1 {
			log.Print("Line corruption found for the word [" + word + "]. Lenght expected " + strconv.Itoa(vectorLength) + " but found " + strconv.Itoa(len(parts)) + ". Word will be skipped.")
			continue
		}

		// pre-allocate a vector for speed.
		vector := make([]float32, vectorLength)

		for i := 1; i <= vectorLength; i++ {
			float, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				t.Fatal(err)
			}
			vector[i-1] = float32(float)
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(vector); err != nil {
			t.Fatal(err)
		}

		err = batch.Put(word, buf.Bytes(), true, true)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = batch.Write("")
	if err != nil {
		t.Fatal(err)
	}

	ref, err := api.PodShare(pod, "", sessionId)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
