package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Qingluan/FrameUtils/tui"
	"github.com/Qingluan/FrameUtils/utils"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/term"
)

var (
	HOME, _  = os.UserHomeDir()
	PWDDIR   = filepath.Join(HOME, ".c.repo")
	PWD      = filepath.Join(HOME, ".c.git.conf")
	USERNAME = ""
	pwd__    = ""
)

func GetRepo() string {
	r, err := ioutil.ReadFile(PWD)
	giturl := ""

	for _, l := range strings.Split(string(r), "\n") {
		if strings.HasSuffix(l, ".git") || strings.HasPrefix(l, "http") {
			giturl = strings.TrimSpace(l)
		} else if strings.TrimSpace(l) != "" {
			USERNAME = strings.TrimSpace(l)
		}
	}

	if err != nil || giturl == "" {
		url := tui.GetPass("set git repo:")
		ioutil.WriteFile(PWD, []byte(url), os.ModePerm)
		if USERNAME == "" {
			USERNAME = tui.GetPass("git username")
		}
		if pwd__ == "" {
			pwd__ = GetPass("git password (" + USERNAME + ")>")
		}
		_, err := git.PlainClone(PWDDIR, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: USERNAME,
				Password: pwd__,
			},
			InsecureSkipTLS:   true,
			Progress:          os.Stderr,
			URL:               giturl,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			log.Fatal("Git clone:", err, " in ", PWDDIR)
		}
		return giturl
	} else {
		if USERNAME == "" {
			USERNAME = tui.GetPass("git username")
		}
		if _, err := os.Stat(PWDDIR); err != nil {
			if pwd__ == "" {
				pwd__ = GetPass("git password (" + USERNAME + ")>")
			}
			_, err := git.PlainClone(PWDDIR, false, &git.CloneOptions{
				Auth: &http.BasicAuth{
					Username: USERNAME,
					Password: pwd__,
				},
				InsecureSkipTLS:   true,
				Progress:          os.Stderr,
				URL:               giturl,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
			if err != nil {
				log.Fatal("Git clone:", err, PWDDIR)
			}
		}
		return string(bytes.TrimSpace(r))
	}
}
func Update() {
	r, err := git.PlainOpen(PWDDIR)
	if err != nil {
		log.Fatal("Git open :", err)
	}

	w, err := r.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("git doing ...")
	if USERNAME == "" {
		USERNAME = tui.GetPass("git username")
	}
	if pwd__ == "" {
		pwd__ = GetPass("git password (" + USERNAME + ")>")
	}

	err = w.Pull(&git.PullOptions{
		RemoteName:      "origin",
		InsecureSkipTLS: true,
		Progress:        os.Stderr,
		Auth: &http.BasicAuth{
			Username: USERNAME,
			Password: pwd__,
		}})
	if err != nil {
		// log.Fatal("pull err:", err)
	} else {

		log.Println("git done")
	}

}

func Upload(password ...string) {
	if USERNAME == "" {
		USERNAME = tui.GetPass("git username")
	}
	if pwd__ == "" {
		pwd__ = GetPass("git password (" + USERNAME + ")>")
	}
	log.Println("Check update...")
	Update()

	log.Println("---- start upload ---")
	r, err := git.PlainOpen(PWDDIR)
	// CheckIfError(err)
	if err != nil {
		log.Fatal("GIt open:", err)
	}
	w, _ := r.Worktree()
	// if w.Status()
	err = w.AddGlob("*.en")
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Commit("upload New ", &git.CommitOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// Info("git push")
	// push using default options

	// password := tui.GetPass("git password (" + USERNAME + ")>")

	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: USERNAME,
			Password: pwd__,
		},
		InsecureSkipTLS: true,
		Progress:        os.Stderr,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type PasswordNote struct {
	Note map[string]interface{} `json:"note"`
}

func GetPass(la string) string {
	fmt.Print(la + "*")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("get passwd err:", err)
	}
	return string(bytePassword)
}

func Load() (pn *PasswordNote) {
	GetRepo()
	passwd := ""
	pn = new(PasswordNote)
	pn.Note = make(map[string]interface{})
	filepath.Walk(PWDDIR, func(p string, f os.FileInfo, e error) error {
		if strings.HasSuffix(p, ".en") {
			if decry, err := ioutil.ReadFile(p); err != nil {
				log.Println(p, "broken !")
			} else {
				if passwd == "" {
					passwd = GetPass("cry :")
				}
				if en, err := AesDecrypt(decry, []byte(passwd)); err != nil {
					log.Println("Err pass or other:", err)
				} else {
					m := make(map[string]interface{})
					if err := json.Unmarshal(en, &m); err != nil {
						log.Println("Err buf !!", err)
					}
					for k, v := range m {
						pn.Note[k] = v
					}
				}
			}

		}
		return nil
	})
	return
}

func (pn *PasswordNote) Save() {
	// must init!

	log.Println("Init Repo ")
	giturl := GetRepo()

	buf, err := json.Marshal(pn.Note)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Generate new notes json...")
	fmt.Print("AES Crypt Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("get passwd err:", err)
	}

	log.Println("Generate crypted file to git ....")
	en, err := AesEncrypt(buf, bytePassword)
	if err != nil {
		log.Fatal("en err:", err)
	}

	ioutil.WriteFile(filepath.Join(PWDDIR, "note.en"), en, os.ModePerm)

	log.Println("pushing ", giturl, "....")
	Upload()
}

type Can string

func (c Can) String() string {
	return string(c)
}

func (c Can) Path() string {
	if strings.HasPrefix(string(c), "[+]") {
		return string(c)[3:]
	} else {
		return string(c)
	}

}

func CanEntry(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}:
		return true
	default:
		return false
	}
}

