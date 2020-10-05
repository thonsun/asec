/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:39:03
 * @Last Modified: thonsun, 2020-10-05  16:39:03
 */

package models

type HitInfo struct {
	TypeID    int64 // 1: CCPolicy  2:GroupPolicy
	PolicyID  int64
	VulnName  string
	Action    PolicyAction
	ClientID  string // for CC/Attack Client ID
	TargetURL string // for CAPTCHA redirect
	BlockTime int64
}

type CaptchaContext struct {
	CaptchaId string
	ClientID  string
}

type OAuthState struct {
	CallbackURL string
	UserID      string
	AccessToken string
}
