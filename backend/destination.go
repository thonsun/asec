/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:21:54
 * @Last Modified: thonsun, 2020-10-05  16:21:54
 */

package backend

//"../models"

func InterfaceContainsDestinationID(destinations []interface{}, destID int64) bool {
	for _, destination := range destinations {
		destMap := destination.(map[string]interface{})
		id := int64(destMap["id"].(float64))
		if id == destID {
			return true
		}
	}
	return false
}
