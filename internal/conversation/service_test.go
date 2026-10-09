package conversation

import "testing"

// Regression test: Marathi STT detection ("mr") must survive the
// Devanagari script heuristic, which defaults Devanagari to Hindi.
func TestResolveReplyLang(t *testing.T) {
	marathi := "नमस्कार, तू कसा आहेस?"
	hindi := "नमस्ते, आप कैसे हैं?"
	tamil := "வணக்கம், எப்படி இருக்கிறீர்கள்?"
	english := "hello, how are you?"

	cases := []struct {
		name string
		text string
		lang string
		want string
	}{
		{"marathi voice keeps mr", marathi, "mr", "mr"},
		{"marathi voice with hi UI keeps hi", marathi, "hi", "hi"},
		{"devanagari text unknown defaults hi", marathi, "", "hi"},
		{"hindi voice keeps hi", hindi, "hi", "hi"},
		{"hindi text with en UI becomes hi", hindi, "en", "hi"},
		{"tamil text with en UI becomes ta", tamil, "en", "ta"},
		{"tamil voice keeps ta", tamil, "ta", "ta"},
		{"english text keeps en", english, "en", "en"},
		{"english text unknown stays unknown", english, "", ""},
	}
	for _, c := range cases {
		if got := resolveReplyLang(c.text, c.lang); got != c.want {
			t.Errorf("%s: resolveReplyLang(%q, %q) = %q, want %q", c.name, c.text, c.lang, got, c.want)
		}
	}
}
