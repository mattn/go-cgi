package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %v /path/to/go-file\n", os.Args[0])
		os.Exit(1)
	}
	tmp := filepath.Join(os.TempDir(), "go-cgi")
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
	if !os.IsExist(err) {
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

	cmd := exec.Command("go", "build", "-o", fname, fname + ".go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, string(out))
	cmd = exec.Command(fname)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
