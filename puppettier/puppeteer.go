package puppettier

import (
	"errors"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"image"
	"os"
	"os/exec"
	"path"
	"strconv"
)

func RenderInPuppeteer(html, filePrefix string, size image.Point) ([]byte, error) {
	pwd := utils.GetRootDir()
	htmlFileName := path.Join(pwd, filePrefix+".html")
	outputFileName := path.Join(pwd, filePrefix+".png")
	err := os.WriteFile(htmlFileName, []byte(html), 0755)
	if err != nil {
		return nil, err
	}
	err = callPuppeteer(htmlFileName, outputFileName, size, pwd)
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

func callPuppeteer(htmlFileName string, pngFileName string, size image.Point, pwd string) error {
	// we pass 4 parameters to the script:
	// source HTML file name
	// target png file name
	// width of the output image
	// height of the output image
	cmd := exec.Command(
		"/usr/bin/node",
		path.Join(pwd, "renderHtml.js"),
		htmlFileName,
		pngFileName,
		strconv.Itoa(size.X),
		strconv.Itoa(size.Y),
	)
	cmd.Dir = pwd
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		print(string(output))
	}
	if err != nil {
		return err
	}
	return nil
}
