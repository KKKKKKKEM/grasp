package pexels

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/KKKKKKKEM/flowkit/stages/download/http/fetcher"
	"github.com/KKKKKKKEM/flowkit/stages/extract"
	"github.com/tidwall/gjson"
)

type APIParser struct {
}

func (p *APIParser) Name() string {
	return "pexels-api-parser"
}

func (p *APIParser) Handlers() []*extract.Parser {
	return []*extract.Parser{
		{
			Pattern:  regexp.MustCompile(`^https://api\.pexels\.com/v1/photos/(\d+)$`),
			Priority: 0,
			Hint:     "pexels image API parser",
			Parse:    p.ParseImageAPI,
		},
		{
			Pattern:  regexp.MustCompile(`^https://www\.pexels\.com/.*?/api/v3/media/(\d+)$`),
			Priority: 0,
			Hint:     "pexels media API parser",
			Parse:    p.ParseMediaAPI,
		},
	}
}

func (p *APIParser) ParseImageAPI(ctx context.Context, task *extract.Task, opts *extract.Opts) ([]extract.Item, error) {

	httpClient := &fetcher.HttpClient{}
	request, err := fetcher.NewRequest("GET", task.URL, task.Headers)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Request(ctx, request, opts.ToDownloaderOpts())
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	alt := gjson.GetBytes(body, "alt").String()
	width := gjson.GetBytes(body, "width").Int()
	height := gjson.GetBytes(body, "height").Int()
	url := gjson.GetBytes(body, "url").String()
	id := gjson.GetBytes(body, "id").Int()

	var items []extract.Item
	gjson.GetBytes(body, "src").ForEach(func(key, value gjson.Result) bool {
		items = append(items, extract.Item{
			Name:     fmt.Sprintf("%s (%s)", alt, key.String()),
			URI:      value.String(),
			IsDirect: true,
			Meta: map[string]any{
				"id":     id,
				"width":  width,
				"height": height,
				"url":    url,
			},
		})
		return true
	})

	return items, nil
}
func (p *APIParser) ParseMediaAPI(ctx context.Context, task *extract.Task, opts *extract.Opts) ([]extract.Item, error) {

	defaultSecretKey := "H2jk9uKnhRmL6WPwh89zBezWvr"
	if task.Headers == nil {
		task.Headers = make(map[string]string)
	}
	// 如果请求头中没有 Authorization，则使用默认的 secret key
	if _, ok := task.Headers["secret-key"]; !ok {
		task.Headers["secret-key"] = defaultSecretKey
	}

	httpClient := &fetcher.HttpClient{}
	request, err := fetcher.NewRequest("GET", task.URL, task.Headers)
	if err != nil {
		return nil, err
	}

	response, err := httpClient.Request(ctx, request, opts.ToDownloaderOpts())
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	title := gjson.GetBytes(body, "data.attributes.title").String()
	width := gjson.GetBytes(body, "data.attributes.width").Int()
	height := gjson.GetBytes(body, "data.attributes.height").Int()
	url := gjson.GetBytes(body, "data.attributes.slug").String()
	id := gjson.GetBytes(body, "data.attributes.id").Int()

	mediaType := gjson.GetBytes(body, "data.type").String()
	var key = "data.attributes.image"
	if mediaType == "video" {
		key = "data.attributes.video.video_files"
	}

	var items []extract.Item
	gjson.GetBytes(body, key).ForEach(func(key, value gjson.Result) bool {
		if mediaType == "video" {
			data := value.Map()
			width := data["width"].Int()
			height := data["height"].Int()
			fps := data["fps"].Float()
			url := data["link"].String()
			items = append(items, extract.Item{
				Name:     fmt.Sprintf("%s - %d*%d (fps: %f)", title, width, height, fps),
				URI:      url,
				IsDirect: true,
				Meta: map[string]any{
					"id":     id,
					"width":  width,
					"height": height,
					"url":    url,
					"fps":    fps,
				},
			})
		} else {
			items = append(items, extract.Item{
				Name:     fmt.Sprintf("%s (%s)", title, key.String()),
				URI:      value.String(),
				IsDirect: true,
				Meta: map[string]any{
					"id":     id,
					"width":  width,
					"height": height,
					"url":    url,
				},
			})
		}

		return true
	})

	return items, nil
}
