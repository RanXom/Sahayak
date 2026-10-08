#!/usr/bin/env bash
# Download the Noto script fonts used by PDF generation. The fonts cover the
# scripts of the 22 languages Sarvam AI supports. Run from any shell:
#   bash scripts/fetch-fonts.sh
#   ./scripts/fetch-fonts.sh
set -euo pipefail

DEST="${FONT_DIR:-$(cd "$(dirname "$0")/.." && pwd)/fonts}"
BASE_URL="https://github.com/notofonts/notofonts.github.io/raw/main/fonts"

FAMILIES=(
  NotoSans
  NotoSansDevanagari
  NotoSansBengali
  NotoSansGujarati
  NotoSansKannada
  NotoSansMalayalam
  NotoSansOriya
  NotoSansGurmukhi
  NotoSansTamil
  NotoSansTelugu
  NotoSansArabic
  NotoSansOlChiki
  NotoSansMeeteiMayek
)

mkdir -p "$DEST"

for family in "${FAMILIES[@]}"; do
  for style in Regular Bold; do
    file="$family-$style.ttf"
    if [ -f "$DEST/$file" ]; then
      echo "skip  $file"
      continue
    fi
    url="$BASE_URL/$family/hinted/ttf/$file"
    if curl -fsSL "$url" -o "$DEST/$file"; then
      echo "fetch $file"
    else
      rm -f "$DEST/$file"
      echo "warn  could not download $file" >&2
    fi
  done
done

echo "fonts written to $DEST"
