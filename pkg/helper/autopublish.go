package helper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/andygrunwald/vdf"
	"github.com/benfiola/game-server-helper/pkg/helperapi"
	"github.com/google/go-github/v68/github"
)

// AutopublisherPublisher is an interface that creates publish jobs for a desired platform.
type AutopublishPublisher interface {
	Publish(data map[string]any) error
	Initialize(api Api) error
}

// GithubPublisher publishes workflows to the github platform
type GithubPublisher struct {
	AutopublishPublisher
	Api                Api
	GetWorkflowPattern func(data map[string]any) string
	Owner              string
	Repo               string
	Token              string `env:"GITHUB_TOKEN"`
	WorkflowFilename   string
}

// Initialize the [GithubPublisher].
// Returns an error if initialization fails.
func (gp *GithubPublisher) Initialize(api Api) error {
	gp.Api = api
	err := gp.Api.ParseEnv(gp)
	if err != nil {
		return err
	}
	if gp.GetWorkflowPattern == nil {
		return fmt.Errorf("get workflow pattern is required")
	}
	return nil
}

// Creates a workflow run using the given data as input.
// Returns a failure if the github API request fails.
// NOTE: Does not wait for the workflow run to exist
func (gp *GithubPublisher) CreateWorkflowRun(data map[string]any) error {
	gp.Api.Logger.Info("create workflow run", "data", data)
	client := github.NewClient(nil).WithAuthToken(gp.Token)
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), gp.Owner, gp.Repo, gp.WorkflowFilename, github.CreateWorkflowDispatchEventRequest{
		Ref:    "main",
		Inputs: data,
	})
	if err != nil {
		return err
	}
	return nil
}

// Finds a workflow run matching the given regex pattern.
// Returns a failure if the github API request fails.
func (gp *GithubPublisher) FindWorkflowRun(nameRegex *regexp.Regexp) (*github.WorkflowRun, error) {
	gp.Api.Logger.Info("list workflow runs")
	client := github.NewClient(nil).WithAuthToken(gp.Token)
	runs, _, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), gp.Owner, gp.Repo, gp.WorkflowFilename, &github.ListWorkflowRunsOptions{})
	if err != nil {
		return nil, err
	}
	var found *github.WorkflowRun
	for _, run := range runs.WorkflowRuns {
		if run.Name == nil {
			continue
		}
		if !nameRegex.Match([]byte(*run.Name)) {
			continue
		}
		found = run
		break
	}
	return found, nil
}

