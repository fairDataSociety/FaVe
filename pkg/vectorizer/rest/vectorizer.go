package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	vctrzr "github.com/fairDataSociety/FaVe/pkg/vectorizer"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

var (
	healthCheckPath = "/health"
	corpiPath       = "/vectorize"

	ErrServiceUnavailable = errors.New("service unavailable")
)

type Vectorizer struct {
	url string
}

func NewVectorizer(url string) (*Vectorizer, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	healthURL := fmt.Sprintf("%s/%s", url, healthCheckPath)
	resp, err := client.Get(healthURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrServiceUnavailable
	}

	return &Vectorizer{url: url}, nil
}

func (v *Vectorizer) Corpi(text []string) (*vctrzr.Vector, error) {
	data := map[string][]string{
		"query": text,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s%s", v.url, corpiPath), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var result map[string][]float32
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	vector := vctrzr.NewVector(result["vector"])
	return &vector, nil
}
