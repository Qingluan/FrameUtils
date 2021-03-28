package task

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/fatih/color"
)

func (tconfig *TaskConfig) uploadFile(w http.ResponseWriter, r *http.Request) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	log.Println(green("File Upload Endpoint Hit"))

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 1 GB files.
	r.ParseMultipartForm(1024 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file

	id := r.FormValue("id")
	// fmt.Println(id, r.Form,r.Fo)
	file, handler, err := r.FormFile(id)
	if err != nil {
		log.Println("Error Retrieving the File", "|", id, "|")
		log.Println(red(err))
		for k, f := range r.MultipartForm.File {
			if k == id {
				log.Println(green("fuck !!! ????", id))
			}
			log.Println(k, f, id)
		}
		return
	} else {
		log.Println("form include:", id)
	}
	defer file.Close()
	log.Println(green(fmt.Sprintf("Uploaded File: %s", handler.Filename), green(fmt.Sprintf("File Size: %d", handler.Size))), yellow(fmt.Sprintf("MIME Header: %+v", handler.Header)))
	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile := filepath.Join(tconfig.LogPath(), id+".log")
	if _, err := os.Stat(tempFile); err != nil {
		if err != nil {
			fmt.Println(err)
		}
		// defer tempFile.Close()
		fp, err := os.Create(tempFile)
		if err != nil {
			jsonWrite(w, TData{
				"state": "fail",
				"log":   err.Error(),
			})
			return
		}
		n, err := io.Copy(fp, file)
		if n != handler.Size && err != nil {
			os.Remove(tempFile)
			jsonWrite(w, TData{
				"state": "fail",
				"log":   err.Error(),
			})
			return
		}
		// write this byte array to our temporary file
		// tempFile.Write(fileBytes)
		// return that we have successfully uploaded our file!
		// fmt.Fprintf(w, "Successfully Uploaded File\n")
		jsonWrite(w, TData{
			"state": "ok",
			"log":   "Successfully Uploaded File\n",
		})
	} else {
		jsonWrite(w, TData{
			"state": "ok",
			"log":   "File exists:" + tempFile,
		})
	}
}

func Upload(id string, fileName string, target string, proxy string) (string, error) {
	sess := jupyter.NewSession()
	// var fi os.FileInfo
	// var err error
	// if fi, err = os.Stat(fileName); err != nil {
	// 	return "", err
	// }
	if res, err := sess.Upload(target, fileName, id, map[string]string{
		"id": id,
	}, false, proxy); err != nil {
		return "", err
	} else {
		ret, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return string(ret), nil
	}

}

func (self *TaskConfig) DealWithUploadFile(w http.ResponseWriter, h *http.Request) {
	if h.Method == "POST" {
		f, _, err := h.FormFile("uploadFile")
		if err != nil {
			jsonWriteErr(w, err)
			return
		}
		buffer := bufio.NewScanner(f)
		buffer.Split(bufio.ScanLines)
		runOk := 0
		for buffer.Scan() {
			line := buffer.Text()
			lineStr := strings.TrimSpace(line)
			// fmt.Println(lineStr, "|")
			if strings.HasPrefix(lineStr, "http") {
				fmt.Println(utils.Green("[http]", lineStr))
			} else if strings.HasPrefix(lineStr, "tcp://") {
				fmt.Println(utils.Blue("[tcp]", lineStr))
			} else if strings.HasPrefix(lineStr, "run,") {
				fmt.Println(utils.Yellow("[cmd]", lineStr))
			} else {
				fmt.Println("[ignore]", lineStr)
				runOk -= 1
			}
		}
		jsonWrite(w, TData{
			"log":   fmt.Sprintf("%d", runOk),
			"state": "ok",
		})
	}
}
