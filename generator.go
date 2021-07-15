package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PackMetaFile struct {
	Pack PackMeta `json:"pack"`
}

type PackMeta struct {
	PackFormat  int    `json:"pack_format"`
	Description string `json:"description"`
}

func HandleError(err error) {
	if err != nil {
		log.Println(err)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Program encountered an error, press enter to exit.")
		reader.ReadString('\n')
		os.Exit(1)
	}
}

func main() {
	// I did this in like 20 minutes don't judge me.
	pack := "pack.zip"
	tempPath := "./temp"
	dataPackPath := "./datapack"

	fmt.Println("Welcome to Minecraft Data Pack Generator")

	fmt.Println("Cleaning wrokspace...")
	err := os.RemoveAll(dataPackPath)
	HandleError(err)
	err = os.RemoveAll(tempPath)
	HandleError(err)

	url := "https://github.com/CanPacis/datapack-template/archive/refs/heads/main.zip"
	fmt.Println("Downloading latest template...")
	err = DownloadFile(pack, url)
	HandleError(err)
	fmt.Println("Downloaded")
	fmt.Println("Extracting latest template...")
	_, err = Unzip(pack, tempPath)
	HandleError(err)
	fmt.Println("Extracted")

	err = os.Rename(tempPath+"/datapack-template-main", dataPackPath)
	HandleError(err)
	err = os.Remove(tempPath)
	HandleError(err)
	err = os.Remove(pack)
	HandleError(err)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Give your datapack a name: ")
	name, err := reader.ReadString('\n')
	HandleError(err)
	fmt.Print("And give it a simple description: ")
	desc, err := reader.ReadString('\n')
	HandleError(err)

	safeDesc := strings.TrimSpace(desc)
	safeName := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "_")
	err = os.Rename(fmt.Sprintf("%s/data/template/", dataPackPath), fmt.Sprintf("%s/data/%s/", dataPackPath, safeName))
	HandleError(err)

	// Look I know this code is not the best but deal with it.
	os.WriteFile(fmt.Sprintf("%s/data/minecraft/tags/functions/load.json", dataPackPath), []byte(fmt.Sprintf("{\"values\": [\"%s:load\"]}", safeName)), 0644)
	os.WriteFile(fmt.Sprintf("%s/data/minecraft/tags/functions/tick.json", dataPackPath), []byte(fmt.Sprintf("{\"values\": [\"%s:tick\"]}", safeName)), 0644)

	meta := PackMetaFile{
		Pack: PackMeta{
			PackFormat:  7,
			Description: safeDesc,
		},
	}

	jsonStr, err := json.Marshal(meta)
	HandleError(err)
	os.WriteFile(fmt.Sprintf("%s/pack.mcmeta", dataPackPath), jsonStr, 0644)

	fmt.Print("Program has finished, press enter to exit.")
	reader.ReadString('\n')
	os.Exit(1)
}

// And below this point, I stole everything from a magical place called 'the internet'

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func Unzip(src string, dest string) ([]string, error) {
	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
