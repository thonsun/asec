/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-05-31 20:01:54
 * @Last Modified: thonsun, 2020-05-31 20:01:54
 */

package gateway

import (
	"fmt"
	"net/http"
)

func SAMLLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SAMLLogin ToDo")
	/*
		cert, err := backend.GetCertificateByDomain(r.URL.Host)
		samlSP, _ := samlsp.New(samlsp.Options{
			URL:            *rootURL,
			Key:            cert.PrivateKey.(*rsa.PrivateKey),
			Certificate:    cert.Leaf,
			IDPMetadataURL: idpMetadataURL,
		})
		samlSP.HandleStartAuthFlow(w, r)
	*/
}
