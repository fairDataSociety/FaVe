package document

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fairDataSociety/FaVe/pkg/vectorizer/rest"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	mock2 "github.com/fairdatasociety/fairOS-dfs/pkg/ensm/eth/mock"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/user"
	"github.com/google/uuid"
	"github.com/jdkato/prose/v2"
	"github.com/sirupsen/logrus"
)

const (
	username = "testuser"
	password = "testpasswordtestpassword"
)

func TestFave(t *testing.T) {
	mockClient := mock.NewMockBeeClient()
	ens := mock2.NewMockNamespaceManager()
	logger := logging.New(os.Stdout, logrus.ErrorLevel)

	users := user.NewUsers(mockClient, ens, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger)
	defer dfsApi.Close()

	_, err := dfsApi.CreateUserV2(username, password, ""+
		"", "")
	if err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		Verbose:       false,
		VectorizerUrl: "http://localhost:9876",
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

	t.Run("test-vectorizer-in-fave", func(t *testing.T) {

		file, err := os.Open("./wiki-15.csv")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		// Create a CSV reader
		reader := csv.NewReader(file)

		// Read all records from the CSV file
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		documents, err := generateDocuments(records)
		if err != nil {
			t.Fatal(err)
		}

		col := &Collection{
			Name: "Wiki",
			Indexes: map[string]collection.IndexType{
				"title":   collection.StringIndex,
				"rawText": collection.StringIndex,
			},
		}

		err = client.CreateCollection(col)
		if err != nil {
			t.Fatal(err)
		}
		err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents...)
		if err != nil {
			t.Fatal(err)
		}
		//// adding second time
		//documents, err = generateDocuments(records)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//
		//err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents...)
		//if err != nil {
		//	t.Fatal(err)
		//}

		//for i, _ := range documents {
		//	s, v, err := dfsApi.KVGet(client.sessionId, client.pod, namespace+col.Name, fmt.Sprintf("%d", i))
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	fmt.Println(s, string(v))
		//}

		//expr := fmt.Sprintf("%s=%d", "hnswId", 13)
		//docs, err := client.api.DocFind(client.sessionId, client.pod, col.Name, expr, 1)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//if len(docs) > 0 {
		//	doc := docs[0]
		//	data := map[string]interface{}{}
		//	err = json.Unmarshal(doc, &data)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	fmt.Println("Found:", data["vector"])
		//	fmt.Println("Found:", data["id"])
		//	fmt.Println("Found:", data["title"])
		//} else {
		//	fmt.Println("Not found")
		//}

		//for id, _ := range records {
		//expr := fmt.Sprintf("%s=%d", "hnswId", 14)
		//docs, err := client.api.DocFind(client.sessionId, client.pod, col.Name, expr, 1)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//if len(docs) > 0 {
		//	doc := docs[0]
		//	data := map[string]interface{}{}
		//	err = json.Unmarshal(doc, &data)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	fmt.Println("Found:", data["vector"])
		//	fmt.Println("Found:", data["id"])
		//	fmt.Println("Found:", data["title"])
		//} else {
		//	fmt.Println("Not found")
		//}
		//}

		//look for documents
		//for _, record := range records {
		//	fmt.Println("Searching for:", record[0])
		//	// Test search
		//	docs, dist, err := client.GetNearDocuments(col.Name, record[0], 1)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	for i, doc := range docs {
		//		props := map[string]interface{}{}
		//		err := json.Unmarshal(doc, &props)
		//		if err != nil {
		//			t.Fatal(err)
		//		}
		//		fmt.Println("Found:", props["title"], dist[i])
		//	}
		//	fmt.Println("=====================================")
		//}

		//for _, record := range records {
		//	fmt.Println("Searching for:", record[0])
		//	// Test search
		//	docs, _, err := client.GetNearDocuments(col.Name, record[0], 1)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	if len(docs) == 0 {
		//		t.Log("No documents found for", record[0])
		//	} else {
		//		props := map[string]interface{}{}
		//		err := json.Unmarshal(docs[0], &props)
		//		if err != nil {
		//			t.Fatal(err)
		//		}
		//		if props["title"] != record[0] {
		//			t.Log("Found:", props["title"], "Expected:", record[0])
		//		}
		//	}
		//
		//	fmt.Println("=====================================")
		//}
		docs, dist, err := client.GetNearDocuments(col.Name, "Bat", 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		for i, doc := range docs {
			props := map[string]interface{}{}
			err := json.Unmarshal(doc, &props)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Found:", props["title"], dist[i])
		}

		client2, err := New(cfg, dfsApi)
		if err != nil {
			t.Fatal(err)
		}

		err = client2.Login(username, password)
		if err != nil {
			t.Fatal(err)
		}

		err = client2.OpenPod("Fave")
		if err != nil {
			t.Fatal(err)
		}

		docs, dist, err = client2.GetNearDocuments(col.Name, "Bat", 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i, doc := range docs {
			props := map[string]interface{}{}
			err := json.Unmarshal(doc, &props)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Found:", props["title"], dist[i])
		}

		colls, err := client2.GetCollections()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Collections:", colls[0].Name)
	})

	t.Run("test-vectorizer-out-of-fave", func(t *testing.T) {
		vctrzr, err := rest.NewVectorizer("http://localhost:9876")
		if err != nil {
			t.Fatal(err)
		}

		file, err := os.Open("./wiki-15.csv")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		// Create a CSV reader
		reader := csv.NewReader(file)

		// Read all records from the CSV file
		records, err := reader.ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		documents, err := generateDocuments(records)
		if err != nil {
			t.Fatal(err)
		}

		for _, doc := range documents {
			vector, err := vctrzr.Corpi([]string{doc.Properties["rawText"].(string)})
			if err != nil {
				t.Fatal(err)
			}
			doc.Properties["vector"] = vector.ToArray()
		}

		col := &Collection{
			Name: "Wiki2",
			Indexes: map[string]collection.IndexType{
				"title":   collection.StringIndex,
				"rawText": collection.StringIndex,
			},
		}

		err = client.CreateCollection(col)
		if err != nil {
			t.Fatal(err)
		}
		err = client.AddDocuments(col.Name, []string{}, documents...)
		if err != nil {
			t.Fatal(err)
		}
		//// adding second time
		//documents, err = generateDocuments(records)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//
		//err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents...)
		//if err != nil {
		//	t.Fatal(err)
		//}

		//for i, _ := range documents {
		//	s, v, err := dfsApi.KVGet(client.sessionId, client.pod, namespace+col.Name, fmt.Sprintf("%d", i))
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	fmt.Println(s, string(v))
		//}

		//expr := fmt.Sprintf("%s=%d", "hnswId", 13)
		//docs, err := client.api.DocFind(client.sessionId, client.pod, col.Name, expr, 1)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//if len(docs) > 0 {
		//	doc := docs[0]
		//	data := map[string]interface{}{}
		//	err = json.Unmarshal(doc, &data)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	fmt.Println("Found:", data["vector"])
		//	fmt.Println("Found:", data["id"])
		//	fmt.Println("Found:", data["title"])
		//} else {
		//	fmt.Println("Not found")
		//}

		//for id, _ := range records {
		//expr := fmt.Sprintf("%s=%d", "hnswId", 14)
		//docs, err := client.api.DocFind(client.sessionId, client.pod, col.Name, expr, 1)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//if len(docs) > 0 {
		//	doc := docs[0]
		//	data := map[string]interface{}{}
		//	err = json.Unmarshal(doc, &data)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	fmt.Println("Found:", data["vector"])
		//	fmt.Println("Found:", data["id"])
		//	fmt.Println("Found:", data["title"])
		//} else {
		//	fmt.Println("Not found")
		//}
		//}

		//look for documents
		//for _, record := range records {
		//	fmt.Println("Searching for:", record[0])
		//	// Test search
		//	docs, dist, err := client.GetNearDocuments(col.Name, record[0], 1)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	for i, doc := range docs {
		//		props := map[string]interface{}{}
		//		err := json.Unmarshal(doc, &props)
		//		if err != nil {
		//			t.Fatal(err)
		//		}
		//		fmt.Println("Found:", props["title"], dist[i])
		//	}
		//	fmt.Println("=====================================")
		//}

		//for _, record := range records {
		//	fmt.Println("Searching for:", record[0])
		//	// Test search
		//	docs, _, err := client.GetNearDocuments(col.Name, record[0], 1)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//	if len(docs) == 0 {
		//		t.Log("No documents found for", record[0])
		//	} else {
		//		props := map[string]interface{}{}
		//		err := json.Unmarshal(docs[0], &props)
		//		if err != nil {
		//			t.Fatal(err)
		//		}
		//		if props["title"] != record[0] {
		//			t.Log("Found:", props["title"], "Expected:", record[0])
		//		}
		//	}
		//
		//	fmt.Println("=====================================")
		//}

		vectorToFind, err := vctrzr.Corpi([]string{"Bat"})
		if err != nil {
			t.Fatal(err)
		}

		docs, dist, err := client.GetNearDocumentsByVector(col.Name, vectorToFind.ToArray(), 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		for i, doc := range docs {
			props := map[string]interface{}{}
			err := json.Unmarshal(doc, &props)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Found:", props["title"], dist[i])
		}

		client2, err := New(cfg, dfsApi)
		if err != nil {
			t.Fatal(err)
		}

		err = client2.Login(username, password)
		if err != nil {
			t.Fatal(err)
		}

		err = client2.OpenPod("Fave")
		if err != nil {
			t.Fatal(err)
		}

		docs, dist, err = client2.GetNearDocuments(col.Name, "Bat", 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i, doc := range docs {
			props := map[string]interface{}{}
			err := json.Unmarshal(doc, &props)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Found:", props["title"], dist[i])
		}

		colls, err := client2.GetCollections()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Collections:", colls[0].Name)
	})
}

func generateDocuments(records [][]string) ([]*Document, error) {
	documents := []*Document{}

	// Print the records
	for _, record := range records {
		docTokens, err := prose.NewDocument(record[1])
		if err != nil {
			return nil, err
		}
		docTokens.Tokens()
		tokens := []string{}
		for _, tok := range docTokens.Tokens() {
			if (tok.Tag == "NN" || tok.Tag == "NNP" || tok.Tag == "NNPS" || tok.Tag == "NNS") && len(tok.Text) > 2 {
				tokens = append(tokens, tok.Text)
			}
		}
		doc := &Document{
			ID: uuid.New().String(),
			Properties: map[string]interface{}{
				"title":   record[0],
				"rawText": strings.Join(tokens, " "),
			},
		}
		fmt.Println("Adding:", doc.Properties["title"], "Raw Text: ", strings.Join(tokens, " "))
		documents = append(documents, doc)
	}
	return documents, nil
}
