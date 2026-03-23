package bot

import (
	"regexp"
	"sort"
	"strings"
)

type CountryInfo struct {
	Code string
	Flag string
}

var CountryMap = map[string]CountryInfo{
	"20":  {"EG", "🇪🇬"},
	"27":  {"ZA", "🇿🇦"},
	"211": {"SS", "🇸🇸"},
	"212": {"MA", "🇲🇦"},
	"213": {"DZ", "🇩🇿"},
	"216": {"TN", "🇹🇳"},
	"218": {"LY", "🇱🇾"},
	"220": {"GM", "🇬🇲"},
	"221": {"SN", "🇸🇳"},
	"222": {"MR", "🇲🇷"},
	"223": {"ML", "🇲🇱"},
	"224": {"GN", "🇬🇳"},
	"225": {"CI", "🇨🇮"},
	"226": {"BF", "🇧🇫"},
	"227": {"NE", "🇳🇪"},
	"228": {"TG", "🇹🇬"},
	"229": {"BJ", "🇧🇯"},
	"230": {"MU", "🇲🇺"},
	"231": {"LR", "🇱🇷"},
	"232": {"SL", "🇸🇱"},
	"233": {"GH", "🇬🇭"},
	"234": {"NG", "🇳🇬"},
	"235": {"TD", "🇹🇩"},
	"236": {"CF", "🇨🇫"},
	"237": {"CM", "🇨🇲"},
	"238": {"CV", "🇨🇻"},
	"239": {"ST", "🇸🇹"},
	"240": {"GQ", "🇬🇶"},
	"241": {"GA", "🇬🇦"},
	"242": {"CG", "🇨🇬"},
	"243": {"CD", "🇨🇩"},
	"244": {"AO", "🇦🇴"},
	"245": {"GW", "🇬🇼"},
	"248": {"SC", "🇸🇨"},
	"249": {"SD", "🇸🇩"},
	"250": {"RW", "🇷🇼"},
	"251": {"ET", "🇪🇹"},
	"252": {"SO", "🇸🇴"},
	"253": {"DJ", "🇩🇯"},
	"254": {"KE", "🇰🇪"},
	"255": {"TZ", "🇹🇿"},
	"256": {"UG", "🇺🇬"},
	"257": {"BI", "🇧🇮"},
	"258": {"MZ", "🇲🇿"},
	"260": {"ZM", "🇿🇲"},
	"261": {"MG", "🇲🇬"},
	"262": {"RE", "🇷🇪"},
	"263": {"ZW", "🇿🇼"},
	"264": {"NA", "🇳🇦"},
	"265": {"MW", "🇲🇼"},
	"266": {"LS", "🇱🇸"},
	"267": {"BW", "🇧🇼"},
	"268": {"SZ", "🇸🇿"},
	"269": {"KM", "🇰🇲"},
	// Asia
	"60":  {"MY", "🇲🇾"},
	"62":  {"ID", "🇮🇩"},
	"63":  {"PH", "🇵🇭"},
	"64":  {"NZ", "🇳🇿"},
	"65":  {"SG", "🇸🇬"},
	"66":  {"TH", "🇹🇭"},
	"81":  {"JP", "🇯🇵"},
	"82":  {"KR", "🇰🇷"},
	"84":  {"VN", "🇻🇳"},
	"86":  {"CN", "🇨🇳"},
	"90":  {"TR", "🇹🇷"},
	"91":  {"IN", "🇮🇳"},
	"92":  {"PK", "🇵🇰"},
	"93":  {"AF", "🇦🇫"},
	"94":  {"LK", "🇱🇰"},
	"95":  {"MM", "🇲🇲"},
	"98":  {"IR", "🇮🇷"},
	"850": {"KP", "🇰🇵"},
	"852": {"HK", "🇭🇰"},
	"853": {"MO", "🇲🇴"},
	"855": {"KH", "🇰🇭"},
	"856": {"LA", "🇱🇦"},
	"960": {"MV", "🇲🇻"},
	"961": {"LB", "🇱🇧"},
	"962": {"JO", "🇯🇴"},
	"963": {"SY", "🇸🇾"},
	"964": {"IQ", "🇮🇶"},
	"965": {"KW", "🇰🇼"},
	"966": {"SA", "🇸🇦"},
	"967": {"YE", "🇾🇪"},
	"968": {"OM", "🇴🇲"},
	"970": {"PS", "🇵🇸"},
	"971": {"AE", "🇦🇪"},
	"972": {"IL", "🇮🇱"},
	"973": {"BH", "🇧🇭"},
	"974": {"QA", "🇶🇦"},
	"975": {"BT", "🇧🇹"},
	"976": {"MN", "🇲🇳"},
	"977": {"NP", "🇳🇵"},
	"992": {"TJ", "🇹🇯"},
	"993": {"TM", "🇹🇲"},
	"994": {"AZ", "🇦🇿"},
	"995": {"GE", "🇬🇪"},
	"996": {"KG", "🇰🇬"},
	"998": {"UZ", "🇺🇿"},
	// Europe
	"30":  {"GR", "🇬🇷"},
	"31":  {"NL", "🇳🇱"},
	"32":  {"BE", "🇧🇪"},
	"33":  {"FR", "🇫🇷"},
	"34":  {"ES", "🇪🇸"},
	"36":  {"HU", "🇭🇺"},
	"39":  {"IT", "🇮🇹"},
	"40":  {"RO", "🇷🇴"},
	"41":  {"CH", "🇨🇭"},
	"43":  {"AT", "🇦🇹"},
	"44":  {"GB", "🇬🇧"},
	"45":  {"DK", "🇩🇰"},
	"46":  {"SE", "🇸🇪"},
	"47":  {"NO", "🇳🇴"},
	"48":  {"PL", "🇵🇱"},
	"49":  {"DE", "🇩🇪"},
	"350": {"GI", "🇬🇮"},
	"351": {"PT", "🇵🇹"},
	"352": {"LU", "🇱🇺"},
	"353": {"IE", "🇮🇪"},
	"354": {"IS", "🇮🇸"},
	"355": {"AL", "🇦🇱"},
	"356": {"MT", "🇲🇹"},
	"357": {"CY", "🇨🇾"},
	"358": {"FI", "🇫🇮"},
	"359": {"BG", "🇧🇬"},
	"370": {"LT", "🇱🇹"},
	"371": {"LV", "🇱🇻"},
	"372": {"EE", "🇪🇪"},
	"373": {"MD", "🇲🇩"},
	"374": {"AM", "🇦🇲"},
	"375": {"BY", "🇧🇾"},
	"376": {"AD", "🇦🇩"},
	"377": {"MC", "🇲🇨"},
	"378": {"SM", "🇸🇲"},
	"380": {"UA", "🇺🇦"},
	"381": {"RS", "🇷🇸"},
	"382": {"ME", "🇲🇪"},
	"383": {"XK", "🇽🇰"},
	"385": {"HR", "🇭🇷"},
	"386": {"SI", "🇸🇮"},
	"387": {"BA", "🇧🇦"},
	"389": {"MK", "🇲🇰"},
	"420": {"CZ", "🇨🇿"},
	"421": {"SK", "🇸🇰"},
	"423": {"LI", "🇱🇮"},
	// Americas
	"1":   {"US", "🇺🇸"},
	"51":  {"PE", "🇵🇪"},
	"52":  {"MX", "🇲🇽"},
	"53":  {"CU", "🇨🇺"},
	"54":  {"AR", "🇦🇷"},
	"55":  {"BR", "🇧🇷"},
	"56":  {"CL", "🇨🇱"},
	"57":  {"CO", "🇨🇴"},
	"58":  {"VE", "🇻🇪"},
	"501": {"BZ", "🇧🇿"},
	"502": {"GT", "🇬🇹"},
	"503": {"SV", "🇸🇻"},
	"504": {"HN", "🇭🇳"},
	"505": {"NI", "🇳🇮"},
	"506": {"CR", "🇨🇷"},
	"507": {"PA", "🇵🇦"},
	"509": {"HT", "🇭🇹"},
	"591": {"BO", "🇧🇴"},
	"592": {"GY", "🇬🇾"},
	"593": {"EC", "🇪🇨"},
	"595": {"PY", "🇵🇾"},
	"597": {"SR", "🇸🇷"},
	"598": {"UY", "🇺🇾"},
}

