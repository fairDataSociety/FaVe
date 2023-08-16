package document

import (
	"context"
	"encoding/json"
	"fmt"
	h "github.com/fairDataSociety/FaVe/pkg/hnsw"
	"github.com/fairDataSociety/FaVe/pkg/hnsw/distancer"
	"github.com/fairDataSociety/FaVe/pkg/vectorizer"
	"github.com/fairDataSociety/FaVe/pkg/vectorizer/rest"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/fairdatasociety/fairOS-dfs/pkg/pod"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

const (
	errLevel   = logrus.ErrorLevel
	debugLevel = logrus.DebugLevel

	hnswIndexName = "hnswId"
)

// Config for fairOS-dfs
type Config struct {
	Verbose         bool
	GloveLevelDBUrl string
}

type Client struct {
	lock          sync.Mutex
	hnswLock      sync.RWMutex
	api           *dfs.API
	indices       map[string]h.VectorIndex
	pod           string
	logger        logging.Logger
	sessionId     string
	podInfo       *pod.Info
	lookup        vectorizer.Vectorizer
	documentCache *lru.Cache
}

type Collection struct {
	Name    string
	Indexes map[string]collection.IndexType
}

type Document struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
}

func New(config Config, api *dfs.API) (*Client, error) {
	// Set the log level
	level := errLevel
	if config.Verbose {
		level = debugLevel
	}
	logger := logging.New(os.Stdout, level)

	client := &Client{
		api:     api,
		logger:  logger,
		indices: map[string]h.VectorIndex{},
	}
	// TODO support multiple languages
	//if config.GlovePodRef != "" {
	//	lkup, err := dfsLookup.New(api, config.GlovePodRef, dfsLookup.GloveStore, dfsLookup.Stopwords["en"])
	//	if err != nil {
	//		logger.Errorf("new lookuper failed :%s\n", err.Error())
	//		return nil, err
	//	}
	//	client.vectorizer = lkup
	//}

	if config.GloveLevelDBUrl == "" {
		logger.Errorf("GLOVE_LEVELDB_URL environment variable is not set")
	}

	// leveldb lookuper
	lkup, err := rest.NewVectorizer(config.GloveLevelDBUrl)
	if err != nil {
		logger.Errorf("new vectorizer failed :%s\n", err.Error())
		return nil, err
	}
	client.lookup = lkup
	documentCache, err := lru.New(1000)
	if err == nil {
		client.documentCache = documentCache
	}
	return client, nil
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if c.sessionId == "" {
		return dfs.ErrUserNotLoggedIn
	}
	if c.podInfo == nil {
		return dfs.ErrPodNotOpen
	}
	col.Indexes[hnswIndexName] = collection.NumberIndex

	vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
		// check if the document is in the cache
		if v, ok := c.documentCache.Get(fmt.Sprintf("%s/%s/%d", c.pod, col.Name, id)); ok {
			return v.([]float32), nil
		}
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
		vector, err := convertToFloat32Slice(data["vector"])
		if err != nil {
			return nil, err
		}
		c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, col.Name, id), vector)
		return vector, err
	}
	kvStore := c.podInfo.GetKVStore()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			return
		default:
		}
		err := c.api.DocCreate(c.sessionId, c.pod, col.Name, col.Indexes, true)
		if err != nil {
			cancel()
			return
		}
		err = c.api.DocOpen(c.sessionId, c.pod, col.Name)
		if err != nil {
			cancel()
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			return
		default:
		}
		err := kvStore.CreateKVTable(col.Name, c.podInfo.GetPodPassword(), collection.StringIndex)
		if err != nil && err != collection.ErrKvTableAlreadyPresent {
			cancel()
			return
		}
		err = kvStore.OpenKVTable(col.Name, c.podInfo.GetPodPassword())
		if err != nil {
			cancel()
			return
		}
	}()

	wg.Wait()
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
	return nil
}

