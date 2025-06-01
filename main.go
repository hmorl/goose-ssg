package main

import (
	"goose-ssg/internal"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/pflag"
)

func main() {
	log.SetFlags(log.Ltime)

	serve := pflag.Bool("serve", false, "Locally serve generated site")
	destination := pflag.String("destination", "dist", "Directory to generate the site in")
	pflag.Parse()

	args := pflag.Args()
	if len(args) < 1 {
		log.Fatalln("Usage: goose-ssg <source-directory> [--serve] [--destination=customDestinationDirectory]")
	}

	sourceDir := args[0]

	templatesPath := filepath.Join(sourceDir, "templates")
	pagesPath := filepath.Join(sourceDir, "pages")
	destinationPath := *destination
	staticContentPath := filepath.Join(sourceDir, "static")

	err := internal.RebuildSite(pagesPath, staticContentPath, templatesPath, destinationPath)
	if err != nil {
		log.Fatalln(err)
	}

	if *serve {
		server := internal.NewServer()

		quitChan := make(chan struct{})

		go func() {
			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt)
			<-signalChan
			close(quitChan)
		}()

		server.ServeAndWatch(destinationPath, sourceDir, func() {
			log.Println("--- Change detected!")

			err := internal.RebuildSite(pagesPath, staticContentPath, templatesPath, destinationPath)
			if err != nil {
				log.Fatalln(err)
			}
		}, quitChan)
	}
}
