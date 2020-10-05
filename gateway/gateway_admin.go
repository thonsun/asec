/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:36:26
 * @Last Modified: thonsun, 2020-10-05  16:36:26
 */

package gateway

import (
	"net/http"
)

// AdminHandlerFunc is for /asec-admin
func AdminHandlerFunc(w http.ResponseWriter, r *http.Request) {
	staticHandler := http.FileServer(http.Dir("static"))
	staticHandler.ServeHTTP(w, r)
}
