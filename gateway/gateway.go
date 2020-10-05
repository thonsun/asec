/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:37:57
 * @Last Modified: thonsun, 2020-10-05  16:37:57
 */

package gateway

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"syscall"
	"time"

	"asec/backend"
	"asec/data"
	"asec/firewall"
	"asec/models"
	"asec/usermgmt"
	"asec/utils"

	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
	"github.com/yookoala/gofast"
	"golang.org/x/net/http2"
)

var (
	store = sessions.NewCookieStore([]byte("asec"))
)

// ReverseHandlerFunc used for reverse handler
func ReverseHandlerFunc(w http.ResponseWriter, r *http.Request) {
	// r.Host may has the format: domain:port, first remove port
	index := strings.IndexByte(r.Host, ':')
	if index > 0 {
		r.Host = r.Host[0:index]
	}
	domain := backend.GetDomainByName(r.Host)
	if domain != nil && domain.Redirect == true {
		RedirectRequest(w, r, domain.Location)
		return
	}
	app := backend.GetApplicationByDomain(r.Host)
	if app == nil {
		hitInfo := &models.HitInfo{PolicyID: 0, VulnName: "Unknown Host"}
		GenerateBlockPage(w, hitInfo)
		return
	}
	if (r.TLS == nil) && (app.RedirectHTTPS == true) {
		RedirectRequest(w, r, "https://"+r.Host+r.URL.Path)
		return
	}
	r.URL.Scheme = app.InternalScheme
	r.URL.Host = r.Host

	// dynamic
	srcIP := GetClientIP(r, app)
	isStatic := firewall.IsStaticResource(r)
	if app.WAFEnabled && !isStatic {
		if isCC, ccPolicy, clientID, needLog := firewall.IsCCAttack(r, app.ID, srcIP); isCC == true {
			targetURL := r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				targetURL += "?" + r.URL.RawQuery
			}
			hitInfo := &models.HitInfo{TypeID: 1,
				PolicyID:  ccPolicy.AppID,
				VulnName:  "CC",
				Action:    ccPolicy.Action,
				ClientID:  clientID,
				TargetURL: targetURL,
				BlockTime: time.Now().Unix()}
			switch ccPolicy.Action {
			case models.Action_Block_100:
				if needLog {
					go firewall.LogCCRequest(r, app.ID, srcIP, ccPolicy)
				}
				if app.ClientIPMethod == models.IPMethod_REMOTE_ADDR {
					go firewall.AddIP2NFTables(srcIP, ccPolicy.BlockSeconds)
				}
				GenerateBlockPage(w, hitInfo)
				return
			case models.Action_BypassAndLog_200:
				if needLog {
					go firewall.LogCCRequest(r, app.ID, srcIP, ccPolicy)
				}
			case models.Action_CAPTCHA_300:
				if needLog {
					go firewall.LogCCRequest(r, app.ID, srcIP, ccPolicy)
				}
				captchaHitInfo.Store(hitInfo.ClientID, hitInfo)
				captchaURL := CaptchaEntrance + "?id=" + hitInfo.ClientID
				http.Redirect(w, r, captchaURL, http.StatusTemporaryRedirect)
				return
			}
		}

		if isHit, policy := firewall.IsRequestHitPolicy(r, app.ID, srcIP); isHit == true {
			switch policy.Action {
			case models.Action_Block_100:
				vulnName, _ := firewall.VulnMap.Load(policy.VulnID)
				hitInfo := &models.HitInfo{TypeID: 2, PolicyID: policy.ID, VulnName: vulnName.(string)}
				go firewall.LogGroupHitRequest(r, app.ID, srcIP, policy)
				GenerateBlockPage(w, hitInfo)
				return
			case models.Action_BypassAndLog_200:
				go firewall.LogGroupHitRequest(r, app.ID, srcIP, policy)
			case models.Action_CAPTCHA_300:
				go firewall.LogGroupHitRequest(r, app.ID, srcIP, policy)
				clientID := GenClientID(r, app.ID, srcIP)
				targetURL := r.URL.Path
				if len(r.URL.RawQuery) > 0 {
					targetURL += "?" + r.URL.RawQuery
				}
				hitInfo := &models.HitInfo{TypeID: 2,
					PolicyID: policy.ID, VulnName: "Group Policy Hit",
					Action: policy.Action, ClientID: clientID,
					TargetURL: targetURL, BlockTime: time.Now().Unix()}
				captchaHitInfo.Store(clientID, hitInfo)
				captchaURL := CaptchaEntrance + "?id=" + clientID
				http.Redirect(w, r, captchaURL, http.StatusTemporaryRedirect)
				return
			default:
				// models.Action_Pass_400 do nothing
			}
		}
	}

	// Check OAuth
	if app.OAuthRequired && data.CFG.PrimaryNode.OAuth.Enabled {
		session, _ := store.Get(r, "asec-token")
		usernameI := session.Values["userid"]
		var url string
		if r.TLS != nil {
			url = "https://" + r.Host + r.URL.Path
		} else {
			url = r.URL.String()
		}
		//fmt.Println("1000", usernameI, url)
		if usernameI == nil {
			// Exec OAuth2 Authentication
			ua := r.UserAgent() //r.Header.Get("User-Agent")
			state := data.SHA256Hash(srcIP + url + ua)
			stateSession := session.Values[state]
			//fmt.Println("1001 state=", state, url)
			if stateSession == nil {
				entranceURL, err := getOAuthEntrance(state)
				if err != nil {
					w.Write([]byte(err.Error()))
					return
				}
				// Save Application URL for CallBack
				oauthState := models.OAuthState{
					CallbackURL: url,
					UserID:      ""}
				usermgmt.OAuthCache.Set(state, oauthState, cache.DefaultExpiration)
				session.Values[state] = state
				session.Options = &sessions.Options{Path: "/", MaxAge: 300}
				session.Save(r, w)
				//fmt.Println("1002 cache state:", oauthState, url, "307 to:", entranceURL)
				http.Redirect(w, r, entranceURL, http.StatusTemporaryRedirect)
				return
			}
			// Has state in session, get UserID from cache
			state = stateSession.(string)
			oauthStateI, found := usermgmt.OAuthCache.Get(state)
			if found == false {
				// Time expired, clear session
				session.Options = &sessions.Options{Path: "/", MaxAge: -1}
				session.Save(r, w)
				http.Redirect(w, r, url, http.StatusTemporaryRedirect)
				return
			}
			// found == true
			oauthState := oauthStateI.(models.OAuthState)
			if oauthState.UserID == "" {
				session.Values["userid"] = nil
				entranceURL, err := getOAuthEntrance(state)
				if err != nil {
					w.Write([]byte(err.Error()))
					return
				}
				http.Redirect(w, r, entranceURL, http.StatusTemporaryRedirect)
				return
			}
			session.Values["userid"] = oauthState.UserID
			session.Values["access_token"] = oauthState.AccessToken
			session.Options = &sessions.Options{Path: "/", MaxAge: int(app.SessionSeconds)}
			session.Save(r, w)
			http.Redirect(w, r, oauthState.CallbackURL, http.StatusTemporaryRedirect)
			return
		}
		// Exist username in session, Forward username to destination
		accessToken := session.Values["access_token"].(string)
		r.Header.Set("Authorization", "Bearer "+accessToken)
		r.Header.Set("X-Auth-User", usernameI.(string))
	}

	dest := backend.SelectBackendRoute(app, r, srcIP)
	if dest == nil {
		w.Write([]byte("Error: No route found, please check the configuration."))
		return
	}

	// Add access log
	utils.AccessLog(r.Host, r.Method, srcIP, r.RequestURI, r.UserAgent())

	if dest.RouteType == models.StaticRoute {
		// Static Web site
		staticHandler := http.FileServer(http.Dir(dest.BackendRoute))
		if strings.HasSuffix(r.URL.Path, "/") {
			targetFile := dest.BackendRoute + strings.Replace(r.URL.Path, dest.RequestRoute, "", 1) + dest.Destination
			http.ServeFile(w, r, targetFile)
			return
		}
		staticHandler.ServeHTTP(w, r)
		return
	} else if dest.RouteType == models.FastCGIRoute {
		// FastCGI
		connFactory := gofast.SimpleConnFactory("tcp", dest.Destination)
		urlPath := utils.GetRoutePath(r.URL.Path)
		newPath := r.URL.Path
		if urlPath != "/" {
			newPath = strings.Replace(r.URL.Path, dest.RequestRoute, "/", 1)
		}
		fastCGIHandler := gofast.NewHandler(
			gofast.NewFileEndpoint(dest.BackendRoute+newPath)(gofast.BasicSession),
			gofast.SimpleClientFactory(connFactory, 0),
		)
		fastCGIHandler.ServeHTTP(w, r)
		return
	}

	// var transport http.RoundTripper
	transport := &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("tcp", dest.Destination)
		},
		DialTLS: func(network, addr string) (net.Conn, error) {
			conn, err := net.Dial("tcp", dest.Destination)
			if err != nil {
				return nil, err
			}
			cfg := &tls.Config{ServerName: r.Host, NextProtos: []string{"h2", "http/1.1"}}
			tlsConn := tls.Client(conn, cfg)
			if err := tlsConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			return tlsConn, nil //net.Dial("tcp", dest)
		},
	}
	http2.ConfigureTransport(transport)

	// Check static cache
	if isStatic {
		// First check Header Range, not cache for range
		rangeValue := r.Header.Get("Range")
		if rangeValue == "" {
			staticRoot := fmt.Sprintf("./static/cdncache/%d", app.ID)
			targetFile := staticRoot + r.URL.Path
			// Check Static Cache
			if fi, err := os.Stat(targetFile); err == nil {
				// Found targetFile
				now := time.Now()
				fiStat := fi.Sys().(*syscall.Stat_t)
				// Use ctime fiStat.Ctim.Sec to mark the last check time
				pastSeconds := now.Unix() - fiStat.Ctim.Sec
				if pastSeconds > 1800 {
					// check update
					go func() {
						backendAddr := fmt.Sprintf("%s://%s%s", app.InternalScheme, dest.Destination, r.RequestURI)
						req, err := http.NewRequest("GET", backendAddr, nil)
						if err != nil {
							utils.DebugPrintln("Check Update NewRequest", err)
						}
						if err == nil {
							req.Header.Set("Host", r.Host)
							modTimeGMT := fi.ModTime().UTC().Format(http.TimeFormat)
							//If-Modified-Since: Sun, 14 Jun 2020 13:54:20 GMT
							req.Header.Set("If-Modified-Since", modTimeGMT)
							client := http.Client{
								Transport: transport,
							}
							resp, err := client.Do(req)
							if err != nil {
								utils.DebugPrintln("Cache update Do", err)
								return
							}
							defer resp.Body.Close()
							if resp.StatusCode == http.StatusOK {
								//fmt.Println("200", backendAddr)
								bodyBuf, _ := ioutil.ReadAll(resp.Body)
								err = ioutil.WriteFile(targetFile, bodyBuf, 0666)
								lastModified, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
								if err != nil {
									utils.DebugPrintln("CDN Parse Last-Modified", targetFile, err)
								}
								err = os.Chtimes(targetFile, now, lastModified)
								if err != nil {
									utils.DebugPrintln("CDN Chtimes", targetFile, err)
								}
							} else if resp.StatusCode == http.StatusNotModified {
								//fmt.Println("304", backendAddr)
								err := os.Chtimes(targetFile, now, fi.ModTime())
								if err != nil {
									utils.DebugPrintln("Cache update access time", err)
								}
							}
						}
					}()
				}
				http.ServeFile(w, r, targetFile)
				return
			}
		}
		// Has Range Header, or resource Not Found, Continue
	}

	// Reverse Proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			//req.URL.Scheme = app.InternalScheme
			//req.URL.Host = r.Host
		},
		Transport:      transport,
		ModifyResponse: rewriteResponse}
	if utils.Debug {
		//dump, err := httputil.DumpRequest(r, true)
		//utils.CheckError("ReverseHandlerFunc DumpRequest", err)
		//fmt.Println(string(dump))
	}
	proxy.ServeHTTP(w, r)
}

