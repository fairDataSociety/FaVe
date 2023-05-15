package document

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	h "github.com/fairDataSociety/FaVe/pkg/hnsw"
	"github.com/fairDataSociety/FaVe/pkg/lookup"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	"github.com/sirupsen/logrus"
	"github.com/weaviate/weaviate/adapters/repos/db/vector/hnsw/distancer"
)

const (
	errLevel   = logrus.ErrorLevel
	debugLevel = logrus.DebugLevel

	hnswIndexName = "hnswId"
)

// Config for fairOS-dfs
type Config struct {
	Verbose     bool
	GlovePodRef string
}

type Client struct {
	lock      sync.Mutex
	hnswLock  sync.RWMutex
	api       *dfs.API
	indices   map[string]h.VectorIndex
	pod       string
	logger    logging.Logger
	sessionId string
	podInfo   *pod.Info
	lookup    *lookup.Lookup
}

type Collection struct {
	Name    string
	Indexes map[string]collection.IndexType
}

type Document struct {
	ID         string
	Properties map[string]interface{}
}

func New(config Config, api *dfs.API) (*Client, error) {
	// Set the log level
	level := errLevel
	if config.Verbose {
		level = debugLevel
	}
	logger := logging.New(os.Stdout, level)

	// TODO support multiple languages
	lkup, err := lookup.New(api, config.GlovePodRef, lookup.GloveStore, lookup.Stopwords["en"])
	if err != nil {
		logger.Errorf("new lookup failed :%s\n", err.Error())
		return nil, err
	}
	return &Client{
		api:     api,
		logger:  logger,
		lookup:  lkup,
		indices: map[string]h.VectorIndex{},
	}, nil
}

func (c *Client) Login(username, password string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	ui, err := c.api.LoginUserV2(username, password, "")
	if err != nil {
		return err
	}
	c.sessionId = ui.UserInfo.GetSessionId()
	return nil
}

func (c *Client) OpenPod(pod string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.pod = pod
	if c.sessionId == "" {
		return dfs.ErrUserNotLoggedIn
	}
	if !c.api.IsPodExist(c.pod, c.sessionId) {
		_, err := c.api.CreatePod(c.pod, c.sessionId)
		if err != nil {
			return err
		}
	}
	pi, err := c.api.OpenPod(c.pod, c.sessionId)
	if err != nil {
		return err
	}
	c.podInfo = pi
	return nil
}

func (c *Client) CreateCollection(col *Collection) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sessionId == "" {
		return dfs.ErrUserNotLoggedIn
	}
	if c.podInfo == nil {
		return dfs.ErrPodNotOpen
	}
	col.Indexes[hnswIndexName] = collection.NumberIndex
	//col.Indexes["vector"] = collection.VectorIndex
	err := c.api.DocCreate(c.sessionId, c.pod, col.Name, col.Indexes, true)
	if err != nil {
		return err
	}

	vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
		expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
		docs, err := c.api.DocFind(c.sessionId, c.pod, col.Name, expr, 1)
		if err != nil {
			return nil, err
		}
		doc := docs[0]
		data := map[string]interface{}{}
		err = json.Unmarshal(doc, &data)
		if err != nil {
			return nil, err
		}
		fmt.Println("====vectorForID ", id, data)
		return data["vector"].([]float32), nil
	}
	kvStore := c.podInfo.GetKVStore()

	err = kvStore.CreateKVTable(col.Name, c.podInfo.GetPodPassword(), collection.StringIndex)
	if err != nil && err != collection.ErrKvTableAlreadyPresent {
		return err
	}

	err = kvStore.OpenKVTable(col.Name, c.podInfo.GetPodPassword())
	if err != nil {
		return err
	}

	makeCL := h.MakeNoopCommitLogger
	index, err := h.New(h.Config{
		RootPath:              "not-used",
		ID:                    "not-used",
		MakeCommitLoggerThunk: makeCL,
		DistanceProvider:      distancer.NewCosineDistanceProvider(),
		VectorForIDThunk:      vectorForID,
		ClassName:             col.Name,
	}, h.UserConfig{
		MaxConnections: 30,
		EFConstruction: 60,
	}, kvStore)
	if err != nil {
		return err
	}

	c.hnswLock.Lock()
	c.indices[col.Name] = index
	c.hnswLock.Unlock()
	return c.api.DocOpen(c.sessionId, c.pod, col.Name)
}

func (c *Client) AddDocuments(collection string, documents ...*Document) error {
	batch, err := c.api.DocBatch(c.sessionId, c.pod, collection)
	if err != nil {
		return err
	}
	c.hnswLock.Lock()
	index := c.indices[collection]
	c.hnswLock.Unlock()
	for id, doc := range documents {
		// vectorise the properties
		// add the vector in the properties before adding the document in the collection
		vectorData := ""
		for _, prop := range doc.Properties {
			vectorData += prop.(string) + " "
		}
		vector, err := c.lookup.Corpi([]string{vectorData})
		if err != nil {
			c.logger.Errorf("corpi failed :%s\n", err.Error())
			continue
		}
		fmt.Println("====AddDocuments vector ", vector.ToArray())
		doc.Properties["vector"] = vector.ToArray()

		doc.Properties[hnswIndexName] = id
		doc.Properties["id"] = doc.ID
		data, err := json.Marshal(doc.Properties)
		if err != nil {
			c.logger.Errorf("marshal document failed :%s\n", err.Error())
			continue
		}
		err = c.api.DocBatchPut(c.sessionId, c.pod, data, batch)
		if err != nil {
			c.logger.Errorf("DocBatchPut failed :%s\n", err.Error())
			return err
		}
		err = index.Add(uint64(id), vector.ToArray())
		if err != nil {
			c.logger.Errorf("index.Add failed :%s\n", err.Error())
			return err
		}
	}
	c.logger.Infof("batch write %d documents\n", len(documents))
	return c.api.DocBatchWrite(c.sessionId, c.pod, batch)
}

func (c *Client) GetNearDocuments(text, collection string, distance float32) ([][]byte, error) {
	vector, err := c.lookup.Corpi([]string{text})
	if err != nil {
		return nil, err
	}
	fmt.Println("====GetNearDocuments , vector:", vector.ToArray())
	c.hnswLock.Lock()
	index := c.indices[collection]
	c.hnswLock.Unlock()
	ids, err := index.KnnSearchByVectorMaxDist(vector.ToArray(), distance, 36, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("====GetNearDocuments , ids:", ids)
	documents := make([][]byte, len(ids))
	for i, id := range ids {
		expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
		docs, err := c.api.DocFind(c.sessionId, c.pod, collection, expr, 1)
		if err != nil {
			return nil, err
		}
		fmt.Println("====GetNearDocuments , docs:", string(docs[0]))
		documents[i] = docs[0]
	}
	return documents, nil
}
