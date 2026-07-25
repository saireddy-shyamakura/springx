package initializr

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Download(project string) error {
	url := fmt.Sprintf(
		"https://start.spring.io/starter.zip?type=maven-project&language=java&bootVersion=3.5.5&groupId=com.example&artifactId=%s&name=%s&packageName=com.example.%s&packaging=jar&javaVersion=21&dependencies=web",
		project,
		project,
		project,
	)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(project + ".zip")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
