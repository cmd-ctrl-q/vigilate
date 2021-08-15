package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/pusher/pusher-http-go"
)

func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	// is this user authenticated?

	// get userID from session
	userID := repo.App.Session.GetInt(r.Context(), "userID")

	// get the user from the db
	u, _ := repo.DB.GetUserById(userID)

	// authenticate to pusher
	params, _ := io.ReadAll(r.Body)

	// create type used by pusher client
	// member info used to connect them to pusher server.
	presenceData := pusher.MemberData{
		UserID: strconv.Itoa(userID),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   strconv.Itoa(userID),
		},
	}

	response, err := app.WsClient.AuthenticatePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	// write json response back to user
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(response)
}
