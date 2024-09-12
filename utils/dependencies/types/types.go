package types

import "os/exec"

type BinaryPathPair struct {
	Binary            string
	BinaryDestination string
	BuildCommand      *exec.Cmd
	BuildArgs         []string
}

type Dependency struct {
	Name       string
	Repository string
	Release    string
	Binaries   []BinaryPathPair
}
