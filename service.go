package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/sync/singleflight"
	"golang.org/x/text/language"
	"sync"
	"time"
)

// Service is a Translator user.
type Service struct {
	translator *TranslatorService
}

func NewService() *Service {
	t := newRandomTranslator(
		100*time.Millisecond,
		500*time.Millisecond,
		0.1,
	)

	return &Service{
		translator: NewTranslatorService(t),
	}
}

//Created a Wrapper TranslatorService around Service so I dont have to change the main.go
//but implement the features required like caching, retrying

type TranslatorService struct {
	sync.Mutex                   //to avoid race conditions across go routines
	cache      map[string]string //a simple in-memory cache
	Translator                   //translation service interface
}

func NewTranslatorService(t Translator) *TranslatorService {
	return &TranslatorService{
		cache:      make(map[string]string),
		Translator: t,
	}
}

const (
	MaxRetries  = 5
	BackoffTime = 2
)

var requestGroup = singleflight.Group{}

// Translate is a wrapper method around translation service
//this implements caching and retry mechanism
//this is created just to avoid changing the main.go file
func (t *TranslatorService) Translate(ctx context.Context, from, to language.Tag, data string) (string, error) {
	key := getTranslateKey(from.String(), to.String(), data)

	//check if same query exist in cache
	// key is a md5(from,to,data)
	if out, ok := t.cache[key]; ok {
		return out, nil
	}

	retryCount := 0

	for {
		//call translate service if query is not found in cache
		out, err, _ := requestGroup.Do(key, func() (interface{}, error) {
			return t.Translator.Translate(ctx, from, to, data)
		})

		if err != nil {
			//retry for max number of times
			if retryCount < MaxRetries {
				retryCount++
				//exponential backoff before retrying
				time.Sleep(time.Duration(retryCount*BackoffTime) * time.Second)
				continue
			}
			return "", err
		}

		//update cache once data is successfully translated
		t.Lock()
		translatedString := out.(string)
		t.cache[key] = translatedString // we know that out is translated string
		t.Unlock()

		return translatedString, nil
	}

}

// generates a md5 hash with from,to language and data string
// this md5 hash is used as a key for caching
func getTranslateKey(from string, to string, data string) string {
	md5KeyHash := md5.Sum([]byte(fmt.Sprintf("%s-%s-%s", from, to, data)))
	return hex.EncodeToString(md5KeyHash[:])
}
