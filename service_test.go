package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
	"sync"
	"testing"
	"time"
)

func TestTranslatorService_Translate(t *testing.T) {
	testCases := []struct {
		ts             *TranslatorService
		errorExpected  bool
		expectedOutput string
		errorString    string
	}{
		{
			ts: NewTranslatorService(newRandomTranslator(
				100*time.Millisecond,
				500*time.Millisecond,
				0.1,
			)),
			errorExpected:  false,
			expectedOutput: "af -> en : testdata -> 6129484611666145821",
		}, {
			ts: NewTranslatorService(newRandomTranslator(
				100*time.Millisecond,
				500*time.Millisecond,
				1.0,
			)),
			errorExpected: true,
			errorString:   "translation failed",
		},
	}

	for _, testCase := range testCases {
		timeNow := time.Now()
		actualOutput, err := testCase.ts.Translate(context.Background(), language.Afrikaans, language.English, "testdata")
		if err != nil {
			if !testCase.errorExpected {
				t.Fatal("no errors expected, but got ", err.Error())
			} else {
				assert.Equal(t, err.Error(), testCase.errorString)
				//if there is an error it rertries for max of 5 times with exponential backoff of 2*retry_count
				//so min delay for 5 retries will be 30, so am asserting that
				if time.Since(timeNow).Seconds() < 30.0 {
					t.Fatal("expected delay of min 30 sec")
				}
			}
		}
		if !testCase.errorExpected {
			assert.Equal(t, actualOutput, testCase.expectedOutput)
			//	check the output is cached
			hashKey := getTranslateKey(language.Afrikaans.String(), language.English.String(), "testdata")
			if _, ok := testCase.ts.cache[hashKey]; !ok {
				t.Fatal("the output is expected to be cached")
			}
		}
	}
}

// i found no way to assert the deduplication logic, since i cant change the method signature
// for translate to return if the requests are shared. I validated the deduplication by simply adding
// a log in the Translate method in translator.go (I have removed it now). The Translate method in translator.go
// is called for only 5 times even when all 5 requests will fail and will be retried for 5 times
// (so in total 25 times the method should be called with deduplication only 5 times the Translate method is invoked)
func TestTranslatorService_DeDuplicateTranslate(t *testing.T) {
	ts := NewTranslatorService(newRandomTranslator(
		100*time.Millisecond,
		500*time.Millisecond,
		1.0))

	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			ts.Translate(context.Background(), language.Afrikaans, language.English, "testdata")
			wg.Done()
		}()
	}
	wg.Wait()
}