func getOAuthEntrance(state string) (entranceURL string, err error) {
	switch data.CFG.PrimaryNode.OAuth.Provider {
	case "wxwork":
		entranceURL = fmt.Sprintf("https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=%s&agentid=%s&redirect_uri=%s&state=%s",
			data.CFG.PrimaryNode.OAuth.Wxwork.CorpID,
			data.CFG.PrimaryNode.OAuth.Wxwork.AgentID,
			data.CFG.PrimaryNode.OAuth.Wxwork.Callback,
			state)
	case "dingtalk":
		entranceURL = fmt.Sprintf("https://oapi.dingtalk.com/connect/qrconnect?appid=%s&response_type=code&scope=snsapi_login&state=%s&redirect_uri=%s",
			data.CFG.PrimaryNode.OAuth.Dingtalk.AppID,
			state,
			data.CFG.PrimaryNode.OAuth.Dingtalk.Callback)
	case "feishu":
		entranceURL = fmt.Sprintf("https://open.feishu.cn/open-apis/authen/v1/index?redirect_uri=%s&app_id=%s&state=%s",
			data.CFG.PrimaryNode.OAuth.Feishu.Callback,
			data.CFG.PrimaryNode.OAuth.Feishu.AppID,
			state)
	case "ldap":
		entranceURL = "/ldap/login?state=" + state
	case "saml":
		entranceURL = "/saml/login?state=" + state
	default:
		//w.Write([]byte("Designated OAuth not supported, please check config.json ."))
		return "", errors.New("the OAuth provider is not supported, please check config.json")
	}
	return entranceURL, nil
}

