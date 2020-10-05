/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:39:15
 * @Last Modified: thonsun, 2020-10-05  16:39:15
 */

package models

type RPCSetGroupPolicy struct {
	Action string       `json:"action"`
	Object *GroupPolicy `json:"object"`
}
