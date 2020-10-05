/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-03-22 10:29:15
 * @Last Modified: thonsun, 2020-03-22 10:29:15
 */

package usermgmt

import (
	"io/ioutil"
	"net/http"

	"asec/data"
	"asec/models"
	"asec/utils"
)

func GetOAuthConfig() (*models.OAuthConfig, error) {
	return &data.CFG.PrimaryNode.OAuth, nil
}

func GetResponse(request *http.Request) (respBytes []byte, err error) {
	request.Header.Set("Accept", "application/json")
	client := http.Client{}
	resp, err := client.Do(request)
	utils.CheckError("GetResponse Do", err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBytes, err = ioutil.ReadAll(resp.Body)
	return respBytes, err
}
