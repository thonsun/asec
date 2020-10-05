/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-10-05 16:25:04
 * @Last Modified: thonsun, 2020-10-05  16:25:04
 */

package data

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"strings"

	"asec/models"
	//"fmt"
)

func NewConfig(filename string) (*models.Config, error) {
	config := new(models.Config)
	configBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configBytes, config)
	if strings.ToLower(config.NodeRole) == "primary" {
		dbPassword := config.PrimaryNode.Database.Password
		if len(dbPassword) <= 32 {
			// Encrypt password
			encryptedPasswordBytes := AES256Encrypt([]byte(dbPassword), true)
			encryptedPassword := hex.EncodeToString(encryptedPasswordBytes)
			encryptedConfig := models.EncryptedConfig(*config)
			encryptedConfig.PrimaryNode.Database.Password = encryptedPassword
			encryptedConfigBytes, _ := json.MarshalIndent(encryptedConfig, "", "\t")
			err = ioutil.WriteFile(filename, encryptedConfigBytes, 0644)
		} else {
			// Decrypt password
			encryptedPassword, err := hex.DecodeString(dbPassword)
			if err != nil {
				return nil, err
			}
			passwordBytes, _ := AES256Decrypt(encryptedPassword, true)
			config.PrimaryNode.Database.Password = string(passwordBytes)
		}
	}
	//fmt.Println("NewConfig config.Database.Password=",config.Database.Password)
	return config, nil
}
