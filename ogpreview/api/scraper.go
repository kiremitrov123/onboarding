package api

import (
	"context"
	"net/http"

	"github.com/kiremitrov123/onboarding/src/ogpreview/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/html"
)

func FetchOGTags(ctx context.Context, url string) (model.OGTags, error) {
	tr := otel.Tracer("og-preview-api")
	_, span := tr.Start(ctx, "fetchOGTags")
	defer span.End()

	span.SetAttributes(attribute.String("url", url))

	resp, err := http.Get(url)
	if err != nil {
		span.RecordError(err)
		return model.OGTags{}, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		span.RecordError(err)
		return model.OGTags{}, err
	}

	var tags model.OGTags
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var prop, content string
			for _, a := range n.Attr {
				if a.Key == "property" {
					prop = a.Val
				}
				if a.Key == "content" {
					content = a.Val
				}
			}
			switch prop {
			case "og:title":
				tags.Title = content
			case "og:description":
				tags.Description = content
			case "og:image":
				tags.Image = content
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)
	return tags, nil
}
