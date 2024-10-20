package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
)

var pdfTotalPages int
var txtSentences []string

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
	splitTextIntoSentences(path.Join("./tmp", fileName), fileName)
	postSentencesToAPI(txtSentences, fileName, pdfLang )
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

	for i := 0; i < pdfTotalPages; i++ {

		fmt.Printf("Converting page %d / %d\n", i+1, pdfTotalPages)
		img, err := pdf.Image(i)
		if err != nil {
			log.Fatalf("Error while converting page Number %d, to image: %v", i+1, err)
		}

		f, err := os.Create(filepath.Join(imgsDirPath, fmt.Sprintf("%s-%04d.png", fileName, i)))
		if err != nil {
			log.Fatalf("Error while saving image Number %d: %v", i+1, err)
		}
		defer f.Close()

		err = png.Encode(f, img)
		if err != nil {
			log.Fatalf("Error while encoding PNG Number %d: %v", i+1, err)
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
		fmt.Printf("Extracting page %d / %d\n", idx+1, pdfTotalPages)
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

func splitTextIntoSentences(src, fileName string) {
	fmt.Println("Splitting text into sentences")
	startTime := time.Now()

	txtFile, err := os.ReadFile(filepath.Join(src, "texts", fmt.Sprintf("%s.txt", fileName)))
	if err != nil {
		log.Fatalf("Error while reading txt file: %v", err)
	}

	txt := string(txtFile)

	txt = strings.ReplaceAll(txt, "\n", " ")
	txt = strings.TrimSpace(txt)

	for _, sentence := range strings.Split(txt, ".") {
		trimmedSentence := strings.TrimSpace(sentence)
		if trimmedSentence != "" {
			txtSentences = append(txtSentences, trimmedSentence)
		}
	}

	durationTime := time.Since(startTime)
	fmt.Println("Finished splitting text")
	fmt.Printf("Took %.2f seconds\n", durationTime.Seconds())
}

func postSentencesToAPI(stcs []string, fileName, lang string) {
	fmt.Println("Generating audios")
	startTime := time.Now()

	_, err := http.Get("http://127.0.0.1:7851/api/ready")
	if err != nil {
		var tryAgain string
		fmt.Println("Not able to connect to http://127.0.0.1:7851")
		fmt.Println("Is the AllTalk Running?")
		fmt.Printf("Error: %s\n", err)
		fmt.Println("Try Again [y/n]?")
		fmt.Scan(&tryAgain)
		if tryAgain == "y" {
			postSentencesToAPI(stcs, fileName, lang)
			return
		} else {
			fmt.Println("Aborted")
			os.Exit(1)
		}
	}

	for idx, stc := range stcs {
    fmt.Printf("Generating audio %d / %d\n", idx+1, len(stcs))
		formData := url.Values{
			"text_input":            {stc},
			"text_filtering":        {"standard"},
			"character_voice_gen":   {"female_01.wav"},
			"narrator_enabled":      {"false"},
			"narrator_voice_gen":    {"male_01.wav"},
			"text_not_inside":       {"character"},
			"language":              {"pt"},
			"output_file_name":      {fmt.Sprintf("%s_%04d", fileName, idx+1)},
			"output_file_timestamp": {"true"},
			"autoplay":              {"false"},
			"autoplay_volume":       {"0.1"},
		}
		postSent, err := http.PostForm(
			"http://127.0.0.1:7851/api/tts-generate",
			formData,
		)
		if err != nil {
			log.Fatalf("Error while posting form to API: %v", err)
		}
		defer postSent.Body.Close()

		bd, err := io.ReadAll(postSent.Body)
		if err != nil {
			log.Fatalf("Error while reading Response Body: %v", err)
		}

    type TTSResponse struct {
      Status         string `json:"status"`
      OutputFilePath string `json:"output_file_path"`
      OutputFileURL  string `json:"output_file_url"`
      OutputCacheURL string `json:"output_cache_url"`
    }

    var resJson TTSResponse

    err = json.Unmarshal(bd, &resJson)
    if err != nil {
      log.Fatalf("Error unmarshaling res: %v", err)
    }

    saveAudioFile(fileName, resJson.OutputFilePath)
	}
	durationTime := time.Since(startTime)
	fmt.Println("Finished Generating Audios")
	fmt.Printf("Took %.2f seconds\n", durationTime.Seconds())
}

func saveAudioFile(fileName, audioSrc string) {
  audiosDirPath := path.Join("./tmp", fileName, "audios")
  if _ , err := os.Stat(audiosDirPath); os.IsNotExist(err){
    err := os.Mkdir(audiosDirPath, os.ModePerm)
    if err != nil {
      log.Fatalf("Error while creating audio dir: %v", err)
    }
  }

  audioName := filepath.Base(audioSrc)

  err := os.Rename(audioSrc, fmt.Sprintf("%s/%s", audiosDirPath, audioName))
  if err != nil {
    log.Fatalf("Error while moving audio: %v", err)
  }
  fmt.Printf("Audio saved at: %s/%s\n", audiosDirPath, audioName)
}
