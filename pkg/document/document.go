package document

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

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
)

const (
	errLevel   = logrus.ErrorLevel
	debugLevel = logrus.DebugLevel

	hnswIndexName = "hnswId"
	namespace     = "fave_"
)

// Config for fairOS-dfs
type Config struct {
	Verbose       bool
	VectorizerUrl string
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

	if config.VectorizerUrl == "" {
		logger.Errorf("VECTORIZER_URL environment variable is not set")
	}

	// leveldb lookuper
	lkup, err := rest.NewVectorizer(config.VectorizerUrl)
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

	//docs, _ := c.api.DocList(c.sessionId, c.pod)
	//fmt.Println("docs", docs)
	//for _, doc := range docs {
	//	err = c.api.DocDelete(c.sessionId, c.pod, doc.Name)
	//	fmt.Println("delete doc", doc.Name, err)
	//}
	//
	//kvs, _ := c.api.KVList(c.sessionId, c.pod)
	//fmt.Println("kvs", kvs)
	//for kv, _ := range kvs {
	//	err = c.api.KVDelete(c.sessionId, c.pod, kv)
	//	fmt.Println("delete kv", kv, err)
	//}
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

	namespacedCollection := namespace + col.Name

	vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
		// check if the document is in the cache
		if v, ok := c.documentCache.Get(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, id)); ok {
			return v.([]float32), nil
		}
		expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
		docs, err := c.api.DocFind(c.sessionId, c.pod, namespacedCollection, expr, 1)
		if err != nil {
			return nil, err
		}
		if len(docs) == 0 {
			return nil, fmt.Errorf("document not found")
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
		c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, id), vector)
		return vector, err
	}
	kvStore := c.podInfo.GetKVStore()
	err := c.api.DocCreate(c.sessionId, c.pod, namespacedCollection, col.Indexes, true)
	if err != nil && err != collection.ErrDocumentDBAlreadyPresent && err != collection.ErrDocumentDBAlreadyOpened {
		return err
	}

	err = kvStore.CreateKVTable(namespacedCollection, c.podInfo.GetPodPassword(), collection.StringIndex)
	if err != nil && err != collection.ErrKvTableAlreadyPresent {
		return err
	}

	err = c.api.DocOpen(c.sessionId, c.pod, namespacedCollection)
	if err != nil && err != collection.ErrDocumentDBAlreadyOpened {
		return err
	}

	err = kvStore.OpenKVTable(namespacedCollection, c.podInfo.GetPodPassword())
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
		ClassName:             namespacedCollection,
	}, h.UserConfig{
		MaxConnections: 30,
		EFConstruction: 60,
	}, kvStore)
	if err != nil {
		return err
	}

	c.hnswLock.Lock()
	c.indices[namespacedCollection] = index
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

	namespacedCollection := namespace + collection

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := c.api.DocDelete(c.sessionId, c.pod, namespacedCollection)
		if err != nil {
			c.logger.Errorf("delete collection failed :%s\n", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		kvStore := c.podInfo.GetKVStore()

		err := kvStore.DeleteKVTable(namespacedCollection, c.podInfo.GetPodPassword())
		if err != nil {
			c.logger.Errorf("delete kv table failed :%s\n", err.Error())
		}
	}()
	wg.Wait()
	c.hnswLock.Lock()
	delete(c.indices, namespacedCollection)
	c.hnswLock.Unlock()
	return nil
}

func (c *Client) GetCollections() ([]*Collection, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.sessionId == "" {
		return nil, dfs.ErrUserNotLoggedIn
	}
	if c.podInfo == nil {
		return nil, dfs.ErrPodNotOpen
	}

	docs, err := c.api.DocList(c.sessionId, c.pod)
	if err != nil {
		return nil, err
	}
	collections := []*Collection{}
	for _, doc := range docs {
		if strings.HasPrefix(doc.Name, namespace) {
			indexes := map[string]collection.IndexType{}
			for _, index := range doc.SimpleIndexes {
				indexes[index.FieldName] = index.FieldType
			}
			for _, index := range doc.ListIndexes {
				indexes[index.FieldName] = index.FieldType
			}
			for _, index := range doc.MapIndexes {
				indexes[index.FieldName] = index.FieldType
			}
			for _, index := range doc.VectorIndexes {
				indexes[index.FieldName] = index.FieldType
			}
			collections = append(collections, &Collection{
				Name:    strings.TrimPrefix(doc.Name, namespace),
				Indexes: indexes,
			})
		}
	}
	return collections, nil
}

