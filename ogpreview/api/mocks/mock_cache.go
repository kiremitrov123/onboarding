// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"github.com/kiremitrov123/onboarding/src/ogpreview/api"
	"github.com/kiremitrov123/onboarding/src/ogpreview/model"
	"sync"
)

// Ensure, that CacheMock does implement api.Cache.
// If this is not the case, regenerate this file with moq.
var _ api.Cache = &CacheMock{}

// CacheMock is a mock implementation of api.Cache.
//
//	func TestSomethingThatUsesCache(t *testing.T) {
//
//		// make and configure a mocked api.Cache
//		mockedCache := &CacheMock{
//			GetTagsFunc: func(url string) (*model.OGTags, error) {
//				panic("mock out the GetTags method")
//			},
//			SetTagsFunc: func(url string, tags model.OGTags) error {
//				panic("mock out the SetTags method")
//			},
//		}
//
//		// use mockedCache in code that requires api.Cache
//		// and then make assertions.
//
//	}
type CacheMock struct {
	// GetTagsFunc mocks the GetTags method.
	GetTagsFunc func(url string) (*model.OGTags, error)

	// SetTagsFunc mocks the SetTags method.
	SetTagsFunc func(url string, tags model.OGTags) error

	// calls tracks calls to the methods.
	calls struct {
		// GetTags holds details about calls to the GetTags method.
		GetTags []struct {
			// URL is the url argument value.
			URL string
		}
		// SetTags holds details about calls to the SetTags method.
		SetTags []struct {
			// URL is the url argument value.
			URL string
			// Tags is the tags argument value.
			Tags model.OGTags
		}
	}
	lockGetTags sync.RWMutex
	lockSetTags sync.RWMutex
}

// GetTags calls GetTagsFunc.
func (mock *CacheMock) GetTags(url string) (*model.OGTags, error) {
	if mock.GetTagsFunc == nil {
		panic("CacheMock.GetTagsFunc: method is nil but Cache.GetTags was just called")
	}
	callInfo := struct {
		URL string
	}{
		URL: url,
	}
	mock.lockGetTags.Lock()
	mock.calls.GetTags = append(mock.calls.GetTags, callInfo)
	mock.lockGetTags.Unlock()
	return mock.GetTagsFunc(url)
}

// GetTagsCalls gets all the calls that were made to GetTags.
// Check the length with:
//
//	len(mockedCache.GetTagsCalls())
func (mock *CacheMock) GetTagsCalls() []struct {
	URL string
} {
	var calls []struct {
		URL string
	}
	mock.lockGetTags.RLock()
	calls = mock.calls.GetTags
	mock.lockGetTags.RUnlock()
	return calls
}

// SetTags calls SetTagsFunc.
func (mock *CacheMock) SetTags(url string, tags model.OGTags) error {
	if mock.SetTagsFunc == nil {
		panic("CacheMock.SetTagsFunc: method is nil but Cache.SetTags was just called")
	}
	callInfo := struct {
		URL  string
		Tags model.OGTags
	}{
		URL:  url,
		Tags: tags,
	}
	mock.lockSetTags.Lock()
	mock.calls.SetTags = append(mock.calls.SetTags, callInfo)
	mock.lockSetTags.Unlock()
	return mock.SetTagsFunc(url, tags)
}

// SetTagsCalls gets all the calls that were made to SetTags.
// Check the length with:
//
//	len(mockedCache.SetTagsCalls())
func (mock *CacheMock) SetTagsCalls() []struct {
	URL  string
	Tags model.OGTags
} {
	var calls []struct {
		URL  string
		Tags model.OGTags
	}
	mock.lockSetTags.RLock()
	calls = mock.calls.SetTags
	mock.lockSetTags.RUnlock()
	return calls
}