func DetectCountry(number string) (string, string) {
	n := strings.ReplaceAll(number, "+", "")
	n = strings.ReplaceAll(n, " ", "")
	n = strings.ReplaceAll(n, "-", "")

	var keys []string
	for k := range CountryMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, k := range keys {
		if strings.HasPrefix(n, k) {
			info := CountryMap[k]
			return info.Code, info.Flag
		}
	}
	return "UN", "🌍"
}

func GetServiceAnimation(service string) string {
	s := strings.ToLower(service)
	switch {
	case containsAny(s, "whatsapp", "ws", "wa", "واتساب", "واتس"):
		return "📞"
	case containsAny(s, "facebook", "fb", "فيسبوك"):
		return "💬"
	case containsAny(s, "telegram", "tg", "تيليجرام", "تلي"):
		return "👉"
	case containsAny(s, "instagram", "ig", "انستقرام", "انستا"):
		return "📷"
	case containsAny(s, "twitter", "x", "تويتر"):
		return "🐦"
	case containsAny(s, "tiktok", "تيك توك", "تيك"):
		return "🎵"
	case containsAny(s, "snapchat", "snap", "سناب"):
		return "👻"
	case containsAny(s, "google", "gmail", "جوجل", "جميل"):
		return "🔍"
	}
	return ""
}

