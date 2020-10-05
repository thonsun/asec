/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:35:13
 * @Last Modified: thonsun, 2020-10-05  16:35:13
 */

package firewall

import (
	"asec/models"
)

// InitFirewall ...
func InitFirewall() {
	InitCCPolicy()
	ccPolicies.Range(func(key, value interface{}) bool {
		appID := key.(int64)
		ccPolicy := value.(*models.CCPolicy)
		if ccPolicy.IsEnabled == true {
			go CCAttackTick(appID)
		}
		return true
	})
	InitVulnType()
	InitGroupPolicy()
	LoadCheckItems()
	InitHitLog()
	InitNFTables()
	go RoutineCleanLogTick()
	go RoutineCleanCacheTick()
}
