/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:23:50
 * @Last Modified: thonsun, 2020-10-05  16:23:50
 */

package data

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"asec/models"
	"asec/utils"

	_ "github.com/lib/pq"
)

// MyDAL used for data access layer
type MyDAL struct {
	db *sql.DB
}

var (
	// DAL is Data Access Layer
	DAL *MyDAL
	// CFG is config
	CFG *models.Config
	// IsPrimary i.e. Is Primary Node
	IsPrimary bool
	// Version of asec
	Version = "1.0.0"
	// NodeKey share with all nodes
	NodeKey []byte
)

// InitDAL init Data Access Layer
func InitDAL() {
	DAL = new(MyDAL)
	var err error
	CFG, err = NewConfig("./config.json")
	utils.CheckError("InitDAL", err)
	if err != nil {
		os.Exit(1)
	}
	nodeRole := strings.ToLower(CFG.NodeRole)
	if nodeRole != "primary" && nodeRole != "replica" {
		fmt.Printf("Error: node_role %s is not supported, it should be primary or replica, please check config.json \n", nodeRole)
		utils.DebugPrintln("Error: node_role ", nodeRole, " is not supported, it should be primary or replica, please check config.json")
		os.Exit(1)
	}
	IsPrimary = (nodeRole == "primary")
	if IsPrimary {
		conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			CFG.PrimaryNode.Database.Host,
			CFG.PrimaryNode.Database.Port,
			CFG.PrimaryNode.Database.User,
			CFG.PrimaryNode.Database.Password,
			CFG.PrimaryNode.Database.DBName)
		DAL.db, err = sql.Open("postgres", conn)
		utils.CheckError("InitDAL sql.Open:", err)
		if err != nil {
			os.Exit(1)
		}
		// Check if the User and Password are Correct
		_, err = DAL.db.Query("select 1")
		utils.CheckError("InitDAL Failed, Error:", err)
		if err != nil {
			utils.DebugPrintln("InitDAL Failed, Please check the database user and password. Error:", err)
			os.Exit(1)
		}

		// Database user and password OK
		DAL.db.SetMaxOpenConns(99)
	} else {
		// Init Node Key (Share with primary node)
		NodeKey = NodeHexKeyToCryptKey(CFG.ReplicaNode.NodeKey)
	}
}

func (dal *MyDAL) ExecSQL(sql string) error {
	_, err := dal.db.Exec(sql)
	return err
}

func (dal *MyDAL) ExistColumnInTable(tableName string, columnName string) bool {
	var count int64
	const sql = `select count(1) from information_schema.columns where table_name=$1 and column_name=$2`
	err := dal.db.QueryRow(sql, tableName, columnName).Scan(&count)
	utils.CheckError("ExistColumnInTable QueryRow", err)
	if count > 0 {
		return true
	}
	return false
}
