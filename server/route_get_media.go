package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
)

var validMediaRegexp = regexp.MustCompile(`^([a-z0-9]{8,128})(\.[a-z0-9]{2,4})$`)

func (s *Server) handleGetMedia(w http.ResponseWriter, r *http.Request) error {
	var err error
	mediaFilename := r.PathValue("filename")
	mediaFilenameMatch := validMediaRegexp.FindStringSubmatch(mediaFilename)
	if mediaFilenameMatch == nil {
		return clientErrorf("not a valid media filename: %s", mediaFilename)
	}
	digest, ext := mediaFilenameMatch[1], mediaFilenameMatch[2]
	widthStr := r.URL.Query().Get("w")
	width := 0
	if widthStr != "" {
		width, err = strconv.Atoi(widthStr)
		if err != nil || width <= 0 || width > 32768 {
			return clientErrorf("not a valid width: %s", widthStr)
		}
		if ext != ".jpg" && ext != ".jpeg" {
			return clientErrorf("cannot specify width for non-jpeg format: %s", ext)
		}
	}
	objectPath := path.Join(s.FilestoreDir, "objects", digest[:2], digest[2:])
	thumbnailPath := path.Join(s.FilestoreDir, "thumbnails", fmt.Sprintf("%05d", width), digest[:2], digest[2:])
	var file *os.File
	if width > 0 {
		file, err = os.Open(thumbnailPath)
		if os.IsNotExist(err) {
			err := s.generateThumbnail(objectPath, thumbnailPath, width)
			if err != nil {
				return fmt.Errorf("generating thumbnail: %w", err)
			}
		}
		file, err = os.Open(thumbnailPath)
	} else {
		file, err = os.Open(objectPath)
	}
	if os.IsNotExist(err) {
		return clientErrorf("no such media: %s", digest)
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
		return clientErrorf("unsupported file extension: %s", ext)
	}
	w.Header().Set("content-type", mimeType)
	w.Header().Set("content-length", fmt.Sprintf("%d", st.Size()))
	w.Header().Set("cache-control", "max-age=31536000, immutable")
	_, err = io.Copy(w, file)
	if err != nil {
		s.Log.Printf("Error while writing media response: %s", err.Error())
	}
	return nil
}
