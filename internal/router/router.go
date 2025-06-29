package router

import (
	"iwut-smartclass-backend/assets"
	"iwut-smartclass-backend/internal/handler"
	"iwut-smartclass-backend/internal/util"
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Home)
	mux.HandleFunc("/getCourse", methodHandler(handler.GetCourse, http.MethodPost))
	mux.HandleFunc("/generateSummary", methodHandler(handler.GenerateSummary, http.MethodPost))

	fileSystem, _ := assets.GetFileSystem()
	mux.Handle("/favicon.ico", http.FileServer(fileSystem))
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
