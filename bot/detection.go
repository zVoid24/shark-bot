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
	// Africa
	"20":  {"EG", "<tg-emoji emoji-id='5222161185138292290'>🇪🇬</tg-emoji>"},
	"27":  {"ZA", "<tg-emoji emoji-id='5224696216570309138'>🇿🇦</tg-emoji>"},
	"211": {"SS", "<tg-emoji emoji-id='5224618146949773268'>🇸🇸</tg-emoji>"},
	"212": {"MA", "<tg-emoji emoji-id='5224530035695693965'>🇲🇦</tg-emoji>"},
	"213": {"DZ", "<tg-emoji emoji-id='5224260376174015500'>🇩🇿</tg-emoji>"},
	"216": {"TN", "<tg-emoji emoji-id='5221991375016310330'>🇹🇳</tg-emoji>"},
	"218": {"LY", "<tg-emoji emoji-id='5222194286451242896'>🇱🇾</tg-emoji>"},
	"220": {"GM", "<tg-emoji emoji-id='5221949872747330159'>🇬🇲</tg-emoji>"},
	"221": {"SN", "<tg-emoji emoji-id='5224358988623130949'>🇸🇳</tg-emoji>"},
	"222": {"MR", "<tg-emoji emoji-id='5224269666188274723'>🇲🇷</tg-emoji>"},
	"223": {"ML", "<tg-emoji emoji-id='5224322352552096671'>🇲🇱</tg-emoji>"},
	"224": {"GN", "<tg-emoji emoji-id='5222337588035073000'>🇬🇳</tg-emoji>"},
	"225": {"CI", "🇨🇮"},
	"226": {"BF", "🇧🇫"},
	"227": {"NE", "<tg-emoji emoji-id='5222099049846420864'>🇳🇪</tg-emoji>"},
	"228": {"TG", "<tg-emoji emoji-id='5222408051268532030'>🇹🇬</tg-emoji>"},
	"229": {"BJ", "<tg-emoji emoji-id='5224515905253291409'>🇧🇯</tg-emoji>"},
	"230": {"MU", "<tg-emoji emoji-id='5224393700548814960'>🇲🇺</tg-emoji>"},
	"231": {"LR", "<tg-emoji emoji-id='5224420995065983217'>🇱🇷</tg-emoji>"},
	"232": {"SL", "<tg-emoji emoji-id='5224420995065983217'>🇸🇱</tg-emoji>"},
	"233": {"GH", "<tg-emoji emoji-id='5224511339703056124'>🇬🇭</tg-emoji>"},
	"234": {"NG", "<tg-emoji emoji-id='5224723614166691638'>🇳🇬</tg-emoji>"},
	"235": {"TD", "<tg-emoji emoji-id='5222060468155204001'>🇹🇩</tg-emoji>"},
	"236": {"CF", "🇨🇫"},
	"237": {"CM", "<tg-emoji emoji-id='5222234560359577687'>🇨🇲</tg-emoji>"},
	"238": {"CV", "<tg-emoji emoji-id='5224567367551428669'>🇨🇻</tg-emoji>"},
	"239": {"ST", "<tg-emoji emoji-id='5221953304426198315'>🇸🇹</tg-emoji>"},
	"240": {"GQ", "<tg-emoji emoji-id='5224455152940886669'>🇬🇶</tg-emoji>"},
	"241": {"GA", "<tg-emoji emoji-id='5222152195771742239'>🇬🇦</tg-emoji>"},
	"242": {"CG", "<tg-emoji emoji-id='5224490444687158452'>🇨🇬</tg-emoji>"},
	"243": {"CD", "<tg-emoji emoji-id='5224490444687158452'>🇨🇩</tg-emoji>"},
	"244": {"AO", "<tg-emoji emoji-id='5224379767674907895'>🇦🇴</tg-emoji>"},
	"245": {"GW", "<tg-emoji emoji-id='5224705704153066489'>🇬🇼</tg-emoji>"},
	"248": {"SC", "<tg-emoji emoji-id='5224467496676896871'>🇸🇨</tg-emoji>"},
	"249": {"SD", "<tg-emoji emoji-id='5224372990216514135'>🇸🇩</tg-emoji>"},
	"250": {"RW", "<tg-emoji emoji-id='5222449197055227754'>🇷🇼</tg-emoji>"},
	"251": {"ET", "<tg-emoji emoji-id='5224467805914542024'>🇪🇹</tg-emoji>"},
	"252": {"SO", "<tg-emoji emoji-id='5222370504664428325'>🇸🇴</tg-emoji>"},
	"253": {"DJ", "<tg-emoji emoji-id='5221991375016310330'>🇩🇯</tg-emoji>"},
	"254": {"KE", "<tg-emoji emoji-id='5222089648163009103'>🇰🇪</tg-emoji>"},
	"255": {"TZ", "<tg-emoji emoji-id='5224397364155923150'>🇹🇿</tg-emoji>"},
	"256": {"UG", "<tg-emoji emoji-id='5222464040462200940'>🇺🇬</tg-emoji>"},
	"257": {"BI", "🇧🇮"},
	"258": {"MZ", "<tg-emoji emoji-id='5222470388423864826'>🇲🇿</tg-emoji>"},
	"260": {"ZM", "<tg-emoji emoji-id='5224646626877911277'>🇿🇲</tg-emoji>"},
	"261": {"MG", "<tg-emoji emoji-id='5222042605386217334'>🇲🇬</tg-emoji>"},
	"262": {"RE", "🇷🇪"},
	"263": {"ZW", "<tg-emoji emoji-id='5222060442385397848'>🇿🇼</tg-emoji>"},
	"264": {"NA", "<tg-emoji emoji-id='5224690826386351746'>🇳🇦</tg-emoji>"},
	"265": {"MW", "<tg-emoji emoji-id='5222470435668505656'>🇲🇼</tg-emoji>"},
	"266": {"LS", "<tg-emoji emoji-id='5224660718665607511'>🇱🇸</tg-emoji>"},
	"267": {"BW", "<tg-emoji emoji-id='5224570532942329532'>🇧🇼</tg-emoji>"},
	"268": {"SZ", "🇸🇿"},
	"269": {"KM", "<tg-emoji emoji-id='5222398735484466247'>🇰🇲</tg-emoji>"},

	// Asia
	"60":  {"MY", "<tg-emoji emoji-id='5224312886444174057'>🇲🇾</tg-emoji>"},
	"62":  {"ID", "<tg-emoji emoji-id='5224405893960969756'>🇮🇩</tg-emoji>"},
	"63":  {"PH", "<tg-emoji emoji-id='5222065042295376892'>🇵🇭</tg-emoji>"},
	"64":  {"NZ", "<tg-emoji emoji-id='5224573595254009705'>🇳🇿</tg-emoji>"},
	"65":  {"SG", "<tg-emoji emoji-id='5224194023224257181'>🇸🇬</tg-emoji>"},
	"66":  {"TH", "<tg-emoji emoji-id='5224638530864556281'>🇹🇭</tg-emoji>"},
	"81":  {"JP", "<tg-emoji emoji-id='5222390089715299207'>🇯🇵</tg-emoji>"},
	"82":  {"KR", "<tg-emoji emoji-id='5222345550904439270'>🇰🇷</tg-emoji>"},
	"84":  {"VN", "<tg-emoji emoji-id='5222359651282071925'>🇻🇳</tg-emoji>"},
	"86":  {"CN", "<tg-emoji emoji-id='5224435456220868088'>🇨🇳</tg-emoji>"},
	"90":  {"TR", "<tg-emoji emoji-id='5224601903383457698'>🇹🇷</tg-emoji>"},
	"91":  {"IN", "<tg-emoji emoji-id='5222300011366200403'>🇮🇳</tg-emoji>"},
	"92":  {"PK", "<tg-emoji emoji-id='5224637061985742245'>🇵🇰</tg-emoji>"},
	"93":  {"AF", "<tg-emoji emoji-id='5222096009009575868'>🇦🇫</tg-emoji>"},
	"94":  {"LK", "<tg-emoji emoji-id='5224277294050192388'>🇱🇰</tg-emoji>"},
	"95":  {"MM", "<tg-emoji emoji-id='5224393700548814960'>🇲🇲</tg-emoji>"},
	"98":  {"IR", "<tg-emoji emoji-id='5224374154152653367'>🇮🇷</tg-emoji>"},
	"850": {"KP", "🇰🇵"},
	"852": {"HK", "🇭🇰"},
	"853": {"MO", "🇲🇴"},
	"855": {"KH", "🇰🇭"},
	"856": {"LA", "🇱🇦"},
	"960": {"MV", "🇲🇻"},
	"961": {"LB", "🇱🇧"},
	"962": {"JO", "<tg-emoji emoji-id='5222229234600130045'>🇯🇴</tg-emoji>"},
	"963": {"SY", "<tg-emoji emoji-id='5224601903383457698'>🇸🇾</tg-emoji>"},
	"964": {"IQ", "<tg-emoji emoji-id='5221980268230882832'>🇮🇶</tg-emoji>"},
	"965": {"KW", "<tg-emoji emoji-id='5222225596762830469'>🇰🇼</tg-emoji>"},
	"966": {"SA", "<tg-emoji emoji-id='5224698145010624573'>🇸🇦</tg-emoji>"},
	"967": {"YE", "<tg-emoji emoji-id='5222300655611294950'>🇾🇪</tg-emoji>"},
	"968": {"OM", "<tg-emoji emoji-id='5222396686785066306'>🇴🇲</tg-emoji>"},
	"970": {"PS", "<tg-emoji emoji-id='5222041677673282461'>🇵🇸</tg-emoji>"},
	"971": {"AE", "<tg-emoji emoji-id='5224565851427976312'>🇦🇪</tg-emoji>"},
	"972": {"IL", "<tg-emoji emoji-id='5224720599099648709'>🇮🇱</tg-emoji>"},
	"973": {"BH", "🇧🇭"},
	"974": {"QA", "🇶🇦"},
	"975": {"BT", "<tg-emoji emoji-id='5222444378101925267'>🇧🇹</tg-emoji>"},
	"976": {"MN", "<tg-emoji emoji-id='5224192257992701543'>🇲🇳</tg-emoji>"},
	"977": {"NP", "🇳🇵"},
	"992": {"TJ", "<tg-emoji emoji-id='5222217865821696536'>🇹🇯</tg-emoji>"},
	"993": {"TM", "<tg-emoji emoji-id='5224256935905208951'>🇹🇲</tg-emoji>"},
	"994": {"AZ", "<tg-emoji emoji-id='5224426544163728284'>🇦🇿</tg-emoji>"},
	"995": {"GE", "🇬🇪"},
	"996": {"KG", "🇰🇬"},
	"998": {"UZ", "<tg-emoji emoji-id='5222404546575219535'>🇺🇿</tg-emoji>"},

	// Europe
	"30":  {"GR", "<tg-emoji emoji-id='5222463490706389920'>🇬🇷</tg-emoji>"},
	"31":  {"NL", "<tg-emoji emoji-id='5224516489368841614'>🇳🇱</tg-emoji>"},
	"32":  {"BE", "<tg-emoji emoji-id='5224520754271366661'>🇧🇪</tg-emoji>"},
	"33":  {"FR", "<tg-emoji emoji-id='5222029789203804982'>🇫🇷</tg-emoji>"},
	"34":  {"ES", "<tg-emoji emoji-id='5222024776976970940'>🇪🇸</tg-emoji>"},
	"36":  {"HU", "<tg-emoji emoji-id='5224691998912427164'>🇭🇺</tg-emoji>"},
	"39":  {"IT", "<tg-emoji emoji-id='5222460101977190141'>🇮🇹</tg-emoji>"},
	"40":  {"RO", "<tg-emoji emoji-id='5222273794885826118'>🇷🇴</tg-emoji>"},
	"41":  {"CH", "<tg-emoji emoji-id='5224707263226194753'>🇨🇭</tg-emoji>"},
	"43":  {"AT", "🇦🇹"},
	"44":  {"GB", "<tg-emoji emoji-id='5224518800061245598'>🇬🇧</tg-emoji>"},
	"45":  {"DK", "<tg-emoji emoji-id='5224245902134226386'>🇩🇰</tg-emoji>"},
	"46":  {"SE", "<tg-emoji emoji-id='5222201098269373561'>🇸🇪</tg-emoji>"},
	"47":  {"NO", "<tg-emoji emoji-id='5224465228934163949'>🇳🇴</tg-emoji>"},
	"48":  {"PL", "<tg-emoji emoji-id='5224670399521892983'>🇵🇱</tg-emoji>"},
	"49":  {"DE", "<tg-emoji emoji-id='5222165617544542414'>🇩🇪</tg-emoji>"},
	"350": {"GI", "🇬🇮"},
	"351": {"PT", "<tg-emoji emoji-id='5224404094369672274'>🇵🇹</tg-emoji>"},
	"352": {"LU", "🇱🇺"},
	"353": {"IE", "<tg-emoji emoji-id='5222233374948602940'>🇮🇪</tg-emoji>"},
	"354": {"IS", "<tg-emoji emoji-id='5222063229819172521'>🇮🇸</tg-emoji>"},
	"355": {"AL", "<tg-emoji emoji-id='5224312057515486246'>🇦🇱</tg-emoji>"},
	"356": {"MT", "🇲🇹"},
	"357": {"CY", "🇨🇾"},
	"358": {"FI", "<tg-emoji emoji-id='5224282903277482188'>🇫🇮</tg-emoji>"},
	"359": {"BG", "🇧🇬"},
	"370": {"LT", "🇱🇹"},
	"371": {"LV", "🇱🇻"},
	"372": {"EE", "🇪🇪"},
	"373": {"MD", "🇲🇩"},
	"374": {"AM", "<tg-emoji emoji-id='5224369957969603463'>🇦🇲</tg-emoji>"},
	"375": {"BY", "<tg-emoji emoji-id='5280820319458707404'>🇧🇾</tg-emoji>"},
	"376": {"AD", "<tg-emoji emoji-id='5221987861733061751'>🇦🇩</tg-emoji>"},
	"377": {"MC", "<tg-emoji emoji-id='5221937224068640464'>🇲🇨</tg-emoji>"},
	"378": {"SM", "🇸🇲"},
	"380": {"UA", "<tg-emoji emoji-id='5222250679371839695'>🇺🇦</tg-emoji>"},
	"381": {"RS", "<tg-emoji emoji-id='5222145396838512729'>🇷🇸</tg-emoji>"},
	"382": {"ME", "<tg-emoji emoji-id='5224463399278096980'>🇲🇪</tg-emoji>"},
	"383": {"XK", "🇽🇰"},
	"385": {"HR", "🇭🇷"},
	"386": {"SI", "🇸🇮"},
	"387": {"BA", "🇧🇦"},
	"389": {"MK", "🇲🇰"},
	"420": {"CZ", "🇨🇿"},
	"421": {"SK", "<tg-emoji emoji-id='5222401879400528047'>🇸🇰</tg-emoji>"},
	"423": {"LI", "🇱🇮"},

	// Americas
	"1":   {"US", "<tg-emoji emoji-id='5224321781321442532'>🇺🇸</tg-emoji>"},
	"51":  {"PE", "<tg-emoji emoji-id='5224482026551258766'>🇵🇪</tg-emoji>"},
	"52":  {"MX", "🇲🇽"},
	"53":  {"CU", "🇨🇺"},
	"54":  {"AR", "🇦🇷"},
	"55":  {"BR", "<tg-emoji emoji-id='5224688610183228070'>🇧🇷</tg-emoji>"},
	"56":  {"CL", "🇨🇱"},
	"57":  {"CO", "🇨🇴"},
	"58":  {"VE", "<tg-emoji emoji-id='5434009132753499322'>🇻🇪</tg-emoji>"},
	"501": {"BZ", "🇧🇿"},
	"502": {"GT", "<tg-emoji emoji-id='5222128302868672826'>🇬🇹</tg-emoji>"},
	"503": {"SV", "🇸🇻"},
	"504": {"HN", "🇭🇳"},
	"505": {"NI", "🇳🇮"},
	"506": {"CR", "🇨🇷"},
	"507": {"PA", "<tg-emoji emoji-id='5222111719999945107'>🇵🇦</tg-emoji>"},
	"509": {"HT", "<tg-emoji emoji-id='5224683146984831315'>🇭🇹</tg-emoji>"},
	"591": {"BO", "🇧🇴"},
	"592": {"GY", "🇬🇾"},
	"593": {"EC", "<tg-emoji emoji-id='5224191188545840926'>🇪🇨</tg-emoji>"},
	"595": {"PY", "<tg-emoji emoji-id='5222152565138929235'>🇵🇾</tg-emoji>"},
	"597": {"SR", "🇸🇷"},
	"598": {"UY", "<tg-emoji emoji-id='5222466849370813232'>🇺🇾</tg-emoji>"},

	// Special
	"un": {"UN", "<tg-emoji emoji-id='5451772687993031127'>🇺🇳</tg-emoji>"},
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
	return "UN", CountryMap["un"].Flag
}

