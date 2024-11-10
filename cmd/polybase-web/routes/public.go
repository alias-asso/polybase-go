package routes

import (
	"log"
	"net/http"
)

func getHome(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Get home - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Get home"))
}

func getLogin(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Get login - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Get login"))
}

func postAuth(ctx *ServerContext, w http.ResponseWriter, r *http.Request) {
	log.Printf("Post auth - Config: %+v, DB: %+v", ctx.Config, ctx.DB)
	w.Write([]byte("Post auth"))
}
