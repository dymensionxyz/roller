package filesystem

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/nxadm/tail"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func DirNotEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if !info.IsDir() {
		return false, fmt.Errorf("%s is not a directory", path)
	}

	files, err := os.ReadDir(path)
	return len(files) > 0, err
}

func CreateDirWithOptionalOverwrite(path string) error {
	isRootExist, err := DirNotEmpty(path)
	if err != nil {
		return err
	}

	if isRootExist {
		msg := fmt.Sprintf("Directory %s is not empty. Do you want to overwrite it?", path)
		shouldOverwrite, err := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).
			WithDefaultValue(false).
			Show()
		if err != nil {
			return err
		}

		if shouldOverwrite {
			err = os.RemoveAll(path)
			if err != nil {
				return err
			}

			err = RemoveServiceFiles(consts.RollappSystemdServices)
			if err != nil {
				return err
			}

			err = os.MkdirAll(path, 0o755)
			if err != nil {
				return err
			}
		} else {
			pterm.Info.Println("cancelled by user")
			os.Exit(0)
		}
	} else {
		err = os.MkdirAll(path, 0o755)
		if err != nil {
			return err
		}
	}

	return nil
}

func MoveFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()
	// nolint:gofumpt
	err = os.MkdirAll(filepath.Dir(dst), 0o750)
	if err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed to delete source file: %w", err)
	}
	return nil
}

func ExpandHomePath(path string) (string, error) {
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path, nil
}

// TODO: download the file in chunks if possible
func DownloadFile(url, fp string) error {
	err := os.MkdirAll(filepath.Dir(fp), 0o755)
	if err != nil {
		return err
	}

	spinner, _ := pterm.DefaultSpinner.
		Start("Downloading ", filepath.Base(fp))
	fmt.Println()

	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		// nolint:errcheck,gosec
		resp.Body.Close()
		spinner.Fail("failed to download file: ", err)
		return err
	}
	// nolint:errcheck
	defer resp.Body.Close()

	out, err := os.Create(fp)
	if err != nil {
		spinner.Fail("failed to download file: ", err)
		return err
	}
	// nolint:errcheck
	defer out.Close()

	spinner.Success(fmt.Sprintf("Successfully downloaded the %s", filepath.Base(fp)))
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func RemoveFileIfExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		c := exec.Command("sudo", "rm", "-rf", filePath)
		_, err := bash.ExecCommandWithStdout(c)
		if err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		fmt.Printf("File %s has been removed\n", filePath)
	} else if os.IsNotExist(err) {
		fmt.Printf("File %s does not exist\n", filePath)
	} else {
		return fmt.Errorf("error checking file: %w", err)
	}
	return nil
}

func TailFile(fp, svcName string, lineNumber int) error {
	tailCfg := tail.Config{
		Follow: true,
		ReOpen: false,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		},
	}

	t, err := tail.TailFile(fp, tailCfg)
	if err != nil {
		return fmt.Errorf("failed to tail file: %v", err)
	}

	infoPrefix := pterm.Info.Prefix
	infoPrefix.Text = svcName
	cp := pterm.PrefixPrinter{
		Prefix: infoPrefix,
	}

	for i := 0; i < lineNumber; i++ {
		<-t.Lines
	}

	for line := range t.Lines {
		cp.Println(line.Text)
	}

	return nil
}

func DoesFileExist(path string) (bool, error) {
	_, err := os.Stat(path)

	if errors.Is(err, fs.ErrNotExist) {
		pterm.Info.Println("existing roller configuration not found")
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
