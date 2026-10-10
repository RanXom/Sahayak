package eligibility

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sahayak/sahayak/internal/document"
	"github.com/sahayak/sahayak/internal/domain"
	"github.com/sahayak/sahayak/internal/location"
	"github.com/sahayak/sahayak/internal/profile"
	"github.com/sahayak/sahayak/internal/scheme"
)

// Service orchestrates eligibility evaluation for a user.
type Service struct {
	profiles  *profile.Repository
	schemes   *scheme.Service
	location  *location.Repository
	documents *document.Service
	engine    *Engine
}

// NewService constructs the eligibility service.
func NewService(profiles *profile.Repository, schemes *scheme.Service, loc *location.Repository, docs *document.Service) *Service {
	return &Service{
		profiles:  profiles,
		schemes:   schemes,
		location:  loc,
		documents: docs,
		engine:    NewEngine(),
	}
}

// ResultView is an evaluated scheme enriched for the API.
type ResultView struct {
	Scheme        *domain.Scheme        `json:"scheme"`
	Status        string                `json:"status"`
	MatchScore    float64               `json:"match_score"`
	RuleResults   []RuleEvaluation      `json:"rule_results"`
	MissingFields []string              `json:"missing_fields,omitempty"`
	Explanation   string                `json:"explanation"`
	Documents     *document.CheckResult `json:"documents,omitempty"`
}

// EvaluateResponse is the response for POST /eligibility/evaluate.
type EvaluateResponse struct {
	EvaluatedAt         time.Time      `json:"evaluated_at"`
	ProfileCompleteness float64        `json:"profile_completeness"`
	CoverageNote        string         `json:"coverage_note"`
	Results             []ResultView   `json:"results"`
	Summary             map[string]int `json:"summary"`
}

// Evaluate runs the eligibility engine for the user's profile.
func (s *Service) Evaluate(ctx context.Context, userID, lang string) (*EvaluateResponse, error) {
	p, err := s.profiles.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, profile.ErrNotFound) {
			return nil, domain.InvalidInput("Please complete your profile before evaluating eligibility")
		}
		return nil, domain.Internal("Could not load profile", err)
	}

	stateCode := ""
	if p.StateID != nil {
		stateCode, _ = s.location.StateCode(ctx, *p.StateID)
	}

	schemes, err := s.schemes.ListForEvaluation(ctx, p.StateID)
	if err != nil {
		return nil, err
	}

	env := FlattenProfile(p.Facts, stateCode)
	results := s.engine.Evaluate(env, schemes)

	views := make([]ResultView, 0, len(results))
	for _, r := range results {
		view := ResultView{
			Scheme:        r.Scheme,
			Status:        r.Status,
			MatchScore:    r.MatchScore,
			RuleResults:   r.RuleResults,
			MissingFields: r.MissingFields,
			Explanation:   buildExplanation(r, lang),
		}
		if r.Status == MatchEligible || r.Status == MatchPotential {
			if check, err := s.documents.CheckScheme(ctx, userID, r.Scheme.ID, lang); err == nil {
				view.Documents = check
			}
		}
		views = append(views, view)
	}

	return &EvaluateResponse{
		EvaluatedAt:         time.Now().UTC(),
		ProfileCompleteness: completeness(p.Facts),
		CoverageNote:        s.coverageNote(ctx, p.StateID),
		Results:             views,
		Summary:             Summarize(results),
	}, nil
}

func (s *Service) coverageNote(ctx context.Context, stateID *int) string {
	if stateID == nil {
		return "Central government schemes are shown. Add your state for state-specific schemes."
	}
	c, err := s.location.GetCoverage(ctx, *stateID)
	if err != nil {
		return "Central government schemes are shown."
	}
	switch c.Level {
	case domain.CoverageFull:
		return "Full scheme coverage for your state."
	case domain.CoveragePartial:
		return "Coverage is partial for your state. More schemes may exist."
	default:
		return "We don't yet have state-specific scheme data for your state. Only central government schemes are shown."
	}
}

// coreFactKeys are used to estimate profile completeness.
var coreFactKeys = []string{
	"age", "gender", "marital_status", "income_annual", "occupation",
	"caste_category", "residence_type", "land_acres", "ration_card_type",
	"has_bank_account", "has_aadhaar",
}

func completeness(facts map[string]any) float64 {
	if len(coreFactKeys) == 0 {
		return 0
	}
	present := 0
	for _, k := range coreFactKeys {
		if v, ok := facts[k]; ok && v != nil {
			present++
		}
	}
	return float64(present) / float64(len(coreFactKeys))
}

