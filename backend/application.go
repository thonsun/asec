/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:21:38
 * @Last Modified: thonsun, 2020-10-05  16:21:38
 */

package backend

import (
	"errors"
	"hash/fnv"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"asec/data"
	"asec/firewall"
	"asec/models"
	"asec/utils"
)

var (
	Apps []*models.Application
)

// SelectDestination deprecated from v0.9.8
/*
func SelectDestination(app *models.Application) string {
	destLen := len(app.Destinations)
	var dest *models.Destination
	if destLen == 1 {
		dest = app.Destinations[0]
	} else if destLen > 1 {
		ns := time.Now().Nanosecond()
		destIndex := ns % len(app.Destinations)
		dest = app.Destinations[destIndex]
	}
	//utils.DebugPrintln("SelectDestination", dest)
	return dest.Destination
}
*/

// SelectBackendRoute will replace SelectDestination
func SelectBackendRoute(app *models.Application, r *http.Request, srcIP string) *models.Destination {
	routePath := utils.GetRoutePath(r.URL.Path)
	var dests []*models.Destination
	hit := false
	if routePath != "/" {
		// First check /abc/
		valueI, ok := app.Route.Load(routePath)
		if ok {
			hit = true
			dests = valueI.([]*models.Destination)
		}
	}

	if !hit {
		// Second check .php
		ext := filepath.Ext(r.URL.Path)
		valueI, ok := app.Route.Load(ext)
		// Third check /
		if !ok {
			valueI, ok = app.Route.Load("/")
		}
		if !ok {
			// lack of route /
			return nil
		}
		dests = valueI.([]*models.Destination)
	}

	destLen := uint32(len(dests))
	var dest *models.Destination
	if destLen == 1 {
		dest = dests[0]
	} else if destLen > 1 {
		// ns := time.Now().Nanosecond() deprecated, change to Hash(IP+UA)
		h := fnv.New32a()
		h.Write([]byte(srcIP + r.UserAgent()))
		hashUInt32 := h.Sum32()
		destIndex := hashUInt32 % destLen
		dest = dests[destIndex]
	}
	if dest.RouteType == models.ReverseProxyRoute {
		if dest.RequestRoute != dest.BackendRoute {
			r.URL.Path = strings.Replace(r.URL.Path, dest.RequestRoute, dest.BackendRoute, 1)
		}
	}
	return dest
}

func GetApplicationByID(appID int64) (*models.Application, error) {
	for _, app := range Apps {
		if app.ID == appID {
			return app, nil
		}
	}
	return nil, errors.New("Not found.")
}

func GetWildDomainName(domain string) string {
	index := strings.Index(domain, ".")
	if index > 0 {
		wildDomain := "*" + domain[index:]
		return wildDomain
	}
	return ""
}

func GetApplicationByDomain(domain string) *models.Application {
	if domainRelation, ok := DomainsMap.Load(domain); ok {
		app := domainRelation.(models.DomainRelation).App //DomainsMap[domain].App
		return app
	}
	wildDomain := GetWildDomainName(domain) // *.asec.com
	if domainRelation, ok := DomainsMap.Load(wildDomain); ok {
		domainRelation2 := domainRelation.(models.DomainRelation)
		app := domainRelation2.App //DomainsMap[domain].App
		DomainsMap.Store(domain, models.DomainRelation{App: app, Cert: domainRelation2.Cert, Redirect: false, Location: ""})
		return app
	}
	return nil
}

func LoadApps() {
	Apps = Apps[0:0]
	if data.IsPrimary {
		dbApps := data.DAL.SelectApplications()
		for _, dbApp := range dbApps {
			app := &models.Application{ID: dbApp.ID,
				Name:           dbApp.Name,
				InternalScheme: dbApp.InternalScheme,
				RedirectHTTPS:  dbApp.RedirectHTTPS,
				HSTSEnabled:    dbApp.HSTSEnabled,
				WAFEnabled:     dbApp.WAFEnabled,
				ClientIPMethod: dbApp.ClientIPMethod,
				Description:    dbApp.Description,
				Destinations:   []*models.Destination{},
				Route:          sync.Map{},
				OAuthRequired:  dbApp.OAuthRequired,
				SessionSeconds: dbApp.SessionSeconds,
				Owner:          dbApp.Owner}
			Apps = append(Apps, app)
		}
	} else {
		// Replica
		rpcApps := RPCSelectApplications()
		if rpcApps != nil {
			Apps = rpcApps
		}
	}
}

