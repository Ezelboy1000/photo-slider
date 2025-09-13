# Photo Slider

A simple Go application that generates an HTML photo slider with customizable colors and author display options.

## Features

- **Image Discovery**: Scans the `images` folder for supported image formats
- **Author/Title Parsing**: Extracts author and title from filename format `author - title`
- **Customizable Colors**: Configure text colors, stroke colors, and image borders via config file
- **Author Display Toggle**: Show or hide author names in captions
- **OBS Integration**: Ready to use as a web source in OBS Studio

## Supported Image Formats

- JPG/JPEG
- PNG
- GIF
- WebP

## Installation

### Download pre-made binary

1. Go to releases and download the latest release
2. Move it to a easy to find location (preferably inside a new empty folder)

### Build it yourself

1. Clone or download this repository
2. Ensure you have Go installed (Go 1.19 or later recommended)
3. Compile the application:
   ```bash
   go build -o photo-slider.exe
   ```

## Usage

### Basic Usage

1. **Images Folder**: Place your images in the `images` folder
2. **Run the Application**: Execute `photo-slider.exe` or `go run main.go`
3. **Use in OBS**: Add `photo.html` as a web source in OBS Studio

### Image Naming Convention

Name your images using the format: `author - title.ext`

Examples:
- `john doe - sunset photo.jpg`
- `artist name - my artwork.png`
- `photographer - landscape view.jpeg`

If you don't use the `author - title` format, the filename will be used as the title and the "Author" won't be displayed.

### Special Characters

- Use `%` in filenames to create line breaks in the displayed text
- Example: `artist - long%title.png` will display as:
  ```
  artist
  long
  title
  ```

## Configuration

The application uses a configuration file `photo-slider.config` to customize behavior and appearance. This file is automatically created on first run with default values.

### Configuration Options

| Option | Description | Default Value | Example |
|--------|-------------|---------------|---------|
| `include_author` | Show/hide author names in captions | `true` | `false` |
| `author_text_color` | Color of author text | `#ffffff` | `#ff0000` |
| `author_stroke_color` | Color of author text stroke | `#803128` | `#000000` |
| `title_text_color` | Color of title text | `#ffffff` | `#00ff00` |
| `title_stroke_color` | Color of title text stroke | `#bd685e` | `#0000ff` |
| `image_border_color` | Color of image border | `#741d34` | `#ffff00` |
| `image_border_style` | Style of image border | `dashed` | `solid` |

### Example Configuration File

```ini
# Photo Slider Configuration
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
```

## Output

The application generates `photo.html` which contains:
- A horizontally scrolling gallery of all images
- Customizable text styling based on your configuration
- Responsive design that works well in streaming applications
- Smooth CSS animations for continuous scrolling

## OBS Studio Integration

1. In OBS Studio, add a new "Browser Source"
2. Set the URL to the full path of `photo.html` (e.g., `file:///C:/path/to/photo.html`)
3. Set the width and height as needed (recommended: 1920x1080 or your stream resolution)
4. The photo slider will display with your custom colors and settings

## File Structure

```
photo-slider/
├── main.go                 # Main application code
├── go.mod                  # Go module file
├── photo-slider.config     # Configuration file (auto-generated)
├── photo.html              # Generated HTML output
├── images/                 # Folder for your images
│   ├── author1 - title1.jpg
│   ├── author2 - title2.png
│   └── ...
└── README.md               # This file
```

## Requirements

- Go 1.19 or later
- A web browser (for viewing the generated HTML)
- OBS Studio (for streaming integration)

## Troubleshooting

### Images Not Showing
- Ensure images are in the `images` folder
- Check that image files have supported extensions (.jpg, .jpeg, .png, .gif, .webp)
- Verify file permissions allow reading the images

### Colors Not Applied
- Check that the config file `photo-slider.config` exists
- Verify color values are in correct hex format (e.g., `#ffffff`)
- Ensure there are no typos in the configuration option names

### OBS Not Displaying
- Use the full file path for the HTML file in OBS
- Try refreshing the browser source in OBS
- Check that the HTML file was generated successfully

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.