// buildExplanation produces a deterministic, grounded explanation. It never
// invents criteria: it only describes the rules that were evaluated.
func buildExplanation(r SchemeResult, lang string) string {
	switch r.Status {
	case MatchEligible:
		if lang == "hi" {
			return "आपकी प्रोफ़ाइल के अनुसार आप इस योजना की सभी पात्रता शर्तें पूरी करते हैं।"
		}
		if lang == "mr" {
			return "तुमच्या प्रोफाइलनुसार तुम्ही या योजनेच्या सर्व पात्रता अटी पूर्ण करता."
		}
		return "Based on your profile, you meet all the eligibility criteria for this scheme."
	case MatchPotential:
		if lang == "hi" {
			return "कुछ जानकारी अभी अपूर्ण है। पात्रता की पुष्टि के लिए कृपया और जानकारी दें: " + joinFields(r.MissingFields)
		}
		if lang == "mr" {
			return "अजून थोडी माहिती अपूर्ण आहे. पात्रता निश्चित करण्यासाठी कृपया आणखी माहिती द्या: " + joinFields(r.MissingFields) + "."
		}
		return "Some information is still needed to confirm eligibility: " + joinFields(r.MissingFields) + "."
	default:
		if lang == "hi" {
			return "आपकी वर्तमान प्रोफ़ाइल इस योजना की कुछ अनिवार्य शर्तें पूरी नहीं करती।"
		}
		if lang == "mr" {
			return "तुमची सध्याची प्रोफाइल या योजनेच्या काही आवश्यक अटी पूर्ण करत नाही."
		}
		return "Your current profile does not meet some of the mandatory criteria for this scheme."
	}
}

func joinFields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	out := ""
	for i, f := range fields {
		if i > 0 {
			out += ", "
		}
		out += f
	}
	return out
}

// VoiceSummary evaluates the user's profile and returns a short, spoken-style
// summary of eligibility for the conversation: the top matching schemes with a
// percentage, the missing documents, and the facts still needed.
func (s *Service) VoiceSummary(ctx context.Context, userID, lang string) (string, error) {
	res, err := s.Evaluate(ctx, userID, lang)
	if err != nil {
		return "", err
	}
	return BuildVoiceSummary(res, lang), nil
}

// BuildVoiceSummary renders the deterministic eligibility response as a
// concise, natural-language summary. It never invents criteria or documents.
func BuildVoiceSummary(res *EvaluateResponse, lang string) string {
	hi := lang == "hi"
	mr := lang == "mr"
	var eligible, potential []ResultView
	for _, r := range res.Results {
		switch r.Status {
		case MatchEligible:
			eligible = append(eligible, r)
		case MatchPotential:
			potential = append(potential, r)
		}
	}

	if len(eligible) == 0 && len(potential) == 0 {
		if hi {
			return "आपकी दी गई जानकारी के आधार पर अभी कोई मेल खाती योजना नहीं मिली। इसका मतलब यह नहीं कि आप किसी सरकारी योजना के पात्र नहीं हैं। कृपया अपनी प्रोफ़ाइल में और जानकारी जोड़ें।"
		}
		if mr {
			return "तुम्ही दिलेल्या माहितीनुसार अजून कोणतीही जुळणारी योजना सापडली नाही. याचा अर्थ असा नाही की तुम्ही कोणत्याही सरकारी योजनेसाठी पात्र नाही. कृपया तुमच्या प्रोफाइलमध्ये आणखी माहिती जोडा."
		}
		return "Based on what you have shared, I could not find a matching scheme yet. That does not mean you are ineligible for all government schemes. Please add more details to your profile."
	}

	var b strings.Builder
	if len(eligible) > 0 {
		if hi {
			b.WriteString("आपकी जानकारी के अनुसार आप इन योजनाओं के लिए पात्र हैं: ")
		} else if mr {
			b.WriteString("तुमच्या माहितीनुसार तुम्ही या योजनांसाठी पात्र आहात: ")
		} else {
			b.WriteString("Based on your information, you are eligible for ")
		}
		b.WriteString(joinScored(eligible, lang, 3))
		if hi {
			b.WriteString("। ")
		} else {
			b.WriteString(". ")
		}
	}
	if len(potential) > 0 {
		if hi {
			b.WriteString("इन योजनाओं के लिए थोड़ी और जानकारी चाहिए: ")
		} else if mr {
			b.WriteString("या योजनांसाठी थोडी आणखी माहिती हवी आहे: ")
		} else {
			b.WriteString("For these schemes I need a little more information: ")
		}
		b.WriteString(joinScored(potential, lang, 3))
		if hi {
			b.WriteString("। ")
		} else {
			b.WriteString(". ")
		}
	}
	if len(eligible) > 0 && eligible[0].Documents != nil && eligible[0].Documents.MissingCount > 0 {
		top := eligible[0]
		if names := missingDocNames(top.Documents, 4); names != "" {
			if hi {
				fmt.Fprintf(&b, "%s के लिए आपके पास ये दस्तावेज़ नहीं हैं: %s। ", top.Scheme.NameHI, names)
			} else if mr {
				fmt.Fprintf(&b, "%s साठी तुमच्याकडे ही कागदपत्रे नाहीत: %s. ", top.Scheme.NameEN, names)
			} else {
				fmt.Fprintf(&b, "For %s you are missing these documents: %s. ", top.Scheme.NameEN, names)
			}
		}
	}
	if len(potential) > 0 && len(potential[0].MissingFields) > 0 {
		labels := humanFieldList(potential[0].MissingFields, lang)
		if hi {
			fmt.Fprintf(&b, "पूरी पात्रता की पुष्टि के लिए कृपया बताएं: %s।", labels)
		} else if mr {
			fmt.Fprintf(&b, "तुमची पूर्ण पात्रता निश्चित करण्यासाठी कृपया सांगा: %s.", labels)
		} else {
			fmt.Fprintf(&b, "To confirm your full eligibility, please tell me: %s.", labels)
		}
	}
	return strings.TrimSpace(b.String())
}