func LoadDestinations() {
	for _, app := range Apps {
		app.Destinations = data.DAL.SelectDestinationsByAppID(app.ID)
		for _, dest := range app.Destinations {
			routeI, ok := app.Route.Load(dest.RequestRoute)
			var route []*models.Destination
			if ok {
				route = routeI.([]*models.Destination)
			}
			route = append(route, dest)
			app.Route.Store(dest.RequestRoute, route)
		}
	}
}

func LoadRoute() {
	for _, app := range Apps {
		for _, dest := range app.Destinations {
			routeI, ok := app.Route.Load(dest.RequestRoute)
			var route []*models.Destination
			if ok {
				route = routeI.([]*models.Destination)
			}
			route = append(route, dest)
			app.Route.Store(dest.RequestRoute, route)
		}
	}
}

/*
func LoadStaticDirs() {
	for i,_ := range Apps {
		app := Apps[i]
		rows,err := DB.Query("select directory from staticdirs where appID=$1", app.ID)
		utils.CheckError(err)
		for rows.Next() {
			var directory string
			err = rows.Scan(&directory)
			utils.CheckError(err)
			app.StaticDirs = append(app.StaticDirs, directory)
		}
	}
}
*/

func LoadAppDomainNames() {
	for _, app := range Apps {
		for _, domain := range Domains {
			if domain.AppID == app.ID {
				app.Domains = append(app.Domains, domain)
			}
		}
	}
}

func GetApplications(authUser *models.AuthUser) ([]*models.Application, error) {
	if authUser.IsAppAdmin {
		return Apps, nil
	}
	var myApps []*models.Application
	for _, app := range Apps {
		if app.Owner == authUser.Username {
			myApps = append(myApps, app)
		}
	}
	return myApps, nil
}

func UpdateDestinations(app *models.Application, destinations []interface{}) {
	//fmt.Println("ToDo UpdateDestinations")
	for _, dest := range app.Destinations {
		// delete outdated destinations from DB
		if !InterfaceContainsDestinationID(destinations, dest.ID) {
			app.Route.Delete(dest.RequestRoute)
			data.DAL.DeleteDestinationByID(dest.ID)
		}
	}
	var newDestinations []*models.Destination
	for _, destinationInterface := range destinations {
		// add new destinations to DB and app
		destMap := destinationInterface.(map[string]interface{})
		destID := int64(destMap["id"].(float64))
		routeType := int64(destMap["route_type"].(float64))
		requestRoute := strings.TrimSpace(destMap["request_route"].(string))
		backendRoute := strings.TrimSpace(destMap["backend_route"].(string))
		destDest := strings.TrimSpace(destMap["destination"].(string))
		appID := app.ID //int64(destMap["appID"].(float64))
		nodeID := int64(destMap["node_id"].(float64))
		if destID == 0 {
			destID, _ = data.DAL.InsertDestination(routeType, requestRoute, backendRoute, destDest, appID, nodeID)
		} else {
			data.DAL.UpdateDestinationNode(routeType, requestRoute, backendRoute, destDest, appID, nodeID, destID)
		}
		dest := &models.Destination{
			ID:           destID,
			RouteType:    models.RouteType(routeType),
			RequestRoute: requestRoute,
			BackendRoute: backendRoute,
			Destination:  destDest,
			AppID:        appID,
			NodeID:       nodeID}
		newDestinations = append(newDestinations, dest)
	}
	app.Destinations = newDestinations

	// Update Route Map
	app.Route.Range(func(key, value interface{}) bool {
		app.Route.Delete(key)
		return true
	})
	for _, dest := range app.Destinations {
		routeI, ok := app.Route.Load(dest.RequestRoute)
		var route []*models.Destination
		if ok {
			route = routeI.([]*models.Destination)
		}
		route = append(route, dest)
		app.Route.Store(dest.RequestRoute, route)
	}
}

