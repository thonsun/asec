/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-05-17 20:21:48
 * @Last Modified: thonsun, 2020-05-17 20:21:48
 */

package usermgmt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"asec/data"
	"asec/models"
	"asec/utils"
)

func toBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}

func toUint32(bytes []byte) uint32 {
	return (uint32(bytes[0]) << 24) + (uint32(bytes[1]) << 16) +
		(uint32(bytes[2]) << 8) + uint32(bytes[3])
}

// getCode ...
func getCode(secretKey string, timestamp int64) (code uint32) {
	secretKeyUpper := strings.ToUpper(secretKey)
	key, err := base32.StdEncoding.DecodeString(secretKeyUpper)
	if err != nil {
		fmt.Println(err)
		return
	}
	hmacSha1 := hmac.New(sha1.New, key)
	hmacSha1.Write(toBytes(timestamp / 30))
	hash := hmacSha1.Sum(nil)
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := toUint32(hashParts)
	code = number % 1000000
	return code
}

// VerifyCode is ok or not
func VerifyCode(secretKey string, code uint32) bool {
	timestamp := time.Now().Unix()
	tempCode := getCode(secretKey, timestamp)
	if code == tempCode {
		return true
	}
	for _, newTimestamp := range []int64{timestamp - 30, timestamp + 30, timestamp - 60, timestamp + 60} {
		tempCode = getCode(secretKey, newTimestamp)
		if code == tempCode {
			return true
		}
	}
	return false
}

func hmacSha1(key, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	if total := len(data); total > 0 {
		h.Write(data)
	}
	return h.Sum(nil)
}

func genKey() string {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, time.Now().UnixNano())
	key := strings.ToUpper(base32.StdEncoding.EncodeToString(hmacSha1(buf.Bytes(), nil)))[0:16]
	return key
}

// GetTOTPByUID return TOTP Key related to uid
func GetTOTPByUID(uid string) (totpItem *models.TOTP, err error) {
	if data.IsPrimary {
		totpItem, err = data.DAL.GetTOTPItemByUID(uid)
		return totpItem, err
	}
	// RPC Get or insert TOTP Item
	totpItem = &models.TOTP{
		ID:           0,
		UID:          uid,
		TOTPKey:      genKey(),
		TOTPVerified: false,
	}
	rpcRequest := &models.RPCRequest{
		Action: "gettotpkey", Object: totpItem}
	resp, err := data.GetRPCResponse(rpcRequest)
	utils.CheckError("GetTOTPByUID", err)
	rpcTOTP := new(models.RPCTOTP)
	if err = json.Unmarshal(resp, rpcTOTP); err != nil {
		utils.CheckError("RPC GetTOTPByUID Unmarshal", err)
	}
	return rpcTOTP.Object, err
}

// GetOrInsertTOTPItem for replica nodes
func GetOrInsertTOTPItem(param map[string]interface{}) (totpItem *models.TOTP, err error) {
	totpI := param["object"].(map[string]interface{})
	uid := totpI["uid"].(string)
	totpItem, err = data.DAL.GetTOTPItemByUID(uid)
	if err != nil {
		totpKey := totpI["totp_key"].(string)
		id, err := data.DAL.InsertTOTPItem(uid, totpKey, false)
		totpItem.ID = id
		return totpItem, err
	}
	return totpItem, nil
}

// UpdateTOTPVerified set verified = true
func UpdateTOTPVerified(id int64) (*models.TOTP, error) {
	if data.IsPrimary {
		data.DAL.UpdateTOTPVerified(true, id)
		return nil, nil
	}
	// RPC called
	rpcRequest := &models.RPCRequest{
		Action: "updatetotp", ObjectID: id, Object: nil}
	_, err := data.GetRPCResponse(rpcRequest)
	utils.CheckError("UpdateTOTPVerified", err)
	return nil, nil
}