// MissingSummary returns a spoken-style summary of what is still missing for
// full eligibility: the facts to confirm and the documents to obtain.
func (s *Service) MissingSummary(ctx context.Context, userID, lang string) (string, error) {
	res, err := s.Evaluate(ctx, userID, lang)
	if err != nil {
		return "", err
	}
	return BuildMissingSummary(res, lang), nil
}

// BuildMissingSummary lists the missing facts and documents across the results.
func BuildMissingSummary(res *EvaluateResponse, lang string) string {
	hi := lang == "hi"
	mr := lang == "mr"
	facts := uniqueMissingFields(res.Results)
	docs := uniqueMissingDocs(res.Results)
	if len(facts) == 0 && len(docs) == 0 {
		if hi {
			return "फ़िलहाल मेरे पास पूरी जानकारी है। आप आवेदन शुरू कर सकते हैं।"
		}
		if mr {
			return "सध्या माझ्याकडे सर्व माहिती आहे. तुम्ही अर्ज सुरू करू शकता."
		}
		return "For now I have everything I need. You can start an application."
	}
	var b strings.Builder
	if len(facts) > 0 {
		labels := humanFieldList(facts, lang)
		if hi {
			fmt.Fprintf(&b, "पूरी पात्रता की पुष्टि के लिए मुझे यह जानकारी चाहिए: %s। ", labels)
		} else if mr {
			fmt.Fprintf(&b, "पूर्ण पात्रता निश्चित करण्यासाठी मला ही माहिती हवी आहे: %s. ", labels)
		} else {
			fmt.Fprintf(&b, "To confirm full eligibility I need: %s. ", labels)
		}
	}
	if len(docs) > 0 {
		names := strings.Join(docs, ", ")
		if hi {
			fmt.Fprintf(&b, "और आवेदन के लिए ये दस्तावेज़ चाहिए: %s।", names)
		} else {
			fmt.Fprintf(&b, "And these documents for your application: %s.", names)
		}
	}
	return strings.TrimSpace(b.String())
}

