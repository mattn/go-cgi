package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func error500(message string) {
	fmt.Print("Status: 500\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	fmt.Println(message)
	os.Exit(1)
}

func tryTmp(tmp string) (string, error) {
	fi, err := os.Lstat(tmp)
	if err != nil {
		if err = os.MkdirAll(tmp, 0755); err != nil {
			return "", err
		}
	} else {
		if fi.Mode().Perm() != 0755 {
			return "", fmt.Errorf("Shouldn't work")
		}
		err = os.Chmod(tmp, 0755)
		if err != nil {
			return "", nil
		}
	}
	return tmp, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v /path/to/go-file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  '%v' is working directory\n", os.TempDir())
		os.Exit(1)
	}

	tmp, err := tryTmp(filepath.Join(os.TempDir(), "go-cgi"))
	if err != nil {
		tmp = filepath.Join(filepath.Dir(os.Args[1]), ".go-cgi")
	}

	envfile := filepath.Join(tmp, "env")
	fi, err := os.Lstat(envfile)
	if err == nil && fi.Mode().IsRegular() {
		b, err := ioutil.ReadFile(envfile)
		if err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "#") {
					continue
				}
				tokens := strings.SplitN(line, "=", 2)
				if len(tokens) == 2 {
					os.Setenv(tokens[0], tokens[1])
				}
			}
		}
	}

	path_hex := fmt.Sprintf("%x", md5.Sum([]byte(os.Args[1])))

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ha := md5.New()
	io.Copy(ha, f)
	f.Seek(0, os.SEEK_SET)
	defer f.Close()

	code_hex := fmt.Sprintf("%x", ha.Sum(nil))

	dname := filepath.Join(tmp, path_hex)
	fname := filepath.Join(tmp, path_hex, code_hex)

	_, err = os.Lstat(dname)
	if err != nil {
		if err = os.Mkdir(dname, 0755); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(2)
		}
	}

	_, err = os.Lstat(fname + ".go")
	if err != nil {
		files, err := filepath.Glob(filepath.Join(dname, "*"))
		if err != nil {
			error500(err.Error())
		}
		for _, file := range files {
			err = os.Remove(file)
			if err != nil {
				error500(err.Error())
			}
		}
		outf, err := os.Create(fname + ".go")
		if err != nil {
			error500(err.Error())
		}
		defer outf.Close()
		buff := bufio.NewReader(f)
		_, err = buff.ReadString('\n')
		if err != nil {
		}
		io.Copy(outf, buff)
	}

	exename := fname
	if runtime.GOOS == "windows" {
		exename += ".exe"
	}
	cmd := command("go", "build", "-o", exename, fname+".go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		error500(err.Error() + "\n" + string(out))
	}

	cmd = command(fname, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	timeout := time.AfterFunc(30 * time.Second, func() {
		cmd.Process.Kill()
		error500("Process was forcely killed")
	})
	err = cmd.Run()
	timeout.Stop()
	if err != nil {
		error500(err.Error())
	}
}
