package dependencies

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	firebaseutils "github.com/dymensionxyz/roller/utils/firebase"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/rollapp"
)

type RollappBinaryInfo struct {
	Bech32Prefix string
	Commit       string
	VMType       string
	DA           string
}

func NewRollappBinaryInfo(bech32Prefix, commit, vmType, da string) RollappBinaryInfo {
	return RollappBinaryInfo{
		Bech32Prefix: bech32Prefix,
		Commit:       commit,
		VMType:       vmType,
		DA:           da,
	}
}

func GetMockDependencies(raVmType string) (map[string]types.Dependency, map[string]types.Dependency, error) {
	goreleaserDeps := map[string]types.Dependency{}

	// @20240913 libwasm is necessary on the host VM to be able to run the prebuilt rollapp binary
	var outputPath string
	var libName string
	libVersion := "v1.2.3"

	if runtime.GOOS == "linux" {
		outputPath = "/usr/lib"
		if runtime.GOARCH == "arm64" {
			libName = "libwasmvm.aarch64.so"
		} else if runtime.GOARCH == "amd64" {
			libName = "libwasmvm.x86_64.so"
		}
	} else if runtime.GOOS == "darwin" {
		outputPath = "/usr/local/lib"
		libName = "libwasmvm.dylib"
	} else {
		return nil, nil, errors.New("unsupported OS")
	}

	downloadPath := fmt.Sprintf(
		"https://github.com/CosmWasm/wasmvm/releases/download/%s/%s",
		libVersion,
		libName,
	)

	fsc := exec.Command("sudo", "mkdir", "-p", outputPath)
	_, err := bash.ExecCommandWithStdout(fsc)
	if err != nil {
		return nil, nil, err
	}

	c := exec.Command("sudo", "wget", "-O", filepath.Join(outputPath, libName), downloadPath)
	_, err = bash.ExecCommandWithStdout(c)
	if err != nil {
		return nil, nil, err
	}

	if raVmType == "evm" {
		goreleaserDeps["rollapp"] = types.Dependency{
			DependencyName:  "rollapp-evm",
			RepositoryOwner: "artemijspavlovs",
			RepositoryName:  "rollapp-evm",
			RepositoryUrl:   "https://github.com/artemijspavlovs/rollapp-evm",
			Release:         "v2.3.4-pg-roller-02",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "rollappd",
					BinaryDestination: consts.Executables.RollappEVM,
				},
			},
		}
	} else if raVmType == "wasm" {
		goreleaserDeps["rollapp"] = types.Dependency{
			DependencyName:  "rollapp-wasm",
			RepositoryOwner: "artemijspavlovs",
			RepositoryName:  "rollapp-wasm",
			RepositoryUrl:   "https://github.com/artemijspavlovs/rollapp-wasm",
			Release:         "v1.0.0-rc04-roller-07",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "rollappd",
					BinaryDestination: consts.Executables.RollappEVM,
				},
			},
		}
	}

	return nil, goreleaserDeps, nil
}

func GetRollappBinaryInfo(rollapp rollapp.Rollapp) (RollappBinaryInfo, error) {
	genesisTmpDir, err := os.MkdirTemp(os.TempDir(), "genesis-file")
	if err != nil {
		return RollappBinaryInfo{}, err
	}
	// nolint: errcheck
	defer os.RemoveAll(genesisTmpDir)

	// TODO refactor, this genesis file fetch is redundand and will slow the process down
	// when the genesis file is big
	err = genesisutils.DownloadGenesis(genesisTmpDir, rollapp.Metadata.GenesisUrl)
	if err != nil {
		pterm.Error.Println("failed to download genesis file: ", err)
		return RollappBinaryInfo{}, err
	}

	as, err := genesisutils.GetGenesisAppState(genesisTmpDir)
	if err != nil {
		return RollappBinaryInfo{}, err
	}

	drsVersion := strconv.Itoa(as.RollappParams.Params.DrsVersion)
	pterm.Info.Println("RollApp drs version from the genesis file : ", drsVersion)
	drsInfo, err := firebaseutils.GetLatestDrsVersionCommit(drsVersion)
	if err != nil {
		return RollappBinaryInfo{}, err
	}

	daType := strings.ToLower(as.RollappParams.Params.Da)
	raVmType := strings.ToLower(rollapp.VmType)
	var raCommit string
	switch raVmType {
	case "evm":
		raCommit = drsInfo.EvmCommit
	case "wasm":
		raCommit = drsInfo.WasmCommit
	}

	pterm.Info.Println(
		"Latest RollApp binary commit for the current DRS version: ",
		raCommit[:6],
	)

	rbi := NewRollappBinaryInfo(
		rollapp.GenesisInfo.Bech32Prefix,
		raCommit,
		raVmType,
		daType,
	)

	return rbi, nil
}

func GetRollappDependencies(rollapp rollapp.Rollapp) (map[string]types.Dependency, map[string]types.Dependency, error) {
	rbi, err := GetRollappBinaryInfo(rollapp)
	if err != nil {
		return nil, nil, err
	}

	buildableDeps := DefaultRollappBuildableDependencies(rbi)
	goreleaserDeps := DefaultRollappPrebuiltDependencies(rbi)

	return buildableDeps, goreleaserDeps, nil
}

func DefaultRollappBuildableDependencies(raBinInfo RollappBinaryInfo) map[string]types.Dependency {
	deps := map[string]types.Dependency{}
	deps["rollapp"] = DefaultRollappDependency(raBinInfo)

	if raBinInfo.DA == "celestia" {
		deps["celestia"] = DefaultCelestiaNodeDependency()
		deps["celestia-app"] = types.Dependency{
			DependencyName: "celestia-app",
			RepositoryUrl:  "https://github.com/celestiaorg/celestia-app",
			Release:        "v2.1.2",
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "celestia-appd",
					BinaryDestination: consts.Executables.CelestiaApp,
					BuildCommand: exec.Command(
						"make",
						"build",
					),
				},
			},
		}
	}

	return deps
}

func DefaultRollappDependency(raBinInfo RollappBinaryInfo) types.Dependency {
	var dep types.Dependency

	switch raBinInfo.VMType {
	case "evm":
		dep = types.Dependency{
			DependencyName:  "rollapp",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "rollapp-evm",
			RepositoryUrl:   "https://github.com/dymensionxyz/rollapp-evm.git",
			Release:         raBinInfo.Commit,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/rollapp-evm",
					BinaryDestination: consts.Executables.RollappEVM,
					BuildCommand: exec.Command(
						"make",
						"build",
						fmt.Sprintf("BECH32_PREFIX=%s", raBinInfo.Bech32Prefix),
					),
				},
			},
		}
	case "wasm":
		dep = types.Dependency{
			DependencyName:  "rollapp",
			RepositoryOwner: "dymensionxyz",
			RepositoryName:  "rollapp-wasm",
			RepositoryUrl:   "https://github.com/dymensionxyz/rollapp-wasm.git",
			Release:         raBinInfo.Commit,
			Binaries: []types.BinaryPathPair{
				{
					Binary:            "./build/rollapp-wasm",
					BinaryDestination: consts.Executables.RollappEVM,
					BuildCommand: exec.Command(
						"make",
						"build",
						fmt.Sprintf("BECH32_PREFIX=%s", raBinInfo.Bech32Prefix),
					),
				},
			},
		}
	default:
		pterm.Warning.Println("unsupported VM type")
	}

	return dep
}

func DefaultRollappPrebuiltDependencies(rbi RollappBinaryInfo) map[string]types.Dependency {
	return map[string]types.Dependency{}
}
