package main

import (
	"net/http"
  "bytes"
  "encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func initServer(){
  r := chi.NewRouter()
  r.Use(middleware.Logger)
  r.Get("/GetConsumerInfo/{ChatID}", GetConsumerInfoAPI)
  r.Get("/GetVacantPeer/", GetVacantPeerAPI)
  r.Get("/", func(w http.ResponseWriter, r *http.Request) {
      w.Write([]byte("Hello World!"))
  })
  http.ListenAndServe(":3000", r)
}

func DrawJSON(w http.ResponseWriter, v interface{}, statusCode int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}

func GetConsumerInfoAPI(w http.ResponseWriter, r *http.Request){
  var consumer ConsGorm
  consumer.ChatID = chi.URLParam(r, "ChatID")
  consumer, err := GetConsumerInfoDB(consumer)
  // DrawJSON(w, consumer, 200)
  if consumer.Username == "" {
    w.WriteHeader(404)
    w.Write([]byte("Consumer not found!"))
    // lg.Printf("Consumer %s not found!", consumer.ChatID)
    return
  }
  if err != nil{
    w.WriteHeader(422)
    w.Write([]byte(err.Error()))
    return
  }
  DrawJSON(w, consumer, 200)
}

func GetVacantPeerAPI(w http.ResponseWriter, r *http.Request){
  var vacantPeer PeerGorm
  vacantPeer, err := GetVacantPeerFromORM()
  if err != nil{
    w.WriteHeader(422)
    lg.Printf("Failed to get vacant peer %s", err)
    return
  }
  if vacantPeer.AllowedIP == ""{
    w.WriteHeader(404)
    lg.Println("No vacant peers")
    return
  }
  DrawJSON(w, vacantPeer, 200)
  lg.Printf("Vacant peer: %s", vacantPeer.Name)

}

