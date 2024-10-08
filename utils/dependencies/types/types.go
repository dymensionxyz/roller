package types

import "os/exec"

type BinaryPathPair struct {
	Binary            string
	BinaryDestination string
	BuildCommand      *exec.Cmd
	BuildArgs         []string
}

type PersistFile struct {
	Source string
	Target string
}

type Dependency struct {
	DependencyName  string
	RepositoryOwner string
	RepositoryName  string
	RepositoryUrl   string
	Release         string
	Binaries        []BinaryPathPair
	PersistFiles    []PersistFile
}
