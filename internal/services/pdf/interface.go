package pdf

import "os"

type HtmlToPdfConverterFiles struct {
	Name         string
	ContentBytes []byte
}

type HtmlToPdfConverter interface {
	ConvertHTMLToPDF(files []HtmlToPdfConverterFiles, filenamePrefix string) (*os.File, error)
}