func (c *Client) DeleteCollection(collection string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sessionId == "" {
		return dfs.ErrUserNotLoggedIn
	}
	if c.podInfo == nil {
		return dfs.ErrPodNotOpen
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := c.api.DocDelete(c.sessionId, c.pod, collection)
		if err != nil {
			c.logger.Errorf("delete collection failed :%s\n", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		kvStore := c.podInfo.GetKVStore()

		err := kvStore.DeleteKVTable(collection, c.podInfo.GetPodPassword())
		if err != nil {
			c.logger.Errorf("delete kv table failed :%s\n", err.Error())
		}
	}()
	wg.Wait()
	c.hnswLock.Lock()
	delete(c.indices, collection)
	c.hnswLock.Unlock()
	return nil
}

func (c *Client) AddDocuments(collection string, propertiesToIndex []string, documents ...*Document) error {
	// check if kv and doc table is open or not
	kvStore := c.podInfo.GetKVStore()
	_, err := kvStore.KVCount(collection)
	if err != nil {
		err = kvStore.OpenKVTable(collection, c.podInfo.GetPodPassword())
		if err != nil {
			return err
		}
	}
	docIsOpen, err := c.api.IsDBOpened(c.sessionId, c.pod, collection)
	if err != nil {
		return err
	}
	if !docIsOpen {
		vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
			// check if the document is in the cache
			if v, ok := c.documentCache.Get(fmt.Sprintf("%s/%s/%d", c.pod, collection, id)); ok {
				return v.([]float32), nil
			}
			expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
			docs, err := c.api.DocFind(c.sessionId, c.pod, collection, expr, 1)
			if err != nil {
				return nil, err
			}
			doc := docs[0]
			data := map[string]interface{}{}
			err = json.Unmarshal(doc, &data)
			if err != nil {
				return nil, err
			}
			vector, err := convertToFloat32Slice(data["vector"])
			if err != nil {
				return nil, err
			}
			c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, collection, id), vector)
			return vector, err
		}

		makeCL := h.MakeNoopCommitLogger
		index, err := h.New(h.Config{
			RootPath:              "not-used",
			ID:                    "not-used",
			MakeCommitLoggerThunk: makeCL,
			DistanceProvider:      distancer.NewCosineDistanceProvider(),
			VectorForIDThunk:      vectorForID,
			ClassName:             collection,
		}, h.UserConfig{
			MaxConnections: 30,
			EFConstruction: 60,
		}, kvStore)
		if err != nil {
			return err
		}

		c.hnswLock.Lock()
		c.indices[collection] = index
		c.hnswLock.Unlock()
		err = c.api.DocOpen(c.sessionId, c.pod, collection)
		if err != nil {
			return err
		}
	}

	c.hnswLock.Lock()
	index := c.indices[collection]
	c.hnswLock.Unlock()
	for id, doc := range documents {
		// vectorize the properties
		// add the vector in the properties before adding the document in the collection
		vectorData := ""
		for _, property := range propertiesToIndex {
			dt, ok := doc.Properties[property]
			if ok {
				vectorData += dt.(string) + " "
			}
		}
		doc.Properties["id"] = doc.ID

		if vectorData != "" {
			vector, err := c.lookup.Corpi([]string{vectorData})
			if err != nil {
				c.logger.Errorf("corpi failed :%s\n", err.Error())
				continue
			}

			doc.Properties["vector"] = vector.ToArray()

			count, err := c.api.KVCount(c.sessionId, c.pod, collection)
			if err != nil {
				return err
			}
			indexId := count.Count + uint64(id)

			doc.Properties[hnswIndexName] = indexId

			err = index.Add(indexId, vector.ToArray())
			if err != nil {
				c.logger.Errorf("index.Add failed :%s\n", err.Error())
				continue
			}

			c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, collection, indexId), vector.ToArray())
		}

		data, err := json.Marshal(doc.Properties)
		if err != nil {
			c.logger.Errorf("marshal document failed :%s\n", err.Error())
			continue
		}

		err = c.api.DocPut(c.sessionId, c.pod, collection, data)
		if err != nil {
			c.logger.Errorf("DocPut failed :%s\n", err.Error())
			continue
		}
	}
	return nil
}

