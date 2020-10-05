/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:23:23
 * @Last Modified: thonsun, 2020-10-05  16:23:23
 */

package backend

import (
	"encoding/json"

	"asec/data"
	"asec/models"
	"asec/utils"
)

func RPCSelectDomains() (dbDomains []*models.DBDomain) {
	rpcRequest := &models.RPCRequest{
		Action: "getdomains", Object: nil}
	resp, err := data.GetRPCResponse(rpcRequest)
	if err != nil {
		utils.CheckError("RPCSelectDomains GetResponse", err)
		return nil
	}
	rpcDBDomains := new(models.RPCDBDomains)
	err = json.Unmarshal(resp, rpcDBDomains)
	if err != nil {
		utils.CheckError("RPCSelectDomains Unmarshal", err)
		return nil
	}
	dbDomains = rpcDBDomains.Object
	return dbDomains
}
