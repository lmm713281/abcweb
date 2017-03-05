package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nullbio/abcweb/config"
	"github.com/spf13/afero"
)

func init() {
	appFS = afero.NewMemMapFs()
}

func TestGetAppPath(t *testing.T) {
	t.Parallel()

	gopath := os.Getenv("GOPATH")
	os.Setenv("GOPATH", "testpath/test")

	appPath, importPath, appName, err := getAppPath([]string{"."})
	if err == nil {
		t.Errorf("expected error, but got none: %s - %s", appPath, appName)
	}

	appPath, importPath, appName, err = getAppPath([]string{"/"})
	if err == nil {
		t.Errorf("expected error, but got none: %s - %s", appPath, appName)
	}

	appPath, importPath, appName, err = getAppPath([]string{"/test"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "testpath/test/src/test" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "test" {
		t.Errorf("mismatch, got %s", appName)
	}
	if importPath != "/test" {
		t.Errorf("mismatch, got %s", importPath)
	}

	appPath, importPath, appName, err = getAppPath([]string{"./stuff/test"})
	if err != nil {
		t.Error(err)
	}
	if appPath != "testpath/test/src/stuff/test" {
		t.Errorf("mismatch, got %s", appPath)
	}
	if appName != "test" {
		t.Errorf("mismatch, got %s", appName)
	}
	if importPath != "stuff/test" {
		t.Errorf("mismatch, got %s", importPath)
	}

	os.Setenv("GOPATH", gopath)
}

func TestGetProcessedPaths(t *testing.T) {
	t.Parallel()

	cfg := newConfig{
		AppPath: "/test/myapp",
		AppName: "myapp",
	}

	inPath := "/lol/" + templatesDirectory + "/file.tmpl"
	cleanPath, fullPath := getProcessedPaths(inPath, "/", cfg)
	if cleanPath != "myapp/file" {
		t.Error("mismatch:", cleanPath)
	}
	if fullPath != "/test/myapp/file" {
		t.Error("mismatch:", fullPath)
	}

	cfg.AppPath = "myapp"
	cfg.AppName = "myapp"

	cleanPath, fullPath = getProcessedPaths(inPath, "/", cfg)
	if cleanPath != "myapp/file" {
		t.Error("mismatch:", cleanPath)
	}
	if fullPath != "myapp/file" {
		t.Error("mismatch:", fullPath)
	}
}

func TestProcessSkips(t *testing.T) {
	cfg := newConfig{
		NoReadme:      true,
		NoGitIgnore:   true,
		NoConfig:      true,
		NoFontAwesome: true,
		Bootstrap:     "none",
		NoBootstrapJS: true,
		NoSessions:    true,
	}

	// check skip basedir
	err := appFS.MkdirAll("/templates", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err := appFS.Stat("/templates")
	if err != nil {
		t.Fatal(err)
	}
	skip, _ := processSkips(cfg, "/templates", "/templates", info)
	if skip == false {
		t.Error("expected to skip base path")
	}

	// check skip skipDirs slice
	err = appFS.MkdirAll("/templates/i18n", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/templates/i18n")
	if err != nil {
		t.Fatal(err)
	}
	skip, err = processSkips(cfg, "/templates", "/templates/i18n", info)
	if skip != true || err == nil {
		t.Error("expected to skip skipDir and receive skipdir err")
	}

	// check skip readme
	f, err := appFS.Create("/templates/README.md")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/README.md", info)
	if skip != true {
		t.Error("expected to skip skip readme")
	}

	// check skip app/sessions.go.tmpl
	f, err = appFS.Create("/templates/app/sessions.go.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/app/sessions.go.tmpl", info)
	if skip != true {
		t.Error("expected to skip skip sessions.go.tmpl")
	}

	// check skip gitignore
	f, err = appFS.Create("/templates/.gitignore")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/.gitignore", info)
	if skip != true {
		t.Error("expected to skip skip readme")
	}

	// check skip cnf.toml
	f, err = appFS.Create("/templates/cnf.toml")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/cnf.toml", info)
	if skip != true {
		t.Error("expected to skip skip cnf.toml")
	}

	// check skip fontawesome files
	f, err = appFS.Create("/templates/font-awesome.css")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/font-awesome.css", info)
	if skip != true {
		t.Error("expected to skip skip font-awesome.css")
	}

	// check skip fontawesome files
	f, err = appFS.Create("/templates/bootstrap.css")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/bootstrap.css", info)
	if skip != true {
		t.Error("expected to skip skip bootstrap.css")
	}

	cfg.Bootstrap = "flex"
	// check skip fontawesome files
	f, err = appFS.Create("/templates/bootstrap.js")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/bootstrap.js", info)
	if skip != true {
		t.Error("expected to skip skip bootstrap.js")
	}

	// check no-skip regular go file
	f, err = appFS.Create("/templates/file.go")
	if err != nil {
		t.Fatal(err)
	}
	info, err = f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	skip, _ = processSkips(cfg, "/templates", "/templates/file.go", info)
	if skip == true {
		t.Error("did not expect skip")
	}

	appFS = afero.NewMemMapFs()
}

func TestNewCmdWalk(t *testing.T) {
	cfg := newConfig{
		AppPath: "/my/app",
		AppName: "app",
		Silent:  true,
	}

	// test skip
	err := appFS.MkdirAll("/templates/i18n", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err := appFS.Stat("/templates/i18n")
	if err != nil {
		t.Fatal(err)
	}
	err = newCmdWalk(cfg, "/templates", "/templates/i18n", info, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if err != filepath.SkipDir {
		t.Fatalf("expected error type filepath.SkipDir, but got %#v", err)
	}

	// check go file write
	err = afero.WriteFile(appFS, "/templates/file.go", []byte("hello"), 0664)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/templates/file.go")
	if err != nil {
		t.Fatal(err)
	}
	err = newCmdWalk(cfg, "/templates", "/templates/file.go", info, nil)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/my/app/file.go")
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() || info.Size() != 5 {
		t.Fatalf("Expected isdir false and size to be 5, got %t and %d", info.IsDir(), info.Size())
	}

	// check template file write
	err = afero.WriteFile(appFS, "/templates/template.go.tmpl", []byte(`package    {{.AppName}}`), 0664)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/templates/template.go.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	err = newCmdWalk(cfg, "/templates", "/templates/template.go.tmpl", info, nil)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/my/app/template.go")
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() || info.Size() != int64(len("package app\n")) {
		b, err := afero.ReadFile(appFS, "/my/app/template.go")
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Expected isdir false and size to be %d, got %t and %d, value: %q", len("package app\n"), info.IsDir(), info.Size(), string(b))
	}

	// check dir write
	err = appFS.MkdirAll("/templates/stuff", 0755)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/templates/stuff")
	if err != nil {
		t.Fatal(err)
	}
	err = newCmdWalk(cfg, "/templates", "/templates/stuff", info, nil)
	if err != nil {
		t.Fatal(err)
	}
	info, err = appFS.Stat("/my/app/stuff")
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("Expected isdir true, got %t", info.IsDir())
	}

	appFS = afero.NewMemMapFs()
}

func TestGenerateTLSCerts(t *testing.T) {
	cfg := newConfig{
		AppPath:       "/out/spiders",
		AppName:       "spiders",
		TLSCommonName: "dragons",
		// attempt to create tls certs twice
		// should fail second time if this is false
		TLSCertsOnly: false,
		Silent:       true,
	}

	err := generateTLSCerts(cfg)
	if err != nil {
		t.Fatal(err)
	}

	info, err := appFS.Stat("/out/spiders/cert.pem")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("expected non-0 size for cert file")
	}

	info, err = appFS.Stat("/out/spiders/private.key")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("expected non-0 size for private key file")
	}

	appFS = afero.NewMemMapFs()
}
