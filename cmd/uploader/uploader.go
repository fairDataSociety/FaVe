package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fairdatasociety/fairOS-dfs/pkg/collection"
	"github.com/fairdatasociety/fairOS-dfs/pkg/contracts"
	"github.com/fairdatasociety/fairOS-dfs/pkg/dfs"
	"github.com/fairdatasociety/fairOS-dfs/pkg/logging"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

const (
	errLevel   = logrus.ErrorLevel
	debugLevel = logrus.DebugLevel
)

var (
	api          *dfs.API
	sessionId    string
	insertedHook = func(word string) {}
)

type Options struct {
	Verbose        bool                 `short:"l" long:"verbose" description:"Show fairos and other debug logs"`
	VectorCSVPath  string               `short:"v" long:"vector-csv-path" description:"Path to the embedding file " required:"true"`
	EnsRPC         string               `short:"r" long:"rpc-endpoint" description:"RPC endpoint for ENS authentication" required:"true"`
	BeeEndpoint    string               `short:"b" long:"bee-api-endpoint" description:"Bee api endpoint" required:"true"`
	StampID        string               `short:"s" long:"stamp" description:"stamp id" required:"true"`
	FairOSUser     string               `short:"u" long:"username" description:"FDP portable username" required:"true"`
	FairOSPassword string               `short:"p" long:"password" description:"account password" required:"true"`
	Pod            string               `short:"d" long:"pod" description:"pod name of the kv store" required:"true"`
	KVStore        string               `short:"k" long:"kv-store" description:"kv store name" required:"true"`
	KVIndexType    collection.IndexType `short:"x" description:"index type for the values in the kv store" default:"2"` // collection.StringIndex in fairos
}

func main() {
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	// Parse the command line flags.
	// If the user asked for help, don't print the error.
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	// Set the log level
	level := errLevel
	if options.Verbose {
		level = debugLevel
	}
	logger := logging.New(os.Stdout, level)

	// start spinner
	s := spinner.New(spinner.CharSets[70], 100*time.Millisecond)
	s.Start()
	insertedHook = func(word string) {
		s.Suffix = fmt.Sprintf("%q inserted", word)
	}
	defer s.Stop()

	// Get config for fairos
	// TODO This will call contracts.TestConfig() eventually
	config, _ := contracts.TestnetConfig()
	config.ProviderBackend = options.EnsRPC

	// init fairos
	var err error
	api, err = dfs.NewDfsAPI(
		context.TODO(),
		options.BeeEndpoint,
		options.StampID,
		config,
		nil,
		logger,
	)
	if err != nil {
		logger.Errorf("new fairos api failed :%s\n", err.Error())
		return
	}
	s.Suffix = "fairOS initialised"

	// login to fairos
	lr, err := api.LoginUserV2(options.FairOSUser, options.FairOSPassword, "")
	if err != nil {
		logger.Errorf("fairos login failed: %s", err.Error())
		return
	}
	s.Suffix = fmt.Sprintf("%s logged in", options.FairOSUser)
	ui := lr.UserInfo
	sessionId = ui.GetSessionId()

	// create pod if it does not exist
	// it is recommended to use a pod that does not have any files or very few files.
	// That will make sure of less time consumption for opening pod
	if !api.IsPodExist(options.Pod, sessionId) {
		logger.Debugf("creating pod %s\n", options.Pod)
		_, err = api.CreatePod(options.Pod, sessionId)
		if err != nil {
			logger.Errorf("failed to create pod: %s\n", err.Error())
			return
		}
	}
	_, err = api.OpenPod(options.Pod, sessionId)
	if err != nil {
		logger.Errorf("failed to open pod: %s\n", err.Error())
		return
	}
	s.Suffix = fmt.Sprintf("pod %s opened", options.Pod)

	// create kv table if it doesn't exist
	err = api.KVCreate(sessionId, options.Pod, options.KVStore, options.KVIndexType)
	if err != nil && err != collection.ErrKvTableAlreadyPresent {
		logger.Errorf("failed to create kv store: %s\n", err.Error())
		return
	}
	// open kv table
	err = api.KVOpen(sessionId, options.Pod, options.KVStore)
	if err != nil {
		logger.Errorf("failed to open kv store: %s\n", err.Error())
		return
	}
	s.Suffix = fmt.Sprintf("kv table %s opened", options.KVStore)

	// get kv batch for inserting vectors in batch
	batch, err := api.KVBatch(sessionId, options.Pod, options.KVStore, []string{})
	if err != nil {
		logger.Errorf("failed to create kv batch action: %s\n", err.Error())
		return
	}

	// open vectors file
	file, err := os.Open(options.VectorCSVPath)
	if err != nil {
		logger.Errorf("failed to open vectors file: %s\n", err.Error())
		return
	}
	defer file.Close()

	var vectorLength = -1
	var count = 0

	// read vectors file line by line and insert in kv store
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count += 1
		parts := strings.Split(scanner.Text(), " ")

		word := parts[0]
		if vectorLength == -1 {
			vectorLength = len(parts) - 1
		}

		if vectorLength != len(parts)-1 {
			logger.Error("vector length mismatch. word will be skipped.")
			continue
		}

		// pre-allocate a vector for speed.
		vector := make([]float32, vectorLength)

		for i := 1; i <= vectorLength; i++ {
			float, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				logger.Errorf("failed to parse vector to float: %s\n", err.Error())
				return
			}
			vector[i-1] = float32(float)
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(vector); err != nil {
			logger.Errorf("failed to encode vector: %s\n", err.Error())
			return
		}
		err = batch.Put(word, buf.Bytes(), true, false)
		if err != nil {
			logger.Errorf("could not put value for %s: %s\n", word, err.Error())
		} else {
			insertedHook(word)
		}

	}
	_, err = batch.Write("")
	if err != nil {
		logger.Errorf("failed to write kv batch: %s\n", err.Error())
		return
	}
	fmt.Println("number of words ", count)
}
