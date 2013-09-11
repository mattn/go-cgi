package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func main() {
	tmp := filepath.Join(os.TempDir(), "go-cgi")
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v /path/to/go-file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %q is working directory\n", tmp)
		os.Exit(1)
	}

	_, err := os.Lstat(tmp)
	if err != nil {
		if err = os.Mkdir(tmp, 0755); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(2)
		}
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
	}

	exename := fname
	if runtime.GOOS == "windows" {
		exename += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", exename, fname + ".go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Print("Status: 500\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
		fmt.Print(string(out))
		os.Exit(1)
	}
	cmd = exec.Command(fname)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stderr
	cmd.Args = os.Args[1:]
	err = cmd.Run()
	if err != nil {
		fmt.Print("Status: 500\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
		fmt.Print(err.Error())
		os.Exit(1)
	}
}
