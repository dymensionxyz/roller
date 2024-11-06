package scripts

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func writeTemplateToFile(tmpl *bytes.Buffer, fp string) error {
	cmd := exec.Command(
		"bash", "-c", fmt.Sprintf(
			"echo '%s' | sudo tee %s",
			tmpl.String(), fp,
		),
	)
	// Need to start and wait instead of run to allow sudo to prompt for password
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func writeDaTemplateToFile(home string) error {
	daDirPath := filepath.Join(home, consts.ConfigDirName.DALightNode)
	pswFp := filepath.Join(home, string(consts.OsKeyringPwdFileNames.Da))
	scriptFp := filepath.Join(home, string(consts.StartupScriptFilePaths.Da))

	daD := DaStartupTemplateData{
		PasswordFilePath: pswFp,
		Binary:           consts.Executables.Celestia,
		HomeDir:          daDirPath,
		StateNode:        "",
		KeyringBackend:   consts.SupportedKeyringBackends.OS,
	}

	raTmpl, err := generateDaStartupScript(daD)
	if err != nil {
		return err
	}

	err = writeTemplateToFile(raTmpl, scriptFp)
	if err != nil {
		return err
	}

	return nil
}

func writeRaTemplateToFile(home string) error {
	raDirPath := filepath.Join(home, consts.ConfigDirName.Rollapp)
	pswFp := filepath.Join(home, string(consts.OsKeyringPwdFileNames.RollApp))
	scriptFp := filepath.Join(home, string(consts.StartupScriptFilePaths.RollApp))

	raD := RaStartupTemplateData{
		PasswordFilePath: pswFp,
		Binary:           consts.Executables.RollappEVM,
		HomeDir:          raDirPath,
		KeyringBackend:   consts.SupportedKeyringBackends.OS,
	}

	raTmpl, err := generateRollappStartupScript(raD)
	if err != nil {
		return err
	}

	err = writeTemplateToFile(raTmpl, scriptFp)
	if err != nil {
		return err
	}

	return nil
}

func CreateRollappStartup(home string) error {
	err := writeDaTemplateToFile(home)
	if err != nil {
		return err
	}

	err = writeRaTemplateToFile(home)
	if err != nil {
		return err
	}

	return nil
}
