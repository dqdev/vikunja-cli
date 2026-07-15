package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const configFileName = ".vikunja.json"

type config struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type apiError struct {
	Message    string
	Code       int
	HTTPStatus int
}

func (e apiError) Error() string {
	return e.Message
}

type responsePayload struct {
	Data       any             `json:"data"`
	Pagination *paginationInfo `json:"_pagination,omitempty"`
}

type paginationInfo struct {
	TotalPages  int `json:"total_pages"`
	ResultCount int `json:"result_count"`
}

type client struct {
	baseURL string
	token   string
	http    *http.Client
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < 1 {
		printError("missing command group")
		return 2
	}

	var err error
	switch args[0] {
	case "config":
		err = runConfig(args[1:])
	case "projects":
		err = runProjects(args[1:])
	case "tasks":
		err = runTasks(args[1:])
	case "labels":
		err = runLabels(args[1:])
	case "comments":
		err = runComments(args[1:])
	case "attachments":
		err = runAttachments(args[1:])
	case "relations":
		err = runRelations(args[1:])
	default:
		err = usageError("unknown command group %q", args[0])
	}

	if err == nil {
		return 0
	}

	var apiErr apiError
	if errors.As(err, &apiErr) {
		payload := map[string]any{"error": apiErr.Message}
		if apiErr.Code != 0 {
			payload["code"] = apiErr.Code
		}
		if apiErr.HTTPStatus != 0 {
			payload["http_status"] = apiErr.HTTPStatus
		}
		printJSON(payload)
		return 1
	}

	printError(err.Error())
	if isUsageError(err) {
		return 2
	}
	return 1
}

func runConfig(args []string) error {
	if len(args) < 1 {
		return usageError("missing config command")
	}

	switch args[0] {
	case "set":
		fs := newFlagSet("config set")
		instanceURL := fs.String("url", "", "Vikunja instance URL")
		token := fs.String("token", "", "API token")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *instanceURL == "" {
			return usageError("missing required option --url")
		}
		if *token == "" {
			return usageError("missing required option --token")
		}
		path, err := saveConfig(config{URL: strings.TrimRight(*instanceURL, "/"), Token: *token})
		if err != nil {
			return err
		}
		return printJSON(responsePayload{Data: map[string]any{
			"message": fmt.Sprintf("Config saved to %s", path),
			"url":     strings.TrimRight(*instanceURL, "/"),
		}})
	case "show":
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		cfg.Token = maskToken(cfg.Token)
		return printJSON(responsePayload{Data: cfg})
	default:
		return usageError("unknown config command %q", args[0])
	}
}

