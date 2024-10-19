package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
)

var pdfTotalPages int

func main() {
	pdfCmd := flag.NewFlagSet("pdf", flag.ExitOnError)
	pdfName := pdfCmd.String("name", "", "Name to be saved")
	pdfPath := pdfCmd.String("path", "", "PDF file path")
	pdfLang := pdfCmd.String("lang", "por", "PDF language")

	if len(os.Args) < 2 {
		fmt.Println("Usage: AllPDF <subcommand> [flags]")
		fmt.Println("Subcommands:")
		fmt.Println("  pdf")
		fmt.Println("Flags:")
		fmt.Println("  -path string")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "pdf":
		pdfCmd.Parse(os.Args[2:])
		fmt.Println("It is recommended to remove any pages of the PDF that you don't want to be Narrated")
		fmt.Println("cover/illustrations/tables/graphs/charts")
		fmt.Println("Is the path correct? [y/n]")
		fmt.Println(*pdfPath)
		var option string
		fmt.Scanln(&option)
		if option == "y" {
			start(*pdfPath, *pdfName, *pdfLang)
		} else {
			fmt.Println("Aborted")
			os.Exit(1)
		}
	default:
		fmt.Println("subcommand is required (pdf)")
		os.Exit(1)
	}
}

func start(src, fileName, pdfLang string) {
	cpFile(src, fileName)
	convertToImgs(path.Join("./tmp", fileName), fileName)
	extractTexts(path.Join("./tmp", fileName), fileName, pdfLang)
}

func cpFile(src, fileName string) (dst string) {
	err := os.RemoveAll("./tmp")
	if err != nil {
		log.Fatalf("Error removing tmp folder: %v", err)
	}

	file, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("Error while reading file: %v", err)
	}

	err = os.MkdirAll(path.Join("./tmp", fileName), os.ModePerm)
	if err != nil {
		log.Fatalf("Error while creating directories: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("./tmp/%s/%s.pdf", fileName, fileName), file, os.ModePerm)
	if err != nil {
		log.Fatalf("Error while saving file: %v", err)
	}

	return fmt.Sprintf("./tmp/%s", fileName)
}

func convertToImgs(src, fileName string) {
	fmt.Println("Starting converting to images")
	startTime := time.Now()
	pdf, err := fitz.New(fmt.Sprintf("%s/%s.pdf", src, fileName))
	if err != nil {
		log.Fatalf("Error while opening file with Fitz: %v", err)
	}
	defer pdf.Close()

	imgsDirPath := path.Join(src, "imgs")
	err = os.Mkdir(imgsDirPath, os.ModePerm)
  pdfTotalPages = pdf.NumPage()

	for i := 0; i < pdfTotalPages / 15; i++ {

		fmt.Printf("Converting page %d / %d\n", i, pdfTotalPages / 15)
		img, err := pdf.Image(i)
		if err != nil {
			log.Fatalf("Error while converting page Number %d, to image: %v", i, err)
		}

		f, err := os.Create(filepath.Join(imgsDirPath, fmt.Sprintf("%s-%04d.png", fileName, i)))
		if err != nil {
			log.Fatalf("Error while saving image Number %d: %v", i, err)
		}
		defer f.Close()

		err = png.Encode(f, img)
		if err != nil {
			log.Fatalf("Error while encoding PNG Number %d: %v", i, err)
		}

	}

	durationTime := time.Since(startTime)
	fmt.Println("Finished converting to images")
	fmt.Printf("Took %.2f seconds\n", durationTime.Seconds())
}

func extractTexts(src, fileName, pdfLang string) {
	fmt.Println("Starting text extraction")
	startTime := time.Now()

	textsDirPath := path.Join(src, "texts")
	err := os.Mkdir(textsDirPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Error while creating text dir path: %v", err)
	}
	txtFile, err := os.OpenFile(filepath.Join(textsDirPath, fmt.Sprintf("%s.txt", fileName)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error while creating text file: %v", err)
	}
	defer txtFile.Close()

	imgsDirPath := path.Join(src, "imgs")
	imgs, err := os.ReadDir(imgsDirPath)
	if err != nil {
		log.Fatalf("Erro while creating getting imgs at dir path: %v", err)
	}

	client := gosseract.NewClient()
	client.SetLanguage(pdfLang)
	defer client.Close()

	for idx, img := range imgs {
    fmt.Printf("Extracting page %d/%d\n", idx, pdfTotalPages / 15)
		client.SetImage(filepath.Join(imgsDirPath, img.Name()))
		txt, err := client.Text()
		if err != nil {
			log.Fatalf("Erro while extracting text from image %d: %v", idx, err)
		}
		txtFile.WriteString(txt)
	}

	durationTime := time.Since(startTime)
	fmt.Println("Finished text extraction")
	fmt.Printf("Took %.2f seconds\n", durationTime.Seconds())
}
