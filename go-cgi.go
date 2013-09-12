package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

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
	fmt.Print("Status: 500\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	return tmp, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v /path/to/go-file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %q is working directory\n", os.TempDir())
		os.Exit(1)
	}

	tmp, err := tryTmp(filepath.Join(os.TempDir(), "go-cgi"))
	if err != nil {
		tmp = filepath.Join(filepath.Dir(os.Args[1]), ".go-cgi")
	}

	ha := md5.New()
	ha.Write([]byte(os.Args[1]))
	path_hex := fmt.Sprintf("%x", ha.Sum(nil))

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ha.Reset()
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
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for _, file := range files {
			err = os.Remove(file)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
		outf, err := os.Create(fname + ".go")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer outf.Close()
		buff := bufio.NewReader(f)
		_, err = buff.ReadString('\n')
		if err != nil {
		}
		io.Copy(outf, buff)

		exename := fname
		if runtime.GOOS == "windows" {
			exename += ".exe"
		}
		cmd := command("go", "build", "-o", exename, fname+".go")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Print("Status: 500\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
			fmt.Print(string(out))
			os.Exit(1)
		}
	}

	cmd := command(fname, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		fmt.Print("Status: 500\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
		fmt.Print(err.Error())
		os.Exit(1)
	}
}
