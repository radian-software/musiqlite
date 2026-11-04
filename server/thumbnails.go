package main

import (
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path"
	"sync"

	"golang.org/x/image/draw"
)

func (s *Server) generateThumbnail(input string, output string, outputWidth int) error {
	// Lock so that two concurrent requests can't cause the same
	// thumbnail file to be written in parallel, which might
	// corrupt the data.
	lockAny, _ := s.lockData.ThumbnailLocks.LoadOrStore(output, &sync.Mutex{})
	lock := lockAny.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()
	inputFile, err := os.Open(input)
	if os.IsNotExist(err) {
		// If original file doesn't exist, then the
		// corresponding thumbnail shouldn't exist either, we
		// call that a non-error result
		return nil
	} else if err != nil {
		return err
	}
	defer inputFile.Close()
	inputImage, _, err := image.Decode(inputFile)
	if err != nil {
		return err
	}
	var outputImage image.Image
	if outputWidth >= inputImage.Bounds().Dx() {
		outputImage = inputImage
	} else {
		outputHeight := inputImage.Bounds().Dy() * outputWidth / inputImage.Bounds().Dx()
		newOutputImage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{outputWidth, outputHeight}})
		draw.CatmullRom.Scale(newOutputImage, newOutputImage.Bounds(), inputImage, inputImage.Bounds(), draw.Src, nil)
		outputImage = newOutputImage
	}
	err = os.MkdirAll(path.Dir(output), 0o777)
	if err != nil {
		return err
	}
	outputFile, err := os.Create(output + ".tmp")
	if err != nil {
		return err
	}
	defer outputFile.Close()
	err = jpeg.Encode(outputFile, outputImage, nil)
	if err != nil {
		return err
	}
	err = outputFile.Close()
	if err != nil {
		return err
	}
	err = os.Rename(output+".tmp", output)
	if err != nil {
		return err
	}
	return nil
}
