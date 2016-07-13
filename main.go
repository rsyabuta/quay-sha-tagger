package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gocraft/web"
	"github.com/uber-go/zap"
)

type Context struct {
}

type QuayBuild struct {
	Repository      string   `json:"repository"`
	Namespace       string   `json:"namespace"`
	Name            string   `json:"name"`
	DockerURL       string   `json:"docker_url"`
	Homepage        string   `json:"homepage"`
	Visibility      string   `json:"visibility"`
	BuildID         string   `json:"build_id"`
	BuildName       string   `json:"build_name"`
	DockerTags      []string `json:"docker_tags"`
	TriggerKind     string   `json:"trigger_kind"`
	TriggerID       string   `json:"trigger_id"`
	TriggerMetadata struct {
		DefaultBranch string `json:"default_branch"`
		Ref           string `json:"ref"`
		Commit        string `json:"commit"`
		CommitInfo    struct {
			URL     string `json:"url"`
			Message string `json:"message"`
			Date    string `json:"date"`
			Author  struct {
				Username  string `json:"username"`
				URL       string `json:"url"`
				AvatarURL string `json:"avatar_url"`
			} `json:"author"`
			Committer struct {
				Username  string `json:"username"`
				URL       string `json:"url"`
				AvatarURL string `json:"avatar_url"`
			} `json:"committer"`
		} `json:"commit_info"`
	} `json:"trigger_metadata"`
}

type QuayImage struct {
	HasAdditional bool `json:"has_additional"`
	Page          int  `json:"page"`
	Tags          []struct {
		Reversion     bool   `json:"reversion"`
		StartTs       int    `json:"start_ts"`
		Name          string `json:"name"`
		DockerImageID string `json:"docker_image_id"`
		EndTs         int    `json:"end_ts,omitempty"`
	} `json:"tags"`
}

var logger zap.Logger
var quayToken string
var quayApiUrl string

func (c *Context) PrintHook(rw web.ResponseWriter, req *web.Request) {

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(string(body))
}

func (c *Context) TagBuild(rw web.ResponseWriter, req *web.Request) {

	defer req.Body.Close()
	body, _ := ioutil.ReadAll(req.Body)
	build := QuayBuild{}
	err := json.Unmarshal(body, &build)
	if err != nil {
		logger.Error("error", zap.Err(err))
		return
	}
	image, err := build.GetImage()
	if err != nil {
		return
	}
	for _, tag := range image.Tags {
		if tag.EndTs == 0 {
			client := &http.Client{}
			req, err := http.NewRequest("PUT", fmt.Sprintf("%s/repository/%s/tag/%s", quayApiUrl, build.Repository, build.BuildName), bytes.NewBufferString(fmt.Sprintf("{\"image\": \"%s\"}", tag.DockerImageID)))
			if err != nil {
				logger.Error("request error", zap.Err(err))
				return
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", quayToken))
			req.Header.Set("Content-Type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				logger.Error("response error", zap.Err(err))
				return
			}
			logger.Info("tagged image", zap.String("repository", build.Repository), zap.String("tag", build.BuildName))
			defer res.Body.Close()
		}
	}
}

func (c *Context) Ping(rw web.ResponseWriter, req *web.Request) {

	fmt.Fprint(rw, "pong\n")
}

func (c *Context) Version(rw web.ResponseWriter, req *web.Request) {

	fmt.Fprint(rw, "v1.0.0\n")
}

func (b *QuayBuild) GetImage() (*QuayImage, error) {

	if len(b.DockerTags) == 0 {
		err := errors.New("build has no tags")
		logger.Error("build error", zap.Err(err))
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repository/%s/tag/?specificTag=%s", quayApiUrl, b.Repository, b.DockerTags[0]), nil)
	if err != nil {
		logger.Error("request error", zap.Err(err))
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", quayToken))
	res, err := client.Do(req)
	if err != nil {
		logger.Error("response error", zap.Err(err))
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	image := &QuayImage{}
	err = json.Unmarshal(body, image)
	if err != nil {
		logger.Error("unmarshal error", zap.Err(err))
		return nil, err
	}
	return image, err
}

func init() {
	flag.StringVar(&quayToken, "quay-token", os.Getenv("QUAY_TOKEN"), "Quay API token. Must have repository read/write access.")
	flag.StringVar(&quayApiUrl, "quay-url", "https://quay.io/api/v1", "Quay API url.")
}

func main() {
	logger = zap.NewJSON()
	flag.Parse()
	router := web.New(Context{}).Post("/tag", (*Context).TagBuild)
	router.Get("/ping", (*Context).Ping)
	router.Get("/version", (*Context).Version)
	http.ListenAndServe("0.0.0.0:3000", router)
}
