/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:25:35
 * @Last Modified: thonsun, 2020-10-05  16:25:35
 */

package data

import (
	"time"

	"asec/models"
	"asec/utils"
)

const (
	sqlCreateTableIfNotExistsCCPolicy = `CREATE TABLE IF NOT EXISTS ccpolicies(app_id bigint primary key,interval_milliseconds bigint,max_count bigint,block_seconds bigint,action bigint,stat_by_url boolean,stat_by_ua boolean,stat_by_cookie boolean,is_enabled boolean)`
	sqlExistsCCPolicy                 = `SELECT coalesce((SELECT 1 FROM ccpolicies LIMIT 1),0)`
	sqlExistsCCPolicyByAppID          = `SELECT coalesce((SELECT 1 FROM ccpolicies WHERE app_id=$1 LIMIT 1),0)`
	sqlInsertCCPolicy                 = `INSERT INTO ccpolicies(app_id,interval_milliseconds,max_count,block_seconds,action,stat_by_url,stat_by_ua,stat_by_cookie,is_enabled) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	sqlSelectCCPolicies               = `SELECT app_id,interval_milliseconds,max_count,block_seconds,action,stat_by_url,stat_by_ua,stat_by_cookie,is_enabled FROM ccpolicies`
	sqlUpdateCCPolicy                 = `UPDATE ccpolicies SET interval_milliseconds=$1,max_count=$2,block_seconds=$3,action=$4,stat_by_url=$5,stat_by_ua=$6,stat_by_cookie=$7,is_enabled=$8 where app_id=$9`
	sqlDeleteCCPolicy                 = `DELETE FROM ccpolicies WHERE app_id=$1`
)

func (dal *MyDAL) CreateTableIfNotExistsCCPolicy() error {
	_, err := dal.db.Exec(sqlCreateTableIfNotExistsCCPolicy)
	return err
}

func (dal *MyDAL) DeleteCCPolicy(appID int64) error {
	stmt, err := dal.db.Prepare(sqlDeleteCCPolicy)
	defer stmt.Close()
	_, err = stmt.Exec(appID)
	utils.CheckError("DeleteCCPolicy", err)
	return err
}

func (dal *MyDAL) UpdateCCPolicy(IntervalMilliSeconds time.Duration, maxCount int64,
	blockSeconds time.Duration, action models.PolicyAction,
	statByUrl bool, statByUA bool, statByCookie bool, isEnabled bool, appID int64) error {
	stmt, err := dal.db.Prepare(sqlUpdateCCPolicy)
	defer stmt.Close()
	_, err = stmt.Exec(IntervalMilliSeconds, maxCount, blockSeconds, action,
		statByUrl, statByUA, statByCookie, isEnabled, appID)
	utils.CheckError("UpdateCCPolicy", err)
	return err
}

func (dal *MyDAL) ExistsCCPolicy() bool {
	var existCCPolicy int
	err := dal.db.QueryRow(sqlExistsCCPolicy).Scan(&existCCPolicy)
	utils.CheckError("ExistsCCPolicy", err)
	if existCCPolicy == 0 {
		return false
	} else {
		return true
	}
}

func (dal *MyDAL) ExistsCCPolicyByAppID(appID int64) bool {
	var existCCPolicy int
	err := dal.db.QueryRow(sqlExistsCCPolicyByAppID, appID).Scan(&existCCPolicy)
	utils.CheckError("ExistsCCPolicyByAppID", err)
	if existCCPolicy == 0 {
		return false
	} else {
		return true
	}
}

func (dal *MyDAL) InsertCCPolicy(appID int64, IntervalMilliSeconds time.Duration, maxCount int64, blockSeconds time.Duration,
	action models.PolicyAction, statByUrl bool, statByUA bool, statByCookie bool, isEnabled bool) error {
	_, err := dal.db.Exec(sqlInsertCCPolicy, appID, IntervalMilliSeconds, maxCount, blockSeconds,
		action, statByUrl, statByUA, statByCookie, isEnabled)
	utils.CheckError("InsertCCPolicy", err)
	return err
}

func (dal *MyDAL) SelectCCPolicies() (ccPolicies []*models.CCPolicy) {
	rows, err := dal.db.Query(sqlSelectCCPolicies)
	utils.CheckError("SelectCCPolicies", err)
	defer rows.Close()
	for rows.Next() {
		ccPolicy := new(models.CCPolicy)
		rows.Scan(&ccPolicy.AppID, &ccPolicy.IntervalMilliSeconds, &ccPolicy.MaxCount, &ccPolicy.BlockSeconds,
			&ccPolicy.Action, &ccPolicy.StatByURL, &ccPolicy.StatByUserAgent, &ccPolicy.StatByCookie, &ccPolicy.IsEnabled)
		ccPolicies = append(ccPolicies, ccPolicy)
	}
	return ccPolicies
}
