package export

// func getBech32Prefix(rlpCfg roller.RollappConfig) (string, error) {
// 	rollappSeqAddrInfo, err := keys.GetAddressInfoBinary(
// 		keys.KeyConfig{
// 			Dir:         consts.ConfigDirName.Rollapp,
// 			ID:          consts.KeysIds.RollappSequencer,
// 			ChainBinary: consts.Executables.RollappEVM,
// 			Type:        "",
// 		}, rlpCfg.Home,
// 	)
// 	if err != nil {
// 		return "", err
// 	}
// 	return strings.Split(rollappSeqAddrInfo.Address, "1")[0], nil
// }
