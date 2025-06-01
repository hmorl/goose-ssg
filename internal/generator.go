package internal

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func copyDirContents(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstDir := filepath.Join(dst, filepath.Dir(relPath))

		const fileMode = 0755
		os.MkdirAll(dstDir, fileMode)

		dstPath := filepath.Join(dstDir, info.Name())
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}

		sourceInfo, err := os.Stat(src)
		if err != nil {
			return err
		}

		return os.Chmod(dst, sourceInfo.Mode())
	})
}

type File struct {
	Name     string
	Path     string
	Contents string
}

func readHtmlFiles(dir string) ([]File, error) {
	var files []File

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)

		if !info.IsDir() && (ext == ".html" || ext == ".gohtml") {
			contents, err := os.ReadFile(path)

			if err != nil {
				return err
			}

			files = append(files, File{info.Name(), path, string(contents)})
		}

		return nil
	})

	return files, err
}

type PageData struct {
	ThisPage string
}

func RebuildSite(pagesPath, staticPath, templatesPath, destinationPath string) error {
	log.Println("Generating site at `" + destinationPath + "`...")
	templates, tmplErr := readHtmlFiles(templatesPath)
	pages, pageErr := readHtmlFiles(pagesPath)

	if err := errors.Join(tmplErr, pageErr); err != nil {
		return err
	}

	combinedTemplate := ""
	for _, tmpl := range templates {
		combinedTemplate += tmpl.Contents
	}

	if os.RemoveAll(destinationPath) != nil {
		msg := fmt.Sprintf("Error removing destination directory: %s. Try removing it manually and try again. \n", destinationPath)
		return errors.New(msg)
	}

	mapFunc := func(root any, args ...any) (map[string]any, error) {
		if len(args)%2 != 0 {
			return nil, errors.New("odd number of args passed into key-value map")
		}

		m := make(map[string]any)
		m["Root"] = root

		for i := 0; i < len(args); i += 2 {
			key, ok := args[i].(string)

			if !ok {
				return nil, errors.New("key must be a string")
			}

			m[key] = args[i+1]
		}

		return m, nil
	}

	for _, page := range pages {
		dstPath := ""
		dir := ""

		if page.Name != "index.html" {
			relPath, _ := filepath.Rel(pagesPath, page.Path)
			dir = strings.TrimSuffix(relPath, filepath.Ext(relPath))
		}

		dstPath = filepath.Join(destinationPath, dir, "index.html")

		data := PageData{ThisPage: "/" + dir}

		tmpl := template.Must(template.New("").Funcs(template.FuncMap{"map": func(args ...any) (map[string]any, error) {
			return mapFunc(data, args...)
		}}).Parse(combinedTemplate))
		tmpl.Parse(page.Contents)

		const fileMode = 0755
		os.MkdirAll(filepath.Dir(dstPath), fileMode)

		file, err := os.Create(dstPath)
		if err != nil {
			return err
		}

		err = tmpl.ExecuteTemplate(file, "base", data)
		if err != nil {
			return err
		}
	}

	_, err := os.Stat(staticPath)
	if os.IsNotExist(err) {
		return nil
	}

	log.Println("Copying static content...")
	err = copyDirContents(staticPath, destinationPath)
	if err != nil {
		return err
	}

	log.Println("Site generated!")
	return nil
}
