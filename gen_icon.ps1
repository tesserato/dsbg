magick -density 376 -background none "logo.svg" "content/logo.webp"

magick -background none "content/logo.webp" -crop 167x167+0+0 "assets/favicon.ico"