// RedirectRequest for example: redirect 80 to 443
func RedirectRequest(w http.ResponseWriter, r *http.Request, location string) {
	if len(r.URL.RawQuery) > 0 {
		location += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, location, http.StatusMovedPermanently)
}

// GenClientID generate unique client id
func GenClientID(r *http.Request, appID int64, srcIP string) string {
	preHashContent := srcIP
	url := r.URL.Path
	preHashContent += url
	ua := r.Header.Get("User-Agent")
	preHashContent += ua
	cookie := r.Header.Get("Cookie")
	preHashContent += cookie
	clientID := data.SHA256Hash(preHashContent)
	return clientID
}

// GetClientIP acquire the client IP address
func GetClientIP(r *http.Request, app *models.Application) (clientIP string) {
	switch app.ClientIPMethod {
	case models.IPMethod_REMOTE_ADDR:
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		return clientIP
	case models.IPMethod_X_FORWARDED_FOR:
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		ips := strings.Split(xForwardedFor, ", ")
		clientIP = ips[len(ips)-1]
	case models.IPMethod_X_REAL_IP:
		clientIP = r.Header.Get("X-Real-IP")
	case models.IPMethod_REAL_IP:
		clientIP = r.Header.Get("Real-IP")
	}
	if len(clientIP) == 0 {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return clientIP
}

func OAuthLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "asec-token")
	session.Options = &sessions.Options{Path: "/", MaxAge: -1}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