func (c *Client) GetNearDocuments(collection, text string, distance float32) ([][]byte, []float32, error) {
	kvStore := c.podInfo.GetKVStore()
	_, err := kvStore.KVCount(collection)
	if err != nil {
		err = kvStore.OpenKVTable(collection, c.podInfo.GetPodPassword())
		if err != nil {
			return nil, nil, err
		}
	}
	docIsOpen, err := c.api.IsDBOpened(c.sessionId, c.pod, collection)
	if err != nil {
		return nil, nil, err
	}
	if !docIsOpen {
		vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
			if v, ok := c.documentCache.Get(fmt.Sprintf("%s/%s/%d", c.pod, collection, id)); ok {
				return v.([]float32), nil
			}
			expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
			docs, err := c.api.DocFind(c.sessionId, c.pod, collection, expr, 1)
			if err != nil {
				return nil, err
			}
			doc := docs[0]
			data := map[string]interface{}{}
			err = json.Unmarshal(doc, &data)
			if err != nil {
				return nil, err
			}
			vector, err := convertToFloat32Slice(data["vector"])
			if err != nil {
				return nil, err
			}
			c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, collection, id), vector)
			return vector, nil
		}

		makeCL := h.MakeNoopCommitLogger
		index, err := h.New(h.Config{
			RootPath:              "not-used",
			ID:                    "not-used",
			MakeCommitLoggerThunk: makeCL,
			DistanceProvider:      distancer.NewCosineDistanceProvider(),
			VectorForIDThunk:      vectorForID,
			ClassName:             collection,
		}, h.UserConfig{
			MaxConnections: 30,
			EFConstruction: 60,
		}, kvStore)
		if err != nil {
			return nil, nil, err
		}

		c.hnswLock.Lock()
		c.indices[collection] = index
		c.hnswLock.Unlock()
		err = c.api.DocOpen(c.sessionId, c.pod, collection)
		if err != nil {
			return nil, nil, err
		}
	}

	vector, err := c.lookup.Corpi([]string{text})
	if err != nil {
		return nil, nil, err
	}
	c.hnswLock.Lock()
	index := c.indices[collection]
	c.hnswLock.Unlock()
	ids, dists, err := index.KnnSearchByVectorMaxDist(vector.ToArray(), distance, 800, nil)
	if err != nil {
		return nil, nil, err
	}

	documents := make([][]byte, len(ids))
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id uint64) {
			defer wg.Done()
			expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
			docs, err := c.api.DocFind(c.sessionId, c.pod, collection, expr, 1)
			if err != nil {
				errCh <- err
				return
			}
			documents[i] = docs[0]
		}(i, id)
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return nil, nil, err
	}

	return documents, dists, nil
}

func (c *Client) GetDocument(collection, property, value string) ([]byte, error) {
	docIsOpen, err := c.api.IsDBOpened(c.sessionId, c.pod, collection)
	if err != nil {
		return nil, err
	}
	if !docIsOpen {
		err = c.api.DocOpen(c.sessionId, c.pod, collection)
		if err != nil {
			return nil, err
		}
	}

	expr := fmt.Sprintf("%s=%s", property, value)
	docs, err := c.api.DocFind(c.sessionId, c.pod, collection, expr, 1)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, fmt.Errorf("document not found")
	}
	return docs[0], nil
}

func convertToFloat32Slice(i interface{}) ([]float32, error) {
	// Check if the underlying value is a slice
	if slice, ok := i.([]interface{}); ok {
		// Create a new slice to store the converted values
		result := make([]float32, len(slice))

		// Convert each element to float32
		for i, v := range slice {
			// Perform a type assertion to ensure the underlying value is float32
			if f, ok := v.(float64); ok {
				result[i] = float32(f)
			} else {
				return nil, fmt.Errorf("value at index %d is not float32", i)
			}
		}

		return result, nil
	}

	return nil, fmt.Errorf("value is not a slice")
}