func GetServiceAnimation(service string) string {
	s := strings.ToLower(service)
	switch {
	case containsAny(s, "whatsapp", "ws", "wa", "واتساب", "واتس"):
		return "<tg-emoji emoji-id='5334998226636390258'>📞</tg-emoji>"
	case containsAny(s, "facebook", "fb", "فيسبوك"):
		return "<tg-emoji emoji-id='5323261730283863478'>💬</tg-emoji>"
	case containsAny(s, "telegram", "tg", "تيليجرام", "تلي"):
		return "<tg-emoji emoji-id='5330237710655306682'>👉</tg-emoji>"
	case containsAny(s, "instagram", "ig", "انستقرام", "انستا"):
		return "<tg-emoji emoji-id='5319160079465857105'>📷</tg-emoji>"
	case containsAny(s, "twitter", "x", "تويتر"):
		return "<tg-emoji emoji-id='5224499567197700690'>🐦</tg-emoji>"
	case containsAny(s, "tiktok", "تيك توك", "تيك"):
		return "<tg-emoji emoji-id='5224601903383457698'>🎵</tg-emoji>"
	case containsAny(s, "snapchat", "snap", "سناب"):
		return "<tg-emoji emoji-id='5222345550904439270'>👻</tg-emoji>"
	case containsAny(s, "google", "gmail", "جوجল", "جميل"):
		return "<tg-emoji emoji-id='5222029789203804982'>🔍</tg-emoji>"
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
		"GOOGLE":    {"google", "gmail", "جوجল", "جميل"},
		"AMAZON":    {"amazon", "امازون"},
		"NETFLIX":   {"netflix", "نتفلكس"},
		"PAYPAL":    {"paypal", "باي بال"},
		"APPLE":     {"apple", "icloud", "ابل"},
		"MICROSOFT": {"microsoft", "outlook", "مايكروسوفت"},
		"UBER":      {"uber", "اوبر"},
		"BINANCE":   {"binance"},
		"COINBASE":  {"coinbase"},
		"SPOTIFY":   {"spotify", "سبوتيفاي"},
		"YOUTUBE":   {"youtube", "يوتيوب"},
		"LINKEDIN":  {"linkedin", "لينكد"},
		"DISCORD":   {"discord", "ديسكورد"},
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
