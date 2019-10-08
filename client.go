package asana

import (
	"encoding/json"
	"github.com/okdewit/go-utils/catch"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

var Token string
var client *http.Client
var host = "https://app.asana.com/api/1.0"

type Response interface {
	Transform()
}

type TaskResponse struct { Data Task }
type ListResponse struct {
	Data []Resource
	Next struct{
		Uri string
	} `json:"next_page"`
}

type StoryListResponse struct {
	ListResponse
	Data []Story
}

type TaskListResponse struct {
	ListResponse
	Data []Task
}

func Call(method string, asanaUrl *url.URL, body io.Reader) (data []byte) {
	if client == nil {
		client = &http.Client{}
	}

	request, err := http.NewRequest(method, asanaUrl.String(), body)
	catch.Check(err)

	request.Header.Add("Authorization", "Bearer " + Token)
	response, err := client.Do(request)
	catch.Check(err)

	data, err = ioutil.ReadAll(response.Body)
	catch.Check(err)
	return
}

func Geturl(endpoint string, gid string) *url.URL {
	asanaurl, _ := url.Parse(host)
	asanaurl.Path += "/" + endpoint
	if gid != "" {
		asanaurl.Path += "/" + gid
	}
	return asanaurl
}

func (response TaskResponse) Transform(data []byte) Task {
	err := json.Unmarshal(data, &response)
	catch.Check(err)
	return response.Data
}

func (response ListResponse) GetNextPageUrl() (uri *url.URL) {
	uri, err := url.Parse(response.Next.Uri)
	catch.Check(err)
	return
}

func (response TaskListResponse) Transform(data []byte) TaskListResponse {
	var newResponse TaskListResponse
	err := json.Unmarshal(data, &newResponse)
	catch.Check(err)
	return newResponse
}

func (response StoryListResponse) GetNextPageUrl() (uri *url.URL) {
	uri, err := url.Parse(response.Next.Uri)
	catch.Check(err)
	return
}

func (response StoryListResponse) Transform(data []byte) StoryListResponse {
	var newResponse StoryListResponse
	err := json.Unmarshal(data, &newResponse)
	catch.Check(err)
	return newResponse
}