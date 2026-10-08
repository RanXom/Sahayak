# Fonts

PDF generation (Section 21 of the implementation plan) requires Unicode fonts
with Devanagari and other Indic script coverage. Place the following font files
in this directory (they are copied into the Docker image at `/app/fonts`):

- `NotoSansDevanagari-Regular.ttf` (Hindi, Marathi, Bhojpuri, Rajasthani)
- `NotoSansTamil-Regular.ttf`
- `NotoSansTelugu-Regular.ttf`
- `NotoSansKannada-Regular.ttf`
- `NotoSansBengali-Regular.ttf`
- `NotoSansGujarati-Regular.ttf`
- `NotoSansGurmukhi-Regular.ttf`
- `NotoSansMalayalam-Regular.ttf`

Download from https://fonts.google.com/noto (SIL Open Font License).
These are not committed to the repository to keep it lightweight.
