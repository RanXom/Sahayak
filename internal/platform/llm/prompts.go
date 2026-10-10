package llm

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FactKeys are the structured fact keys the extractor may return.
var FactKeys = []string{
	"age", "gender", "marital_status", "income_annual", "occupation",
	"caste_category", "residence_type", "state", "district", "land_acres",
	"ration_card_type", "has_bank_account", "has_aadhaar", "education_level",
	"disability_status", "minority_status", "family_members",
	"has_pucca_house", "has_electricity", "employment_status",
	"is_pregnant_or_lactating", "has_girl_child", "girl_child_age",
}

const extractionSystemPrompt = `You are a structured information extraction assistant for Sahayak, an Indian welfare scheme discovery platform.

Your ONLY job is to extract factual information from what the user says and identify what information is still missing.

RULES:
1. Extract facts as a JSON object. Allowed keys: %s.
2. Only extract facts the user actually stated. Never assume or infer.
3. Normalize values to these controlled vocabularies where applicable:
   - gender: male | female | other
   - marital_status: single | married | widowed | divorced | separated
   - caste_category: general | obc | sc | st | ews
   - residence_type: rural | urban
   - ration_card_type: apl | bpl | aay | none
   - education_level: illiterate | primary | secondary | higher_secondary | graduate | post_graduate
   - disability_status: none | moderate | severe | multiple
   - employment_status: employed | unemployed | self_employed | student | retired | homemaker
   Numbers must be numeric (age, income_annual, land_acres).
4. If a statement is ambiguous, list it under "ambiguities" instead of guessing.
5. List fact keys you could not determine but that were asked about under "needs_follow_up".
6. Do NOT evaluate eligibility.
7. Do NOT mention specific schemes.
8. Do NOT fabricate information.
9. Respond with JSON only, matching the requested schema.

Use EXACTLY these keys and canonical values. If unsure, omit the fact.

EXAMPLES:
"I am a 45 year old widow from Rajasthan, I farm 2 acres and have a BPL card"
-> {"age":45,"gender":"female","marital_status":"widowed","state":"Rajasthan","occupation":"farmer","land_acres":2.0,"ration_card_type":"bpl"}
"मेरी उम्र 50 साल है और मैं विधवा हूँ"
-> {"age":50,"gender":"female","marital_status":"widowed"}
"I do daily wage work in a village and have no bank account"
-> {"occupation":"labourer","residence_type":"rural","has_bank_account":false}`

// ExtractionSystemPrompt builds the system prompt for fact extraction.
func ExtractionSystemPrompt() string {
	return fmt.Sprintf(extractionSystemPrompt, strings.Join(FactKeys, ", "))
}

// BuildExtractionUserPrompt builds the user turn for fact extraction.
func BuildExtractionUserPrompt(req ExtractFactsRequest) string {
	var b strings.Builder
	b.WriteString("Current profile (already known facts): ")
	if len(req.CurrentProfile) == 0 {
		b.WriteString("{}")
	} else {
		b.WriteString(jsonString(req.CurrentProfile))
	}
	b.WriteString("\n\nUser's message: ")
	b.WriteString(req.UserText)
	b.WriteString("\n\nRespond with JSON: {\"facts\":{...},\"confidence\":0.0,\"ambiguities\":[],\"needs_follow_up\":[]}")
	return b.String()
}