func UpdateAppDomains(app *models.Application, appDomains []interface{}) {
	newAppDomains := []*models.Domain{}
	newDomainNames := []string{}
	for _, domainMap := range appDomains {
		domain := UpdateDomain(app, domainMap)
		newAppDomains = append(newAppDomains, domain)
		newDomainNames = append(newDomainNames, domain.Name)
	}
	for _, oldDomain := range app.Domains {
		if !InterfaceContainsDomainID(appDomains, oldDomain.ID) {
			DomainsMap.Delete(oldDomain.Name)
			data.DAL.DeleteDomainByDomainID(oldDomain.ID)
		}
	}
	app.Domains = newAppDomains
}

func UpdateApplication(param map[string]interface{}) (*models.Application, error) {
	application := param["object"].(map[string]interface{})
	appID := int64(application["id"].(float64))
	appName := application["name"].(string)
	internalScheme := application["internal_scheme"].(string)
	redirectHttps := application["redirect_https"].(bool)
	hstsEnabled := application["hsts_enabled"].(bool)
	wafEnabled := application["waf_enabled"].(bool)
	ipMethod := models.IPMethod(application["ip_method"].(float64))
	var description string
	var ok bool
	if description, ok = application["description"].(string); !ok {
		description = ""
	}
	oauthRequired := application["oauth_required"].(bool)
	sessionSeconds := int64(application["session_seconds"].(float64))
	owner := application["owner"].(string)
	var app *models.Application
	if appID == 0 {
		// new application
		newID := data.DAL.InsertApplication(appName, internalScheme, redirectHttps, hstsEnabled, wafEnabled, ipMethod, description, oauthRequired, sessionSeconds, owner)
		app = &models.Application{
			ID: newID, Name: appName,
			InternalScheme: internalScheme,
			//Destinations:   []*models.Destination{},
			Route:          sync.Map{},
			Domains:        []*models.Domain{},
			RedirectHTTPS:  redirectHttps,
			HSTSEnabled:    hstsEnabled,
			WAFEnabled:     wafEnabled,
			ClientIPMethod: ipMethod,
			Description:    description,
			OAuthRequired:  oauthRequired,
			SessionSeconds: sessionSeconds,
			Owner:          owner}
		Apps = append(Apps, app)
	} else {
		app, _ = GetApplicationByID(appID)
		if app != nil {
			data.DAL.UpdateApplication(appName, internalScheme, redirectHttps, hstsEnabled, wafEnabled, ipMethod, description, oauthRequired, sessionSeconds, owner, appID)
			app.Name = appName
			app.InternalScheme = internalScheme
			app.RedirectHTTPS = redirectHttps
			app.HSTSEnabled = hstsEnabled
			app.WAFEnabled = wafEnabled
			app.ClientIPMethod = ipMethod
			app.Description = description
			app.OAuthRequired = oauthRequired
			app.SessionSeconds = sessionSeconds
			app.Owner = owner
		} else {
			return nil, errors.New("Application not found.")
		}
	}
	destinations := application["destinations"].([]interface{})
	UpdateDestinations(app, destinations)
	appDomains := application["domains"].([]interface{})
	UpdateAppDomains(app, appDomains)
	data.UpdateBackendLastModified()
	return app, nil
}

func GetApplicationIndex(appID int64) int {
	for i := 0; i < len(Apps); i++ {
		if Apps[i].ID == appID {
			return i
		}
	}
	return -1
}

func DeleteDestinationsByApp(appID int64) {
	data.DAL.DeleteDestinationsByAppID(appID)
}

func DeleteApplicationByID(appID int64) error {
	app, err := GetApplicationByID(appID)
	if err != nil {
		return err
	}
	DeleteDomainsByApp(app)
	DeleteDestinationsByApp(appID)
	firewall.DeleteCCPolicyByAppID(appID)
	err = data.DAL.DeleteApplication(appID)
	if err != nil {
		return err
	}
	i := GetApplicationIndex(appID)
	Apps = append(Apps[:i], Apps[i+1:]...)
	data.UpdateBackendLastModified()
	return nil
}
