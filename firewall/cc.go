/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:33:22
 * @Last Modified: thonsun, 2020-10-05  16:33:22
 */

package firewall

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"asec/data"
	"asec/models"
)

var (
	ccPoliciesList []*models.CCPolicy
	ccPolicies     sync.Map //map[int64]*models.CCPolicy // key: appID==0  Global Policy
	ccCounts       sync.Map //map[int64]*(map[string]*models.ClientStat) // appID, clientID, ClientStat
	ccTickers      sync.Map //map[int64]*time.Ticker
)

// ClearCCStatByClientID clear CC stat by client id
func ClearCCStatByClientID(policyAppID int64, clientID string) {
	if ccCount, ok := ccCounts.Load(policyAppID); ok {
		appCCCount := ccCount.(*sync.Map)
		appCCCount.Delete(clientID)
	}
}

// CCAttackTick CC tick
func CCAttackTick(appID int64) {
	if appCCTicker, ok := ccTickers.Load(appID); ok {
		ccTicker := appCCTicker.(*time.Ticker)
		ccTicker.Stop()
	}
	ccPolicyMap, _ := ccPolicies.Load(appID)
	ccPolicy := ccPolicyMap.(*models.CCPolicy)
	ccTicker := time.NewTicker(ccPolicy.IntervalMilliSeconds * time.Millisecond)

	ccTickers.Store(appID, ccTicker)
	for range ccTicker.C {
		ccCount, _ := ccCounts.LoadOrStore(appID, &sync.Map{})
		//fmt.Println("CCAttackTick AppID=", appID, time.Now())
		appCCCount := ccCount.(*sync.Map)
		appCCCount.Range(func(key, value interface{}) bool {
			clientID := key.(string)
			stat := value.(*models.ClientStat)
			//fmt.Println("CCAttackTick:", appID, clientID, stat)
			if stat.IsBadIP == true {
				stat.RemainSeconds -= ccPolicy.IntervalMilliSeconds
				if stat.RemainSeconds <= 0 {
					appCCCount.Delete(clientID)
				}
			} else if stat.Count >= ccPolicy.MaxCount {
				stat.Count = 0
				stat.IsBadIP = true
				stat.RemainSeconds = ccPolicy.BlockSeconds
			} else {
				appCCCount.Delete(clientID)
			}
			return true
		})
	}
}

// GetCCPolicyByAppID get CC policy by app id
func GetCCPolicyByAppID(appID int64) *models.CCPolicy {
	if ccPolicy, ok := ccPolicies.Load(appID); ok {
		return ccPolicy.(*models.CCPolicy)
	}
	ccPolicy, _ := ccPolicies.Load(int64(0))
	return ccPolicy.(*models.CCPolicy)
}

// GetCCPolicies get all CC policies
func GetCCPolicies() ([]*models.CCPolicy, error) {
	return ccPoliciesList, nil
}

// GetCCPolicyRespByAppID get CC policy by app id
func GetCCPolicyRespByAppID(appID int64) (*models.CCPolicy, error) {
	ccPolicy := GetCCPolicyByAppID(appID)
	return ccPolicy, nil
}

// IsCCAttack to judge a request is CC attack
func IsCCAttack(r *http.Request, appID int64, srcIP string) (bool, *models.CCPolicy, string, bool) {
	ccPolicy := GetCCPolicyByAppID(appID)
	if ccPolicy.IsEnabled == false {
		return false, nil, "", false
	}
	if ccPolicy.AppID == 0 {
		appID = 0 // Important: stat within general policy
	}
	ccCount, _ := ccCounts.LoadOrStore(appID, &sync.Map{})
	appCCCount := ccCount.(*sync.Map)
	preHashContent := srcIP
	if ccPolicy.StatByURL == true {
		preHashContent += r.URL.Path
	}
	if ccPolicy.StatByUserAgent == true {
		ua := r.Header.Get("User-Agent")
		preHashContent += ua
	}
	if ccPolicy.StatByCookie == true {
		cookie := r.Header.Get("Cookie")
		preHashContent += cookie
	}
	clientID := data.SHA256Hash(preHashContent)
	clientIDStat, _ := appCCCount.LoadOrStore(clientID, &models.ClientStat{Count: 0, IsBadIP: false, RemainSeconds: 0})
	clientStat := clientIDStat.(*models.ClientStat)
	if clientStat.IsBadIP == true {
		needLog := false
		if clientStat.Count == 0 {
			clientStat.Count++
			needLog = true
		}
		return true, ccPolicy, clientID, needLog
	}
	clientStat.Count++
	//fmt.Println("IsCCAttack:", r.URL.Path, clientID, clientStat.Count, clientStat.IsBadIP, clientStat.RemainSeconds)
	return false, nil, "", false
}

