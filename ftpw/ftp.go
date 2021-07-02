package ftpw

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"github.com/sgybupt/goUtils/fastFixer"
	"log"
	"os"
	"path"
	"time"
)

var config Config

type Config struct {
	Addr               string
	Timeout            time.Duration
	Username, Password string
	FTPDir             string
}

var ftpConn *ftp.ServerConn

func init() {
}

func SetConfig(c Config) {
	config = c
}

func newFTPConn(addr string, timeout time.Duration) {
	var err error
	if ftpConn != nil {
		err = ftpConn.NoOp()
		if err == nil {
			return
		}
		_ = ftpConn.Quit()
		ftpConn = nil
	}

	// ftpConn == nil  need to re connect
	var c *ftp.ServerConn
	config.Addr = addr
	config.Timeout = timeout

	maxTime := 3
	for ; maxTime > 0; maxTime-- {
		c, err = ftp.Dial(addr, ftp.DialWithTimeout(timeout))
		if err != nil {
			_ = fmt.Errorf("connect to server %s error: %s", addr, err.Error())
		} else {
			ftpConn = c
			if err = ftpConn.Login(config.Username, config.Password); err != nil {
				log.Fatal("username and password is incorrect")
			}
			if err = ftpConn.ChangeDir(config.FTPDir); err != nil {
				p := path.Clean(config.FTPDir)
				log.Fatal("change dir failed", p)
			}
			break
		}
	}
	if ftpConn == nil {
		log.Fatal("can not connect to server")
	}
	return
}

func GetFTPConn() (c *ftp.ServerConn) {
	var err error
	if ftpConn != nil {
		err = ftpConn.NoOp()
	}

	bad := ftpConn == nil || err != nil
	fastFixer.Do(bad,
		func() {
		}, func() {
			newFTPConn(config.Addr, config.Timeout)
		})
	return ftpConn
}

func GetFileList(dir string) ([]*ftp.Entry, error) {
	return GetFTPConn().List(dir)
}

type EntryWithPath struct {
	Path string
	ftp.Entry
}

func GetAllFilesInDirWalk(dir string) ([]EntryWithPath, error) {
	res := make([]EntryWithPath, 0)
	w := GetFTPConn().Walk(dir)
	for w.Next() {
		res = append(res, EntryWithPath{
			Path:  w.Path(),
			Entry: *w.Stat(),
		})
	}
	return res, w.Err()
}

func Quit() {
	_ = GetFTPConn().Quit()
}

func ChangeDir(dir string) error {
	return GetFTPConn().ChangeDir(dir)
}

func makeDir(dir string) error {
	return GetFTPConn().MakeDir(dir)
}

func checkDirExist(dp string) (bool, error) {
	now, err := GetFTPConn().CurrentDir()
	if err != nil {
		return false, err
	}
	defer func() {
		_ = GetFTPConn().ChangeDir(now)
	}()
	err = GetFTPConn().ChangeDir(dp)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// mkdir -p
func MakeDirsP(dir string) error {
	dirs := make([]string, 0, 2)
	var dfs func(string)
	dfs = func(d string) {
		d = path.Clean(d)
		_d, _p := path.Split(d)
		dirs = append(dirs, _p)
		if _d != "" && _d != string(os.PathSeparator) {
			dfs(_d)
		}
	}
	dfs(dir)
	prePath := ""
	for i := len(dirs) - 1; i >= 0; i-- {
		nowPath := path.Join(prePath, dirs[i])
		has, err := checkDirExist(nowPath)
		if err != nil {
			return err
		}
		if !has {
			if err := makeDir(nowPath); err != nil {
				return err
			}
		}
		prePath = nowPath
	}
	return nil
}

func UploadFile(localFp string, aimPath string) error {
	file, err := os.Open(localFp)
	if err != nil {
		return err
	}
	return GetFTPConn().Stor(aimPath, file)
}
