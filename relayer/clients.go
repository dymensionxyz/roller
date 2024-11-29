package relayer

// func (r *Relayer) CheckClientsExist() (bool, error) {
// 	rlyCfg, err := ReadRlyConfig(r.RollerHome)
// 	if err != nil {
// 		return false, err
// 	}
// 	clientIDRollapp_raw, err := roller_utils.GetNestedValue(
// 		rlyCfg,
// 		[]string{"paths", consts.DefaultRelayerPath, "dst", "client-id"},
// 	)
// 	if err != nil {
// 		return false, err
// 	}
//
// 	clientIDHub_raw, err := roller_utils.GetNestedValue(
// 		rlyCfg,
// 		[]string{"paths", consts.DefaultRelayerPath, "src", "client-id"},
// 	)
// 	if err != nil {
// 		return false, err
// 	}
//
// 	//nolint:errcheck
// 	clientIDRollapp := clientIDRollapp_raw.(string)
// 	//nolint:errcheck
// 	clientIDHub := clientIDHub_raw.(string)
//
// 	if clientIDRollapp == "" || clientIDHub == "" {
// 		r.logger.Printf("can't find clients in the config for both rollapp and hub")
// 		return false, nil
// 	}
// 	return true, nil
// }
