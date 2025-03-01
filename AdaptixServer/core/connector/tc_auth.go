package connector

import (
	"AdaptixServer/core/utils/krypt"
	"AdaptixServer/core/utils/logs"
	"AdaptixServer/core/utils/token"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccessJWT struct {
	AccessToken string `json:"access_token"`
}

func (tc *TsConnector) tcLogin(ctx *gin.Context) {
	var (
		creds Credentials
		err   error
	)

	err = ctx.ShouldBindJSON(&creds)
	if err != nil {
		_ = ctx.Error(errors.New("invalid credentials"))
		return
	}

	recvHash := krypt.SHA256([]byte(creds.Password))

	if recvHash != tc.Hash {
		_ = ctx.Error(errors.New("incorrect password"))
		return
	}

	accessToken, err := token.GenerateAccessToken(creds.Username)
	if err != nil {
		_ = ctx.Error(errors.New("could not generate access token"))
		return
	}

	refreshToken, err := token.GenerateRefreshToken(creds.Username)
	if err != nil {
		_ = ctx.Error(errors.New("could not generate refresh token"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

func (tc *TsConnector) tcConnect(ctx *gin.Context) {
	var (
		wsUpgrader websocket.Upgrader
		wsConn     *websocket.Conn
		err        error
	)

	wsConn, err = wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logs.Error("", "WebSocket upgrade error: "+err.Error())
		return
	}

	if wsConn == nil {
		logs.Error("", "WebSocket is nil")
		return
	}

	go tc.tcWebsocketConnect(wsConn)
}

func (tc *TsConnector) tcWebsocketConnect(wsConn *websocket.Conn) {
	var (
		body        []byte
		err         error
		structToken AccessJWT
		username    string
	)

	_, body, err = wsConn.ReadMessage()
	if err != nil {
		logs.Error("", "Failed ReadMessage from WebSocket: "+err.Error())
		return
	}

	err = json.Unmarshal(body, &structToken)
	if err != nil {
		logs.Error("", "JSON Unmarshal error: "+err.Error())
		return
	}

	username, err = token.GetUsernameFromJWT(structToken.AccessToken)
	if err != nil {
		logs.Error("", "Invalid JWT error: "+err.Error())
		return
	}

	tc.teamserver.TsClientConnect(username, wsConn)

	for {
		if _, _, err = wsConn.ReadMessage(); err == nil {
			continue
		}

		logs.Debug("Client '%s' disconnected: %s\n", username, err.Error())

		tc.teamserver.TsClientDisconnect(username)
		break
	}
}
