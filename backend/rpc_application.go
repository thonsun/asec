/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:23:12
 * @Last Modified: thonsun, 2020-10-05  16:23:12
 */

package backend

import (
	"encoding/json"

	"asec/data"
	"asec/models"
	"asec/utils"
)

func RPCSelectApplications() (apps []*models.Application) {
	rpcRequest := &models.RPCRequest{Action: "getapps", Object: nil}
	resp, err := data.GetRPCResponse(rpcRequest)
	if err != nil {
		utils.CheckError("RPCSelectApplications GetResponse", err)
		return nil
	}
	rpcApps := new(models.RPCApplications)
	err = json.Unmarshal(resp, rpcApps)
	if err != nil {
		utils.CheckError("RPCSelectApplications Unmarshal", err)
		return nil
	}
	applications := rpcApps.Object
	return applications
}
