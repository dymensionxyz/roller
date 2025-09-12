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
	"strings"

	"github.com/nxadm/tail"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
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

func CreateRollerRootWithOptionalOverride(home string, forceOverwrite bool) error {
	isRootExist, err := DirNotEmpty(home)
	if err != nil {
		return err
	}

	if isRootExist {
		fmt.Printf("Directory %s is not empty.\n", home)

		var shouldOverwrite bool
		if !forceOverwrite {
			shouldOverwrite, err = pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Do you want to overwrite %s?", home)).
				WithDefaultValue(false).
				Show()
			if err != nil {
				return err
			}
		}

		if shouldOverwrite || forceOverwrite {
			err = os.RemoveAll(home)
			if err != nil {
				return err
			}

			err = RemoveServiceFiles(consts.RollappSystemdServices)
			if err != nil {
				return err
			}

			err = os.MkdirAll(home, 0o755)
			if err != nil {
				return err
			}
		} else {
			pterm.Info.Println("cancelled by user")
			os.Exit(0)
		}
	} else {
		err = os.MkdirAll(home, 0o755)
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

func DownloadGenesisFile(genesisUrl, destinationPath string) error {
	if genesisUrl == "" {
		return fmt.Errorf("RollApp's genesis url field is empty, contact the rollapp owner")
	}

	// Check for override via environment variable
	if override := os.Getenv("ROLLER_GENESIS_OVERRIDE"); override != "" {
		genesisUrl = override
		pterm.Warning.Printf("Using genesis override from ROLLER_GENESIS_OVERRIDE: %s\n", override)
	}

	// Support local files
	if strings.HasPrefix(genesisUrl, "file://") || strings.HasPrefix(genesisUrl, "/") {
		localPath := strings.TrimPrefix(genesisUrl, "file://")

		input, err := os.ReadFile(localPath)
		if err != nil {
			return fmt.Errorf("failed to read local genesis file: %w", err)
		}

		err = os.MkdirAll(filepath.Dir(destinationPath), 0o755)
		if err != nil {
			return err
		}

		err = os.WriteFile(destinationPath, input, 0o644)
		if err != nil {
			return fmt.Errorf("failed to write genesis file: %w", err)
		}

		pterm.Success.Printf("Copied local genesis file from %s\n", localPath)
		return nil
	}

	// Original HTTP download
	return DownloadFile(genesisUrl, destinationPath)
}

func RemoveFileIfExists(filePath string) error {
	_, err := os.Stat(filePath)
	if err == nil {
		c := exec.Command("sudo", "rm", "-rf", filePath)
		err := c.Run()
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
		pterm.Info.Printfln("%s does not exist", path)
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// ReadFromFile reads the contents of a file and returns it as a string
func ReadFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	// nolint:errcheck
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), err
}

func GetOsKeyringPswFileName(command string) (consts.OsKeyringPwdFileName, error) {
	var pswFileName consts.OsKeyringPwdFileName
	switch command {
	case consts.Executables.Celestia:
		pswFileName = consts.OsKeyringPwdFileNames.Da
	case consts.Executables.CelKey:
		pswFileName = consts.OsKeyringPwdFileNames.Da
	case consts.Executables.RollappEVM:
		pswFileName = consts.OsKeyringPwdFileNames.RollApp
	case consts.Executables.Dymension:
		pswFileName = consts.OsKeyringPwdFileNames.RollApp
	default:
		return "", fmt.Errorf("unsupported command: %s", command)
	}
	return pswFileName, nil
}

func ReadOsKeyringPswFile(home, command string) (string, error) {
	pswFileName, err := GetOsKeyringPswFileName(command)
	if err != nil {
		pterm.Error.Println("failed to get os keyring psw file name", err)
		return "", err
	}
	fp := filepath.Join(home, string(pswFileName))
	psw, err := ReadFromFile(fp)
	if err != nil {
		pterm.Error.Println("failed to read keyring passphrase file", err)
		return "", err
	}

	return psw, nil
}
