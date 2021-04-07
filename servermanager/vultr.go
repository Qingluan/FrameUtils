package servermanager

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Qingluan/FrameUtils/utils"
	jupyter "github.com/Qingluan/jupyter/http"
	"github.com/Qingluan/merkur"
	"github.com/machinebox/progress"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type Vps struct {
	IP     string
	USER   string
	PWD    string
	TAG    string
	Region string
	Proxy  string
}

type Vultr struct {
	API     string //
	Servers map[string]Vps
}

func (vps Vps) String() string {
	return fmt.Sprintf("%s(Loc:%s Tag:%s)", vps.IP, vps.Region, vps.TAG)
}
func (vps Vps) Connect() (client *ssh.Client, sess *ssh.Session, err error) {
	sshConfig := &ssh.ClientConfig{
		User: vps.USER,
		Auth: []ssh.AuthMethod{ssh.Password(vps.PWD)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	ip := vps.IP
	if !strings.Contains(ip, ":") {
		ip += ":22"
	}
	if vps.Proxy != "" {
		if dialer := merkur.NewProxyDialer(vps.Proxy); dialer != nil {
			if conn, err := dialer.Dial("tcp", ip); err == nil {
				conn, chans, reqs, err := ssh.NewClientConn(conn, ip, sshConfig)
				if err != nil {
					return nil, nil, err
				}
				log.Println(utils.Green("Use Proxy:", vps.Proxy))
				client = ssh.NewClient(conn, chans, reqs)
			} else {
				return nil, nil, err
			}
		} else {
			return nil, nil, fmt.Errorf("%v", "no proxy dialer create ok!")
		}
	} else {
		client, err = ssh.Dial("tcp", ip, sshConfig)
	}
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

func (vps Vps) Rm(file string) bool {
	if _, sess, err := vps.Connect(); err != nil {
		return false
	} else {
		if err := sess.Run("rm " + filepath.Join("/tmp", file)); err != nil {
			return false
		} else {
			return true
		}
	}
}

func (vps Vps) Shell() bool {
	if conn, session, err := vps.Connect(); err != nil {
		log.Fatal(err)
		return false
	} else {
		if runtime.GOOS != "windows" {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
			ctx, cancel := context.WithCancel(context.Background())

			// session, err := conn.NewSession()
			// if err != nil {
			// 	return fmt.Errorf("cannot open new session: %v", err)
			// }

			run := func(ctx context.Context, conn *ssh.Client, session *ssh.Session) error {
				defer session.Close()

				go func() {
					<-ctx.Done()
					conn.Close()
				}()

				// fd := int(os.Stdin.Fd())
				fd := int(os.Stdout.Fd())

				state, err := terminal.MakeRaw(fd)
				if err != nil {
					return fmt.Errorf("terminal make raw: %s", err)
				}
				defer terminal.Restore(fd, state)

				w, h, err := terminal.GetSize(fd)
				if err != nil {
					return fmt.Errorf("terminal get size: %s", err)
				}

				modes := ssh.TerminalModes{
					ssh.ECHO:          1,
					ssh.TTY_OP_ISPEED: 14400,
					ssh.TTY_OP_OSPEED: 14400,
				}

				term := os.Getenv("TERM")
				if term == "" {
					term = "xterm-256color"
				}
				if err := session.RequestPty(term, h, w, modes); err != nil {
					return fmt.Errorf("session xterm: %s", err)
				}

				session.Stdout = os.Stdout
				session.Stderr = os.Stderr
				session.Stdin = os.Stdin

				if err := session.Shell(); err != nil {
					return fmt.Errorf("session shell: %s", err)
				}

				if err := session.Wait(); err != nil {
					if e, ok := err.(*ssh.ExitError); ok {
						switch e.ExitStatus() {
						case 130:
							return nil
						}
					}
					return fmt.Errorf("ssh: %s", err)
				}
				return nil
			}

			go func() {
				if err := run(ctx, conn, session); err != nil {
					log.Print(err)
				}
				cancel()
			}()

			select {
			case <-sig:
				cancel()
			case <-ctx.Done():
			}

		} else {
			// session.Stdout = os.Stdout
			defer session.Close()

			// StdinPipe for commands
			stdin, err := session.StdinPipe()
			if err != nil {
				log.Fatal(err)
			}

			// Uncomment to store output in variable
			//var b bytes.Buffer
			//session.Stdout = &b
			//session.Stderr = &b

			// Enable system stdout
			// Comment these if you uncomment to store in variable
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr

			// Start remote shell
			err = session.Shell()
			if err != nil {
				log.Fatal(err)
			}

			// send the commands
			buffer := bufio.NewReader(os.Stdin)
			for {
				// nowCwd := fmt
				time.Sleep(1 * time.Second)
				fmt.Printf("%s >", utils.Green(vps))
				line, _, _ := buffer.ReadLine()

				_, err = fmt.Fprintf(stdin, "%s\n", line)
				if err != nil {
					log.Fatal(err)
				}
			}

			// Wait for session to finish
			// err = session.Wait()
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// Uncomment to store in variable
			//fmt.Println(b.String())

		}
		return true
	}
}

func (vps Vps) Upload(file string, canexcute bool) error {
	if cli, _, err := vps.Connect(); err != nil {
		log.Fatal(err)
		return err
	} else {
		if sftpChannel, err := sftp.NewClient(cli); err != nil {

			log.Println(err)
			return err

		} else {
			fileName := filepath.Base(file)
			sftpChannel.Remove("/tmp/" + fileName)
			fp, err := sftpChannel.OpenFile("/tmp/"+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR)

			if err != nil {

				log.Println(err)
				return err
			}
			localState, err := os.Stat(file)
			if err != nil {

				log.Println(err)
				return err
			}

			startAt := int64(0)
			defer fp.Close()
			if state, err := fp.Stat(); err == nil {
				startAt = state.Size()
				if startAt == localState.Size() {
					log.Println("Already upload !")
					return nil
				}
				if startAt != 0 {
					log.Println("Continued at:", float64(startAt)/float64(1024)/float64(1024), "MB")
				}
			}
			localFp, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
			if err != nil {
				log.Println(err)
				return err
			}
			defer localFp.Close()
			_, err = localFp.Seek(startAt, os.SEEK_SET)
			if err != nil {
				log.Println(err)
				return err
			}
			// ctx := context.Background()

			// get a reader and the total expected number of bytes
			// s := `Now that's what I call progress`
			size := localState.Size()
			r := progress.NewReader(localFp)
			// Start a goroutine printing progress
			go func() {
				ctx := context.Background()
				progressChan := progress.NewTicker(ctx, r, size, 5*time.Second)
				for p := range progressChan {
					fmt.Printf("\r[%.3f%%] %.3f MB %s %v remaining...", p.Percent(), float64(p.Size())/float64(1024)/float64(1024), fileName, p.Remaining().Round(time.Second))
				}
				fmt.Println("\rdownload is completed")
			}()

			io.Copy(fp, r)
			if canexcute {
				fp.Chmod(os.ModeExclusive)
			}
		}

	}
	return nil
}

func (vps Vps) Deploy(file []string, run string) string {
	var wait sync.WaitGroup
	// for _, f := range file {
	for i, f := range file {
		wait.Add(1)
		go func(w *sync.WaitGroup, fp string) {
			if err := vps.Upload(fp, true); err != nil {
				log.Println(err)
			}
			w.Done()

		}(&wait, f)

		if i%3 == 0 && i != 0 {
			wait.Wait()
			wait = sync.WaitGroup{}
		}
	}
	time.Sleep(1 * time.Second)
	wait.Wait()
	if _, sess, err := vps.Connect(); err != nil {
		return err.Error()
	} else {
		if out, err := sess.Output(run); err != nil {
			return err.Error()
		} else {
			return string(out)
		}
	}
}

func NewVultr(api string) (v *Vultr) {
	v = new(Vultr)
	v.Servers = make(map[string]Vps)
	v.API = api
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
	sess.SetHeader("API-Key", strings.TrimSpace(v.API))
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
