package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
)

var validMediaRegexp = regexp.MustCompile(`^([a-z0-9]{8,128})(\.[a-z0-9]{2,4})$`)

func (s *Server) handleGetMedia(w http.ResponseWriter, r *http.Request) error {
	mediaFilename := r.PathValue("filename")
	mediaFilenameMatch := validMediaRegexp.FindStringSubmatch(mediaFilename)
	if mediaFilenameMatch == nil {
		return clientError(fmt.Errorf("not a valid media filename: %s", mediaFilename), http.StatusBadRequest)
	}
	digest, ext := mediaFilenameMatch[1], mediaFilenameMatch[2]
	objectPath := path.Join(s.ObjectDir, digest[:2], digest[2:])
	file, err := os.Open(objectPath)
	if os.IsNotExist(err) {
		return clientError(fmt.Errorf("no such media: %s", digest), http.StatusNotFound)
	} else if err != nil {
		return err
	}
	defer file.Close()
	st, err := file.Stat()
	if err != nil {
		return err
	}
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return clientError(fmt.Errorf("unsupported file extension: %s", ext), http.StatusBadRequest)
	}
	w.Header().Set("content-type", mimeType)
	w.Header().Set("content-length", fmt.Sprintf("%d", st.Size()))
	_, err = io.Copy(w, file)
	if err != nil {
		s.Log.Printf("Error while writing media response: %s", err.Error())
	}
	return nil
}
