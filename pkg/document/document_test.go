package document

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	mockpost "github.com/ethersphere/bee/pkg/postage/mock"
	mockstorer "github.com/ethersphere/bee/pkg/storer/mock"
	"github.com/fairDataSociety/FaVe/pkg/vectorizer/rest"
	"github.com/fairdatasociety/fairOS-dfs/pkg/blockstore/bee"
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

var (
	stopWords = []string{
		"this",
		"This",
		"that",
		"That",
		"the",
		"The",
		"an",
		"An",
		"of",
		"Of",
		"in",
		"In",
		"and",
		"And",
		"to",
		"To",
		"was",
		"Was",
		"is",
		"Is",
		"for",
		"For",
		"on",
		"On",
		"as",
		"As",
	}
)

type Metadata struct {
	Title   string `json:"Title"`
	Content struct {
		ArticleHTML     string `json:"article.html"`
		ArticleWikitext string `json:"article.wikitext"`
		ArticleTxt      string `json:"article.txt"`
	} `json:"Content"`
}

func getContent(file *zip.File) ([]byte, error) {
	fileReader, err := file.Open()
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer fileReader.Close()
	buffer, err := io.ReadAll(fileReader)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}
	return buffer, nil
}

func TestFave(t *testing.T) {
	storer := mockstorer.New()
	beeUrl := mock.NewTestBeeServer(t, mock.TestServerOptions{
		Storer:          storer,
		PreventRedirect: true,
		Post:            mockpost.New(mockpost.WithAcceptAll()),
	})
	fmt.Println(beeUrl)
	logger := logging.New(io.Discard, logrus.DebugLevel)
	mockClient := bee.NewBeeClient(beeUrl, mock.BatchOkStr, true, logger)
	ens := mock2.NewMockNamespaceManager()

	users := user.NewUsers(mockClient, ens, 1000, time.Minute*60, logger)
	dfsApi := dfs.NewMockDfsAPI(mockClient, users, logger)
	defer dfsApi.Close()

	//t.Run("test-zwis-in-fave", func(t *testing.T) {
	//	_, err := dfsApi.CreateUserV2(username+"ww", password, ""+
	//		"", "")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	cfg := Config{
	//		Verbose:       false,
	//		VectorizerUrl: "http://localhost:9876",
	//	}
	//	client, err := New(cfg, dfsApi)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	err = client.Login(username+"ww", password)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = client.OpenPod("Fave")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	documents := []*Document{}
	//
	//	entries, err := os.ReadDir("/Users/sabyasachipatra/Downloads/citi")
	//	if err != nil {
	//		t.Fatal("Error opening zwi source file:", err)
	//	}
	//	for _, entry := range entries {
	//
	//		zipFile, err := zip.OpenReader(filepath.Join("/Users/sabyasachipatra/Downloads/citi", entry.Name()))
	//		if err != nil {
	//			fmt.Println("Error opening ZIP file:", err)
	//			continue
	//		}
	//		defer zipFile.Close()
	//		doc := &Document{
	//			ID:         uuid.New().String(),
	//			Properties: map[string]interface{}{},
	//		}
	//		for _, file := range zipFile.File {
	//
	//			if file.Name == "article.txt" || file.Name == "metadata.json" || file.Name == "article.html" {
	//				buffer, err := getContent(file)
	//				if err != nil {
	//					fmt.Println("Error reading file:", err)
	//					continue
	//				}
	//				switch file.Name {
	//				case "article.html":
	//					doc.Properties["html"] = string(buffer)
	//				case "article.txt":
	//					doc.Properties["rawText"] = string(buffer)
	//					re := regexp.MustCompile(`\|.*`)
	//					filteredText := re.ReplaceAllString(string(buffer), "")
	//
	//					re2 := regexp.MustCompile(`(?m)^This editable Main Article.*$`)
	//					filteredText = re2.ReplaceAllString(filteredText, "")
	//
	//					re3 := regexp.MustCompile(`(?m)^This article.*$`)
	//					filteredText = re3.ReplaceAllString(filteredText, "")
	//
	//					doc.Properties["article"] = filteredText
	//				case "metadata.json":
	//					metadata := &Metadata{}
	//					err = json.Unmarshal(buffer, metadata)
	//					if err != nil {
	//						fmt.Println("Error unmarshalling JSON:", err)
	//						continue
	//					}
	//					doc.Properties["title"] = metadata.Title
	//				}
	//			}
	//		}
	//
	//		if doc.Properties["article"] == "" {
	//			log.Println("article.txt not found")
	//			continue
	//		}
	//		if doc.Properties["title"] == "" {
	//			log.Println("metadata.json not found in zwi file", entry.Name())
	//			continue
	//		}
	//		if doc.Properties["html"] == "" {
	//			log.Println("article.html not found in zwi file", entry.Name())
	//			continue
	//		}
	//		documents = append(documents, doc)
	//	}
	//
	//	col := &Collection{
	//		Name: "zwis",
	//		Indexes: map[string]collection.IndexType{
	//			"title": collection.StringIndex,
	//		},
	//	}
	//
	//	err = client.CreateCollection(col)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = client.AddDocuments(col.Name, []string{"article"}, documents...)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	<-time.After(time.Second * 30)
	//	//for i, _ := range documents {
	//	//	s, v, err := dfsApi.KVGet(client.sessionId, client.pod, namespace+col.Name, fmt.Sprintf("%d", i))
	//	//	if err != nil {
	//	//		t.Log(i, err)
	//	//	}
	//	//	fmt.Println(s, string(v))
	//	//}
	//	//
	//	//for id, _ := range documents {
	//	//	expr := fmt.Sprintf("%s=%d", "hnswId", id)
	//	//	docsFound, err := client.api.DocFind(client.sessionId, client.pod, "fave_zwis", expr, 1)
	//	//	if err != nil {
	//	//		t.Fatal(err)
	//	//	}
	//	//	for _, doc := range docsFound {
	//	//		data := map[string]interface{}{}
	//	//		err = json.Unmarshal(doc, &data)
	//	//		if err != nil {
	//	//			t.Fatal(err)
	//	//		}
	//	//
	//	//		fmt.Println("doc Found:", data["vector"])
	//	//		fmt.Println("doc Found:", data["id"])
	//	//		fmt.Println("doc Found:", data["title"])
	//	//		fmt.Println("doc Found:", data["hnswId"])
	//	//	}
	//	//}
	//
	//	match, mismatch := 0, 0
	//	for i, ddoc := range documents {
	//		if i%10 == 0 {
	//			<-time.After(time.Second * 30)
	//		}
	//		fmt.Println("========= GetNearDocuments", ddoc.Properties["title"])
	//		docs, _, err := client.GetNearDocuments(col.Name, fmt.Sprintf("%s", ddoc.Properties["title"]), 1, 30)
	//		if err != nil {
	//			t.Fatal(err, ddoc.Properties["title"])
	//		}
	//		found := false
	//		for _, doc := range docs {
	//			props := map[string]interface{}{}
	//			err := json.Unmarshal(doc, &props)
	//			if err != nil {
	//				t.Fatal(err)
	//			}
	//			if props["title"] == ddoc.Properties["title"] {
	//				found = true
	//				break
	//			}
	//		}
	//
	//		if found {
	//			match++
	//		} else {
	//			fmt.Println("Mismatch for", ddoc.Properties["title"])
	//			mismatch++
	//		}
	//	}
	//	fmt.Println("Match:", match, "Mismatch:", mismatch)
	//	//expr := fmt.Sprintf("%s>%s", "title", "")
	//	//docsFound, err := client.api.DocFind(client.sessionId, client.pod, "fave_zwis", expr, 1000)
	//	//if err != nil {
	//	//	t.Fatal(err)
	//	//}
	//	//for _, doc := range docsFound {
	//	//	data := map[string]interface{}{}
	//	//	err = json.Unmarshal(doc, &data)
	//	//	if err != nil {
	//	//		t.Fatal(err)
	//	//	}
	//	//
	//	//	//embeddingStr := fmt.Sprintf("\"%s\" %v\n", data["title"], data["vector"])
	//	//	//
	//	//	//// Write the embedding string to the file
	//	//	//_, err := writer.WriteString(embeddingStr)
	//	//	//if err != nil {
	//	//	//	fmt.Println("Error writing to file:", err)
	//	//	//	return
	//	//	//}
	//	//
	//	//	fmt.Println("doc Found:", data["vector"])
	//	//	fmt.Println("doc Found:", data["id"])
	//	//	fmt.Println("doc Found:", data["title"])
	//	//	fmt.Println("doc Found:", data["hnswId"])
	//	//}
	//	//writer.Flush()
	//
	//})

	//t.Run("test-zwis-in-fave-different-client", func(t *testing.T) {
	//	_, err := dfsApi.CreateUserV2(username+"ww", password, ""+
	//		"", "")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	cfg := Config{
	//		Verbose:       false,
	//		VectorizerUrl: "http://localhost:9876",
	//	}
	//	client, err := New(cfg, dfsApi)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	err = client.Login(username+"ww", password)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = client.OpenPod("Fave")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	documents := []*Document{}
	//
	//	entries, err := os.ReadDir("/Users/sabyasachipatra/Downloads/citizendium")
	//	if err != nil {
	//		t.Fatal("Error opening zwi source file:", err)
	//	}
	//	for _, entry := range entries {
	//
	//		zipFile, err := zip.OpenReader(filepath.Join("/Users/sabyasachipatra/Downloads/citizendium", entry.Name()))
	//		if err != nil {
	//			fmt.Println("Error opening ZIP file:", err)
	//			continue
	//		}
	//		defer zipFile.Close()
	//		doc := &Document{
	//			ID:         uuid.New().String(),
	//			Properties: map[string]interface{}{},
	//		}
	//		for _, file := range zipFile.File {
	//
	//			if file.Name == "article.txt" || file.Name == "metadata.json" || file.Name == "article.html" {
	//				buffer, err := getContent(file)
	//				if err != nil {
	//					fmt.Println("Error reading file:", err)
	//					continue
	//				}
	//				switch file.Name {
	//				case "article.html":
	//					doc.Properties["html"] = string(buffer)
	//				case "article.txt":
	//					doc.Properties["rawText"] = string(buffer)
	//					re := regexp.MustCompile(`\|.*`)
	//					filteredText := re.ReplaceAllString(string(buffer), "")
	//
	//					re2 := regexp.MustCompile(`(?m)^This editable Main Article.*$`)
	//					filteredText = re2.ReplaceAllString(filteredText, "")
	//
	//					re3 := regexp.MustCompile(`(?m)^This article.*$`)
	//					filteredText = re3.ReplaceAllString(filteredText, "")
	//
	//					//stopWordsPattern := strings.Join(stopWords, "|")
	//					//re4 := regexp.MustCompile(`\b(` + stopWordsPattern + `)\b`)
	//					//filteredText = re4.ReplaceAllString(stripmd.Strip(filteredText), "")
	//
	//					doc.Properties["article"] = filteredText
	//				case "metadata.json":
	//					metadata := &Metadata{}
	//					err = json.Unmarshal(buffer, metadata)
	//					if err != nil {
	//						fmt.Println("Error unmarshalling JSON:", err)
	//						continue
	//					}
	//					doc.Properties["title"] = metadata.Title
	//				}
	//			}
	//		}
	//
	//		if doc.Properties["article"] == "" {
	//			log.Println("article.txt not found")
	//			continue
	//		}
	//		if doc.Properties["title"] == "" {
	//			log.Println("metadata.json not found in zwi file", entry.Name())
	//			continue
	//		}
	//		if doc.Properties["html"] == "" {
	//			log.Println("article.html not found in zwi file", entry.Name())
	//			continue
	//		}
	//		documents = append(documents, doc)
	//	}
	//
	//	col := &Collection{
	//		Name: "zwis",
	//		Indexes: map[string]collection.IndexType{
	//			"title": collection.StringIndex,
	//		},
	//	}
	//
	//	err = client.CreateCollection(col)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = client.AddDocuments(col.Name, []string{"article"}, documents...)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	//for i, _ := range documents {
	//	//	s, v, err := dfsApi.KVGet(client.sessionId, client.pod, namespace+col.Name, fmt.Sprintf("%d", i))
	//	//	if err != nil {
	//	//		log.Fatal(err)
	//	//	}
	//	//	fmt.Println(s, string(v))
	//	//}
	//
	//	//fmt.Println("========= DocFind")
	//	//expr := fmt.Sprintf("%s>%s", "title", "")
	//	//docsFound, err := client.api.DocFind(client.sessionId, client.pod, "fave_zwis", expr, 1000)
	//	//if err != nil {
	//	//	t.Fatal(err)
	//	//}
	//	//for _, doc := range docsFound {
	//	//	data := map[string]interface{}{}
	//	//	err = json.Unmarshal(doc, &data)
	//	//	if err != nil {
	//	//		t.Fatal(err)
	//	//	}
	//	//
	//	//	fmt.Println("doc Found:", data["vector"])
	//	//	fmt.Println("doc Found:", data["id"])
	//	//	fmt.Println("doc Found:", data["title"])
	//	//	fmt.Println("doc Found:", data["hnswId"])
	//	//}
	//
	//	client2, err := New(cfg, dfsApi)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = client2.Login(username+"ww", password)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//
	//	err = client2.OpenPod("Fave")
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	<-time.After(time.Second * 30)
	//	fmt.Println()
	//	fmt.Println()
	//	//for i, _ := range documents {
	//	//	s, v, err := dfsApi.KVGet(client2.sessionId, client2.pod, namespace+col.Name, fmt.Sprintf("%d", i))
	//	//	if err != nil {
	//	//		log.Fatal(err)
	//	//	}
	//	//	fmt.Println(s, string(v))
	//	//}
	//	//	count := 0
	//	//redo:
	//	//	if count == 5 {
	//	//		t.Fatal("failed to find document")
	//	//	}
	//	//	docs, _, err := client2.GetNearDocuments(col.Name, "One-way encryption", 1, 10)
	//	//	if err != nil {
	//	//		t.Fatal(err)
	//	//	}
	//	//	found := false
	//	//	for _, doc := range docs {
	//	//		props := map[string]interface{}{}
	//	//		err := json.Unmarshal(doc, &props)
	//	//		if err != nil {
	//	//			t.Fatal(err)
	//	//		}
	//	//		if props["title"] == "One-way encryption" {
	//	//			found = true
	//	//			break
	//	//		}
	//	//	}
	//	//	if !found {
	//	//		count++
	//	//		goto redo
	//	//	}
	//	//	fmt.Println(count)
	//
	//	match, mismatch := 0, 0
	//	for i, ddoc := range documents {
	//		if i%10 == 0 {
	//			<-time.After(time.Second * 30)
	//		}
	//		count := 0
	//	redo:
	//		fmt.Println("========= GetNearDocuments", ddoc.Properties["title"], i, i%10)
	//		docs, _, err := client2.GetNearDocuments(col.Name, fmt.Sprintf("%s", ddoc.Properties["title"]), 1, 10)
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//		found := false
	//		for _, doc := range docs {
	//			props := map[string]interface{}{}
	//			err := json.Unmarshal(doc, &props)
	//			if err != nil {
	//				t.Fatal(err)
	//			}
	//			if props["title"] == ddoc.Properties["title"] {
	//				found = true
	//				break
	//			}
	//		}
	//		if !found && count < 5 {
	//			count++
	//			goto redo
	//		}
	//		if found {
	//			match++
	//		} else {
	//			fmt.Println("Mismatch for", ddoc.Properties["title"])
	//			mismatch++
	//		}
	//	}
	//	fmt.Println("Match:", match, "Mismatch:", mismatch)
	//})

	t.Run("test-vectorizer-in-fave", func(t *testing.T) {
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
		file, err := os.Open("./wiki-100.csv")
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
		fmt.Println("=====================================")

		<-time.After(50 * time.Second)
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

		docs, dist, err = client2.GetNearDocuments(col.Name, "Bat", 1, 10)
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
		fmt.Println("=====================================")
		docs, dist, err = client2.GetNearDocuments(col.Name, "Tiger", 1, 10)
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
		_, err := dfsApi.CreateUserV2(username+"v", password, ""+
			"", "")
		if err != nil {
			t.Fatal(err)
		}

		cfg := Config{
			Verbose: false,
		}
		client, err := New(cfg, dfsApi)
		if err != nil {
			t.Fatal(err)
		}
		err = client.Login(username+"v", password)
		if err != nil {
			t.Fatal(err)
		}

		err = client.OpenPod("Fave")
		if err != nil {
			t.Fatal(err)
		}

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
		<-time.After(50 * time.Second)
		client2, err := New(cfg, dfsApi)
		if err != nil {
			t.Fatal(err)
		}

		err = client2.Login(username+"v", password)
		if err != nil {
			t.Fatal(err)
		}

		err = client2.OpenPod("Fave")
		if err != nil {
			t.Fatal(err)
		}

		colls, err := client2.GetCollections()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Collections:", colls[0].Name)
	})

	t.Run("test-vectorizer-in-fave-add-to-collection-multiple-times", func(t *testing.T) {
		_, err := dfsApi.CreateUserV2(username+"w", password, ""+
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
		err = client.Login(username+"w", password)
		if err != nil {
			t.Fatal(err)
		}

		err = client.OpenPod("Fave")
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.Open("./wiki-100.csv")
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
		//fmt.Println(len(documents))
		//err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents[0:10]...)
		//if err != nil {
		//	t.Fatal(err)
		//}
		//fmt.Println("Added 10 documents")
		//err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents[10:]...)
		//if err != nil {
		//	t.Fatal(err)
		//}
		for i := 0; i < 10; i++ {
			err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents[i*10:(i*10)+10]...)
			if err != nil {
				t.Fatal(err)
			}
		}
		err = client.AddDocuments(col.Name, []string{"title", "rawText"}, documents[100:]...)
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

		for i, _ := range documents {
			s, v, err := dfsApi.KVGet(client.sessionId, client.pod, namespace+col.Name, fmt.Sprintf("%d", i))
			if err != nil {
				t.Log(i, err)
			}
			fmt.Println(s, string(v))
		}

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
		//}

		fmt.Println("===================================== mammals with true flight")
		docs, dist, err := client.GetNearDocuments(col.Name, "mammals with true flight", 1, 10)
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
		fmt.Println("===================================== largest living cat")
		docs, dist, err = client.GetNearDocuments(col.Name, "largest living cat", 1, 10)
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
		fmt.Println("===================================== Tiger")
		docs, dist, err = client.GetNearDocuments(col.Name, "Tiger", 1, 10)
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
		fmt.Println("===================================== dark vertical stripes on orange")
		docs, dist, err = client.GetNearDocuments(col.Name, "dark vertical stripes on orange", 1, 10)
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
		<-time.After(50 * time.Second)
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
		fmt.Println("===================================== 2 mammals with true flight")
		docs, dist, err = client2.GetNearDocuments(col.Name, "mammals with true flight", 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		for i, doc := range docs {
			props := map[string]interface{}{}
			err := json.Unmarshal(doc, &props)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Found: ", props["title"], dist[i])
		}
		fmt.Println("===================================== 2 largest living cat")
		docs, dist, err = client2.GetNearDocuments(col.Name, "largest living cat", 1, 10)
		if err != nil {
			t.Fatal(err)
		}
		for i, doc := range docs {
			props := map[string]interface{}{}
			err := json.Unmarshal(doc, &props)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println("Found: ", props["title"], dist[i])
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
