/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:38:41
 * @Last Modified: thonsun, 2020-10-05  16:38:41
 */

package models

import (
	"crypto/tls"
	"database/sql"
	"sync"
)

type Application struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	InternalScheme string         `json:"internal_scheme"` // http, https
	Destinations   []*Destination `json:"destinations"`

	// Route: map[string][]*Destination
	// {"/abc/": ["192.168.1.1:8800", "192.168.1.2:8800"], ".php": [...], "/": [...]}
	Route sync.Map `json:"-"`

	Domains        []*Domain `json:"domains"`
	RedirectHTTPS  bool      `json:"redirect_https"`
	HSTSEnabled    bool      `json:"hsts_enabled"`
	WAFEnabled     bool      `json:"waf_enabled"`
	ClientIPMethod IPMethod  `json:"ip_method"`
	Description    string    `json:"description"`
	OAuthRequired  bool      `json:"oauth_required"`
	SessionSeconds int64     `json:"session_seconds"`
	Owner          string    `json:"owner"`
}

type DBApplication struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	InternalScheme string   `json:"internal_scheme"` // http, https
	RedirectHTTPS  bool     `json:"redirect_https"`
	HSTSEnabled    bool     `json:"hsts_enabled"`
	WAFEnabled     bool     `json:"waf_enabled"`
	ClientIPMethod IPMethod `json:"ip_method"`
	Description    string   `json:"description"`
	OAuthRequired  bool     `json:"oauth_required"`
	SessionSeconds int64    `json:"session_seconds"`
	Owner          string   `json:"owner"`
}

type DomainRelation struct {
	App      *Application
	Cert     *CertItem
	Redirect bool
	Location string
}

type Domain struct {
	ID       int64        `json:"id"`
	Name     string       `json:"name"`
	AppID    int64        `json:"app_id"`
	CertID   int64        `json:"cert_id"`
	Redirect bool         `json:"redirect"`
	Location string       `json:"location"`
	App      *Application `json:"-"`
	Cert     *CertItem    `json:"-"`
}

type DBDomain struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	AppID    int64  `json:"app_id"`
	CertID   int64  `json:"cert_id"`
	Redirect bool   `json:"redirect"`
	Location string `json:"location"`
}

// RouteType used for backend routing
type RouteType int64

const (
	// ReverseProxyRoute used for secondary application /abc/ /xyz/
	ReverseProxyRoute RouteType = 1

	// FastCGIRoute used for PHP etc.
	FastCGIRoute RouteType = 1 << 1

	// StaticRoute used for static web server
	StaticRoute RouteType = 1 << 2
)

// Destination is used for backend routing
type Destination struct {
	ID int64 `json:"id"`

	// 0.9.8+
	RouteType RouteType `json:"route_type"`

	// 0.9.8+
	RequestRoute string `json:"request_route"`

	// 0.9.8+
	BackendRoute string `json:"backend_route"`

	// Destination is backend IP:Port , or static directory
	Destination string `json:"destination"`

	AppID  int64 `json:"app_id"`
	NodeID int64 `json:"node_id"`
}

type CertItem struct {
	ID             int64           `json:"id"`
	CommonName     string          `json:"common_name"`
	CertContent    string          `json:"cert_content"`
	PrivKeyContent string          `json:"priv_key_content"`
	TlsCert        tls.Certificate `json:"-"`
	ExpireTime     int64           `json:"expire_time"`
	Description    string          `json:"description"`
}

type DBCertItem struct {
	ID               int64
	CommonName       string
	CertContent      string
	EncryptedPrivKey []byte
	ExpireTime       int64
	Description      sql.NullString
}

// AuthUser used for Authentication in Memory
type AuthUser struct {
	UserID        int64  `json:"user_id"`
	Username      string `json:"username"`
	Logged        bool   `json:"logged"`
	IsSuperAdmin  bool   `json:"is_super_admin"`
	IsCertAdmin   bool   `json:"is_cert_admin"`
	IsAppAdmin    bool   `json:"is_app_admin"`
	NeedModifyPWD bool   `json:"need_modify_pwd"`
}

// AppUser used for DB Storage
type AppUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	HashPwd       string `json:"-"`
	Salt          string `json:"-"`
	Email         string `json:"email"`
	IsSuperAdmin  bool   `json:"is_super_admin"`
	IsCertAdmin   bool   `json:"is_cert_admin"`
	IsAppAdmin    bool   `json:"is_app_admin"`
	NeedModifyPWD bool   `json:"need_modify_pwd"`
}

// QueryAppUser not include password and salt
type QueryAppUser struct {
	ID           int64
	Username     string
	Email        sql.NullString
	IsSuperAdmin bool
	IsCertAdmin  bool
	IsAppAdmin   bool
}

// TOTP Authenticator
type TOTP struct {
	ID           int64  `json:"id"`
	UID          string `json:"uid"`
	TOTPKey      string `json:"totp_key"`
	TOTPVerified bool   `json:"totp_verified"`
}

type Setting struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type IPMethod int64

const (
	IPMethod_REMOTE_ADDR     IPMethod = 1
	IPMethod_X_FORWARDED_FOR IPMethod = 1 << 1
	IPMethod_X_REAL_IP       IPMethod = 1 << 2
	IPMethod_REAL_IP         IPMethod = 1 << 3
)
