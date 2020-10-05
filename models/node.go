/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:39:09
 * @Last Modified: thonsun, 2020-10-05  16:39:09
 */

package models

type Node struct {
	ID              int64  `json:"id"`
	Version         string `json:"version"`
	LastIP          string `json:"last_ip"`
	LastRequestTime int64  `json:"last_req_time"`
}

type DBNode struct {
	ID              int64  `json:"id"`
	Version         string `json:"version"`
	LastIP          string `json:"last_ip"`
	LastRequestTime int64  `json:"last_req_time"`
}

type NodeAuth struct {
	CurTime int64 `json:"cur_time"`
}

type NodesKey struct {
	HexEncryptedKey string `json:"nodes_key"`
}
