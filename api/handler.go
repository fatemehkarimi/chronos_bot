package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/repository"
)

type Handler interface {
	GetUpdates(w http.ResponseWriter, r *http.Request)
}

type HttpHandler struct {
	db       repository.Repository
	updateId int
}

func NewHttpHandler(db repository.Repository) Handler {
	return &HttpHandler{db: db}
}

func (h *HttpHandler) GetUpdates(w http.ResponseWriter, r *http.Request) {
	var update entities.Update
	err := json.NewDecoder(r.Body).Decode(&update)

	fmt.Println("here error", err)

	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	fmt.Println("here body = ", update.UpdateId, update.CallbackQuery.Id, update.CallbackQuery.From, update.CallbackQuery.Message, update.CallbackQuery.Data)
	w.WriteHeader(http.StatusOK)
}
