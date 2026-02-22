package ds_service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"chain-love/domain/ds"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/service/models"
)

// SaveBookFile saves book metadata and pages as JSON files under configured folder.
// Structure: <bookRoot>/<bookID>/v<version>/meta.json and <bookRoot>/<bookID>/v<version>/<idx>.json
// Also saves <bookRoot>/<bookID>/index.json with the latest version.
func SaveBookFile(book *ds.Book, pages []models.PageData) error {
	// use configured root, fallback to runtime/books if empty
	root := setting.Config.App.FilePath.Book
	e.PanicIf(root == "", "book文件路径未配置！")

	bookDir := filepath.Join(root, fmt.Sprintf("%d", book.Id))
	versionStr := fmt.Sprintf("v%d", book.Version)
	versionDir := filepath.Join(bookDir, versionStr)

	// create version folder
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return err
	}

	// save meta.json
	bookBytes, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(versionDir, "meta.json"), bookBytes, 0o644); err != nil {
		return err
	}

	// save pages as <idx>.json
	for _, p := range pages {
		pBytes, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}
		fn := filepath.Join(versionDir, fmt.Sprintf("%d.json", p.Idx))
		if err := os.WriteFile(fn, pBytes, 0o644); err != nil {
			return err
		}
	}

	// save index.json
	indexData := map[string]string{"version": versionStr}
	indexBytes, _ := json.MarshalIndent(indexData, "", "  ")
	if err := os.WriteFile(filepath.Join(bookDir, "index.json"), indexBytes, 0o644); err != nil {
		return err
	}

	// cleanup old versions (keep latest 3)
	entries, err := os.ReadDir(bookDir)
	if err == nil {
		var versions []int
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "v") {
				if v, err := strconv.Atoi(entry.Name()[1:]); err == nil {
					versions = append(versions, v)
				}
			}
		}
		// sort descending
		sort.Slice(versions, func(i, j int) bool {
			return versions[i] > versions[j]
		})
		// keep 3
		if len(versions) > 3 {
			for _, v := range versions[3:] {
				os.RemoveAll(filepath.Join(bookDir, fmt.Sprintf("v%d", v)))
			}
		}
	}

	return nil
}
