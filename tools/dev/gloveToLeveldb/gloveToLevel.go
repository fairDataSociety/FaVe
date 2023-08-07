package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/syndtr/goleveldb/leveldb"
)

type Options struct {
	VectorCSVPath string `short:"c" long:"vector-csv-path" description:"Path to the output file of Glove" required:"true"`
	TempDBPath    string `short:"t" long:"temp-db-path" description:"Location for the temporary database" default:".tmp_import"`
}

type WordVectorInfo struct {
	numberOfWords int
	vectorWidth   int
}

func main() {
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	db, err := leveldb.OpenFile(options.TempDBPath, nil)
	defer db.Close()

	if err != nil {
		log.Fatalf("Could not open temporary database file %+v", err)
	}

	file, err := os.Open(options.VectorCSVPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.Print("Processing and ordering raw trained data")
	info := readGloveVectorsFromFileAndInsertIntoLevelDB(db, file)

	fmt.Println("Number of words: ", info.numberOfWords, "with vector width: ", info.vectorWidth)
}

// read word vectors, insert them into level db, also return the dimension of the vectors.
func readGloveVectorsFromFileAndInsertIntoLevelDB(db *leveldb.DB, file *os.File) WordVectorInfo {
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
				log.Fatal("Error parsing float")
			}

			vector[i-1] = float32(float)
		}

		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(vector); err != nil {
			log.Fatal("Could not encode vector for temp db storage")
		}

		db.Put([]byte(word), buf.Bytes(), nil)
	}

	return WordVectorInfo{numberOfWords: nrWords, vectorWidth: vectorLength}
}
