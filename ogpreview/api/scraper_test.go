package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kiremitrov123/onboarding/src/ogpreview/api"
	"github.com/stretchr/testify/assert"
)

func TestFetchOGTags(t *testing.T) {
	// Mock a page with OG tags
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
				<head>
					<meta property="og:title" content="Test Title" />
					<meta property="og:description" content="Test Description" />
					<meta property="og:image" content="http://example.com/image.jpg" />
				</head>
			</html>
		`))
	}))
	defer server.Close()

	tags, err := api.FetchOGTags(context.Background(), server.URL)
	assert.NoError(t, err)
	assert.Equal(t, "Test Title", tags.Title)
	assert.Equal(t, "Test Description", tags.Description)
	assert.Equal(t, "http://example.com/image.jpg", tags.Image)
}
