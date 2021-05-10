package task

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/fatih/color"
)

func (tconfig *TaskConfig) uploadFile(w http.ResponseWriter, r *http.Request) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	// yellow := color.New(color.FgYellow).SprintFunc()
	// log.Println(green("File Upload Endpoint Hit"))

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 1 GB files.
	r.ParseMultipartForm(1024 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file

	id := r.FormValue("id")
	// fmt.Println(id, r.Form,r.Fo)
	if v, ok := tconfig.depatch[id]; ok {
		tconfig.depatch[id] = v + "-Finished"
		// log.Println("Finish:", utils.Green(id), " from : ", utils.Yellow(v))
	}
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
		// log.Println("form include:", id)
	}
	defer file.Close()
	// log.Println(green(fmt.Sprintf("Uploaded File: %s", handler.Filename), green(fmt.Sprintf("File Size: %d", handler.Size))), yellow(fmt.Sprintf("MIME Header: %+v", handler.Header)))

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
		/*
			对于远程部署的任务返回，改变部署状态
		*/
		tconfig.DeployedSwitchState(id, "Finished")

		// write this byte array to our temporary file
		// tempFile.Write(fileBytes)
		// return that we have successfully uploaded our file!
		// fmt.Fprintf(w, "Successfully Uploaded File\n")
		jsonWrite(w, TData{
			"state": "ok",
			"log":   "Successfully Uploaded File\n",
		})
	} else {
		/*
			对于远程部署的任务返回，改变部署状态
		*/
		tconfig.DeployedSwitchState(id, "Finished")

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
			log.Println("Upload err:", err)
			return "", err
		}
		return string(ret), nil
	}

}
