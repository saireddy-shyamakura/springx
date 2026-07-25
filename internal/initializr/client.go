package initializr

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func Download(project string) (string, error) {
	url := fmt.Sprintf(
		"https://start.spring.io/starter.zip?type=maven-project&language=java&groupId=com.example&artifactId=%s&name=%s&packageName=com.example.%s&packaging=jar&javaVersion=21&dependencies=web",
		project,
		project,
		project,
	)

	filename := project + ".zip"

	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spring initializr returned %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}
