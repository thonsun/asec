/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:35:48
 * @Last Modified: thonsun, 2020-10-05  16:35:48
 */

package firewall

import (
	"asec/data"
	"asec/models"
	"asec/utils"
)

// RPCGroupHitLog ...
func RPCGroupHitLog(regexHitLog *models.GroupHitLog) {
	rpcRequest := &models.RPCRequest{
		Action: "log_group_hit", Object: regexHitLog}
	_, err := data.GetRPCResponse(rpcRequest)
	utils.CheckError("RPCRegexHitLog", err)
}

// RPCCCLog ...
func RPCCCLog(ccLog *models.CCLog) {
	rpcRequest := &models.RPCRequest{
		Action: "log_cc", Object: ccLog}
	_, err := data.GetRPCResponse(rpcRequest)
	utils.CheckError("RPCCCLog", err)
}
