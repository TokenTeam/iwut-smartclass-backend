package handler

import (
	"iwut-smart-timetable-backend/internal/util"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	util.WriteResponse(w, http.StatusForbidden, nil)
}
