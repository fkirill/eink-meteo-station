package puppettier

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fkirill.org/eink-meteo-station/renderable/utils"
	"github.com/rotisserie/eris"
	"github.com/tidwall/go-node"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"
)

type HTMLRenderer interface {
	Render(html string, size image.Point) (image.Image, error)
	Close()
}

// ps -ae | grep chromium | awk '{print $1}' | xargs sudo kill -9
func NewHTMLRenderer() (HTMLRenderer, error) {
	res := &htmlRenderer{
		wg:             &sync.WaitGroup{},
		syncExec:       &sync.Mutex{},
		renderTemplate: template.Must(template.New("render").Parse(renderTemplateText)),
	}
	res.nodeVm = node.New(&node.Options{
		OnLog: func(s string) {
			res.logResp = s
			res.wg.Done()
		},
		Dir: utils.GetRootDir(),
	})
	cmdRes := res.nodeVm.Run("puppeteer = require('puppeteer');")
	if cmdRes.Error() != nil {
		return nil, eris.Wrap(cmdRes.Error(), "Error loading puppeteer library. Did you run 'npm install'?")
	}
	cmdRes = res.nodeVm.Run(`browserPromise = puppeteer.launch({
        headless: true,
        executablePath: '/usr/bin/chromium',
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });`)
	if cmdRes.Error() != nil {
		return nil, eris.Wrap(cmdRes.Error(), "Error starting chromium")
	}
	return res, nil
}

type htmlRenderer struct {
	wg             *sync.WaitGroup
	syncExec       *sync.Mutex
	nodeVm         node.VM
	logResp        string
	renderTemplate *template.Template
}

type renderData struct {
	HtmlContent   string
	Width, Height int
}

var renderTemplateText = "(async () => {\n" +
	"    const htmlContent = {{ .HtmlContent }};\n" +
	"    const browser = await browserPromise;\n" +
	"    const page = await browser.newPage();\n" +
	"    await page.setViewport({ width: {{ .Width }}, height: {{ .Height }} });\n" +
	"    await page.setContent(htmlContent, {options: {waitUntil:\"networkidle0\"}});\n" +
	"    const buf = await page.screenshot({ encoding: \"base64\", type: \"png\"});\n" +
	"    await page.close();\n" +
	"    console.log(buf);\n" +
	"})();"

func (h *htmlRenderer) Render(htmlContent string, size image.Point) (image.Image, error) {
	h.syncExec.Lock()
	defer h.syncExec.Unlock()
	h.wg.Add(1)
	jsBuf := bytes.Buffer{}
	err := h.renderTemplate.Execute(&jsBuf, renderData{
		HtmlContent: "`" + htmlContent + "`",
		Width:       size.X,
		Height:      size.Y,
	})
	js := string(jsBuf.Bytes())
	cmdRes := h.nodeVm.Run(js)
	if cmdRes.Error() != nil {
		h.wg.Done()
		return nil, eris.Wrap(cmdRes.Error(), "Error executing javascript to render the page")
	}
	h.wg.Wait()
	imgBuf, err := base64.StdEncoding.DecodeString(h.logResp)
	if err != nil {
		return nil, eris.Wrap(err, "Error decoding png image from base64")
	}
	img, _, err := image.Decode(bytes.NewReader(imgBuf))
	if err != nil {
		return nil, eris.Wrap(err, "Error loading png image")
	}
	return img, nil
}

func (h *htmlRenderer) Close() {
	h.syncExec.Lock()
	defer h.syncExec.Unlock()
	h.nodeVm.Run("browserPromise.then((b) => { b.close(); }")
}

var renderer HTMLRenderer

func getHTMLRenderer() HTMLRenderer {
	if renderer != nil {
		return renderer
	}
	var err error
	renderer, err = NewHTMLRenderer()
	if err != nil {
		panic(err)
	}
	return renderer
}

const writePngFiles = false

func RenderInPuppeteer(html, filePrefix string, size image.Point) ([]byte, error) {
	//pwd := utils.GetRootDir()
	//htmlFileName := path.Join(pwd, filePrefix+".html")
	//outputFileName := path.Join(pwd, filePrefix+".png")
	//err := os.WriteFile(htmlFileName, []byte(html), 0755)
	//if err != nil {
	//	return nil, err
	//}
	//err = callPuppeteer(htmlFileName, outputFileName, size, pwd)
	//if err != nil {
	//	return nil, err
	//}
	//err = os.Remove(htmlFileName)
	//if err != nil {
	//	return nil, err
	//}
	//img, err := utils.LoadImage(outputFileName)
	img, err := getHTMLRenderer().Render(html, size)
	if err != nil {
		return nil, err
	}
	if img.Bounds().Size().X != size.X || img.Bounds().Size().Y != size.Y {
		return nil, errors.New("Unexpected png image size")
	}
	if writePngFiles {
		fileName := time.Now().Format(time.RFC3339Nano) + ".png"
		filePath := filepath.Join(utils.GetRootDir(), fileName)
		file, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		err = png.Encode(file, img)
		if err != nil {
			panic(err)
		}
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}
	//err = os.Remove(outputFileName)
	//if err != nil {
	//	return nil, err
	//}
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
