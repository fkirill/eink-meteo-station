package puppettier

import (
	"errors"
	"image"
	"io/ioutil"
	"os"
	"os/exec"
	"renderable/utils"
	"strconv"
)

func RenderInPuppeteer(html, filePrefix string, size image.Point) ([]byte, error) {
	htmlFileName := filePrefix + ".html"
	outputFileName := filePrefix + ".png"
	err := ioutil.WriteFile(htmlFileName, []byte(html), 0755)
	if err != nil {
		return nil, err
	}
	err = callPuppeteer(htmlFileName, outputFileName, size)
	if err != nil {
		return nil, err
	}
	err = os.Remove(htmlFileName)
	if err != nil {
		return nil, err
	}
	img, err := utils.LoadImage(outputFileName)
	if err != nil {
		return nil, err
	}
	if img.Bounds().Size().X != size.X || img.Bounds().Size().Y != size.Y {
		return nil, errors.New("Unexpected png image size")
	}
	err = os.Remove(outputFileName)
	if err != nil {
		return nil, err
	}
	raster, err := utils.ConvertToGrayScale(img)
	if err != nil {
		return nil, err
	}
	return raster, nil
}

func callPuppeteer(htmlFileName string, pngFileName string, size image.Point) error {
	// we pass 4 parameters to the script:
	// source HTML file name
	// target png file name
	// width of the output image
	// height of the output image
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command(
		"/usr/bin/node",
		"renderHtml.js",
		pwd+"/"+htmlFileName,
		pwd+"/"+pngFileName,
		strconv.Itoa(size.X),
		strconv.Itoa(size.Y))
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		print(string(output))
	}
	if err != nil {
		return err
	}
	return nil
}
