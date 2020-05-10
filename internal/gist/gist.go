package gist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// Defines different constants used
// GIT_IO_URL is the Github's URL shortner
// API v3 is the current version of GitHub API
const (
	GITHUB_API_URL = "https://api.github.com/gists"
)

//
type Option func(*options)

//
type options struct {
	anonymous   bool
	update      string
	description string
	public      bool
	username    string
	password    string
	files       GistFiles
}

// To update
func GistId(gistId string) func(*options) {
	return func(options *options) {
		options.update = gistId
	}
}

//
func Anonymous(anonymous bool) func(*options) {
	return func(options *options) {
		options.anonymous = anonymous
	}
}

//
func Description(description string) func(*options) {
	return func(options *options) {
		options.description = description
	}
}

//
func Public(public bool) func(*options) {
	return func(options *options) {
		options.public = public
	}
}

//
func Credentials(token string) func(*options) {
	parts := strings.Split(strings.TrimSpace(token), ":")
	return func(options *options) {
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return
		}
		options.anonymous = false
		options.username, options.password = parts[0], parts[1]
	}
}

// The top-level struct for a gist file
type GistFile struct {
	Content string `json:"content"`
}

//
type GistFiles map[string]GistFile

// The required structure for POST data for API purposes
type Gist struct {
	Description string    `json:"description,omitempty"`
	Public      bool      `json:"public"`
	GistFile    GistFiles `json:"files"`
	config      options   `json:"-"`
}

//
type User struct {
	Login             string `json:"login"`
	Id                int    `json:"id"`
	NodeId            string `json:"node_id"`
	AvatarUrl         string `json:"avatar_url"`
	GravatarId        string `json:"gravatar_id"`
	Url               string `json:"url"`
	HtmlUrl           string `json:"html_url"`
	FollowersUrl      string `json:"followers_url"`
	FollowingUrl      string `json:"following_url"`
	GistsUrl          string `json:"gists_url"`
	StarredUrl        string `json:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url"`
	ReposUrl          string `json:"repos_url"`
	EventsUrl         string `json:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}

type Response struct {
	Url         string                 `json:"url"`
	ForksUrl    string                 `json:"forks_url"`
	CommitsUrl  string                 `json:"commits_url"`
	Id          string                 `json:"id"`
	NodeId      string                 `json:"node_id"`
	GitPullUrl  string                 `json:"git_pull_url"`
	GitPushUrl  string                 `json:"git_push_url"`
	HtmlUrl     string                 `json:"html_url"`
	Files       map[string]interface{} `json:"files"`
	Public      bool                   `json:"public"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Description string                 `json:"description"`
	Comments    int                    `json:"comments"`
	User        User                   `json:"user"`
	CommentsUrl string                 `json:"comments_url"`
	Owner       map[string]interface{} `json:"owner"`
	Truncated   bool                   `json:"truncated"`
	Forks       []interface{}          `json:"forks"`
	History     []interface{}          `json:"history"`
}

//
type Errors struct {
	Message string
	Errors  map[string]error
}

func newErrors(message ...string) *Errors {
	return &Errors{Message: strings.Join(message, " "), Errors: map[string]error{}}
}

//
func (err Errors) Error() string {
	b := bytes.Buffer{}
	_, _ = fmt.Fprintf(&b, "ERROR: %s\n", err.Message)
	for filename, err := range err.Errors {
		_, _ = fmt.Fprintf(&b, "%s: %s\n", filename, err)
	}
	return b.String()
}

//
func MustFiles(files ...string) func(*options) {
	in, err := Files(files...)
	if err != nil {
		log.Fatal(err)
	}
	return in
}

//
func Files(files ...string) (func(*options), error) {
	errors := newErrors()
	gistFiles := GistFiles{}

	for _, filename := range files {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			errors.Errors[filename] = err
			continue
		}

		name := filepath.Base(filename)
		gistFiles[name] = GistFile{string(content)}
	}

	if len(errors.Errors) == 0 {
		errors = nil
	}

	return func(options *options) {
		options.files = gistFiles
	}, nil
}

//
func New(config ...Option) (gist Gist, err error) {
	opt := options{}
	for _, o := range config {
		o(&opt)
	}

	return Gist{
		config:      opt,
		Description: opt.description,
		Public:      opt.public,
		GistFile:    opt.files,
	}, nil
}

//
func (gist Gist) Send() (response Response, err error) {
	opts := gist.config

	if opts.username == "" && !opts.anonymous {
		return response, fmt.Errorf("no credentials provided and anonymous is false")
	}

	pfile, err := json.Marshal(gist)
	if err != nil {
		return response, err
	}

	b := bytes.NewBuffer(pfile)
	log.Println("Sending...")

	// Send request to API
	apiUrl := GITHUB_API_URL
	if opts.update != "" {
		apiUrl += "/" + opts.update
	}
	req, err := http.NewRequest("POST", apiUrl, b)
	if err != nil {
		return response, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if opts.username != "" {
		req.SetBasicAuth(opts.username, opts.password)
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	var decoded map[string]interface{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	// log.Printf("%+v", string(body))

	err = json.Unmarshal(body, &decoded)
	if err != nil {
		return response, err
	}

	if errors, ok := decoded["errors"]; ok {
		errs := Errors{}
		for _, messages := range errors.([]interface{}) {
			for reason, message := range messages.(map[string]interface{}) {
				errs.Errors[reason] = fmt.Errorf("%v", message)
			}
		}
		return response, errs
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}
