package go_asana

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/civil"
	"fmt"
	"github.com/okdewit/go-utils/catch"
	"github.com/okdewit/go-utils/httputils"
	"time"
)

type Events struct { Events []Event }

type Event struct {
	HasTimeStamps
	User User
	Action string
	Parent Resource `json:"parent"`
	Task Task `json:"resource"`
}

type Resource struct {
	HasTimeStamps
	Gid string
	Name string
	ResourceType string `json:"resource_type"`
	ResourceSubtype string `json:"resource_subtype"`
	CreatedBy User `json:"created_by"`
}

type Tasks []Task
type Task struct {
	Resource
	Assignee User
	AssigneeStatus string `json:"assignee_status"`
	Stories Stories
	Completed bool
	CompletedAt string `json:"completed_at"`
}

type Stories []Story
type Story struct {
	Resource
	Text string
}

type HasTimeStamps struct {
	CreatedAt string `json:"created_at"`
}

type User struct {
	Gid string
	Name string
	ResourceType string `json:"resource_type"`
}

func GetTasks(project string) []Task {
	uri := Geturl("tasks", "")
	httputils.AddParams(uri, map[string]string{
		"project": project,
		"opt_fields": "created_at",
		"limit": "100",
	})

	var tasks []Task

	for {
		data := Call("GET", uri, nil)
		response := TaskListResponse{}.Transform(data)
		tasks = append(tasks, response.Data...)
		if response.Next.Uri == "" { break }
		uri = response.GetNextPageUrl()
	}

	return tasks
}

func (task *Task) Enrich() Task {
	fmt.Printf("enriching task %v\n", task.Gid)
	data := Call("GET", Geturl("tasks", task.Gid), nil)
	return TaskResponse{}.Transform(data)
}

func (stories Stories) Filter(storyTypes map[string]bool) (filteredStories Stories) {
	for _, story := range stories {
		func(story Story) {
			if storyTypes[story.ResourceSubtype] {
				filteredStories = append(filteredStories, story)
			}
		}(story)
	}

	return filteredStories
}

func (timestampable *HasTimeStamps) GetCreatedAt() (t time.Time) {
	t, err := time.Parse(time.RFC3339, timestampable.CreatedAt)
	catch.Check(err)
	return
}

func (task *Task) GetCompletedAt() (t time.Time) {
	t, err := time.Parse(time.RFC3339, task.CompletedAt)
	catch.Check(err)
	return
}

func (task *Task) LastEvent(max time.Time) (lastEvent time.Time) {
	lastEvent = time.Now()
	if task.Completed {
		lastEvent = task.GetCompletedAt()
	}
	if max.Before(lastEvent) {
		lastEvent = max
	}
	return
}

func (resource *Resource) Load(vs []bigquery.Value, s bigquery.Schema) error{
	for _, v := range vs {
		switch value := v.(type) {
		case string: resource.Gid = value
		case civil.DateTime: resource.CreatedAt = value.In(time.UTC).Format(time.RFC3339)
		}
	}
	return nil
}
