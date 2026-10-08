package notification

import (
	"strings"
)

// Template keys.
const (
	TemplateDeadlineReminder = "deadline_reminder"
	TemplateStatusUpdate     = "status_update"
	TemplateFollowUpReminder = "follow_up_reminder"
	TemplateNewSchemeAlert   = "new_scheme_alert"
	TemplateApplicationReady = "application_ready"
)

// Render produces the message body for a template in the given language.
// Content is pre-approved and never LLM-generated.
func Render(templateKey, lang string, params map[string]string) string {
	hi := lang == "hi"
	switch templateKey {
	case TemplateDeadlineReminder:
		if hi {
			return "अनुस्मारक: " + param(params, "scheme", "योजना") + " के आवेदन की अंतिम तिथि " + param(params, "date", "") + " है। कृपया समय पर आवेदन करें।"
		}
		return "Reminder: the application deadline for " + param(params, "scheme", "the scheme") + " is " + param(params, "date", "") + ". Please apply in time."
	case TemplateStatusUpdate:
		if hi {
			return "आपके आवेदन की स्थिति अपडेट हुई है: " + param(params, "status", "")
		}
		return "Your application status has been updated: " + param(params, "status", "")
	case TemplateFollowUpReminder:
		if hi {
			return "आपने अपनी सहायक प्रोफ़ाइल पूरी नहीं की। पात्र योजनाएँ खोजने के लिए कृपया इसे पूरा करें।"
		}
		return "You have not completed your Sahayak profile. Please complete it to discover schemes you may be eligible for."
	case TemplateNewSchemeAlert:
		if hi {
			return "आपकी प्रोफ़ाइल से मेल खाने वाली एक नई योजना उपलब्ध है: " + param(params, "scheme", "")
		}
		return "A new scheme matching your profile is available: " + param(params, "scheme", "")
	case TemplateApplicationReady:
		if hi {
			return "आपका आवेदन पत्र तैयार है: " + param(params, "scheme", "") + ". आप इसे सहायक ऐप से डाउनलोड कर सकते हैं।"
		}
		return "Your application form is ready: " + param(params, "scheme", "") + ". You can download it from the Sahayak app."
	default:
		return param(params, "message", "")
	}
}

func param(params map[string]string, key, fallback string) string {
	if v, ok := params[key]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
