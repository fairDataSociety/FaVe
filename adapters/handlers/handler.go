package handlers

import (
	"context"
	"os"
	"time"

	"github.com/fairDataSociety/FaVe/pkg/document"
	"github.com/fairDataSociety/FaVe/restapi/operations"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/go-openapi/runtime/middleware"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	doc *document.Client
}

type HandlerConfig struct {
	Verbose         bool
	BeeAPI          string
	RPCEndpoint     string
	StampId         string
	GloveLevelDBUrl string
	Username        string
	Password        string
	Pod             string
}

func NewHandler(ctx context.Context, config *HandlerConfig) (*Handler, error) {
	FromEnv(config)
	SetDefaults(config)
	documentConfig := document.Config{
		Verbose:       config.Verbose,
		VectorizerUrl: config.GloveLevelDBUrl,
	}
	ensConf, _ := contracts.TestnetConfig(contracts.Sepolia)
	//ensConf, _ := contracts.PlayConfig()
	ensConf.ProviderBackend = config.RPCEndpoint
	level := logrus.ErrorLevel
	if config.Verbose {
		level = logrus.DebugLevel
	}
	logger := logging.New(os.Stdout, level)

	dfsOpts := &dfs.Options{
		BeeApiEndpoint:     config.BeeAPI,
		Stamp:              config.StampId,
		EnsConfig:          ensConf,
		SubscriptionConfig: nil,
		Logger:             logger,
		FeedCacheSize:      10000,
		FeedCacheTTL:       time.Minute * 60,
	}

	dfsApi, err := dfs.NewDfsAPI(ctx, dfsOpts)
	if err != nil {
		return nil, err
	}
	d, err := document.New(documentConfig, dfsApi)
	if err != nil {
		return nil, err
	}

	err = d.Login(config.Username, config.Password)
	if err != nil {
		return nil, err
	}
	err = d.OpenPod(config.Pod)
	if err != nil {
		return nil, err
	}

	return &Handler{
		doc: d,
	}, nil
}

func (s *Handler) FaveRootHandler(_ operations.FaveRootParams) middleware.Responder {
	return operations.NewFaveRootOK()
}
