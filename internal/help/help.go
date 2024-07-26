package help

import (
	"embed"
	"fmt"
	"os"
	"runtime/debug"
	"text/template"
	"time"
)

type About struct {
	AppName   string
	Version   string
	Commit    string
	Bin       string
	UserAgent string
}

var (
	about = About{
		AppName: "Senta - tlgbot",
	}

	//go:embed templates
	templateFS embed.FS
	templates  = template.Must(template.ParseFS(templateFS, "templates/*.txt"))
)

func init() {
	var versionDate = "-"

	about.Bin = os.Args[0]
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				about.Commit = setting.Value[:8]
			case "vcs.time":
				if build, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					versionDate = build.Format("2006-01-02 15:04:05")
				}
			}
		}
	}

	about.Version = fmt.Sprintf("%s / %s", versionDate, about.Commit)
	about.UserAgent = about.AppName + "/" + about.Version
}

func PrintUsage() {
	_ = templates.ExecuteTemplate(os.Stdout, "usage", &about)
}

func PrintVersion() {
	_ = templates.ExecuteTemplate(os.Stdout, "version", &about)
}
