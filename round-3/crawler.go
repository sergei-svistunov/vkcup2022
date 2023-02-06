package main

import (
	"bytes"
	"container/list"
	"fmt"
	"image"
	"image/png"
	"net/url"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"round-3/mypng"
)

type Crawler struct {
	httpClient    *fasthttp.Client
	id            int
	useStdDecoder bool
}

type ImgTask struct {
	Id       int
	Location string
}

func NewCrawler(useStdDecoder bool) *Crawler {
	return &Crawler{
		httpClient: &fasthttp.Client{
			NoDefaultUserAgentHeader:      true,
			MaxConnsPerHost:               runtime.NumCPU() * 2,
			MaxIdleConnDuration:           time.Minute,
			ReadBufferSize:                64 * 1024,
			WriteBufferSize:               8 * 1024,
			DisableHeaderNamesNormalizing: true,
			DisablePathNormalizing:        true,
		},
		useStdDecoder: useStdDecoder,
	}
}

func (c *Crawler) Walk(addr string, imgCh chan ImgTask) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}

	visited := map[string]struct{}{}
	queue := list.New()
	queue.PushBack(u.Path)

	var buf [8 * 1024]byte

	for queue.Len() > 0 {
		location := queue.Front().Value.(string)
		queue.Remove(queue.Front())

		if _, exists := visited[location]; exists {
			continue
		}
		visited[location] = struct{}{}

		code, body, err := c.httpClient.Get(buf[:], u.Scheme+"://"+u.Host+location)
		if err != nil {
			return err
		}
		if code != 200 {
			return fmt.Errorf("http error: %d", code)
		}

		links, images, err := parseHTML(string(body))
		if err != nil {
			return err
		}

		for _, link := range links {
			// We do not follow external links
			if strings.Index(link, "://") != -1 {
				linkU, err := url.Parse(link)
				if err != nil || linkU.Host != u.Host {
					continue
				}
				link = u.Path
			}
			// Process relative links
			if link[0] != '/' {
				link = path.Join(path.Dir(location), link)
			}
			queue.PushBack(link)
		}

		for _, img := range images {
			if img[0] != '/' {
				img = path.Join(path.Dir(location), img)
			}
			imgCh <- ImgTask{c.id, u.Scheme + "://" + u.Host + img}
			c.id++
		}
	}

	return nil
}

func (c *Crawler) GetImage(addr string) (image.Image, error) {
	var buf [8 * 1024]byte
	code, body, err := c.httpClient.Get(buf[:], addr)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("http error: %d", code)
	}

	if c.useStdDecoder {
		return png.Decode(bytes.NewBuffer(body))
	}

	return mypng.Decode(bytes.NewBuffer(body))
}