// InitCCPolicy init CC policy
func InitCCPolicy() {
	//var cc_policies_list []*models.CCPolicy
	if data.IsPrimary {
		data.DAL.CreateTableIfNotExistsCCPolicy()
		existCCPolicy := data.DAL.ExistsCCPolicy()
		if existCCPolicy == false {
			data.DAL.InsertCCPolicy(0, 100, 5, 7200, models.Action_Block_100, true, false, false, true)
		}
		ccPoliciesList = data.DAL.SelectCCPolicies()
	} else {
		ccPoliciesList = RPCSelectCCPolicies()
	}
	for _, ccPolicy := range ccPoliciesList {
		ccPolicies.Store(ccPolicy.AppID, ccPolicy)
		//fmt.Println("InitCCPolicy:", ccPolicy.AppID, ccPolicy)
	}
}

// UpdateCCPolicy update CC policy
func UpdateCCPolicy(param map[string]interface{}) error {
	ccPolicyMap := param["object"].(map[string]interface{})
	appID := int64(param["id"].(float64))
	intervalMilliSeconds := time.Duration(ccPolicyMap["interval_milliseconds"].(float64))
	maxCount := int64(ccPolicyMap["max_count"].(float64))
	blockSeconds := time.Duration(ccPolicyMap["block_seconds"].(float64))
	action := models.PolicyAction(ccPolicyMap["action"].(float64))
	statByURL := ccPolicyMap["stat_by_url"].(bool)
	statByUA := ccPolicyMap["stat_by_ua"].(bool)
	statByCookie := ccPolicyMap["stat_by_cookie"].(bool)
	isEnabled := ccPolicyMap["is_enabled"].(bool)
	existAppID := data.DAL.ExistsCCPolicyByAppID(appID)
	if existAppID == false {
		// new policy
		err := data.DAL.InsertCCPolicy(appID, intervalMilliSeconds, maxCount, blockSeconds, action, statByURL, statByUA, statByCookie, isEnabled)
		if err != nil {
			return err
		}
		ccPolicy := &models.CCPolicy{
			AppID:                appID,
			IntervalMilliSeconds: intervalMilliSeconds, MaxCount: maxCount, BlockSeconds: blockSeconds,
			Action: action, StatByURL: statByURL, StatByUserAgent: statByUA, StatByCookie: statByCookie,
			IsEnabled: isEnabled}
		ccPolicies.Store(appID, ccPolicy)
		if ccPolicy.IsEnabled == true {
			go CCAttackTick(appID)
		}
	} else {
		// update policy
		err := data.DAL.UpdateCCPolicy(intervalMilliSeconds, maxCount, blockSeconds, action, statByURL, statByUA, statByCookie, isEnabled, appID)
		if err != nil {
			return err
		}
		ccPolicy := GetCCPolicyByAppID(appID)
		if ccPolicy.IntervalMilliSeconds != intervalMilliSeconds {
			ccPolicy.IntervalMilliSeconds = intervalMilliSeconds
			appCCTicker, _ := ccTickers.Load(appID)
			ccTicker := appCCTicker.(*time.Ticker)
			ccTicker.Stop()
		}
		ccPolicy.MaxCount = maxCount
		ccPolicy.BlockSeconds = blockSeconds
		ccPolicy.StatByURL = statByURL
		ccPolicy.StatByUserAgent = statByUA
		ccPolicy.StatByCookie = statByCookie
		ccPolicy.Action = action
		ccPolicy.IsEnabled = isEnabled
		if ccPolicy.IsEnabled == true {
			go CCAttackTick(appID)
		}
	}
	data.UpdateFirewallLastModified()
	return nil
}

// DeleteCCPolicyByAppID delete CC policy by app id
func DeleteCCPolicyByAppID(appID int64) error {
	if appID == 0 {
		return errors.New("Global CC policy cannot be deleted")
	}
	data.DAL.DeleteCCPolicy(appID)
	ccPolicies.Delete(appID)
	if appCCTicker, ok := ccTickers.Load(appID); ok {
		ccTicker := appCCTicker.(*time.Ticker)
		if ccTicker != nil {
			ccTicker.Stop()
		}
	}
	data.UpdateFirewallLastModified()
	return nil
}
