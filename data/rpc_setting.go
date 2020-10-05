/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:31:19
 * @Last Modified: thonsun, 2020-10-05  16:31:19
 */

package data

import (
	"encoding/json"

	"asec/models"
	"asec/utils"
)

func RPCGetSettings() []*models.Setting {
	rpcRequest := &models.RPCRequest{
		Action: "getsettings", Object: nil}
	resp, err := GetRPCResponse(rpcRequest)
	utils.CheckError("RPCGetSettings", err)
	rpcSettings := new(models.RPCSettings)
	if err = json.Unmarshal(resp, rpcSettings); err != nil {
		utils.CheckError("RPCGetSettings Unmarshal", err)
	}
	return rpcSettings.Object
}

func RPCGetOAuthConfig() *models.OAuthConfig {
	rpcRequest := &models.RPCRequest{
		Action: "getoauthconf", Object: nil}
	resp, err := GetRPCResponse(rpcRequest)
	utils.CheckError("RPCGetOAuthConfig", err)
	rpcOAuthConf := new(models.RPCOAuthConfig)
	if err = json.Unmarshal(resp, rpcOAuthConf); err != nil {
		utils.CheckError("RPCGetOAuthConfig Unmarshal", err)
	}
	//fmt.Println("RPCGetOAuthConfig", rpcOAuthConf.Object)
	return rpcOAuthConf.Object
}