func runProjects(args []string) error {
	if len(args) < 1 {
		return usageError("missing projects command")
	}

	switch args[0] {
	case "list":
		fs := newFlagSet("projects list")
		page := fs.Int("page", 1, "Page number")
		perPage := fs.Int("per-page", 50, "Items per page")
		search := fs.String("search", "", "Search projects by title")
		archived := fs.Bool("archived", false, "Include archived projects")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		params := url.Values{"page": {strconv.Itoa(*page)}, "per_page": {strconv.Itoa(*perPage)}}
		addParam(params, "s", *search)
		if *archived {
			params.Set("is_archived", "true")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get("/projects", params, true)
	case "get":
		id, err := singleIntArg(args[1:], "PROJECT_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get(fmt.Sprintf("/projects/%d", id), nil, false)
	case "create":
		fs := newFlagSet("projects create")
		title := fs.String("title", "", "Project title")
		description := fs.String("description", "", "Project description")
		parentID := fs.Int("parent-id", 0, "Parent project ID")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *title == "" {
			return usageError("missing required option --title")
		}
		body := map[string]any{"title": *title}
		addBodyString(body, "description", *description)
		if *parentID != 0 {
			body["parent_project_id"] = *parentID
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put("/projects", body)
	case "update":
		id, rest, err := leadingIntArg(args[1:], "PROJECT_ID")
		if err != nil {
			return err
		}
		fs := newFlagSet("projects update")
		title := fs.String("title", "", "New title")
		description := fs.String("description", "", "New description")
		if err := parseFlags(fs, rest); err != nil {
			return err
		}
		body := map[string]any{}
		addBodyString(body, "title", *title)
		addBodyString(body, "description", *description)
		if len(body) == 0 {
			return errors.New("Provide at least one of --title or --description")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.post(fmt.Sprintf("/projects/%d", id), body)
	case "delete":
		id, err := singleIntArg(args[1:], "PROJECT_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.delete(fmt.Sprintf("/projects/%d", id))
	default:
		return usageError("unknown projects command %q", args[0])
	}
}

func runTasks(args []string) error {
	if len(args) < 1 {
		return usageError("missing tasks command")
	}

	switch args[0] {
	case "list":
		fs := newFlagSet("tasks list")
		projectID := fs.Int("project-id", 0, "Project ID")
		page := fs.Int("page", 1, "Page number")
		perPage := fs.Int("per-page", 50, "Items per page")
		search := fs.String("search", "", "Search tasks by text")
		filter := fs.String("filter", "", "Vikunja filter expression")
		sortBy := fs.String("sort-by", "", "Sort field")
		orderBy := fs.String("order-by", "", "Sort order")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *orderBy != "" && *orderBy != "asc" && *orderBy != "desc" {
			return usageError("--order-by must be asc or desc")
		}
		params := url.Values{"page": {strconv.Itoa(*page)}, "per_page": {strconv.Itoa(*perPage)}}
		addParam(params, "s", *search)
		addParam(params, "sort_by", *sortBy)
		addParam(params, "order_by", *orderBy)
		filterParts := []string{}
		if *projectID != 0 {
			filterParts = append(filterParts, fmt.Sprintf("project_id = %d", *projectID))
		}
		if *filter != "" {
			filterParts = append(filterParts, fmt.Sprintf("(%s)", *filter))
		}
		if len(filterParts) > 0 {
			params.Set("filter", strings.Join(filterParts, " && "))
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get("/tasks", params, true)
	case "get":
		id, err := singleIntArg(args[1:], "TASK_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get(fmt.Sprintf("/tasks/%d", id), nil, false)
	case "create":
		fs := newFlagSet("tasks create")
		projectID := fs.Int("project-id", 0, "Project ID")
		title := fs.String("title", "", "Task title")
		description := fs.String("description", "", "Task description")
		dueDate := fs.String("due-date", "", "Due date")
		priority := fs.Int("priority", -1, "Priority 0-5")
		labelIDs := fs.String("label-ids", "", "Comma-separated label IDs")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *projectID == 0 {
			return usageError("missing required option --project-id")
		}
		if *title == "" {
			return usageError("missing required option --title")
		}
		if *priority < -1 || *priority > 5 {
			return usageError("--priority must be between 0 and 5")
		}
		body := map[string]any{"title": *title}
		addBodyString(body, "description", *description)
		addBodyString(body, "due_date", *dueDate)
		if *priority != -1 {
			body["priority"] = *priority
		}
		if *labelIDs != "" {
			ids, err := parseCSVInts(*labelIDs)
			if err != nil {
				return err
			}
			labels := make([]map[string]int, 0, len(ids))
			for _, id := range ids {
				labels = append(labels, map[string]int{"id": id})
			}
			body["labels"] = labels
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put(fmt.Sprintf("/projects/%d/tasks", *projectID), body)
	case "update":
		id, rest, err := leadingIntArg(args[1:], "TASK_ID")
		if err != nil {
			return err
		}
		fs := newFlagSet("tasks update")
		title := fs.String("title", "", "New title")
		description := fs.String("description", "", "New description")
		dueDate := fs.String("due-date", "", "New due date or empty string to clear")
		priority := fs.Int("priority", -1, "Priority 0-5")
		done := fs.String("done", "", "true or false")
		percentDone := fs.Float64("percent-done", -1, "Completion percentage")
		if err := parseFlags(fs, rest); err != nil {
			return err
		}
		if *priority < -1 || *priority > 5 {
			return usageError("--priority must be between 0 and 5")
		}
		if *percentDone < -1 || *percentDone > 1 {
			return usageError("--percent-done must be between 0.0 and 1.0")
		}
		body := map[string]any{}
		addBodyString(body, "title", *title)
		addBodyString(body, "description", *description)
		if flagProvided(fs, "due-date") {
			if *dueDate == "" {
				body["due_date"] = nil
			} else {
				body["due_date"] = *dueDate
			}
		}
		if *priority != -1 {
			body["priority"] = *priority
		}
		if *done != "" {
			switch *done {
			case "true":
				body["done"] = true
			case "false":
				body["done"] = false
			default:
				return usageError("--done must be true or false")
			}
		}
		if *percentDone != -1 {
			body["percent_done"] = *percentDone
		}
		if len(body) == 0 {
			return errors.New("Provide at least one field to update")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.post(fmt.Sprintf("/tasks/%d", id), body)
	case "delete":
		id, err := singleIntArg(args[1:], "TASK_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.delete(fmt.Sprintf("/tasks/%d", id))
	default:
		return usageError("unknown tasks command %q", args[0])
	}
}

func runLabels(args []string) error {
	if len(args) < 1 {
		return usageError("missing labels command")
	}

	switch args[0] {
	case "list":
		fs := newFlagSet("labels list")
		page := fs.Int("page", 1, "Page number")
		perPage := fs.Int("per-page", 50, "Items per page")
		search := fs.String("search", "", "Search labels by title")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		params := url.Values{"page": {strconv.Itoa(*page)}, "per_page": {strconv.Itoa(*perPage)}}
		addParam(params, "s", *search)
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get("/labels", params, true)
	case "get":
		id, err := singleIntArg(args[1:], "LABEL_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get(fmt.Sprintf("/labels/%d", id), nil, false)
	case "create":
		fs := newFlagSet("labels create")
		title := fs.String("title", "", "Label title")
		color := fs.String("color", "", "Hex color")
		description := fs.String("description", "", "Description")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *title == "" {
			return usageError("missing required option --title")
		}
		body := map[string]any{"title": *title}
		if *color != "" {
			body["hex_color"] = strings.TrimPrefix(*color, "#")
		}
		addBodyString(body, "description", *description)
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put("/labels", body)
	case "update":
		id, rest, err := leadingIntArg(args[1:], "LABEL_ID")
		if err != nil {
			return err
		}
		fs := newFlagSet("labels update")
		title := fs.String("title", "", "New title")
		color := fs.String("color", "", "New hex color")
		description := fs.String("description", "", "New description")
		if err := parseFlags(fs, rest); err != nil {
			return err
		}
		body := map[string]any{}
		addBodyString(body, "title", *title)
		if *color != "" {
			body["hex_color"] = strings.TrimPrefix(*color, "#")
		}
		if flagProvided(fs, "description") {
			body["description"] = *description
		}
		if len(body) == 0 {
			return errors.New("Provide at least one of --title, --color, or --description")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put(fmt.Sprintf("/labels/%d", id), body)
	case "delete":
		id, err := singleIntArg(args[1:], "LABEL_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.delete(fmt.Sprintf("/labels/%d", id))
	case "apply":
		fs := newFlagSet("labels apply")
		taskID := fs.Int("task-id", 0, "Task ID")
		labelID := fs.Int("label-id", 0, "Label ID")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *taskID == 0 {
			return usageError("missing required option --task-id")
		}
		if *labelID == 0 {
			return usageError("missing required option --label-id")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put(fmt.Sprintf("/tasks/%d/labels", *taskID), map[string]any{"label_id": *labelID})
	case "remove":
		fs := newFlagSet("labels remove")
		taskID := fs.Int("task-id", 0, "Task ID")
		labelID := fs.Int("label-id", 0, "Label ID")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *taskID == 0 {
			return usageError("missing required option --task-id")
		}
		if *labelID == 0 {
			return usageError("missing required option --label-id")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.delete(fmt.Sprintf("/tasks/%d/labels/%d", *taskID, *labelID))
	default:
		return usageError("unknown labels command %q", args[0])
	}
}

func runComments(args []string) error {
	if len(args) < 1 {
		return usageError("missing comments command")
	}

	switch args[0] {
	case "list":
		taskID, err := requiredTaskID(args[1:], "comments list")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get(fmt.Sprintf("/tasks/%d/comments", taskID), nil, false)
	case "create":
		fs := newFlagSet("comments create")
		taskID := fs.Int("task-id", 0, "Task ID")
		text := fs.String("text", "", "Comment text")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *taskID == 0 {
			return usageError("missing required option --task-id")
		}
		if *text == "" {
			return usageError("missing required option --text")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put(fmt.Sprintf("/tasks/%d/comments", *taskID), map[string]any{"comment": *text})
	case "delete":
		fs := newFlagSet("comments delete")
		taskID := fs.Int("task-id", 0, "Task ID")
		commentID := fs.Int("comment-id", 0, "Comment ID")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *taskID == 0 {
			return usageError("missing required option --task-id")
		}
		if *commentID == 0 {
			return usageError("missing required option --comment-id")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.delete(fmt.Sprintf("/tasks/%d/comments/%d", *taskID, *commentID))
	default:
		return usageError("unknown comments command %q", args[0])
	}
}

func runAttachments(args []string) error {
	if len(args) < 1 {
		return usageError("missing attachments command")
	}

	switch args[0] {
	case "list":
		taskID, err := requiredTaskID(args[1:], "attachments list")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get(fmt.Sprintf("/tasks/%d/attachments", taskID), nil, false)
	case "download":
		fs := newFlagSet("attachments download")
		taskID := fs.Int("task-id", 0, "Task ID")
		attachmentID := fs.Int("attachment-id", 0, "Attachment ID")
		outputDir := fs.String("output-dir", ".", "Output directory")
		if err := parseFlags(fs, args[1:]); err != nil {
			return err
		}
		if *taskID == 0 {
			return usageError("missing required option --task-id")
		}
		if *attachmentID == 0 {
			return usageError("missing required option --attachment-id")
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.download(fmt.Sprintf("/tasks/%d/attachments/%d", *taskID, *attachmentID), *outputDir)
	default:
		return usageError("unknown attachments command %q", args[0])
	}
}

func runRelations(args []string) error {
	if len(args) < 1 {
		return usageError("missing relations command")
	}

	switch args[0] {
	case "list":
		id, err := singleIntArg(args[1:], "TASK_ID")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.get(fmt.Sprintf("/tasks/%d/relations", id), nil, false)
	case "create":
		taskID, otherID, kind, err := relationFlags(args[1:], "relations create")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.put(fmt.Sprintf("/tasks/%d/relations", taskID), map[string]any{
			"relation_kind": kind,
			"other_task_id": otherID,
		})
	case "delete":
		taskID, otherID, kind, err := relationFlags(args[1:], "relations delete")
		if err != nil {
			return err
		}
		c, err := newClient()
		if err != nil {
			return err
		}
		return c.delete(fmt.Sprintf("/tasks/%d/relations/%s/%d", taskID, kind, otherID))
	default:
		return usageError("unknown relations command %q", args[0])
	}
}

func newClient() (*client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return &client{
		baseURL: strings.TrimRight(cfg.URL, "/") + "/api/v1",
		token:   cfg.Token,
		http:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (c *client) get(path string, params url.Values, paginated bool) error {
	return c.request(http.MethodGet, path, params, nil, paginated)
}

func (c *client) put(path string, body map[string]any) error {
	return c.request(http.MethodPut, path, nil, body, false)
}

func (c *client) post(path string, body map[string]any) error {
	return c.request(http.MethodPost, path, nil, body, false)
}

func (c *client) delete(path string) error {
	return c.request(http.MethodDelete, path, nil, nil, false)
}

func (c *client) request(method, path string, params url.Values, body map[string]any, paginated bool) error {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp, respBody)
	}

	var data any
	if resp.StatusCode == http.StatusNoContent || len(respBody) == 0 {
		data = nil
	} else if err := json.Unmarshal(respBody, &data); err != nil {
		return err
	}

	payload := responsePayload{Data: data}
	if paginated {
		payload.Pagination = &paginationInfo{
			TotalPages:  headerInt(resp.Header, "x-pagination-total-pages", 1),
			ResultCount: headerInt(resp.Header, "x-pagination-result-count", 0),
		}
	}
	return printJSON(payload)
}

func (c *client) download(path, outputDir string) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}
		return parseAPIError(resp, respBody)
	}

	filename := extractFilename(resp.Header.Get("content-disposition"))
	if filename == "" {
		parts := strings.Split(strings.TrimRight(path, "/"), "/")
		filename = parts[len(parts)-1]
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	outputPath := filepath.Join(outputDir, filename)
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}
	return printJSON(responsePayload{Data: map[string]any{"saved_to": absPath, "filename": filename}})
}

func loadConfig() (config, error) {
	path, err := configPath()
	if err != nil {
		return config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return config{}, fmt.Errorf("Config not found at %s. Run: vikunja config set --url URL --token TOKEN", path)
		}
		return config{}, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	if cfg.URL == "" {
		return config{}, errors.New("Config missing required field 'url'. Run: vikunja config set --url URL --token TOKEN")
	}
	if cfg.Token == "" {
		return config{}, errors.New("Config missing required field 'token'. Run: vikunja config set --url URL --token TOKEN")
	}
	cfg.URL = strings.TrimRight(cfg.URL, "/")
	return cfg, nil
}

func saveConfig(cfg config) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	payload = append(payload, '\n')
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return "", err
	}
	_ = os.Chmod(path, 0o600)
	return path, nil
}

func configPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolvedExecutable, err := filepath.EvalSymlinks(executable)
	if err == nil {
		executable = resolvedExecutable
	}
	return filepath.Join(filepath.Dir(executable), configFileName), nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	return nil
}

func singleIntArg(args []string, name string) (int, error) {
	if len(args) != 1 {
		return 0, usageError("expected %s", name)
	}
	return parseIntArg(args[0], name)
}

func leadingIntArg(args []string, name string) (int, []string, error) {
	if len(args) < 1 {
		return 0, nil, usageError("expected %s", name)
	}
	id, err := parseIntArg(args[0], name)
	return id, args[1:], err
}

func parseIntArg(value, name string) (int, error) {
	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, usageError("%s must be an integer", name)
	}
	return id, nil
}

func requiredTaskID(args []string, command string) (int, error) {
	fs := newFlagSet(command)
	taskID := fs.Int("task-id", 0, "Task ID")
	if err := parseFlags(fs, args); err != nil {
		return 0, err
	}
	if *taskID == 0 {
		return 0, usageError("missing required option --task-id")
	}
	return *taskID, nil
}

func relationFlags(args []string, command string) (int, int, string, error) {
	fs := newFlagSet(command)
	taskID := fs.Int("task-id", 0, "Source task ID")
	otherID := fs.Int("other-id", 0, "Related task ID")
	kind := fs.String("kind", "", "Relation kind")
	if err := parseFlags(fs, args); err != nil {
		return 0, 0, "", err
	}
	if *taskID == 0 {
		return 0, 0, "", usageError("missing required option --task-id")
	}
	if *otherID == 0 {
		return 0, 0, "", usageError("missing required option --other-id")
	}
	if !validRelationKind(*kind) {
		return 0, 0, "", usageError("invalid relation kind %q", *kind)
	}
	return *taskID, *otherID, *kind, nil
}

func validRelationKind(kind string) bool {
	switch kind {
	case "subtask", "parenttask", "related", "duplicateof", "duplicates",
		"blocking", "blocked", "precedes", "follows", "copiedfrom", "copiedto":
		return true
	default:
		return false
	}
}

func flagProvided(fs *flag.FlagSet, name string) bool {
	provided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			provided = true
		}
	})
	return provided
}

func parseCSVInts(value string) ([]int, error) {
	parts := strings.Split(value, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			return nil, usageError("label IDs must be integers")
		}
		out = append(out, id)
	}
	return out, nil
}

func addParam(params url.Values, key, value string) {
	if value != "" {
		params.Set(key, value)
	}
}

func addBodyString(body map[string]any, key, value string) {
	if value != "" {
		body[key] = value
	}
}

func parseAPIError(resp *http.Response, body []byte) error {
	var raw map[string]any
	_ = json.Unmarshal(body, &raw)

	message, _ := raw["message"].(string)
	if message == "" {
		message = resp.Status
	}
	code := 0
	switch value := raw["code"].(type) {
	case float64:
		code = int(value)
	case int:
		code = value
	}
	return apiError{Message: message, Code: code, HTTPStatus: resp.StatusCode}
}

func headerInt(headers http.Header, name string, fallback int) int {
	value := headers.Get(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:6] + "..." + token[len(token)-4:]
}

func extractFilename(contentDisposition string) string {
	if contentDisposition == "" {
		return ""
	}
	re := regexp.MustCompile(`(?i)filename\*?=(?:UTF-8''|["']?)([^;"'\r\n]+)`)
	matches := re.FindStringSubmatch(contentDisposition)
	if len(matches) < 2 {
		return ""
	}
	return strings.Trim(matches[1], `"'`)
}

func printJSON(payload any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func printError(message string) {
	_ = printJSON(map[string]any{"error": message})
}

type usageErr struct {
	message string
}

func (e usageErr) Error() string {
	return e.message
}

func usageError(format string, args ...any) error {
	return usageErr{message: fmt.Sprintf(format, args...)}
}

func isUsageError(err error) bool {
	var target usageErr
	return errors.As(err, &target)
}
