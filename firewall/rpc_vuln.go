/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:35:55
 * @Last Modified: thonsun, 2020-10-05  16:35:55
 */

package firewall

import (
	"encoding/json"

	"asec/data"
	"asec/models"
	"asec/utils"
)

// RPCSelectVulntypes ...
func RPCSelectVulntypes() (vulnTypes []*models.VulnType) {
	rpcRequest := &models.RPCRequest{
		Action: "getvulntypes", Object: nil}
	resp, err := data.GetRPCResponse(rpcRequest)
	if err != nil {
		utils.CheckError("RPCSelectVulntypes GetResponse", err)
		return nil
	}
	rpcVulnTypes := new(models.RPCVulntypes)
	if err := json.Unmarshal(resp, rpcVulnTypes); err != nil {
		utils.CheckError("RPCSelectVulntypes Unmarshal", err)
		return nil
	}
	vulnTypes = rpcVulnTypes.Object
	return vulnTypes
}