// QuestionForField returns a deterministic, localized question for a fact key.
// It is used both as the LLM prompt hint and as a fallback when the LLM is down.
func QuestionForField(field, lang string) string {
	type q struct{ en, hi, mr string }
	questions := map[string]q{
		"state":             {"Which state do you live in?", "आप किस राज्य में रहते हैं?", "तुम्ही कोणत्या राज्यात राहता?"},
		"gender":            {"Please tell me whether you are male or female.", "कृपया बताएं कि आप पुरुष या नारी हैं।", "कृपया सांगा, तुम्ही पुरुष आहात की स्त्री आहात."},
		"age":               {"How old are you?", "आपकी उम्र कितनी है?", "तुमचे वय किती आहे?"},
		"marital_status":    {"What is your marital status?", "आपकी वैवाहिक स्थिति क्या है?", "तुमची वैवाहिक स्थिती काय आहे?"},
		"income_annual":     {"What is your annual family income?", "आपकी वार्षिक पारिवारिक आय कितनी है?", "तुमचे वार्षिक कौटुंबिक उत्पन्न किती आहे?"},
		"occupation":        {"What is your occupation?", "आपका व्यवसाय क्या है?", "तुमचा व्यवसाय काय आहे?"},
		"caste_category":    {"Which social category do you belong to (General, OBC, SC, ST)?", "आप किस सामाजिक वर्ग से हैं (सामान्य, ओबीसी, अनुसूचित जाति, अनुसूचित जनजाति)?", "तुम्ही कोणत्या सामाजिक प्रवर्गातील आहात (सर्वसाधारण, ओबीसी, अनुसूचित जाती, अनुसूचित जमाती)?"},
		"residence_type":    {"Do you live in a rural or urban area?", "आप ग्रामीण या शहरी क्षेत्र में रहते हैं?", "तुम्ही ग्रामीण की शहरी भागात राहता?"},
		"ration_card_type":  {"What type of ration card does your household have?", "आपके परिवार के पास किस प्रकार का राशन कार्ड है?", "तुमच्या कुटुंबाकडे कोणत्या प्रकारचे रेशन कार्ड आहे?"},
		"land_acres":        {"How much agricultural land does your family own (in acres)?", "आपके परिवार के पास कितनी कृषि भूमि है (एकड़ में)?", "तुमच्या कुटुंबाकडे किती शेतजमीन आहे (एकरमध्ये)?"},
		"has_bank_account":  {"Do you have a bank account?", "क्या आपका बैंक खाता है?", "तुमचे बँक खाते आहे का?"},
		"has_aadhaar":       {"Do you have an Aadhaar card?", "क्या आपके पास आधार कार्ड है?", "तुमच्याकडे आधार कार्ड आहे का?"},
		"education_level":   {"What is your education level?", "आपकी शिक्षा का स्तर क्या है?", "तुमची शैक्षणिक पात्रता काय आहे?"},
		"family_members":    {"Tell me about your family members and their ages.", "अपने परिवार के सदस्यों और उनकी उम्र के बारे में बताइए।", "तुमच्या कुटुंबातील सदस्यांबद्दल आणि त्यांच्या वयाबद्दल सांगा."},
		"has_pucca_house":   {"Does your family own a pucca (permanent) house?", "क्या आपके परिवार के पास पक्का घर है?", "तुमच्या कुटुंबाकडे पक्के (कायमस्वरूपी) घर आहे का?"},
		"has_electricity":   {"Does your household have an electricity connection?", "क्या आपके घर में बिजली कनेक्शन है?", "तुमच्या घरात वीज जोडणी आहे का?"},
		"employment_status": {"Are you currently employed, unemployed, or a student?", "आप वर्तमान में कार्यरत, बेरोजगार या विद्यार्थी हैं?", "तुम्ही सध्या नोकरदार, बेरोजगार की विद्यार्थी आहात?"},
	}
	if lang == "hi" {
		if v, ok := questions[field]; ok {
			return v.hi
		}
		return "कृपया इसके बारे में अधिक जानकारी दें: " + field
	}
	if lang == "mr" {
		if v, ok := questions[field]; ok {
			return v.mr
		}
		return "कृपया याबद्दल अधिक माहिती द्या: " + field
	}
	if v, ok := questions[field]; ok {
		return v.en
	}
	return "Could you tell me more about your " + strings.ReplaceAll(field, "_", " ") + "?"
}

