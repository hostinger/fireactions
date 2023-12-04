package fireactions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/hostinger/fireactions/version"
)

var (
	defaultUserAgent = fmt.Sprintf("fireactions/%s", version.Version)
	defaultEndpoint  = "http://127.0.0.1:8080"
)

// Client manages communication with the Fireactions API.
type Client interface {
	SetUserAgent(userAgent string)
	SetEndpoint(endpoint string)
	SetClient(client *http.Client)
	ListNodes(ctx context.Context, opts *NodesListOptions) ([]*Node, *Response, error)
	GetNode(ctx context.Context, id string) (*Node, *Response, error)
	RegisterNode(ctx context.Context, nodeRegisterRequest *NodeRegisterRequest) (*NodeRegistrationInfo, *Response, error)
	DeregisterNode(ctx context.Context, id string) (*Response, error)
	CordonNode(ctx context.Context, id string) (*Response, error)
	UncordonNode(ctx context.Context, id string) (*Response, error)
	GetNodeRunners(ctx context.Context, id string) ([]*Runner, *Response, error)
	GetRunner(ctx context.Context, id string) (*Runner, *Response, error)
	CreateRunner(ctx context.Context, createRunnerRequest CreateRunnerRequest) ([]*Runner, *Response, error)
	ListRunners(ctx context.Context, opts *RunnersListOptions) ([]*Runner, *Response, error)
	SetRunnerStatus(ctx context.Context, id string, setRunnerStatusRequest SetRunnerStatusRequest) (*Response, error)
	DeleteRunner(ctx context.Context, id string) (*Response, error)
}

type clientImpl struct {
	client *http.Client

	// Endpoint is the Fireactions API endpoint.
	Endpoint string

	// UserAgent is the User-Agent header to send when communicating with the
	// Fireactions API.
	UserAgent string
}

// ClientOpt is an option for a new Fireactions client.
type ClientOpt func(Client)

// WithHTTPClient returns a ClientOpt that specifies the HTTP client to use when
// making requests to the Fireactions API.
func WithHTTPClient(client *http.Client) ClientOpt {
	f := func(c Client) {
		c.SetClient(client)
	}

	return f
}

// WithEndpoint returns a ClientOpt that specifies the Fireactions API endpoint
// to use when making requests to the Fireactions API.
func WithEndpoint(endpoint string) ClientOpt {
	f := func(c Client) {
		c.SetEndpoint(endpoint)
	}

	return f
}

// WithUserAgent returns a ClientOpt that specifies the User-Agent header to use
// when making requests to the Fireactions API.
func WithUserAgent(userAgent string) ClientOpt {
	f := func(c Client) {
		c.SetUserAgent(userAgent)
	}

	return f
}

