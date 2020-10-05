/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:24:35
 * @Last Modified: thonsun, 2020-10-05  16:24:35
 */

package data

import (
	"asec/models"
	"asec/utils"
)

func (dal *MyDAL) UpdateDestinationNode(routeType int64, requestRoute string, backendRoute string, destination string, appID int64, nodeID int64, id int64) error {
	const sqlUpdateDestinationNode = `UPDATE destinations SET route_type=$1,request_route=$2,backend_route=$3,destination=$4,app_id=$5,node_id=$6 WHERE id=$7`
	stmt, err := dal.db.Prepare(sqlUpdateDestinationNode)
	defer stmt.Close()
	_, err = stmt.Exec(routeType, requestRoute, backendRoute, destination, appID, nodeID, id)
	utils.CheckError("UpdateDestinationNode", err)
	return err
}

func (dal *MyDAL) ExistsDestinationID(id int64) bool {
	var exist int
	const sqlExistsDestinationID = `SELECT coalesce((SELECT 1 FROM destinations WHERE id=$1 limit 1),0)`
	err := dal.db.QueryRow(sqlExistsDestinationID, id).Scan(&exist)
	utils.CheckError("ExistsDestinationID", err)
	if exist == 0 {
		return false
	} else {
		return true
	}
}

func (dal *MyDAL) CreateTableIfNotExistsDestinations() error {
	const sqlCreateTableIfNotExistsDestinations = `CREATE TABLE IF NOT EXISTS destinations(id bigserial PRIMARY KEY,route_type bigint default 1,request_route varchar(128) default '/',backend_route varchar(128) default '/',destination varchar(128) NOT NULL,app_id bigint NOT NULL,node_id bigint NOT NULL)`
	_, err := dal.db.Exec(sqlCreateTableIfNotExistsDestinations)
	return err
}

func (dal *MyDAL) SelectDestinationsByAppID(app_id int64) (dests []*models.Destination) {
	const sqlSelectDestinationsByAppID = `SELECT id,route_type,request_route,backend_route,destination,node_id FROM destinations WHERE app_id=$1`
	rows, err := dal.db.Query(sqlSelectDestinationsByAppID, app_id)
	utils.CheckError("SelectDestinationsByAppID", err)
	if err != nil {
		return dests
	}
	defer rows.Close()
	for rows.Next() {
		dest := &models.Destination{AppID: app_id}
		rows.Scan(&dest.ID, &dest.RouteType, &dest.RequestRoute, &dest.BackendRoute, &dest.Destination, &dest.NodeID)
		dests = append(dests, dest)
	}
	return dests
}

func (dal *MyDAL) InsertDestination(routeType int64, requestRoute string, backendRoute string, dest string, appID int64, nodeID int64) (newID int64, err error) {
	const sqlInsertDestination = `INSERT INTO destinations(route_type,request_route,backend_route,destination,app_id,node_id) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`
	err = dal.db.QueryRow(sqlInsertDestination, routeType, requestRoute, backendRoute, dest, appID, nodeID).Scan(&newID)
	utils.CheckError("InsertDestination", err)
	return newID, err
}

func (dal *MyDAL) DeleteDestinationByID(id int64) error {
	const sqlDeleteDestinationByID = `DELETE FROM destinations WHERE id=$1`
	stmt, err := dal.db.Prepare(sqlDeleteDestinationByID)
	defer stmt.Close()
	_, err = stmt.Exec(id)
	utils.CheckError("DeleteDestinationByID", err)
	return err
}

func (dal *MyDAL) DeleteDestinationsByAppID(appID int64) error {
	const sqlDeleteDestinationsByAppID = `DELETE FROM destinations WHERE app_id=$1`
	stmt, err := dal.db.Prepare(sqlDeleteDestinationsByAppID)
	defer stmt.Close()
	_, err = stmt.Exec(appID)
	utils.CheckError("DeleteDestinationsByAppID", err)
	return err
}
