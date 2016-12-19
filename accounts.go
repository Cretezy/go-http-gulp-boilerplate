package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func GetAccount(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	templateManager.RenderView(w, r, "account", nil, nil)
}