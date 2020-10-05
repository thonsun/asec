/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-03-23 21:02:39
 * @Last Modified: thonsun, 2020-03-23 21:02:39
 */

package usermgmt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"asec/utils"

	"asec/data"
	"asec/models"

	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
)

type FeishuAccessToken struct {
	Code           int64  `json:"code"`
	Msg            string `json:"msg"`
	AppAccessToken string `json:"app_access_token"`
	Expire         int    `json:"expire"`
}

// https://open.feishu.cn/document/ukTMukTMukTM/uEDO4UjLxgDO14SM4gTN
type FeishuUserReqBody struct {
	AppAccessToken string `json:"app_access_token"`
	GrantType      string `json:"grant_type"`
	Code           string `json:"code"`
}

// https://open.feishu.cn/document/ukTMukTMukTM/uEDO4UjLxgDO14SM4gTN
type FeishuUser struct {
	Code int64          `json:"code"`
	Msg  string         `json:"msg"`
	Data FeishuAuthData `json:"data"`
}

type FeishuAuthData struct {
	AccessToken string `json:"access_token"`
	EnName      string `json:"en_name"`
}

// Doc: https://open.feishu.cn/document/ukTMukTMukTM/ukzN4UjL5cDO14SO3gTN
// Step 1: GET https://open.feishu.cn/open-apis/authen/v1/index?redirect_uri={REDIRECT_URI}&app_id={APPID}&state={STATE}
// If state==admin, for asec-admin; else for frontend applications
func FeishuCallbackWithCode(w http.ResponseWriter, r *http.Request) {
	// Step 2.1: Callback with code and state, http://gate.asec.com/?code=BM8k8U6RwtQtNY&state=admin
	code := r.FormValue("code")
	state := r.FormValue("state")
	// Step 2.2: Within Callback, get app_access_token
	// Doc: https://open.feishu.cn/document/ukTMukTMukTM/uADN14CM0UjLwQTN
	// POST https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal/
	// {"app_id":"cli_slkdasd", "app_secret":"dskLLdkasdKK"}
	accessTokenURL := "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal/"
	body := fmt.Sprintf(`{"app_id":"%s", "app_secret":"%s"}`,
		data.CFG.PrimaryNode.OAuth.Feishu.AppID,
		data.CFG.PrimaryNode.OAuth.Feishu.AppSecret)
	request, _ := http.NewRequest("POST", accessTokenURL, bytes.NewReader([]byte(body)))
	resp, err := GetResponse(request)
	if err != nil {
		utils.DebugPrintln("FeishuCallbackWithCode GetResponse", err)
	}
	tokenResponse := FeishuAccessToken{}
	json.Unmarshal(resp, &tokenResponse)
	// Step 2.3: Get User name
	// https://open.feishu.cn/document/ukTMukTMukTM/uEDO4UjLxgDO14SM4gTN
	userURL := "https://open.feishu.cn/open-apis/authen/v1/access_token"
	feishuUserReqBody := FeishuUserReqBody{
		AppAccessToken: tokenResponse.AppAccessToken,
		GrantType:      "authorization_code",
		Code:           code,
	}
	bytesData, _ := json.Marshal(feishuUserReqBody)
	request, _ = http.NewRequest("POST", userURL, bytes.NewReader(bytesData))
	request.Header.Set("Content-Type", "application/json")

	resp, err = GetResponse(request)
	if err != nil {
		utils.DebugPrintln("FeishuCallbackWithCode GetResponse", err)
	}
	feishuUser := FeishuUser{}
	json.Unmarshal(resp, &feishuUser)
	if state == "admin" {
		// Insert into db if not existed
		id, _ := data.DAL.InsertIfNotExistsAppUser(feishuUser.Data.EnName, "", "", "", false, false, false, false)
		// create session
		authUser := &models.AuthUser{
			UserID:        id,
			Username:      feishuUser.Data.EnName,
			Logged:        true,
			IsSuperAdmin:  false,
			IsCertAdmin:   false,
			IsAppAdmin:    false,
			NeedModifyPWD: false}
		session, _ := store.Get(r, "sessionid")
		session.Values["authuser"] = authUser
		session.Options = &sessions.Options{Path: "/asec-admin/", MaxAge: tokenResponse.Expire}
		session.Save(r, w)
		http.Redirect(w, r, data.CFG.PrimaryNode.Admin.Portal, http.StatusFound)
		return
	}
	// Gateway OAuth for employees and internal application
	oauthStateI, found := OAuthCache.Get(state)
	if found {
		oauthState := oauthStateI.(models.OAuthState)
		oauthState.UserID = feishuUser.Data.EnName
		oauthState.AccessToken = feishuUser.Data.AccessToken
		OAuthCache.Set(state, oauthState, cache.DefaultExpiration)
		http.Redirect(w, r, oauthState.CallbackURL, http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
