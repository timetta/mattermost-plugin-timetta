package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	source := flag.String("source", "", "directory containing the unpacked plugin")
	output := flag.String("output", "", "target .tar.gz file")
	root := flag.String("root", "", "top-level directory name in the bundle")
	flag.Parse()
	if *source == "" || *output == "" || *root == "" {
		fatalf("source, output, and root are required")
	}
	if strings.ContainsAny(*root, `/\\`) || *root == "." || *root == ".." {
		fatalf("root must be a single safe directory name")
	}
	if err := createBundle(*source, *output, *root); err != nil {
		fatalf("create plugin bundle: %v", err)
	}
}

func createBundle(source, output, root string) (returnErr error) {
	sourcePath, err := filepath.Abs(source)
	if err != nil {
		return err
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	archive, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() {
		if err := archive.Close(); returnErr == nil && err != nil {
			returnErr = err
		}
	}()

	gzipWriter := gzip.NewWriter(archive)
	defer func() {
		if err := gzipWriter.Close(); returnErr == nil && err != nil {
			returnErr = err
		}
	}()
	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		if err := tarWriter.Close(); returnErr == nil && err != nil {
			returnErr = err
		}
	}()

	return filepath.WalkDir(sourcePath, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(sourcePath, filePath)
		if err != nil {
			return err
		}
		archiveName := root
		if relative != "." {
			archiveName = path.Join(root, filepath.ToSlash(relative))
		}
		if entry.IsDir() {
			archiveName += "/"
		}
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return err
		}
		header.Name = archiveName
		header.Uid = 0
		header.Gid = 0
		header.Uname = ""
		header.Gname = ""
		if entry.IsDir() || isPluginExecutable(relative) {
			header.Mode = 0o755
		} else {
			header.Mode = 0o644
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tarWriter, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func isPluginExecutable(relative string) bool {
	normalized := filepath.ToSlash(relative)
	return strings.HasPrefix(normalized, "server/dist/")
}

func fatalf(format string, arguments ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", arguments...)
	os.Exit(1)
}