func DetectServiceFromMessage(message string) string {
	m := strings.ToLower(message)
	patterns := map[string][]string{
		"WHATSAPP":  {"whatsapp", "wa", "واتساب", "واتس"},
		"FACEBOOK":  {"facebook", "fb", "فيسبوك"},
		"TELEGRAM":  {"telegram", "tg", "تيليجرام", "تلي"},
		"INSTAGRAM": {"instagram", "ig", "انستقرام", "انستا"},
		"TWITTER":   {"twitter", "x.com", "تويتر"},
		"TIKTOK":    {"tiktok", "تيك توك", "تيك"},
		"SNAPCHAT":  {"snapchat", "سناب"},
		"GOOGLE":    {"google", "gmail", "جوجل", "جميل"},
	}

	for service, keywords := range patterns {
		for _, kw := range keywords {
			if strings.Contains(m, kw) {
				return service
			}
		}
	}
	return "UNKNOWN"
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func MaskPhoneNumber(number string) string {
	n := strings.TrimSpace(number)
	if len(n) < 8 {
		return n
	}
	return n[:5] + "SHARK" + n[len(n)-3:]
}

var otpCodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(\d{2,4})[\s\-:]+(\d{2,4})`),
	regexp.MustCompile(`\b\d{4,8}\b`),
	regexp.MustCompile(`\d+`),
}

func ExtractOTPCode(message string) string {
	if message == "" {
		return "N/A"
	}
	t := strings.ReplaceAll(message, "\n", " ")

	for _, p := range otpCodePatterns {
		matches := p.FindStringSubmatch(t)
		if len(matches) > 1 {
			// Handle segmented codes
			combo := matches[1] + matches[2]
			if len(combo) >= 4 && len(combo) <= 8 {
				return combo
			}
		} else if len(matches) == 1 {
			code := matches[0]
			if len(code) >= 4 && len(code) <= 8 {
				return code
			}
			return code
		}
	}
	return "N/A"
}
