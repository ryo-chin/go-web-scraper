package p

import (
	"golang.org/x/net/context"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckResale(t *testing.T) {
	err := SetupCredentials()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	err = CheckResale(ctx, PubSubMessage{})

	if err != nil {
		t.Fatal(err)
	}
}

func SetupCredentials() error {
	p, _ := os.Getwd()
	abs, _ := filepath.Abs(p + "/../../")
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", abs+"/secret.json")
	return err
}