// FollowUpSystemPrompt builds the system prompt used to phrase the next turn.
// The model is allowed to answer questions about Sahayak itself, then must ask
// exactly one intake question. It may never decide eligibility.
func FollowUpSystemPrompt(lang string) string {
	return "You are Sahayak, a warm, respectful assistant that helps Indian citizens " +
		"discover welfare schemes and build their profile by voice or text. " +
		"The user may ask questions about you, what you do, how voice input works, or " +
		"which languages you support; when they do, answer briefly and helpfully first.\n\n" +
		"RULES:\n" +
		"1. Always write in " + languageName(lang) + ".\n" +
		"2. If the user's latest message is a question or greeting, answer it briefly and accurately first.\n" +
		"3. Then, if the next piece of information is still needed, ask exactly one short, clear question about it.\n" +
		"4. Do not ask again for information the user has already given.\n" +
		"5. Never decide or explain eligibility. Never name specific schemes. Never invent facts, documents or URLs.\n" +
		"6. Keep the whole reply to 1-3 short, natural, respectful sentences.\n" +
		"7. Respond with JSON only: {\"question\":\"<full reply>\",\"question_about\":\"<fact key or empty>\",\"is_last\":false}."
}

// BuildFollowUpUserPrompt builds the user turn for the next reply. It carries the
// user's latest message, the known profile, the next field to ask about and the
// recent conversation so the model can both answer and continue the intake.
func BuildFollowUpUserPrompt(req FollowUpRequest) string {
	field := ""
	if len(req.MissingFields) > 0 {
		field = req.MissingFields[0]
	}
	var b strings.Builder
	b.WriteString("User's latest message: ")
	b.WriteString(req.UserText)
	b.WriteString("\n\nKnown profile facts: ")
	if len(req.ProfileFacts) == 0 {
		b.WriteString("{}")
	} else {
		b.WriteString(jsonString(req.ProfileFacts))
	}
	b.WriteString("\n\nNext information needed (field key): ")
	b.WriteString(field)
	b.WriteString("\nAsk for this field naturally in ")
	b.WriteString(languageName(req.Language))
	b.WriteString(". Do NOT copy the example phrasing; use your own natural words that fit the context.")
	if len(req.History) > 0 {
		b.WriteString("\n\nRecent conversation (oldest first):")
		for _, m := range req.History {
			b.WriteString("\n")
			b.WriteString(m.Role)
			b.WriteString(": ")
			b.WriteString(m.Content)
		}
	}
	b.WriteString("\n\nWrite the reply in " + languageName(req.Language) +
		": first answer the user's message if it is a question, then ask the next question. Respond with JSON only.")
	return b.String()
}

// FallbackExplanation builds a deterministic explanation when the LLM is unavailable.
func FallbackExplanation(req ExplanationRequest) string {
	if req.Language == "hi" {
		switch req.Status {
		case "eligible":
			return "आपकी प्रोफ़ाइल के अनुसार आप इस योजना की सभी पात्रता शर्तें पूरी करते हैं।"
		case "potentially_eligible":
			return "कुछ जानकारी अभी अपूर्ण है। पात्रता की पुष्टि के लिए कृपया और जानकारी दें।"
		default:
			return "आपकी वर्तमान प्रोफ़ाइल इस योजना की कुछ अनिवार्य शर्तें पूरी नहीं करती।"
		}
	}
	if req.Language == "mr" {
		switch req.Status {
		case "eligible":
			return "तुमच्या प्रोफाइलनुसार तुम्ही या योजनेच्या सर्व पात्रता अटी पूर्ण करता."
		case "potentially_eligible":
			return "अजून थोडी माहिती अपूर्ण आहे. पात्रता निश्चित करण्यासाठी कृपया आणखी माहिती द्या."
		default:
			return "तुमची सध्याची प्रोफाइल या योजनेच्या काही आवश्यक अटी पूर्ण करत नाही."
		}
	}
	switch req.Status {
	case "eligible":
		return "Based on your profile, you meet all the eligibility criteria for this scheme."
	case "potentially_eligible":
		return "Some information is still needed to confirm your eligibility for this scheme."
	default:
		return "Your current profile does not meet some of the mandatory criteria for this scheme."
	}
}

func jsonString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
