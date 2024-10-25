This code is a command-line application written in Go that performs the following tasks:

1. Convert a PDF file into images.
2. Extract text from the images.
3. Split the extracted text into sentences.
4. Send each sentence to an API ([alltalk_tts](https://github.com/erew123/alltalk_tts)) to generate audio files.
5. Concatenate the generated audio files into a single audio file.

To run the application, use the following command:
```shell
go run main.go <subcommand> [flags]
```

Available subcommands:
- `pdf`: Converts a PDF file into images, extracts text, splits text into sentences, sends sentences to the API, and concatenates the generated audio files.

Flags:
- `-name`: Specifies the name of the output file.
- `-path`: Specifies the path of the PDF file.
- `-lang`: Specifies the language of the PDF file.
- `-delay`: Specifies the delay between sentences in milliseconds.

Example usage:
```shell
go run main.go pdf -name=mybook -path=/path/to/mybook.pdf -lang=pt -delay=1000
```

Note:
- Before running the application, make sure to have the necessary dependencies installed.
- The application assumes the availability of a local API at `http://127.0.0.1:7851/api`.
- The generated audio files will be stored in a `./tmp` directory.

Please ensure that you have the necessary permissions and dependencies in place before running this application.


