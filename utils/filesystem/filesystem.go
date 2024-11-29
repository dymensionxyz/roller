package filesystem

import (
	"encoding/json"
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

func CreateRollerRootWithOptionalOverride(home string) error {
	isRootExist, err := DirNotEmpty(home)
	if err != nil {
		return err
	}

	if isRootExist {
		fmt.Printf("Directory %s is not empty.\n", home)

		shouldOverwrite, err := pterm.DefaultInteractiveConfirm.WithDefaultText(fmt.Sprintf("Do you want to overwrite %s?", home)).
			WithDefaultValue(false).
			Show()
		if err != nil {
			return err
		}

		if shouldOverwrite {
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
// func DownloadFile(url, fp string) error {
// 	err := os.MkdirAll(filepath.Dir(fp), 0o755)
// 	if err != nil {
// 		return err
// 	}

// 	spinner, _ := pterm.DefaultSpinner.
// 		Start("Downloading ", filepath.Base(fp))
// 	fmt.Println()

// 	// nolint:gosec
// 	resp, err := http.Get(url)
// 	if err != nil || resp.StatusCode != http.StatusOK {
// 		// nolint:errcheck,gosec
// 		resp.Body.Close()
// 		spinner.Fail("failed to download file: ", err)
// 		return err
// 	}
// 	// nolint:errcheck
// 	defer resp.Body.Close()

// 	out, err := os.Create(fp)
// 	if err != nil {
// 		spinner.Fail("failed to download file: ", err)
// 		return err
// 	}
// 	// nolint:errcheck
// 	defer out.Close()

// 	spinner.Success(fmt.Sprintf("Successfully downloaded the %s", filepath.Base(fp)))
// 	_, err = io.Copy(out, resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func DownloadFile(url, fp string) error {
	err := os.MkdirAll(filepath.Dir(fp), 0o755)
	if err != nil {
		return err
	}

	spinner, _ := pterm.DefaultSpinner.
		Start("Accessing ", filepath.Base(fp))
	fmt.Println()

	// Check if the URL is a local file path
	if _, err := os.Stat(url); err == nil {
		// The URL is a local file; copy it to the target location
		sourceFile, err := os.Open(url)
		if err != nil {
			spinner.Fail("failed to access local file: ", err)
			return err
		}
		defer sourceFile.Close()

		out, err := os.Create(fp)
		if err != nil {
			spinner.Fail("failed to create output file: ", err)
			return err
		}
		defer out.Close()

		// Copy the local file to the target path
		_, err = io.Copy(out, sourceFile)
		if err != nil {
			spinner.Fail("failed to copy local file: ", err)
			return err
		}
		spinner.Success(fmt.Sprintf("Successfully accessed the local file %s", filepath.Base(fp)))
		return nil
	}

	// If not a local file, proceed with downloading from the URL
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		spinner.Fail("failed to download file: ", err)
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(fp)
	if err != nil {
		spinner.Fail("failed to create output file: ", err)
		return err
	}
	defer out.Close()

	spinner.Success(fmt.Sprintf("Successfully downloaded the %s", filepath.Base(fp)))
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Genesis json.RawMessage `json:"genesis"`
	} `json:"result"`
}

// func DownloadFile(url, fp string) error {
// 	// Create directories if not exist
// 	err := os.MkdirAll(filepath.Dir(fp), 0o755)
// 	if err != nil {
// 		return err
// 	}

// 	// Start the spinner
// 	spinner, _ := pterm.DefaultSpinner.Start("Downloading ", filepath.Base(fp))
// 	fmt.Println()

// 	// nolint:gosec
// 	resp, err := http.Get(url)
// 	if err != nil || resp.StatusCode != http.StatusOK {
// 		if resp != nil {
// 			// nolint:errcheck,gosec
// 			resp.Body.Close()
// 		}
// 		spinner.Fail("failed to download file: ", err)
// 		return err
// 	}
// 	// nolint:errcheck
// 	defer resp.Body.Close()

// 	// Read the response body into memory
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		spinner.Fail("failed to read response body: ", err)
// 		return err
// 	}

// 	// Check if the response is a JSON-RPC response
// 	var jsonResponse JSONRPCResponse
// 	writeData := body // Default to writing the whole body
// 	if err := json.Unmarshal(body, &jsonResponse); err == nil && jsonResponse.JSONRPC != "" {
// 		// JSON-RPC response detected
// 		if len(jsonResponse.Result.Genesis) > 0 {
// 			spinner.Success("Found genesis value and writing it to the file")
// 			writeData, err = json.MarshalIndent(jsonResponse.Result.Genesis, "", "  ") // Indent for readability

// 			if err != nil {
// 				spinner.Fail("failed to format genesis data: ", err)
// 				return err
// 			}
// 		} else {
// 			spinner.Warning("JSON-RPC response does not contain a 'genesis' value in 'result'")
// 		}
// 	} else {
// 		spinner.Success("Successfully downloaded the file")
// 	}

// 	// Add a newline to the end of writeData
// 	writeData = append(writeData, '\n')

// 	// Write the data to the file (either genesis value or full body)
// 	out, err := os.Create(fp)
// 	if err != nil {
// 		spinner.Fail("failed to create file: ", err)
// 		return err
// 	}
// 	// nolint:errcheck
// 	defer out.Close()

// 	// Write the data to the file
// 	_, err = out.Write(writeData)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

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
