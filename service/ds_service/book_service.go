package ds_service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"chain-love/domain/ds"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/service/models"
)

// SaveBookFile saves book metadata and pages as JSON files under configured folder.
// Structure: <bookRoot>/<bookID>/book.json and <bookRoot>/<bookID>/<idx>.json
// If the folder exists it will be removed first.
func SaveBookFile(book *ds.Book, pages []models.PageData) error {
	// use configured root, fallback to runtime/books if empty
	root := setting.Config.App.FilePath.Book
	e.PanicIf(root == "", "book文件路径未配置！")

	bookDir := filepath.Join(root, fmt.Sprintf("%d", book.Id))

	// remove existing folder if any
	if err := os.RemoveAll(bookDir); err != nil {
		return err
	}
	// create folder
	if err := os.MkdirAll(bookDir, 0o755); err != nil {
		return err
	}

	// save book.json
	bookBytes, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(bookDir, "meta.json"), bookBytes, 0o644); err != nil {
		return err
	}

	// save pages as <idx>.json
	for _, p := range pages {
		pBytes, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		fn := filepath.Join(bookDir, fmt.Sprintf("%d.json", p.Idx))
		if err := os.WriteFile(fn, pBytes, 0o644); err != nil {
			return err
		}
	}

	return nil
}
