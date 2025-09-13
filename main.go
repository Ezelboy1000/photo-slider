package main

import (
	"bufio"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const (
	imageFolder = "images"
	outputFile  = "photo.html"
	configFile  = "photo-slider.config"
)

var allowedExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
}

type imageMeta struct {
	relPath string
	author  string
	title   string
}

type config struct {
	includeAuthor     bool
	authorTextColor   string
	authorStrokeColor string
	titleTextColor    string
	titleStrokeColor  string
	imageBorderColor  string
	imageBorderStyle  string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Read config file
	cfg, err := readConfig()
	if err != nil {
		return err
	}
	// Ensure images directory exists
	if _, err := os.Stat(imageFolder); errors.Is(err, fs.ErrNotExist) {
		if mkErr := os.MkdirAll(imageFolder, 0o755); mkErr != nil {
			return fmt.Errorf("failed to create %s: %w", imageFolder, mkErr)
		}
		fmt.Printf("Creating %s folder...\n", imageFolder)
		fmt.Printf("Please place your images in the %s folder and run this program again.\n", imageFolder)
		return nil
	}

	// Discover images
	images, err := findImages(imageFolder)
	if err != nil {
		return err
	}

	// Randomize order for output
	rand.Shuffle(len(images), func(i, j int) { images[i], images[j] = images[j], images[i] })

	metas := make([]imageMeta, 0, len(images))
	for _, path := range images {
		base := filepath.Base(path)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		author, title := parseAuthorTitle(name)
		metas = append(metas, imageMeta{relPath: filepath.ToSlash(path), author: author, title: title})
	}

	if err := writeHTML(outputFile, metas, cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Generated %s with %d images from %s folder.\n", outputFile, len(metas), imageFolder)
	fmt.Println()
	fmt.Println("Instructions:")
	fmt.Printf("1. Place your images in the \"%s\" folder\n", imageFolder)
	fmt.Printf("2. Run this program to generate the HTML (edit %s to hide author)\n", configFile)
	fmt.Printf("3. Add %s as web source in OBS to view the photo slider\n", outputFile)
	fmt.Println()
	return nil
}

func findImages(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", root, err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if _, ok := allowedExt[ext]; !ok {
			continue
		}
		out = append(out, filepath.Join(root, e.Name()))
	}
	return out, nil
}

func parseAuthorTitle(filename string) (string, string) {
	// Expect format: "author - title"
	// If missing, author defaults to "Author" and title uses the filename
	if strings.Contains(filename, "-") {
		parts := strings.SplitN(filename, "-", 2)
		rawAuthor := strings.TrimSpace(parts[0])
		rawTitle := ""
		if len(parts) > 1 {
			rawTitle = strings.TrimSpace(parts[1])
		}
		repAuthor := strings.Replace(rawAuthor, "%", "<br>", -1)
		repTitle := strings.Replace(rawTitle, "%", "<br>", -1)
		author := strings.TrimSpace(repAuthor)
		title := strings.TrimSpace(repTitle)
		if author == "" {
			author = ""
		}
		if title == "" {
			title = ""
		}
		return author, title
	}
	filename = strings.Replace(filename, "%", "<br>", -1)
	return "", filename
}

func readConfig() (config, error) {
	// Default config values
	cfg := config{
		includeAuthor:     true,
		authorTextColor:   "#ffffff",
		authorStrokeColor: "#803128",
		titleTextColor:    "#ffffff",
		titleStrokeColor:  "#bd685e",
		imageBorderColor:  "#741d34",
		imageBorderStyle:  "dashed",
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); errors.Is(err, fs.ErrNotExist) {
		// Create default config file
		if err := createDefaultConfig(); err != nil {
			return cfg, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	// Read config file
	content, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "include_author":
				cfg.includeAuthor = value == "true"
			case "author_text_color":
				cfg.authorTextColor = value
			case "author_stroke_color":
				cfg.authorStrokeColor = value
			case "title_text_color":
				cfg.titleTextColor = value
			case "title_stroke_color":
				cfg.titleStrokeColor = value
			case "image_border_color":
				cfg.imageBorderColor = value
			case "image_border_style":
				cfg.imageBorderStyle = value
			}
		}
	}

	return cfg, nil
}

