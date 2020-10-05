/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:35:30
 * @Last Modified: thonsun, 2020-10-05  16:35:30
 */

package firewall

import (
	"io/ioutil"
	"os"
	"syscall"
	"time"

	"asec/data"
	"asec/utils"
)

// RoutineCleanLogTick Clear Expired Logs
func RoutineCleanLogTick() {
	if data.IsPrimary {
		logExpireSeconds, err := data.DAL.SelectIntSetting("Log_Expire_Seconds")
		utils.CheckError("RoutineTick", err)
		//fmt.Println("RoutineTick log_expire_seconds:", log_expire_seconds)
		routineTicker := time.NewTicker(time.Duration(5*60) * time.Second)
		for range routineTicker.C {
			//fmt.Println("RoutineTick", time.Now())
			expiredTime := time.Now().Unix() - logExpireSeconds
			data.DAL.DeleteHitLogsBeforeTime(expiredTime)
			data.DAL.DeleteCCLogsBeforeTime(expiredTime)
		}
	}
}

// RoutineCleanCacheTick Clean expired cdn files
func RoutineCleanCacheTick() {
	routineTicker := time.NewTicker(time.Duration(7200) * time.Second)
	for range routineTicker.C {
		ClearExpiredFiles("./static/cdncache/", time.Now())
	}
}

// ClearExpiredFiles clear expired static cdn files
func ClearExpiredFiles(path string, now time.Time) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		utils.DebugPrintln("ClearExpiredFiles", err)
	}
	for _, file := range fs {
		if file.IsDir() {
			ClearExpiredFiles(path+file.Name()+"/", now)
		} else {
			targetFile := path + file.Name()
			if fi, err := os.Stat(targetFile); err == nil {
				fiStat := fi.Sys().(*syscall.Stat_t)
				// Use ctime fiStat.Ctim.Sec to mark the last check time
				pastSeconds := now.Unix() - fiStat.Ctim.Sec
				if pastSeconds >= 86400*7 {
					err = os.Remove(targetFile)
					if err != nil {
						utils.DebugPrintln("ClearExpiredFiles Remove", targetFile, err)
					}
				}
			}
		}
	}
}
