# Assets

## beot.svg

Source SVG for the application icon. Features an Anglo-Saxon inspired design with:
- Gold "B" letterform with knotwork styling
- Timer arc suggesting the Pomodoro technique
- Dark background with gold accents

## Generating beot.ico

The ICO file needs to be generated from the SVG. Use one of these methods:

### Option 1: ImageMagick (recommended)
```bash
# Install ImageMagick, then:
magick convert beot.svg -define icon:auto-resize=256,128,64,48,32,16 beot.ico
```

### Option 2: Online converter
- https://convertio.co/svg-ico/
- https://cloudconvert.com/svg-to-ico

Upload beot.svg and select multiple sizes (16, 32, 48, 64, 128, 256).

### Option 3: Inkscape
1. Open beot.svg in Inkscape
2. Export as PNG at 256x256
3. Use an ICO converter to create multi-resolution ICO

## Icon sizes included in ICO

Windows recommends these sizes for best display:
- 16x16 - Small icon (taskbar, file lists)
- 32x32 - Medium icon
- 48x48 - Large icon
- 64x64 - Extra large
- 128x128 - Jumbo
- 256x256 - Thumbnail/high DPI
