package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const token = "???"
const signature = "C0AC25D3DB05BDAF758A4E0A002F25F63F2FC93A"

func main() {
  err := os.Setenv("PATH", "/usr/local/bin:/usr/local/go/bin:" + os.Getenv("PATH"))
  if err != nil {
		log.Fatal(err)
	}

	err = os.Setenv("DRIPCAP_DARWIN_SIGN", signature)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Setenv("GOROOT", "/usr/local/go")
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := ioutil.TempDir("", "drip")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Setenv("GOPATH", tmp + "/gosrc")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Setenv("GOBIN", tmp + "/gosrc/bin")
	if err != nil {
		log.Fatal(err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	release, _, err := client.Repositories.GetLatestRelease("dripcap", "dripcap")
	if err != nil {
		log.Fatal(err)
	}

	for _, as := range release.Assets {
		if strings.Contains(*as.Name, "darwin") {
			log.Println("aleardy exists")
			os.Exit(0)
		}
	}

	c := exec.Command(os.Getenv("SHELL"), "-c", fmt.Sprintf(`
    cd %s
    git clone --depth=1 -b %s %s
    cd dripcap
    npm install
    gulp darwin-sign
  `, tmp, *release.TagName, "https://github.com/dripcap/dripcap.git"))

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(path.Join(tmp, "dripcap", "dripcap-darwin.zip"))
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = client.Repositories.UploadReleaseAsset("dripcap", "dripcap", *release.ID, &github.UploadOptions{Name: "dripcap-darwin.zip"}, f)
	if err != nil {
		log.Fatal(err)
	}
}
