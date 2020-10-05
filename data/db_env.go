/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:25:23
 * @Last Modified: thonsun, 2020-10-05  16:25:23
 */

package data

import (
	"asec/utils"
)

const (
	sqlSetIDSeqStartWith = `SELECT setval($1,$2,false)`
)

func (dal *MyDAL) SetIDSeqStartWith(tableName string, seq int64) error {
	tableIDSeq := tableName + `_id_seq`
	_, err := dal.db.Exec(sqlSetIDSeqStartWith, tableIDSeq, seq)
	utils.CheckError("SetIDSeqStartWith", err)
	return err
}
