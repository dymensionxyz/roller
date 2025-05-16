package celestia

type BD struct {
	Height    string `yaml:"height"`
	StateRoot string `yaml:"stateRoot"`
}

type StateInfo struct {
	BDs struct {
		BD []BD `yaml:"BD"`
	} `yaml:"BDs"`
	DAPath         string `yaml:"DAPath"`
	CreationHeight string `yaml:"creationHeight"`
	NumBlocks      string `yaml:"numBlocks"`
	Sequencer      string `yaml:"sequencer"`
	StartHeight    string `yaml:"startHeight"`
	StateInfoIndex struct {
		Index     string `yaml:"index"`
		RollappId string `yaml:"rollappId"`
	} `yaml:"stateInfoIndex"`
	Status string `yaml:"status"`
}

type RollappStateResponse struct {
	StateInfo StateInfo `yaml:"stateInfo"`
}
