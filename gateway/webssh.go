/*
 * @Copyright Reserved By asec (https://www.asec.com/).
 * @Author: thonsun
 * @Date: 2020-02-10 22:07:47
 * @Last Modified: thonsun, 2020-02-10 22:07:47
 */

package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"asec/data"

	"asec/usermgmt"
	"asec/utils"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type HostInfo struct {
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// SSH build connection
func SSH(sshInput *io.WriteCloser, sshOutput *io.Reader, host *HostInfo, errChan chan<- error) {
	sshClient, err := ssh.Dial("tcp", host.IP+":"+host.Port, &ssh.ClientConfig{
		User:            host.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(host.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		errChan <- err
		utils.CheckError("errChan", err)
		return
	}
	sshSession, err := sshClient.NewSession()
	utils.CheckError("new ssh session", err)
	defer sshSession.Close()
	*sshInput, err = sshSession.StdinPipe()
	utils.CheckError("sshInput", err)
	*sshOutput, err = sshSession.StdoutPipe()
	utils.CheckError("sshOuput", err)
	errChan <- err
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = sshSession.RequestPty("xterm", 25, 80, modes)
	utils.CheckError("request pty", err)
	err = sshSession.Shell()
	utils.CheckError("start shell", err)
	err = sshSession.Wait()
	errChan <- err
}

func RoutineOutput(outputTicker *time.Ticker, wsConn *websocket.Conn, sshOutput *io.Reader) {
	for range outputTicker.C {
		cmdOutput := make([]byte, 1024*10)
		n, err := (*sshOutput).Read(cmdOutput)
		if err != nil {
			// EOF
			return
		}
		if n > 0 {
			err := wsConn.WriteMessage(websocket.TextMessage, cmdOutput)
			if err != nil {
				return
			}
		}
	}
}

func WebSSHHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var isLogin bool
	isLogin, _ = usermgmt.IsLogIn(w, r)
	if isLogin == false {
		GenResponseByObject(w, nil, errors.New("Please login!"))
		return
	}
	username := usermgmt.GetLoginUsername(r)
	var sshInput io.WriteCloser
	var sshOutput io.Reader //bytes.Buffer
	wsConn, err := websocket.Upgrade(w, r, nil, 1024, 1024*10)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer wsConn.Close()
	// Read SSH Parameters
	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		utils.CheckError("ReadMessage SSH Parameters Error:", err)
		return
	}
	if data.CFG.PrimaryNode.Admin.WebSSHEnabled == false {
		wsConn.WriteMessage(websocket.TextMessage, []byte("WebSSH disabled in config.json!\r\n"))
		return
	}
	var host HostInfo
	json.Unmarshal(msg, &host)
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte("Connecting "+host.IP+":"+host.Port+" ... Please wait a moment!\r\n")); err != nil {
		return
	}
	errChan := make(chan error)
	go SSH(&sshInput, &sshOutput, &host, errChan)
	err = <-errChan
	if err != nil {
		wsConn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}
	var logBuf bytes.Buffer
	outputTicker := time.NewTicker(100 * time.Millisecond)
	go RoutineOutput(outputTicker, wsConn, &sshOutput)
	for {
		select {
		case <-errChan:
			wsConn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			return
		default:
			if wsConn == nil {
				return
			}
			_, msg, err := wsConn.ReadMessage()
			if err != nil {
				return
			}
			//log.Printf("Received: %s %v\n", string(msg), msg)
			if sshInput != nil {
				go CmdLog(&logBuf, username, &host, &msg)
				if _, err := sshInput.Write(msg); err != nil {
					return
				}
			}
		}
	}
}

func CmdLog(logBuf *bytes.Buffer, username string, host *HostInfo, cmdChars *[]byte) {
	for i := 0; i < len(*cmdChars); i++ {
		cmdChar := (*cmdChars)[i]
		switch cmdChar {
		case '\r', '\n':
			cmdStr := logBuf.String()
			hostInfo := host.Username + "@" + host.IP + ":" + host.Port
			utils.DebugPrintln("WebSSH User:", username, hostInfo, "Command:", cmdStr)
			logBuf.Reset()
		default:
			//fmt.Println("Char:", cmdChar)
			logBuf.WriteByte(cmdChar)
		}
	}
}
