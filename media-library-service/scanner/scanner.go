package scanner

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type MediaItem struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Category     string  `json:"category"`
	Path         string  `json:"path"`
	HLSUrl       *string `json:"hls_url,omitempty"`
	Size         int64   `json:"size"`
	ModifiedTime float64 `json:"modified_time"`
}

type Scanner struct {
	mediaDir string
}

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".webm": true,
	".flv":  true,
	".wmv":  true,
}

var skipKeywords = []string{"trailer", "teaser", "sample", "featurette"}

func New(mediaDir string) *Scanner {
	return &Scanner{mediaDir: mediaDir}
}

func (s *Scanner) ScanCategory(category, dirName string) ([]MediaItem, error) {
	items := []MediaItem{}
	categoryPath := filepath.Join(s.mediaDir, dirName)

	// Check if directory exists (case-insensitive)
	if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
		// Try to find case-insensitive match
		entries, err := os.ReadDir(s.mediaDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && strings.EqualFold(entry.Name(), dirName) {
					categoryPath = filepath.Join(s.mediaDir, entry.Name())
					break
				}
			}
		}
	}

	// Walk directory
	err := filepath.Walk(categoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden files/directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check extension
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !videoExtensions[ext] {
			return nil
		}

		// Skip trailers, teasers, etc.
		lowerName := strings.ToLower(info.Name())
		for _, keyword := range skipKeywords {
			if strings.Contains(lowerName, keyword) {
				return nil
			}
		}

		// Get relative path
		relPath, err := filepath.Rel(s.mediaDir, path)
		if err != nil {
			return nil
		}

		// Generate ID
		itemID := strings.ReplaceAll(relPath, string(os.PathSeparator), "_")

		// Clean title
		title := cleanTitle(info.Name())

		// Check for HLS
		hlsPath := filepath.Join(s.mediaDir, "hls", itemID, "master.m3u8")
		var hlsURL *string
		if _, err := os.Stat(hlsPath); err == nil {
			url := "/media/hls/" + itemID + "/master.m3u8"
			hlsURL = &url
		}

		items = append(items, MediaItem{
			ID:           itemID,
			Title:        title,
			Category:     category,
			Path:         filepath.ToSlash(relPath),
			HLSUrl:       hlsURL,
			Size:         info.Size(),
			ModifiedTime: float64(info.ModTime().Unix()),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Scanner) ScanAll() ([]MediaItem, error) {
	allItems := []MediaItem{}

	categories := []struct {
		name    string
		dirName string
	}{
		{"Movies", "Movies"},
		{"TvShows", "TvShows"},
		{"Anime", "Anime"},
	}

	for _, cat := range categories {
		items, err := s.ScanCategory(cat.name, cat.dirName)
		if err != nil {
			// Log but continue
			continue
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

func (s *Scanner) Search(query, category string) ([]MediaItem, error) {
	var items []MediaItem
	var err error

	if category != "" {
		// Scan specific category
		dirName := getCategoryDir(category)
		items, err = s.ScanCategory(category, dirName)
	} else {
		// Scan all
		items, err = s.ScanAll()
	}

	if err != nil {
		return nil, err
	}

	// Filter by query
	queryLower := strings.ToLower(query)
	filtered := []MediaItem{}
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Title), queryLower) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

func (s *Scanner) GetMediaPath(relativePath string) string {
	return filepath.Join(s.mediaDir, filepath.FromSlash(relativePath))
}

func (s *Scanner) GetHLSPath(fileID, filename string) string {
	return filepath.Join(s.mediaDir, "hls", fileID, filename)
}

func cleanTitle(filename string) string {
	// Remove extension
	title := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Remove year in parentheses/brackets
	yearPattern := regexp.MustCompile(`\s*[\(\[]?\d{4}[\)\]]?\s*`)
	title = yearPattern.ReplaceAllString(title, "")

	// Clean up
	title = strings.TrimSpace(title)

	return title
}

func getCategoryDir(category string) string {
	switch strings.ToLower(category) {
	case "movies":
		return "Movies"
	case "tvshows", "series":
		return "TvShows"
	case "anime":
		return "Anime"
	default:
		return "Movies"
	}
}
