/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:22:46
 * @Last Modified: thonsun, 2020-10-05  16:22:46
 */

package backend

import (
	"asec/data"

	_ "github.com/lib/pq"
)

func InitDatabase() {
	dal := data.DAL
	dal.CreateTableIfNotExistsCertificates()
	dal.CreateTableIfNotExistsApplications()
	dal.CreateTableIfNotExistsDomains()
	dal.CreateTableIfNotExistsDestinations()
	dal.CreateTableIfNotExistsSettings()
	dal.CreateTableIfNotExistsAppUsers()
	dal.InsertIfNotExistsAppUser(`admin`, `313128e4aa423207518ff1e856bb23bb495e5ca8915cdd16b8baa8a5e13a89c8`,
		`afa8bae009c9dbf4135f62e165847227`, ``, true, true, true, true)
	dal.CreateTableIfNotExistsNodes()
	dal.CreateTableIfNotExistsTOTP()
	// Upgrade to latest version
	if dal.ExistColumnInTable("domains", "redirect") == false {
		// v0.9.6+ required
		dal.ExecSQL(`alter table domains add column redirect boolean default false, add column location varchar(256) default null`)
	}
	if dal.ExistColumnInTable("applications", "oauth_required") == false {
		// v0.9.7+ required
		dal.ExecSQL(`alter table applications add column oauth_required boolean default false, add column session_seconds bigint default 7200, add column owner varchar(128)`)
	}
	if dal.ExistColumnInTable("destinations", "route_type") == false {
		// v0.9.8+ required
		dal.ExecSQL(`alter table destinations add column route_type bigint default 1, add column request_route varchar(128) default '/', add column backend_route varchar(128) default '/'`)
	}
	if dal.ExistColumnInTable("ccpolicies", "interval_seconds") == true {
		// v0.9.9 interval_seconds, v0.9.10 interval_milliseconds
		dal.ExecSQL(`ALTER TABLE ccpolicies RENAME COLUMN interval_seconds TO interval_milliseconds`)
		dal.ExecSQL(`UPDATE ccpolicies SET interval_milliseconds=interval_milliseconds*1000`)
	}
}

func LoadAppConfiguration() {
	LoadCerts()
	LoadApps()
	if data.IsPrimary {
		LoadDestinations()
		LoadDomains()
		LoadAppDomainNames()
		LoadNodes()
	} else {
		LoadRoute()
		LoadDomains()
	}
}
