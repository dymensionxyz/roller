package migrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/version"
)

func RequireRollerMigrateIfNeeded(rlpCfg roller.RollappConfig) {
	currentRollerVersion := version.TrimVersionStr(version.BuildVersion)
	configRollerVersion := version.TrimVersionStr(rlpCfg.RollerVersion)

	if configRollerVersion != currentRollerVersion {
		//nolint:errcheck,gosec
		pterm.Warning.Printf(
			"💈 Your rollapp config version ('%s') is older than your"+
				" installed roller version ('%s'),"+
				" please run 'roller migrate' to update your config.\n", configRollerVersion, currentRollerVersion,
		)
	}

	os.Exit(1)
}

func RequireRollappMigrateIfNeeded(current, last, vmType string) error {
	if current == last {
		pterm.Info.Println("versions match")

		return nil
	}
	var currentVersionTimestamp time.Time
	var currentVersionCommit string
	var lastVersionTimestamp time.Time
	var err error
	var rollappType string

	// this is a good utility function
	switch vmType {
	case string(consts.WASM_ROLLAPP):
		rollappType = "rollapp-wasm"
	case string(consts.EVM_ROLLAPP):
		rollappType = "rollapp-evm"
	default:
		err = fmt.Errorf("invalid rollapp type: %s", vmType)
		return err
	}

	if strings.HasPrefix(current, "v") {
		currentVersionTimestamp, err = GetCommitTimestampByTag(
			"dymensionxyz",
			rollappType,
			current,
		)

		commit, err := GetCommitFromTag(
			"dymensionxyz",
			rollappType,
			current,
		)
		if err != nil {
			return err
		}

		currentVersionCommit = commit
	} else {
		currentVersionTimestamp, err = GetCommitTimestamp(
			"dymensionxyz",
			rollappType,
			current,
		)

		currentVersionCommit = current
	}
	if err != nil {
		return err
	}

	if strings.HasPrefix(last, "v") {
		lastVersionTimestamp, err = GetCommitTimestampByTag(
			"dymensionxyz",
			rollappType,
			last,
		)
	} else {
		lastVersionTimestamp, err = GetCommitTimestamp(
			"dymensionxyz",
			rollappType,
			last,
		)
	}

	if err != nil {
		return err
	}
	// end good utility function

	if lastVersionTimestamp.Before(currentVersionTimestamp) {
		msg := pterm.Sprintf(
			"The last used rollapp version commit (%s)"+
				" is older than the currently installed version commit (%s)"+
				" please run %s to migrate the binary and configuration vefore being able to start the RollApp\n",
			last,
			currentVersionCommit,
			pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
				Sprintf("roller rollapp migrate %s", currentVersionCommit),
		)

		pterm.Info.Println(msg)
	}

	return fmt.Errorf("a migration from %s to %s is required", last, currentVersionCommit)
}

func GetCommitTimestamp(owner, repo, sha string) (time.Time, error) {
	sha = strings.ToLower(sha)
	sha = strings.TrimSuffix(sha, "\n")
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, sha)
	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return time.Time{}, err
	}
	// nolint: errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("failed to fetch commit data: %s", resp.Status)
	}

	var commit Commit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return time.Time{}, err
	}

	return commit.Commit.Committer.Date, nil
}

func GetCommitFromTag(owner, repo, tag string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/tags/%s", owner, repo, tag)

	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// nolint: errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch tag data: %s", resp.Status)
	}

	var tagData Tag
	if err := json.NewDecoder(resp.Body).Decode(&tagData); err != nil {
		return "", err
	}

	return tagData.Commit.SHA, nil
}

func GetCommitTimestampByTag(owner, repo, tag string) (time.Time, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/tags/%s", owner, repo, tag)

	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return time.Time{}, err
	}
	// nolint: errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("failed to fetch tag data: %s", resp.Status)
	}

	var tagData Tag
	if err := json.NewDecoder(resp.Body).Decode(&tagData); err != nil {
		return time.Time{}, err
	}

	return GetCommitTimestamp(owner, repo, tagData.Commit.SHA)
}

// nolint: unused
func GetCommitTimestampByRelease(owner, repo, releaseTag string) (time.Time, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/tags/%s",
		owner,
		repo,
		releaseTag,
	)

	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return time.Time{}, err
	}
	// nolint: errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("failed to fetch release data: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return time.Time{}, err
	}

	return GetCommitTimestampByTag(owner, repo, release.TagName)
}