// NewClient returns a new Fireactions Client implementation.
func NewClient(opts ...ClientOpt) *clientImpl {
	c := &clientImpl{
		Endpoint:  defaultEndpoint,
		UserAgent: defaultUserAgent,
		client:    http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewRequestWithContext returns a new HTTP request with a context.
func (c *clientImpl) NewRequestWithContext(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.Endpoint, endpoint), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}

// Do sends an HTTP request and returns an HTTP response.
func (c *clientImpl) Do(req *http.Request, v interface{}) (*Response, error) {
	rsp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	response := &Response{Response: rsp}
	switch rsp.StatusCode {
	case
		http.StatusOK,
		http.StatusNoContent,
		http.StatusCreated:

		if v != nil {
			if w, ok := v.(io.Writer); ok {
				io.Copy(w, rsp.Body)
			} else {
				if err := json.NewDecoder(rsp.Body).Decode(v); err != nil {
					return response, err
				}
			}
		}

		return response, nil
	default:
		var apiErr Error
		if err := json.NewDecoder(rsp.Body).Decode(&apiErr); err != nil {
			return response, fmt.Errorf("%v %v: %d", req.Method, req.URL, rsp.StatusCode)
		}

		return response, &apiErr
	}
}

// Error represents an error returned by the Fireactions API.
type Error struct {
	Message string `json:"error"`
}

// Error returns the error message. Implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Response wraps an HTTP response.
type Response struct {
	*http.Response
}

// HasNextPage returns true if the response has a next page.
func (r *Response) HasNextPage() bool {
	return r.Header.Get("Link") != ""
}

// NextPage returns the next page URL.
func (r *Response) NextPage() (string, error) {
	link := r.Header.Get("Link")
	if link == "" {
		return "", nil
	}

	return "", nil
}

// ListOptions specifies the optional parameters to various List methods that
// support pagination.
type ListOptions struct {
	Page    int
	PerPage int
}

// Apply modifies the request to include the optional pagination parameters.
func (o *ListOptions) Apply(req *http.Request) {
	q := req.URL.Query()

	if o.Page != 0 {
		q.Set("page", fmt.Sprintf("%d", o.Page))
	}
	if o.PerPage != 0 {
		q.Set("per_page", fmt.Sprintf("%d", o.PerPage))
	}

	req.URL.RawQuery = q.Encode()
}

// NodesListOptions specifies the optional parameters to the
// NodesClient.List method.
type NodesListOptions struct {
	ListOptions
}

// ListNodes returns a list of Nodes.
func (c *clientImpl) ListNodes(ctx context.Context, opts *NodesListOptions) ([]*Node, *Response, error) {
	type Root struct {
		Nodes []*Node `json:"nodes"`
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/nodes", nil)
	if err != nil {
		return nil, nil, err
	}

	if opts != nil {
		opts.Apply(req)
	}

	var root Root
	response, err := c.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Nodes, response, nil
}

// GetNode returns a Node by ID.
func (c *clientImpl) GetNode(ctx context.Context, id string) (*Node, *Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/nodes/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var node Node
	response, err := c.Do(req, &node)
	if err != nil {
		return nil, response, err
	}

	return &node, response, nil
}

// NodeRegisterRequest represents a request to register a Node.
type NodeRegisterRequest struct {
	Name               string            `json:"name" binding:"required"`
	ReconcileInterval  time.Duration     `json:"reconcile_interval" binding:"required"`
	Labels             map[string]string `json:"labels" binding:"required"`
	CpuOvercommitRatio float64           `json:"cpu_overcommit_ratio" binding:"required"`
	CpuCapacity        int64             `json:"cpu_capacity" binding:"required"`
	RamOvercommitRatio float64           `json:"ram_overcommit_ratio" binding:"required"`
	RamCapacity        int64             `json:"ram_capacity" binding:"required"`
}

// NodeRegistrationInfo represents information about a registered Node. This
// information is returned when registering a Node.
type NodeRegistrationInfo struct {
	ID string `json:"id"`
}

// RegisterNode registers a Node.
func (c *clientImpl) RegisterNode(ctx context.Context, nodeRegisterRequest *NodeRegisterRequest) (*NodeRegistrationInfo, *Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/nodes", nodeRegisterRequest)
	if err != nil {
		return nil, nil, err
	}

	var nodeRegistrationInfo NodeRegistrationInfo
	response, err := c.Do(req, &nodeRegistrationInfo)
	if err != nil {
		return nil, response, err
	}

	return &nodeRegistrationInfo, response, nil
}

// DeregisterNode deregisters a Node by ID.
func (c *clientImpl) DeregisterNode(ctx context.Context, id string) (*Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/nodes/%s", id), nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, nil)
}

// CordonNode cordons a Node by ID.
func (c *clientImpl) CordonNode(ctx context.Context, id string) (*Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/cordon", id), nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, nil)
}

// UncordonNode uncordons a Node by ID.
func (c *clientImpl) UncordonNode(ctx context.Context, id string) (*Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/api/v1/nodes/%s/uncordon", id), nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, nil)
}

// GetNodeRunners returns a list of Runners for a Node by ID.
func (c *clientImpl) GetNodeRunners(ctx context.Context, id string) ([]*Runner, *Response, error) {
	type Root struct {
		Runners []*Runner `json:"runners"`
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/nodes/%s/runners", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var root Root
	response, err := c.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Runners, response, nil
}

// GetRunner returns a Runner by ID.
func (c *clientImpl) GetRunner(ctx context.Context, id string) (*Runner, *Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/runners/%s", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var runner Runner
	response, err := c.Do(req, &runner)
	if err != nil {
		return nil, response, err
	}

	return &runner, response, nil
}

// RunnersListOptions specifies the optional parameters to the
// RunnersClient.List method.
type RunnersListOptions struct {
	ListOptions
}

// ListRunners returns a list of Runners.
func (c *clientImpl) ListRunners(ctx context.Context, opts *RunnersListOptions) ([]*Runner, *Response, error) {
	type Root struct {
		Runners []*Runner `json:"runners"`
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/runners", nil)
	if err != nil {
		return nil, nil, err
	}

	var root Root
	response, err := c.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Runners, response, nil
}

// RunnerSetStatusRequest represents a request to set the status of a Runner by
// ID.
type SetRunnerStatusRequest struct {
	State       RunnerState `json:"phase" binding:"required"`
	Description string      `json:"description"`
}

// SetRunnerStatus sets the status of a Runner by ID.
func (c *clientImpl) SetRunnerStatus(ctx context.Context, id string, setRunnerStatusRequest SetRunnerStatusRequest) (*Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/runners/%s/status", id), setRunnerStatusRequest)
	if err != nil {
		return nil, err
	}

	return c.Do(req, nil)
}

// DeleteRunner deletes a Runner by ID. This is a soft delete. The Runner will be
// marked as deleted but will not be removed from the database.
func (c *clientImpl) DeleteRunner(ctx context.Context, id string) (*Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/runners/%s", id), nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req, nil)
}

type CreateRunnerRequest struct {
	Organisation string `json:"organisation" binding:"required"`
	JobLabel     string `json:"job_label" binding:"required"`
	Count        int    `json:"count" binding:"required"`
}

func (c *clientImpl) CreateRunner(ctx context.Context, createRunnerRequest CreateRunnerRequest) ([]*Runner, *Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/runners", createRunnerRequest)
	if err != nil {
		return nil, nil, err
	}

	type Root struct {
		Runners []*Runner `json:"runners"`
	}

	var root Root
	response, err := c.Do(req, &root)
	if err != nil {
		return nil, response, err
	}

	return root.Runners, response, nil
}

// SetUserAgent sets the User-Agent header to use when making requests to the
// Fireactions API.
func (c *clientImpl) SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
}

// SetEndpoint sets the Fireactions API endpoint to use when making requests to
// the Fireactions API.
func (c *clientImpl) SetEndpoint(endpoint string) {
	c.Endpoint = endpoint
}

// SetClient sets the HTTP client to use when making requests to the
// Fireactions API.
func (c *clientImpl) SetClient(client *http.Client) {
	c.client = client
}

type WorkflowRunStats struct {
	Stats []*WorkflowRunStat `json:"stats"`
}

// GetTotal returns the total number of workflow runs.
func (wrs *WorkflowRunStats) GetTotal() int {
	var total int
	for _, stat := range wrs.Stats {
		total += stat.Total
	}

	return total
}

// GetTotalDuration returns the total duration of the workflow runs.
func (wrs *WorkflowRunStats) GetTotalDuration() time.Duration {
	var total time.Duration
	for _, stat := range wrs.Stats {
		total += stat.TotalDuration
	}

	return total
}

// GetSucceeded returns the number of succeeded workflow runs.
func (wrs *WorkflowRunStats) GetSucceeded() int {
	var total int
	for _, stat := range wrs.Stats {
		total += stat.Succeeded
	}

	return total
}

// GetFailed returns the number of failed workflow runs.
func (wrs *WorkflowRunStats) GetFailed() int {
	var total int
	for _, stat := range wrs.Stats {
		total += stat.Failed
	}

	return total
}

// GetCancelled returns the number of cancelled workflow runs.
func (wrs *WorkflowRunStats) GetCancelled() int {
	var total int
	for _, stat := range wrs.Stats {
		total += stat.Cancelled
	}

	return total
}

// GetAverageDuration returns the average duration of the workflow runs.
func (wrs *WorkflowRunStats) GetAverageDuration() time.Duration {
	if len(wrs.Stats) == 0 {
		return 0
	}

	return time.Duration(int64(wrs.GetTotalDuration()) / int64(wrs.GetTotal()))
}

// GetSuccessRatio returns the success ratio of the workflow runs.
func (wrs *WorkflowRunStats) GetSuccessRatio() float64 {
	if len(wrs.Stats) == 0 {
		return 0
	}

	return float64(wrs.GetSucceeded()) / float64(wrs.GetTotal()) * 100
}

// GetFailureRatio returns the failure ratio of the workflow runs.
func (wrs *WorkflowRunStats) GetFailureRatio() float64 {
	if len(wrs.Stats) == 0 {
		return 0
	}

	return float64(wrs.GetFailed()) / float64(wrs.GetTotal()) * 100
}

// Sort sorts the workflow run stats by the given key and order.
func (wrs *WorkflowRunStats) Sort(key string, order string) error {
	switch key {
	case "TOTAL":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].Total < wrs.Stats[j].Total
			}

			return wrs.Stats[i].Total > wrs.Stats[j].Total
		})
	case "TOTAL_DURATION":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].TotalDuration < wrs.Stats[j].TotalDuration
			}

			return wrs.Stats[i].TotalDuration > wrs.Stats[j].TotalDuration
		})
	case "SUCCEEDED":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].Succeeded < wrs.Stats[j].Succeeded
			}

			return wrs.Stats[i].Succeeded > wrs.Stats[j].Succeeded
		})
	case "CANCELLED":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].Cancelled < wrs.Stats[j].Cancelled
			}

			return wrs.Stats[i].Cancelled > wrs.Stats[j].Cancelled
		})
	case "FAILED":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].Failed < wrs.Stats[j].Failed
			}

			return wrs.Stats[i].Failed > wrs.Stats[j].Failed
		})
	case "AVERAGE_DURATION":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].GetAverageDuration() < wrs.Stats[j].GetAverageDuration()
			}

			return wrs.Stats[i].GetAverageDuration() > wrs.Stats[j].GetAverageDuration()
		})
	case "SUCCESS_RATE":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].GetSuccessRatio() < wrs.Stats[j].GetSuccessRatio()
			}

			return wrs.Stats[i].GetSuccessRatio() > wrs.Stats[j].GetSuccessRatio()
		})
	case "FAILURE_RATE":
		sort.Slice(wrs.Stats, func(i, j int) bool {
			if order == "asc" {
				return wrs.Stats[i].GetFailureRatio() < wrs.Stats[j].GetFailureRatio()
			}

			return wrs.Stats[i].GetFailureRatio() > wrs.Stats[j].GetFailureRatio()
		})
	default:
		return fmt.Errorf("unrecognized sort key: %s, valid keys are: \"TOTAL\", \"TOTAL_DURATION\", \"SUCCEEDED\", \"CANCELLED\", \"FAILED\", \"AVERAGE_DURATION\", \"SUCCESS_RATIO\", \"FAILURE_RATIO\"", key)
	}

	return nil
}