func (c *Client) AddDocuments(collection string, propertiesToIndex []string, documents ...*Document) error {
	namespacedCollection := namespace + collection

	// check if kv and doc table is open or not
	kvStore := c.podInfo.GetKVStore()
	_, err := kvStore.KVCount(namespacedCollection)
	if err != nil {
		err = kvStore.OpenKVTable(namespacedCollection, c.podInfo.GetPodPassword())
		if err != nil {
			return err
		}
	}
	docIsOpen, err := c.api.IsDBOpened(c.sessionId, c.pod, namespacedCollection)
	if err != nil {
		return err
	}
	if !docIsOpen {
		vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
			// check if the document is in the cache
			if v, ok := c.documentCache.Get(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, id)); ok {
				return v.([]float32), nil
			}
			expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
			docs, err := c.api.DocFind(c.sessionId, c.pod, namespacedCollection, expr, 1)
			if err != nil {
				return nil, err
			}
			if len(docs) == 0 {
				return nil, fmt.Errorf("document not found")
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
			c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, id), vector)
			return vector, err
		}

		makeCL := h.MakeNoopCommitLogger
		index, err := h.New(h.Config{
			RootPath:              "not-used",
			ID:                    "not-used",
			MakeCommitLoggerThunk: makeCL,
			DistanceProvider:      distancer.NewCosineDistanceProvider(),
			VectorForIDThunk:      vectorForID,
			ClassName:             namespacedCollection,
		}, h.UserConfig{
			MaxConnections: 30,
			EFConstruction: 60,
		}, kvStore)
		if err != nil {
			return err
		}

		c.hnswLock.Lock()
		c.indices[namespacedCollection] = index
		c.hnswLock.Unlock()
		err = c.api.DocOpen(c.sessionId, c.pod, namespacedCollection)
		if err != nil {
			return err
		}
	}

	c.hnswLock.Lock()
	index := c.indices[namespacedCollection]
	c.hnswLock.Unlock()
	count, err := c.api.KVCount(c.sessionId, c.pod, namespacedCollection)
	if err != nil {
		return err
	}

	indexId := count.Count
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

			doc.Properties[hnswIndexName] = indexId

			err = index.Add(indexId, vector.ToArray())
			if err != nil {
				c.logger.Errorf("index.Add failed :%s\n", err.Error())
				continue
			}
			c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, indexId), vector.ToArray())

			indexId++
		} else {
			doc.Properties[hnswIndexName] = -1
		}

		data, err := json.Marshal(doc.Properties)
		if err != nil {
			c.logger.Errorf("marshal document failed :%s\n", err.Error())
			continue
		}

		err = c.api.DocPut(c.sessionId, c.pod, namespacedCollection, data)
		if err != nil {
			c.logger.Errorf("DocPut failed :%s, %+v\n", err.Error(), doc.Properties)
			return err
		}
		fmt.Println("added document", id)
	}

	return index.Flush()
}

func (c *Client) GetNearDocuments(collection, text string, distance float32, limit int) ([][]byte, []float32, error) {
	namespacedCollection := namespace + collection

	kvStore := c.podInfo.GetKVStore()
	_, err := kvStore.KVCount(namespacedCollection)
	if err != nil {
		err = kvStore.OpenKVTable(namespacedCollection, c.podInfo.GetPodPassword())
		if err != nil {
			return nil, nil, err
		}
	}
	docIsOpen, err := c.api.IsDBOpened(c.sessionId, c.pod, namespacedCollection)
	if err != nil {
		return nil, nil, err
	}
	if !docIsOpen {
		vectorForID := func(ctx context.Context, id uint64) ([]float32, error) {
			if v, ok := c.documentCache.Get(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, id)); ok {
				return v.([]float32), nil
			}
			expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
			docs, err := c.api.DocFind(c.sessionId, c.pod, namespacedCollection, expr, 1)
			if err != nil {
				return nil, err
			}
			if len(docs) == 0 {
				return nil, fmt.Errorf("document not found")
			}
			doc := docs[0]
			data := map[string]interface{}{}
			err = json.Unmarshal(doc, &data)
			if err != nil {
				return nil, err
			}
			if data["vector"] == nil {
				return nil, fmt.Errorf("vector is nil")
			}
			vector, err := convertToFloat32Slice(data["vector"])
			if err != nil {
				return nil, err
			}
			c.documentCache.Add(fmt.Sprintf("%s/%s/%d", c.pod, namespacedCollection, id), vector)
			return vector, nil
		}

		makeCL := h.MakeNoopCommitLogger
		index, err := h.New(h.Config{
			RootPath:              "not-used",
			ID:                    "not-used",
			MakeCommitLoggerThunk: makeCL,
			DistanceProvider:      distancer.NewCosineDistanceProvider(),
			VectorForIDThunk:      vectorForID,
			ClassName:             namespacedCollection,
		}, h.UserConfig{
			MaxConnections: 30,
			EFConstruction: 60,
		}, kvStore)
		if err != nil {
			return nil, nil, err
		}

		c.hnswLock.Lock()
		c.indices[namespacedCollection] = index
		c.hnswLock.Unlock()
		err = c.api.DocOpen(c.sessionId, c.pod, namespacedCollection)
		if err != nil {
			return nil, nil, err
		}
	}
	vector, err := c.lookup.Corpi([]string{text})
	if err != nil {
		return nil, nil, err
	}
	c.hnswLock.Lock()
	index := c.indices[namespacedCollection]
	c.hnswLock.Unlock()
	err = index.LoadEntrypoint()
	if err != nil {
		return nil, nil, err
	}
	ids, dists, err := index.KnnSearchByVectorMaxDist(vector.ToArray(), distance, 800, nil)
	if err != nil {
		return nil, nil, err
	}
	if limit != 0 && len(ids) > limit {
		ids = ids[:limit]
		dists = dists[:limit]
	}

	documents := make([][]byte, len(ids))
	wg := sync.WaitGroup{}
	errCh := make(chan error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(i int, id uint64) {
			defer wg.Done()
			expr := fmt.Sprintf("%s=%d", hnswIndexName, id)
			docs, err := c.api.DocFind(c.sessionId, c.pod, namespacedCollection, expr, 1)
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
	namespacedCollection := namespace + collection
	docIsOpen, err := c.api.IsDBOpened(c.sessionId, c.pod, namespacedCollection)
	if err != nil {
		return nil, err
	}
	if !docIsOpen {
		err = c.api.DocOpen(c.sessionId, c.pod, namespacedCollection)
		if err != nil {
			return nil, err
		}
	}

	expr := ""
	if property != "" && value != "" {
		expr = fmt.Sprintf("%s=%s", property, value)
	}
	docs, err := c.api.DocFind(c.sessionId, c.pod, namespacedCollection, expr, 1)
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
