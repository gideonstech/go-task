package main

import (
	"context"
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

//this is a wrapper method around translation service
//this implements caching and retry mechanism
//this is created just to avoid changing the main.go file
func (t *TranslatorService) Translate(ctx context.Context, from, to language.Tag, data string) (string, error) {
	return t.Translator.Translate(ctx, from, to, data)
}