// WorkflowRunStat represents statistics for a workflow run across multiple
// repositories.
type WorkflowRunStat struct {
	Repository    string        `json:"repository"`
	Total         int           `json:"total"`
	TotalDuration time.Duration `json:"total_duration"`
	Succeeded     int           `json:"succeeded"`
	Failed        int           `json:"failed"`
	Cancelled     int           `json:"cancelled"`
}

// GetAverageDuration returns the average duration of the workflow runs.
func (wrs *WorkflowRunStat) GetAverageDuration() time.Duration {
	if wrs.Total == 0 {
		return 0
	}

	return time.Duration(int64(wrs.TotalDuration) / int64(wrs.Total))
}

// GetSuccessRatio returns the success ratio of the workflow runs.
func (wrs *WorkflowRunStat) GetSuccessRatio() float64 {
	if wrs.Total == 0 {
		return 0
	}

	return float64(wrs.Succeeded) / float64(wrs.Total) * 100
}

// GetFailureRatio returns the failure ratio of the workflow runs.
func (wrs *WorkflowRunStat) GetFailureRatio() float64 {
	if wrs.Total == 0 {
		return 0
	}

	return float64(wrs.Failed) / float64(wrs.Total) * 100
}

// WorkflowRunStatsQuery represents a query to get workflow run statistics.
type WorkflowRunStatsQuery struct {
	Repositories string    `form:"repositories"`
	Start        time.Time `form:"start"`
	End          time.Time `form:"end"`
	SortOrder    string    `form:"sort_order"`
	Sort         string    `form:"sort"`
	Limit        int       `form:"limit"`
}

