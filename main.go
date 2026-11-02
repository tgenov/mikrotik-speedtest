// Command speedtest-launcher downloads the Ookla Speedtest CLI for the host
// architecture on first run, then replaces itself with it. Ookla's licence
// forbids redistributing the CLI, so the image cannot ship it.
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

const dir = "/opt/speedtest"

var ooklaArch = map[string]string{
	"amd64": "x86_64",
	"arm64": "aarch64",
	"arm":   "armel",
	"386":   "i386",
}

func main() {
	bin := filepath.Join(dir, "speedtest")
	if _, err := os.Stat(bin); err != nil {
		if err := install(bin); err != nil {
			fmt.Fprintln(os.Stderr, "install speedtest:", err)
			os.Exit(1)
		}
	}
	args := append([]string{bin, "--accept-license", "--accept-gdpr"}, os.Args[1:]...)
	if err := syscall.Exec(bin, args, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "exec speedtest:", err)
		os.Exit(1)
	}
}

func install(bin string) error {
	arch, ok := ooklaArch[runtime.GOARCH]
	if !ok {
		return fmt.Errorf("unsupported architecture %s", runtime.GOARCH)
	}
	version := os.Getenv("SPEEDTEST_VERSION")
	if version == "" {
		version = "1.2.0"
	}
	url := fmt.Sprintf("https://install.speedtest.net/app/cli/ookla-speedtest-%s-linux-%s.tgz", version, arch)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("no speedtest binary in %s", url)
		}
		if err != nil {
			return err
		}
		if hdr.Name != "speedtest" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(bin+".tmp", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		return os.Rename(bin+".tmp", bin)
	}
}