func createDefaultConfig() error {
	content := `# Photo Slider Configuration
# Set include_author to true to show author names, false to hide them
include_author=true

# Color customization (use hex color codes like #ffffff)
author_text_color=#ffffff
author_stroke_color=#803128
title_text_color=#ffffff
title_stroke_color=#bd685e
image_border_color=#741d34

# Border style options: none, solid, dashed, dotted, double, groove, ridge, inset, outset
image_border_style=dashed
`
	return os.WriteFile(configFile, []byte(content), 0o644)
}

func writeHTML(path string, metas []imageMeta, cfg config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	// Begin HTML
	mustWrite(w, "<!DOCTYPE html>\n")
	mustWrite(w, "<html>\n")
	mustWrite(w, "  <head>\n")
	mustWrite(w, "    <title>Photo Slider</title>\n")
	mustWrite(w, "    <link rel=\"preconnect\" href=\"https://fonts.googleapis.com\">\n")
	mustWrite(w, "    <link rel=\"preconnect\" href=\"https://fonts.gstatic.com\" crossorigin>\n")
	mustWrite(w, "    <link href=\"https://fonts.googleapis.com/css2?family=Nunito:ital,wght@1,800&display=swap\" rel=\"stylesheet\">\n")
	mustWrite(w, "    <style>\n")
	mustWrite(w, "      html, body {\n")
	mustWrite(w, "        display: flex;\n")
	mustWrite(w, "        flex-direction: column;\n")
	mustWrite(w, "        width: 100%;\n")
	mustWrite(w, "        height: 100%;\n")
	mustWrite(w, "        margin: 0px;\n")
	mustWrite(w, "        padding: 0px;\n")
	mustWrite(w, "        overflow: hidden;\n")
	mustWrite(w, "        max-width: 100%;\n")
	mustWrite(w, "        overflow-x: hidden;\n")
	mustWrite(w, "        scrollbar-width: none;\n")
	mustWrite(w, "        -ms-overflow-style: none;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "     html::-webkit-scrollbar, body::-webkit-scrollbar {\n")
	mustWrite(w, "       display: none;\n")
	mustWrite(w, "     }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      *, *::before, *::after {\n")
	mustWrite(w, "        box-sizing: border-box;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas {\n")
	mustWrite(w, "        height: 750px;\n")
	mustWrite(w, "        position: absolute;\n")
	mustWrite(w, "        overflow: hidden;\n")
	mustWrite(w, "        overflow-y: hidden;\n")
	mustWrite(w, "        white-space: nowrap;\n")
	mustWrite(w, "        left: 0;\n")
	mustWrite(w, "        animation-name: scroll;\n")
	mustWrite(w, fmt.Sprintf("        animation-duration: %ds;\n", len(metas)*5))
	mustWrite(w, "        animation-iteration-count: infinite;\n")
	mustWrite(w, "        animation-timing-function: linear;\n")
	mustWrite(w, "        display: flex;\n")
	mustWrite(w, "        width: max-content;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas .scroll-content {\n")
	mustWrite(w, "        display: flex;\n")
	mustWrite(w, "        white-space: nowrap;\n")
	mustWrite(w, "        flex-shrink: 0;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas .scroll-content-duplicate {\n")
	mustWrite(w, "        display: flex;\n")
	mustWrite(w, "        white-space: nowrap;\n")
	mustWrite(w, "        flex-shrink: 0;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      .image-container {\n")
	mustWrite(w, "        display: inline-block;\n")
	mustWrite(w, "        margin-top: 32px;\n")
	mustWrite(w, "        margin-right: 80px;\n")
	mustWrite(w, "        text-align: center;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas img {\n")
	mustWrite(w, "        height: 500px;\n")
	mustWrite(w, "        border-radius: 12px;\n")
	mustWrite(w, "        display: block;\n")
	mustWrite(w, "        margin-bottom: 10px;\n")
	mustWrite(w, fmt.Sprintf("        outline: 5px %s %s;\n", cfg.imageBorderStyle, cfg.imageBorderColor))
	mustWrite(w, "        outline-offset: 16px;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas .caption {\n")
	mustWrite(w, "        font-family: \"Nunito\", sans-serif;\n")
	mustWrite(w, "        white-space: normal;\n")
	mustWrite(w, "        overflow: hidden;\n")
	mustWrite(w, "        text-overflow: ellipsis;\n")
	mustWrite(w, "        max-width: 100%;\n")
	mustWrite(w, "        text-align: center;\n")
	mustWrite(w, "        margin: 0 auto;\n")
	mustWrite(w, "        margin-top: 32px;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas .author {\n")
	mustWrite(w, "        font-size: 48px;\n")
	mustWrite(w, fmt.Sprintf("        color: %s;\n", cfg.authorTextColor))
	mustWrite(w, fmt.Sprintf("        -webkit-text-stroke: 10px %s;\n", cfg.authorStrokeColor))
	mustWrite(w, "        paint-order: stroke fill;\n")
	mustWrite(w, "        font-weight: bold;\n")
	mustWrite(w, "        display: block;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      #permas .title {\n")
	mustWrite(w, "        font-size: 40px;\n")
	mustWrite(w, "        display: block;\n")
	mustWrite(w, fmt.Sprintf("        color: %s;\n", cfg.titleTextColor))
	mustWrite(w, fmt.Sprintf("        -webkit-text-stroke: 10px %s;\n", cfg.titleStrokeColor))
	mustWrite(w, "        paint-order: stroke fill;\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "\n")
	mustWrite(w, "      @keyframes scroll {\n")
	mustWrite(w, "        0% {\n")
	mustWrite(w, "          transform: translateX(0);\n")
	mustWrite(w, "        }\n")
	mustWrite(w, "        100% {\n")
	mustWrite(w, "          transform: translateX(-50%);\n")
	mustWrite(w, "        }\n")
	mustWrite(w, "      }\n")
	mustWrite(w, "    </style>\n")
	mustWrite(w, "  </head>\n")
	mustWrite(w, "  <body>\n")
	mustWrite(w, "    <div id=\"permas\">\n")
	mustWrite(w, "      <div class=\"scroll-content\">\n")

	for _, m := range metas {
		writeImageContainer(w, m, cfg)
	}

	mustWrite(w, "      </div>\n")
	mustWrite(w, "      <div class=\"scroll-content-duplicate\">\n")

	for _, m := range metas {
		writeImageContainer(w, m, cfg)
	}

	mustWrite(w, "      </div>\n")
	mustWrite(w, "    </div>\n")
	mustWrite(w, "  </body>\n")
	mustWrite(w, "</html>\n")

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush %s: %w", path, err)
	}
	return nil
}

func writeImageContainer(w *bufio.Writer, m imageMeta, cfg config) {
	mustWrite(w, "        <div class=\"image-container\">\n")
	mustWrite(w, fmt.Sprintf("          <img class=\"scroller\" src=\"%s\">\n", html.EscapeString(filepath.ToSlash(m.relPath))))
	mustWrite(w, "          <div class=\"caption\">\n")
	if cfg.includeAuthor {
		mustWrite(w, fmt.Sprintf("            <div class=\"author\">%s</div>\n", m.author))
	}
	mustWrite(w, fmt.Sprintf("            <div class=\"title\">%s</div>\n", m.title))
	mustWrite(w, "          </div>\n")
	mustWrite(w, "        </div>\n")
}

func mustWrite(w *bufio.Writer, s string) {
	if _, err := w.WriteString(s); err != nil {
		panic(err)
	}
}
