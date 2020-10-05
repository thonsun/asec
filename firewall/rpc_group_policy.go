/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:35:42
 * @Last Modified: thonsun, 2020-10-05  16:35:42
 */

package firewall

import (
	"encoding/json"

	"asec/data"
	"asec/models"
	"asec/utils"
)

// RPCSelectGroupPolicies ...
func RPCSelectGroupPolicies() (groupPolicies []*models.GroupPolicy) {
	rpcRequest := &models.RPCRequest{
		Action: "getgrouppolicies", Object: nil}
	resp, err := data.GetRPCResponse(rpcRequest)
	if err != nil {
		utils.CheckError("RPCSelectGroupPolicies GetResponse", err)
		return nil
	}
	rpcGroupPolicies := new(models.RPCGroupPolicies)
	if err := json.Unmarshal(resp, rpcGroupPolicies); err != nil {
		utils.CheckError("RPCSelectGroupPolicies Unmarshal", err)
		return nil
	}
	groupPolicies = rpcGroupPolicies.Object
	return groupPolicies
}
