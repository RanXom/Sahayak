# Fonts

PDF generation (Section 21 of the implementation plan) requires Unicode fonts
with Indic script coverage. These fonts cover the scripts of the 22 languages
Sarvam AI supports. They are copied into the Docker image at `/app/fonts`.

Fetch them with:

```bash
make fetch-fonts      # or: bash scripts/fetch-fonts.sh
```

| File | Script | Languages |
|------|--------|-----------|
| `NotoSans-Regular.ttf`, `NotoSans-Bold.ttf` | Latin | English |
| `NotoSansDevanagari-*.ttf` | Devanagari | Hindi, Marathi, Nepali, Sanskrit, Bodo, Dogri, Konkani, Maithili |
| `NotoSansBengali-*.ttf` | Bengali | Bengali, Assamese |
| `NotoSansGujarati-*.ttf` | Gujarati | Gujarati |
| `NotoSansKannada-*.ttf` | Kannada | Kannada |
| `NotoSansMalayalam-*.ttf` | Malayalam | Malayalam |
| `NotoSansOriya-*.ttf` | Odia | Odia |
| `NotoSansGurmukhi-*.ttf` | Gurmukhi | Punjabi |
| `NotoSansTamil-*.ttf` | Tamil | Tamil |
| `NotoSansTelugu-*.ttf` | Telugu | Telugu |
| `NotoSansArabic-*.ttf` | Arabic | Urdu, Kashmiri, Sindhi |
| `NotoSansOlChiki-*.ttf` | Ol Chiki | Santali |
| `NotoSansMeeteiMayek-*.ttf` | Meetei Mayek | Manipuri |

Source: https://github.com/notofonts/notofonts.github.io (SIL Open Font License).
These are not committed to the repository to keep it lightweight.
