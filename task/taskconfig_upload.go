package task

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
)

func (tconfig *TaskConfig) uploadFile(w http.ResponseWriter, r *http.Request) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

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
		log.Println("Error Retrieving the File")
		log.Println(red(err))
		for k, f := range r.MultipartForm.File {
			if k == id {
				log.Println(green("fuck !!! ????", id))
			}
			log.Println(k, f, id)
		}
		return
	}
	defer file.Close()
	log.Println(green(fmt.Sprintf("Uploaded File: %+v\n", handler.Filename), green(fmt.Sprintf("File Size: %+v\n", handler.Size))), green(fmt.Sprintf("MIME Header: %+v\n", handler.Header)))
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

func Upload(id string, fileName string, target string) (string, error) {

	fp, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	var fi os.FileInfo
	if fi, err = fp.Stat(); err != nil {
		log.Fatal(err)
	}
	bar := pb.New64(fi.Size()).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.Start()

	defer fp.Close()
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)

	go func() {
		var part io.Writer
		defer w.Close()
		defer fp.Close()

		w1, _ := mpw.CreateFormField("id")
		w1.Write([]byte(id))

		if part, err = mpw.CreateFormFile(id, fi.Name()); err != nil {
			log.Fatal(err)
		}
		part = io.MultiWriter(part, bar)
		if _, err = io.Copy(part, fp); err != nil {
			log.Fatal(err)
		}
		if err = mpw.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	resp, err := http.Post(target, mpw.FormDataContentType(), r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}