func (pn *PasswordNote) CHoose() string {
	keys := []tui.CanString{}
	canSave := false
	lastKeys := []string{}
	var value interface{}
	for k, v := range pn.Note {
		switch v.(type) {
		case map[string]interface{}:
			keys = append(keys, Can("[+]"+k))
		default:
			keys = append(keys, Can(k))
		}
	}
	keys = append(keys, Can(utils.Red("<new >>>")), Can(utils.Red("<Exit>")), Can(utils.Red("<save>")))
	var one tui.CanString
	var ok bool
	value = pn.Note
	for {
		if one, ok = tui.SelectOne("See notes:", keys); ok {
			if one.String() == utils.Red("<new >>>") {
				value = pn.Add(value.(map[string]interface{}))
				canSave = true
			} else if one.String() == utils.Red("<Exit>") {

				break
			} else if one.String() == utils.Red("<save>") {
				canSave = true
				break
			} else if one.String() == utils.Red("<last page>") {
				value = pn.Note
				// fmt.Println("ll:", lastKeys)
				lastKeys = lastKeys[:len(lastKeys)-1]
				for _, k := range lastKeys {
					// tui.GetPass(fmt.Sprint("test:", k))
					value = value.(map[string]interface{})[Can(k).Path()]
				}

			} else {
				value = value.(map[string]interface{})[one.(Can).Path()]
				// if CanEntry(value) {
				lastKeys = append(lastKeys, one.String())
				// }
				// tui.GetPass(fmt.Sprint("test:", lastKeys))
			}
		} else {
			continue
		}
		if !CanEntry(value) {
			// tui.GetPass(fmt.Sprint(value, lastKeys))
			break
		}

		// lastKeys = keys
		keys = []tui.CanString{}
		for k, v := range value.(map[string]interface{}) {
			switch v.(type) {
			case map[string]interface{}:
				keys = append(keys, Can("[+]"+k))
			default:
				keys = append(keys, Can(k))
			}
		}
		if len(lastKeys) == 0 {
			keys = append(keys, Can(utils.Red("<Exit>")), Can(utils.Red("<new >>>")), Can(utils.Red("<save>")))

		} else {

			keys = append(keys, Can(utils.Red("<last page>")), Can(utils.Red("<new >>>")), Can(utils.Red("<save>")))

		}

	}

	if canSave {
		if op, ok := tui.SelectOne("Save? ", []tui.CanString{Can("no"), Can("yes")}); ok && op.String() == "yes" {
			pn.Save()
		}

	}
	if value == nil {
		return ""
	}
	switch value.(type) {
	case map[string]interface{}:
		return ""
	}
	return value.(string)

}

func (pn *PasswordNote) Add(dict map[string]interface{}) map[string]interface{} {

	res := tui.GetPass("'K = V' / 'K' >>> ")
	if strings.Contains(res, "=") {
		fs := strings.SplitN(res, "=", 2)
		nkey := strings.TrimSpace(fs[0])
		nval := strings.TrimSpace(fs[1])
		if nval == "" {
			delete(dict, nkey)
		} else {

			dict[nkey] = nval
		}
		return dict
	} else {
		dict[strings.TrimSpace(res)] = map[string]interface{}{}
		return dict
	}

}
