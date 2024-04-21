package handlers

import "github.com/DrC0ns0le/bind-api/rdb"

type responseBody struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var bd *rdb.BindData

func Init(data *rdb.BindData) {
	bd = data
}