// Validate validates the query.
func (q *WorkflowRunStatsQuery) Validate() error {
	if !q.Start.IsZero() && !q.End.IsZero() && q.Start.After(q.End) {
		return fmt.Errorf("start date can't be after end date")
	}

	if !q.End.IsZero() && q.End.After(time.Now()) {
		return fmt.Errorf("end date can't be after current date")
	}

	if !q.Start.IsZero() && q.Start.After(time.Now()) {
		return fmt.Errorf("start date can't be after current date")
	}

	return nil
}

// GetWorkflowRunStats returns workflow run statistics.
func (c *clientImpl) GetWorkflowRunStats(ctx context.Context, organisation string, query *WorkflowRunStatsQuery) (*WorkflowRunStats, *Response, error) {
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("/api/v1/workflow-runs/%s/stats", organisation), nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()
	if !query.Start.IsZero() {
		q.Set("start", query.Start.Format(time.RFC3339))
	}

	if !query.End.IsZero() {
		q.Set("end", query.End.Format(time.RFC3339))
	}

	if query.Repositories != "" {
		q.Set("repositories", query.Repositories)
	}

	if query.Sort != "" {
		q.Set("sort", query.Sort)
	}

	if query.SortOrder != "" {
		q.Set("sort_order", query.SortOrder)
	}

	if query.Limit != 0 {
		q.Set("limit", fmt.Sprintf("%d", query.Limit))
	}

	req.URL.RawQuery = q.Encode()

	var workflowRunStats WorkflowRunStats
	response, err := c.Do(req, &workflowRunStats)
	if err != nil {
		return nil, response, err
	}

	return &workflowRunStats, response, nil
}
