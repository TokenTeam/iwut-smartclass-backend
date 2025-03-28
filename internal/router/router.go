package router

import (
	"iwut-smart-timetable-backend/internal/handler"
	"iwut-smart-timetable-backend/internal/util"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/getCourse", methodHandler(handler.GetCourse, http.MethodPost))
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/favicon.ico")
	})
	return mux
}

// methodHandler HTTP 方法处理
func methodHandler(h http.HandlerFunc, allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, method := range allowedMethods {
			if r.Method == method {
				h.ServeHTTP(w, r)
				return
			}
		}
		util.WriteResponse(w, http.StatusMethodNotAllowed, nil)
	}
}
