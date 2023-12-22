package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"go.uber.org/zap"
)

type HttpWriter struct {
	logger *zap.Logger
}

func (writer *HttpWriter) writeError(w http.ResponseWriter, code int, err error) {
	writer.logger.Error("HTTP request error", zap.Error(err))

	http.Error(w, err.Error(), code)
}

func (writer *HttpWriter) writeResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	
	_, err := w.Write([]byte(message))
	if err != nil {
		err = fmt.Errorf("Http writing error: %w", err)
		writer.writeError(w, http.StatusInternalServerError, err)
		return
	}

	_, err = w.Write([]byte("\n"))
	if err != nil {
		err = fmt.Errorf("Http writing error: %w", err)
		writer.writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func (writer *HttpWriter) writeJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("Data marshal error: %w", err)
		writer.writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	writer.writeResponse(w, status, string(response))
}

func NewHttpWriter(logger *zap.Logger) *HttpWriter {

	return &HttpWriter{
		logger: logger,
	}
}