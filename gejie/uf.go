package gejie

import (
	"fmt"
	"sync"
)

type URLStatus int

const (
	StatusPending URLStatus = iota
	StatusVisited
	StatusFailed
)

// URLFrontierInterface defines the behavior of a URL frontier
type URLFrontierInterface interface {
	BulkAdd(urls []string)
	Add(url string)
	MarkVisited(url string)
	MarkFailed(url string)
	GetNext() (string, bool)
	Count() int
	CountRemaining() int
}

// URLFrontier implements URLFrontierInterface
type URLFrontier struct {
	urls map[string]URLStatus
	mu   sync.Mutex
}

func NewURLFrontier() URLFrontierInterface {
	return &URLFrontier{
		urls: make(map[string]URLStatus),
	}
}

func (f *URLFrontier) Count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.urls)
}

// BulkAdd adds multiple URLs to the frontier
func (f *URLFrontier) BulkAdd(urls []string) {
	for _, url := range urls {
		f.Add(url)
	}
}

// remaining methods stay the same, just ensure URLFrontier implements URLFrontierInterface
func (f *URLFrontier) Add(url string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.urls[url]; !exists {
		f.urls[url] = StatusPending
	} else {
		fmt.Printf("url already exists (%s)", url)
	}
}

func (f *URLFrontier) MarkVisited(url string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.urls[url] = StatusVisited
}

func (f *URLFrontier) MarkFailed(url string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.urls[url] = StatusFailed
}

func (f *URLFrontier) GetNext() (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for url, status := range f.urls {
		if status == StatusPending {
			return url, true
		}
	}
	return "", false
}

// CountRemaining returns the number of pending URLs that have not been visited
func (f *URLFrontier) CountRemaining() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, status := range f.urls {
		if status == StatusPending {
			count++
		}
	}
	return count
}
