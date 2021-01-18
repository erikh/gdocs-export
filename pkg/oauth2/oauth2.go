package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/erikh/gdocs-export/pkg/cli"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func configDir() string {
	dir, err := homedir.Dir()
	if err != nil {
		cli.ErrExit("Cannot locate or access home directory: %v", err)
	}

	return filepath.Join(dir, ".gdexport")
}

func ensureConfigDir() string {
	dir := configDir()
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		return dir
	}

	if err := os.MkdirAll(dir, 0701); err != nil {
		cli.ErrExit("Could not make settings directory: %v", err)
	}

	return dir
}

// ImportCredentials imports the credentials.json the user supplied.
func ImportCredentials(in io.Reader) error {
	dir := ensureConfigDir()

	out, err := os.Create(filepath.Join(dir, "credentials.json"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// GetClient retrieves a token, saves the token, then returns the generated client.
func GetClient() *http.Client {
	dir := ensureConfigDir()

	b, err := ioutil.ReadFile(filepath.Join(dir, "credentials.json"))
	if err != nil {
		cli.ErrExit("Unable to read client secret file; did you import-credentials yet? (err: %v)", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/documents.readonly")
	if err != nil {
		cli.ErrExit("Unable to parse client secret file to config: %v", err)
	}

	tokFile := filepath.Join(dir, "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// GetTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		cli.ErrExit("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		cli.ErrExit("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// TokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// SaveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		cli.ErrExit("Unable to cache OAuth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}
