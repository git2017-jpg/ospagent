package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/openspacee/ospagent/pkg/utils"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"time"
)

type SendResponse func(interface{}, string, string)

type WebSocket struct {
	Url          *url.URL
	Token        string
	RequestChan  chan *utils.Request
	ResponseChan chan *utils.TResponse
	Conn         *websocket.Conn
}

func NewWebSocket(
	url *url.URL,
	token string,
	requestChan chan *utils.Request,
	responseChan chan *utils.TResponse) *WebSocket {
	return &WebSocket{
		Url:          url,
		Token:        token,
		RequestChan:  requestChan,
		ResponseChan: responseChan,
	}
}

func (ws *WebSocket) ReadRequest() {
	defer ws.Conn.Close()

	ws.reconnectServer()
	for {
		_, data, err := ws.Conn.ReadMessage()
		if err != nil {
			klog.Error("read err:", err)
			ws.Conn.Close()
			ws.reconnectServer()
			continue
		}
		klog.V(1).Infof("request data: %s", string(data))
		request := &utils.Request{}
		err = json.Unmarshal(data, request)
		if err != nil {
			klog.Errorf("unserializer request data error: %s", err)
		} else {
			ws.RequestChan <- request
		}
	}
}

func (ws *WebSocket) reconnectServer() {
	err := ws.connectServer()
	if err == nil {
		return
	}
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := ws.connectServer()
			if err == nil {
				return
			}
		}
	}
}

func (ws *WebSocket) connectServer() error {
	klog.Info("start connect to server", ws.Url.String())
	wsHeader := http.Header{}
	wsHeader.Add("token", ws.Token)
	conn, _, err := websocket.DefaultDialer.Dial(ws.Url.String(), wsHeader)
	if err != nil {
		klog.Infof("connect to server %s error: %v, retry after 5 seconds\n", ws.Url.String(), err)
		return err
	} else {
		klog.Infof("connect to server %s success\n", ws.Url.String())
		ws.Conn = conn
		return nil
	}
}

func (ws *WebSocket) WriteResponse() {
	for {
		select {
		case resp, ok := <-ws.ResponseChan:
			if ok {
				respMsg, err := resp.Serializer()
				if err != nil {
					klog.Errorf("response %v serializer error: %s", resp, err)
				}
				err = ws.Conn.WriteMessage(websocket.TextMessage, respMsg)
				if err != nil {
					klog.Errorf("write response %s err: %s", string(respMsg), err)
					continue
				}
				klog.V(1).Infof("write response %s success", string(respMsg))
			}
		}
	}
}

func (ws *WebSocket) SendResponse(resp interface{}, requestId, resType string) {
	if ws.Conn != nil {
		tResp := &utils.TResponse{RequestId: requestId, Data: resp, ResType: resType}
		ws.ResponseChan <- tResp
	}
}
