/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:35:35
 * @Last Modified: thonsun, 2020-10-05  16:35:35
 */

package firewall

import (
	"encoding/json"

	"asec/data"
	"asec/models"
	"asec/utils"
)

// RPCSelectCCPolicies ...
func RPCSelectCCPolicies() (ccPolicies []*models.CCPolicy) {
	rpcRequest := &models.RPCRequest{
		Action: "getccpolicies", Object: nil}
	resp, err := data.GetRPCResponse(rpcRequest)
	if err != nil {
		utils.CheckError("RPCSelectCCPolicies GetResponse", err)
		return nil
	}
	rpcCCPolicies := new(models.RPCCCPolicies)
	if err := json.Unmarshal(resp, rpcCCPolicies); err != nil {
		utils.CheckError("RPCSelectCCPolicies Unmarshal", err)
		return nil
	}
	ccPolicies = rpcCCPolicies.Object
	return ccPolicies
}