// Publishes a workflow to github.
// Returns an error if the workflow publish fails.
func (gp *GithubPublisher) Publish(data map[string]any) error {
	gp.Api.Logger.Info("publish to github", "data", data)
	workflowPattern := gp.GetWorkflowPattern(data)
	workflowRegex, err := regexp.Compile(workflowPattern)
	if err != nil {
		return err
	}
	workflowRun, err := gp.FindWorkflowRun(workflowRegex)
	if err != nil {
		return err
	}
	if workflowRun != nil {
		return nil
	}
	err = gp.CreateWorkflowRun(data)
	if err != nil {
		return nil
	}
	for workflowRun != nil {
		workflowRun, err = gp.FindWorkflowRun(workflowRegex)
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

// AutopublishChecker checks a remote source for the latest version of software
type AutopublishChecker interface {
	GetData() (map[string]any, error)
	Initialize(api Api) error
}

// SteamChecker checks steam for updated manifests for a given application.
type SteamChecker struct {
	AutopublishChecker
	Api        Api
	AppId      string
	BranchName string
	DepotId    string
	Password   string `env:"STEAM_PASSWORD"`
	Username   string `env:"STEAM_USERNAME"`
}

// Helper method to run a steamcmd script.  Returns the output of the command.
// Returns an error if a temporary directory (holding the steamcmd script) cannot be created.
// Returns an error if the script cannot be written to the temporary file.
// Returns an error if the steamcmd fails
func (sc *SteamChecker) RunSteamCmd(command []string) (string, error) {
	sc.Api.Logger.Info("run steam cmd", "command", command)
	commands := []string{fmt.Sprintf("login %s %s", sc.Username, sc.Password)}
	commands = append(commands, strings.Join(command, " "))
	commands = append(commands, "quit")
	output := ""
	err := sc.Api.CreateTempDir(func(tempDir string) error {
		tempFile := filepath.Join(tempDir, "steamcmd.txt")
		content := []byte(strings.Join(commands, "\n"))
		err := os.WriteFile(tempFile, content, 0755)
		if err != nil {
			return err
		}
		output, err = sc.Api.RunCommand([]string{"steamcmd", "+runscript", tempFile}, helperapi.CmdOpts{})
		if err != nil {
			return err
		}
		return nil
	})
	return output, err
}

// Retrieves application data from steam.
// Returns an error if the data fetch fails
func (sc *SteamChecker) GetAppInfo() (map[string]any, error) {
	fail := func(err error) (map[string]any, error) {
		return map[string]any{}, err
	}
	sc.Api.Logger.Info("get steam app info", "app", sc.AppId)
	output, err := sc.RunSteamCmd([]string{"app_info_print", sc.AppId})
	if err != nil {
		return fail(err)
	}
	appInfoString := ""
	marker := fmt.Sprintf("\"%s\"", sc.AppId)
	lines := strings.Split(output, "\n")
	for index, line := range lines {
		if !strings.HasPrefix(line, marker) {
			continue
		}
		appInfoString = strings.Join(lines[index:], "\n")
		break
	}
	if appInfoString == "" {
		return fail(fmt.Errorf("data not found in steamcmd output"))
	}
	parser := vdf.NewParser(strings.NewReader(appInfoString))
	parsed, err := parser.Parse()
	if err != nil {
		return fail(err)
	}
	appInfo, ok := parsed[sc.AppId].(map[string]any)
	if !ok {
		return fail(fmt.Errorf("app id %s not found in app info", sc.AppId))
	}
	return appInfo, nil
}

// Parses steam app info to find the current manifest id for the given app id, depot id, branch name
// Returns an error if the manifest id cannot be found.
func (sc *SteamChecker) GetManifestId(appInfo map[string]any) (string, error) {
	fail := func(err error) (string, error) {
		return "", err
	}
	sc.Api.Logger.Info("get steam app manifest id", "app", sc.AppId, "branch", sc.BranchName, "depot", sc.DepotId)
	depots, ok := appInfo["depots"].(map[string]any)
	if !ok {
		return fail(fmt.Errorf("app info contains no depots"))
	}
	depot, ok := depots[sc.DepotId].(map[string]any)
	if !ok {
		return fail(fmt.Errorf("depot %s not found", sc.DepotId))
	}
	manifestsData, ok := depot["manifests"].(map[string]any)
	if !ok {
		return fail(fmt.Errorf("depot %s contains no manifests", sc.DepotId))
	}
	manifestData, ok := manifestsData[sc.BranchName].(map[string]any)
	if !ok {
		return fail(fmt.Errorf("depot %s does not contain branch %s", sc.DepotId, sc.BranchName))
	}
	manifestId, ok := manifestData["gid"].(string)
	if !ok {
		return fail(fmt.Errorf("depot %s, branch %s does not contain manifest gid", sc.DepotId, sc.BranchName))
	}
	return manifestId, nil
}

// Returns a data payload to be used with the autopublisher.
// Returns an error if the data payload cannot be found.
func (sc *SteamChecker) GetData() (map[string]any, error) {
	sc.Api.Logger.Info("get steam data")
	fail := func(err error) (map[string]any, error) {
		return map[string]any{}, err
	}
	appInfo, err := sc.GetAppInfo()
	if err != nil {
		return fail(err)
	}
	manifestId, err := sc.GetManifestId(appInfo)
	if err != nil {
		return fail(err)
	}
	return map[string]any{
		"ManifestId": manifestId,
	}, nil
}

// Initializes the [SteamChecker]
func (sc *SteamChecker) Initialize(api Api) error {
	sc.Api = api

	err := sc.Api.ParseEnv(sc)
	if err != nil {
		return err
	}

	return nil
}

// Autopublisher uses an [AutopublishChecker] and [AutopublishPublisher] to check for new versions and publish them.
type Autopublisher struct {
	Checker   AutopublishChecker
	Publisher AutopublishPublisher
}

// Runs the autopublish task using the configured [Autopublisher].
func (h *Helper) Autopublish(ctx context.Context, api Api) error {
	api.Logger.Info("autopublish")

	ap := h.Autopublisher
	if ap == (Autopublisher{}) {
		return fmt.Errorf("autopublisher not configured")
	}

	err := ap.Checker.Initialize(api)
	if err != nil {
		return err
	}

	err = ap.Publisher.Initialize(api)
	if err != nil {
		return err
	}

	data, err := ap.Checker.GetData()
	if err != nil {
		return err
	}

	err = ap.Publisher.Publish(data)
	if err != nil {
		return err
	}

	return nil
}