func uniqueMissingFields(rs []ResultView) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rs {
		if r.Status != MatchPotential {
			continue
		}
		for _, f := range r.MissingFields {
			if !seen[f] {
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	return out
}

func uniqueMissingDocs(rs []ResultView) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rs {
		if r.Documents == nil {
			continue
		}
		for _, d := range r.Documents.Documents {
			if !d.UserHas && d.IsMandatory && !seen[d.DocumentName] {
				seen[d.DocumentName] = true
				out = append(out, d.DocumentName)
			}
		}
	}
	return out
}

func joinScored(rs []ResultView, lang string, max int) string {
	parts := make([]string, 0, max)
	for i, r := range rs {
		if i >= max {
			break
		}
		name := r.Scheme.NameEN
		if lang == "hi" {
			name = r.Scheme.NameHI
		}
		entry := fmt.Sprintf("%s (%d%%", name, matchPercent(r))
		if r.Scheme.BenefitAmount != "" {
			entry += " · " + r.Scheme.BenefitAmount
		}
		entry += ")"
		parts = append(parts, entry)
	}
	return strings.Join(parts, ", ")
}

// matchPercent expresses confidence as a percentage: 100% when all mandatory
// rules pass, otherwise the share of rules already satisfied.
func matchPercent(r ResultView) int {
	if r.Status == MatchEligible {
		return 100
	}
	total := len(r.RuleResults)
	if total == 0 {
		return 0
	}
	passed := 0
	for _, rr := range r.RuleResults {
		if rr.Status == StatusPass {
			passed++
		}
	}
	p := int(float64(passed) / float64(total) * 100)
	if p >= 100 {
		p = 99
	}
	return p
}

func missingDocNames(c *document.CheckResult, max int) string {
	names := make([]string, 0, max)
	for _, d := range c.Documents {
		if !d.UserHas && d.IsMandatory {
			names = append(names, d.DocumentName)
			if len(names) >= max {
				break
			}
		}
	}
	return strings.Join(names, ", ")
}

var fieldLabelsEN = map[string]string{
	"income_annual":            "your annual family income",
	"has_pucca_house":          "whether you own a pucca house",
	"caste_category":           "your social category",
	"land_acres":               "how much land you own",
	"ration_card_type":         "your ration card type",
	"has_bank_account":         "whether you have a bank account",
	"has_aadhaar":              "whether you have an Aadhaar card",
	"residence_type":           "whether you live in a rural or urban area",
	"occupation":               "what work you do",
	"education_level":          "your education level",
	"disability_status":        "your disability status",
	"minority_status":          "whether you belong to a minority community",
	"has_electricity":          "whether your home has an electricity connection",
	"employment_status":        "your employment status",
	"family_members":           "your family members",
	"is_pregnant_or_lactating": "whether you are pregnant or lactating",
	"has_girl_child":           "whether you have a girl child",
	"girl_child_age":           "your girl child's age",
	"state_code":               "your state",
}

var fieldLabelsHI = map[string]string{
	"income_annual":            "आपकी वार्षिक पारिवारिक आय",
	"has_pucca_house":          "क्या आपके पास पक्का घर है",
	"caste_category":           "आपकी सामाजिक श्रेणी",
	"land_acres":               "आपके पास कितनी ज़मीन है",
	"ration_card_type":         "आपके राशन कार्ड का प्रकार",
	"has_bank_account":         "क्या आपका बैंक खाता है",
	"has_aadhaar":              "क्या आपके पास आधार कार्ड है",
	"residence_type":           "आप ग्रामीण या शहरी क्षेत्र में रहते हैं",
	"occupation":               "आप क्या काम करते हैं",
	"education_level":          "आपकी शिक्षा का स्तर",
	"disability_status":        "आपकी विकलांगता स्थिति",
	"minority_status":          "क्या आप अल्पसंख्यक समुदाय से हैं",
	"has_electricity":          "क्या आपके घर में बिजली कनेक्शन है",
	"employment_status":        "आपकी रोज़गार स्थिति",
	"family_members":           "आपके परिवार के सदस्य",
	"is_pregnant_or_lactating": "क्या आप गर्भवती या स्तनपान कराने वाली माँ हैं",
	"has_girl_child":           "क्या आपकी बालिका संतान है",
	"girl_child_age":           "आपकी बालिका की आयु",
	"state_code":               "आपका राज्य",
}

var fieldLabelsMR = map[string]string{
	"income_annual":            "तुमचे वार्षिक कौटुंबिक उत्पन्न",
	"has_pucca_house":          "तुमच्याकडे पक्के घर आहे का",
	"caste_category":           "तुमचा सामाजिक प्रवर्ग",
	"land_acres":               "तुमच्याकडे किती जमीन आहे",
	"ration_card_type":         "तुमच्या रेशन कार्डचा प्रकार",
	"has_bank_account":         "तुमचे बँक खाते आहे का",
	"has_aadhaar":              "तुमच्याकडे आधार कार्ड आहे का",
	"residence_type":           "तुम्ही ग्रामीण की शहरी भागात राहता",
	"occupation":               "तुम्ही काय काम करता",
	"education_level":          "तुमची शैक्षणिक पात्रता",
	"disability_status":        "तुमची दिव्यांग स्थिती",
	"minority_status":          "तुम्ही अल्पसंख्याक समाजातील आहात का",
	"has_electricity":          "तुमच्या घरात वीज जोडणी आहे का",
	"employment_status":        "तुमची रोजगार स्थिती",
	"family_members":           "तुमचे कुटुंबीय",
	"is_pregnant_or_lactating": "तुम्ही गर्भवती किंवा स्तनदा माता आहात का",
	"has_girl_child":           "तुम्हाला मुलगी आहे का",
	"girl_child_age":           "तुमच्या मुलीचे वय",
	"state_code":               "तुमचे राज्य",
}

func humanFieldList(fields []string, lang string) string {
	m := fieldLabelsEN
	if lang == "hi" {
		m = fieldLabelsHI
	} else if lang == "mr" {
		m = fieldLabelsMR
	}
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if v, ok := m[f]; ok {
			out = append(out, v)
		} else {
			out = append(out, strings.ReplaceAll(f, "_", " "))
		}
	}
	return strings.Join(out, ", ")
}
