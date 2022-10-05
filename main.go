package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var phpVersions = map[string]string{
	"7.2": "php7.2",
	"7.4": "php7.4",
	"8.1": "php8.1",
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	content := ""

	if request.IsBase64Encoded {
		bytes, err := base64.StdEncoding.DecodeString(request.Body)

		if err != nil {
			return errorResponse("Invalid request data"), nil
		}

		content = string(bytes)
	} else {
		content = request.Body
	}

	requestedPhpVersion, ok := request.QueryStringParameters["version"]

	if !ok {
		return errorResponse("Missing version parameter"), nil
	}

	phpExecuteable, ok := phpVersions[requestedPhpVersion]

	if !ok {
		return errorResponse(fmt.Sprintf("Cannot find given php version: %s", requestedPhpVersion)), nil
	}

	mediaType, params, err := mime.ParseMediaType(request.Headers["content-type"])

	if err != nil {
		return errorResponse(fmt.Sprintf("Invalid Content-Type header: %s", err.Error())), nil
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return errorResponse(fmt.Sprintf("Invalid Content-Type header: %s", mediaType)), nil
	}

	mr := multipart.NewReader(strings.NewReader(content), params["boundary"])

	for {
		p, err := mr.NextPart()

		if err == io.EOF {
			break
		}

		if err != nil {
			return errorResponse(fmt.Sprintf("Invalid request data: %s", err.Error())), nil
		}

		if p.FormName() == "file" {
			fileContent, err := io.ReadAll(p)

			if err != nil {
				return errorResponse(fmt.Sprintf("Cannot read file: %s", err.Error())), nil
			}

			errors := validateAllFiles(fileContent, phpExecuteable)

			msg, _ := json.Marshal(map[string][]string{"errors": errors})

			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       string(msg),
			}, nil
		} else {
			return errorResponse(fmt.Sprintf("Invalid form field: %s", p.FormName())), nil
		}
	}
	return errorResponse("Invalid request body sent"), nil
}

func validateAllFiles(p []byte, phpExecuteable string) []string {
	zipReader, err := zip.NewReader(bytes.NewReader(p), int64(len(p)))

	if err != nil {
		return []string{fmt.Sprintf("Invalid zip given: %s", err.Error())}
	}

	errors := []string{}
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, ".php") {
			if err := validateFile(file, phpExecuteable); err != nil {
				errors = append(errors, fmt.Sprintf("File %s: %s", file.Name, err.Error()))
			}
		}
	}

	return errors
}

func validateFile(zipFile *zip.File, phpExecuteable string) error {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "php-syntax-checker")

	if err != nil {
		return err
	}

	opendZipFile, err := zipFile.Open()

	if err != nil {
		return err
	}

	_, err = io.Copy(tmpFile, opendZipFile)

	_ = opendZipFile.Close()

	if err != nil {
		return err
	}

	defer os.Remove(tmpFile.Name())

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(path.Join(cwd, phpExecuteable), "-l", tmpFile.Name())
	msg, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s", msg)
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return fmt.Errorf("%s", msg)
	}

	return nil

}

func main() {
	lambda.Start(handler)
}

func errorResponse(message string) events.APIGatewayProxyResponse {
	msg, _ := json.Marshal(map[string]string{"message": message})

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusForbidden,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(msg),
	}
}
