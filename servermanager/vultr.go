package servermanager

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/machinebox/progress"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Vps struct {
	IP     string
	USER   string
	PWD    string
	TAG    string
	Region string
}

type Vultr struct {
	API     string //
	Servers map[string]Vps
}

func (vps Vps) String() string {
	return fmt.Sprintf("%s(%s)", vps.IP, vps.Region)
}
func (vps Vps) Connect() (client *ssh.Client, sess *ssh.Session, err error) {
	sshConfig := &ssh.ClientConfig{
		User: vps.USER,
		Auth: []ssh.AuthMethod{ssh.Password(vps.PWD)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	client, err = ssh.Dial("tcp", vps.IP, sshConfig)
	if err != nil {
		return nil, nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, session, nil
}

func (vps Vps) Upload(file string, canexcute bool) bool {
	if cli, _, err := vps.Connect(); err != nil {
		return false
	} else {
		if sftpChannel, err := sftp.NewClient(cli); err != nil {

			log.Println(err)
			return false

		} else {
			fileName := filepath.Base(file)
			fp, err := sftpChannel.OpenFile(filepath.Join("/tmp", fileName), os.O_APPEND|os.O_CREATE|os.O_RDWR)

			if err != nil {

				log.Println(err)
				return false
			}
			localState, err := os.Stat(file)
			if err != nil {

				log.Println(err)
				return false
			}

			startAt := int64(0)
			defer fp.Close()
			if state, err := fp.Stat(); err == nil {
				startAt = state.Size()
				if startAt == localState.Size() {
					log.Println("Already upload !")
					return true
				}
			}
			localFp, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
			if err != nil {
				log.Println(err)
				return false
			}
			defer localFp.Close()
			_, err = localFp.Seek(startAt, os.SEEK_SET)
			if err != nil {
				log.Println(err)
				return false
			}
			// ctx := context.Background()

			// get a reader and the total expected number of bytes
			// s := `Now that's what I call progress`
			size := localState.Size()
			r := progress.NewReader(localFp)
			// Start a goroutine printing progress
			go func() {
				ctx := context.Background()
				progressChan := progress.NewTicker(ctx, r, size, 10*time.Second)
				for p := range progressChan {
					fmt.Printf("\r%s %v remaining...", fileName, p.Remaining().Round(time.Second))
				}
				fmt.Println("\rdownload is completed")
			}()

			io.Copy(fp, r)
			if canexcute {
				fp.Chmod(os.ModeExclusive)
			}
		}

	}
	return false
}

func NewVultr(api string) (v *Vultr) {
	v = new(Vultr)
	v.Servers = make(map[string]Vps)
	return
}

func (v *Vultr) GetServers() (vs []Vps) {
	for _, e := range v.Servers {
		vs = append(vs, e)
	}
	return

}

func (v *Vultr) Update() (err error) {
	if v.API == "" {
		return fmt.Errorf("%v", "Need api key!!: ")
	}
	sess := jupyter.NewSession()
	sess.SetHeader("API-Key", v.API)
	if res, err := sess.Get("https://api.vultr.com/v1/server/list"); err != nil {
		return err
	} else {
		data := res.Json()
		servers := make(map[string]Vps)
		for _, val := range data {
			vps := Vps{}
			server := val.(map[string]interface{})
			if vals, ok := server["main_ip"]; ok {
				vps.IP = vals.(string)
			}
			if vals, ok := server["default_password"]; ok {
				vps.PWD = vals.(string)
			}
			if vals, ok := server["tag"]; ok {
				vps.TAG = vals.(string)
			}
			if vals, ok := server["location"]; ok {
				vps.Region = vals.(string)
			}

			vps.USER = "root"
			servers[vps.IP] = vps
		}
		if len(servers) > 0 {
			v.Servers = nil
			v.Servers = servers
		}
	}
	return
}
