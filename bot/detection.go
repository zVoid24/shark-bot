package bot

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type CountryInfo struct {
	Code string
	Flag string
	Name string
}

var CountryMap = map[string]CountryInfo{
	// ================= AFRICA =================
	"20":  {"EG", "<tg-emoji emoji-id='5222161185138292290'>🇪🇬</tg-emoji>", "Egypt"},
	"27":  {"ZA", "<tg-emoji emoji-id='5224696216570309138'>🇿🇦</tg-emoji>", "South Africa"},
	"211": {"SS", "<tg-emoji emoji-id='5224618146949773268'>🇸🇸</tg-emoji>", "South Sudan"},
	"212": {"MA", "<tg-emoji emoji-id='5224530035695693965'>🇲🇦</tg-emoji>", "Morocco"},
	"213": {"DZ", "<tg-emoji emoji-id='5224260376174015500'>🇩🇿</tg-emoji>", "Algeria"},
	"216": {"TN", "<tg-emoji emoji-id='5221991375016310330'>🇹🇳</tg-emoji>", "Tunisia"},
	"218": {"LY", "<tg-emoji emoji-id='5222194286451242896'>🇱🇾</tg-emoji>", "Libya"},
	"220": {"GM", "<tg-emoji emoji-id='5221949872747330159'>🇬🇲</tg-emoji>", "Gambia"},
	"221": {"SN", "<tg-emoji emoji-id='5224358988623130949'>🇸🇳</tg-emoji>", "Senegal"},
	"222": {"MR", "<tg-emoji emoji-id='5224269666188274723'>🇲🇷</tg-emoji>", "Mauritania"},
	"223": {"ML", "<tg-emoji emoji-id='5224322352552096671'>🇲🇱</tg-emoji>", "Mali"},
	"224": {"GN", "<tg-emoji emoji-id='5222337588035073000'>🇬🇳</tg-emoji>", "Guinea"},
	"225": {"CI", "<tg-emoji emoji-id='5411283953984218884'>🇨🇮</tg-emoji>", "Ivory Coast"},
	"226": {"BF", "🇧🇫", "Burkina Faso"},
	"227": {"NE", "<tg-emoji emoji-id='5222099049846420864'>🇳🇪</tg-emoji>", "Niger"},
	"228": {"TG", "<tg-emoji emoji-id='5222408051268532030'>🇹🇬</tg-emoji>", "Togo"},
	"229": {"BJ", "<tg-emoji emoji-id='5224515905253291409'>🇧🇯</tg-emoji>", "Benin"},
	"230": {"MU", "<tg-emoji emoji-id='5224393700548814960'>🇲🇺</tg-emoji>", "Mauritius"},
	"231": {"LR", "<tg-emoji emoji-id='5224420995065983217'>🇱🇷</tg-emoji>", "Liberia"},
	"232": {"SL", "<tg-emoji emoji-id='5224420995065983217'>🇸🇱</tg-emoji>", "Sierra Leone"},
	"233": {"GH", "<tg-emoji emoji-id='5224511339703056124'>🇬🇭</tg-emoji>", "Ghana"},
	"234": {"NG", "<tg-emoji emoji-id='5224723614166691638'>🇳🇬</tg-emoji>", "Nigeria"},
	"235": {"TD", "<tg-emoji emoji-id='5222060468155204001'>🇹🇩</tg-emoji>", "Chad"},
	"236": {"CF", "<tg-emoji emoji-id='5222060468155204001'>🇨🇫</tg-emoji>", "Central African Republic"},
	"237": {"CM", "<tg-emoji emoji-id='5222234560359577687'>🇨🇲</tg-emoji>", "Cameroon"},
	"238": {"CV", "<tg-emoji emoji-id='5224567367551428669'>🇨🇻</tg-emoji>", "Cape Verde"},
	"239": {"ST", "<tg-emoji emoji-id='5221953304426198315'>🇸🇹</tg-emoji>", "Sao Tome"},
	"240": {"GQ", "<tg-emoji emoji-id='5224455152940886669'>🇬🇶</tg-emoji>", "Equatorial Guinea"},
	"241": {"GA", "<tg-emoji emoji-id='5222152195771742239'>🇬🇦</tg-emoji>", "Gabon"},
	"242": {"CG", "<tg-emoji emoji-id='5224490444687158452'>🇨🇬</tg-emoji>", "Congo"},
	"243": {"CD", "<tg-emoji emoji-id='5224490444687158452'>🇨🇩</tg-emoji>", "DR Congo"},
	"244": {"AO", "<tg-emoji emoji-id='5224379767674907895'>🇦🇴</tg-emoji>", "Angola"},
	"245": {"GW", "<tg-emoji emoji-id='5224705704153066489'>🇬🇼</tg-emoji>", "Guinea-Bissau"},
	"248": {"SC", "<tg-emoji emoji-id='5224467496676896871'>🇸🇨</tg-emoji>", "Seychelles"},
	"249": {"SD", "<tg-emoji emoji-id='5224372990216514135'>🇸🇩</tg-emoji>", "Sudan"},
	"250": {"RW", "<tg-emoji emoji-id='5222449197055227754'>🇷🇼</tg-emoji>", "Rwanda"},
	"251": {"ET", "<tg-emoji emoji-id='5224467805914542024'>🇪🇹</tg-emoji>", "Ethiopia"},
	"252": {"SO", "<tg-emoji emoji-id='5222370504664428325'>🇸🇴</tg-emoji>", "Somalia"},
	"253": {"DJ", "<tg-emoji emoji-id='5221991375016310330'>🇩🇯</tg-emoji>", "Djibouti"},
	"254": {"KE", "<tg-emoji emoji-id='5222089648163009103'>🇰🇪</tg-emoji>", "Kenya"},
	"255": {"TZ", "<tg-emoji emoji-id='5224397364155923150'>🇹🇿</tg-emoji>", "Tanzania"},
	"256": {"UG", "<tg-emoji emoji-id='5222464040462200940'>🇺🇬</tg-emoji>", "Uganda"},
	"257": {"BI", "<tg-emoji emoji-id='5224490444687158452'>🇧🇮</tg-emoji>", "Burundi"},
	"258": {"MZ", "<tg-emoji emoji-id='5222470388423864826'>🇲🇿</tg-emoji>", "Mozambique"},
	"260": {"ZM", "<tg-emoji emoji-id='5224646626877911277'>🇿🇲</tg-emoji>", "Zambia"},
	"261": {"MG", "<tg-emoji emoji-id='5222042605386217334'>🇲🇬</tg-emoji>", "Madagascar"},
	"262": {"RE", "<tg-emoji emoji-id='5222042605386217334'>🇷🇪</tg-emoji>", "Reunion"},
	"263": {"ZW", "<tg-emoji emoji-id='5222060442385397848'>🇿🇼</tg-emoji>", "Zimbabwe"},
	"264": {"NA", "<tg-emoji emoji-id='5224690826386351746'>🇳🇦</tg-emoji>", "Namibia"},
	"265": {"MW", "<tg-emoji emoji-id='5222470435668505656'>🇲🇼</tg-emoji>", "Malawi"},
	"266": {"LS", "<tg-emoji emoji-id='5224660718665607511'>🇱🇸</tg-emoji>", "Lesotho"},
	"267": {"BW", "<tg-emoji emoji-id='5224570532942329532'>🇧🇼</tg-emoji>", "Botswana"},
	"268": {"SZ", "<tg-emoji emoji-id='5224269666188274723'>🇸🇿</tg-emoji>", "Eswatini"},
	"269": {"KM", "<tg-emoji emoji-id='5222398735484466247'>🇰🇲</tg-emoji>", "Comoros"},

	// ================= ASIA =================
	"60":  {"MY", "<tg-emoji emoji-id='5224312886444174057'>🇲🇾</tg-emoji>", "Malaysia"},
	"62":  {"ID", "<tg-emoji emoji-id='5224405893960969756'>🇮🇩</tg-emoji>", "Indonesia"},
	"63":  {"PH", "<tg-emoji emoji-id='5222065042295376892'>🇵🇭</tg-emoji>", "Philippines"},
	"64":  {"NZ", "<tg-emoji emoji-id='5224573595254009705'>🇳🇿</tg-emoji>", "New Zealand"},
	"65":  {"SG", "<tg-emoji emoji-id='5224194023224257181'>🇸🇬</tg-emoji>", "Singapore"},
	"66":  {"TH", "<tg-emoji emoji-id='5224638530864556281'>🇹🇭</tg-emoji>", "Thailand"},
	"81":  {"JP", "<tg-emoji emoji-id='5222390089715299207'>🇯🇵</tg-emoji>", "Japan"},
	"82":  {"KR", "<tg-emoji emoji-id='5222345550904439270'>🇰🇷</tg-emoji>", "South Korea"},
	"84":  {"VN", "<tg-emoji emoji-id='5222359651282071925'>🇻🇳</tg-emoji>", "Vietnam"},
	"86":  {"CN", "<tg-emoji emoji-id='5224435456220868088'>🇨🇳</tg-emoji>", "China"},
	"90":  {"TR", "<tg-emoji emoji-id='5224601903383457698'>🇹🇷</tg-emoji>", "Turkey"},
	"91":  {"IN", "<tg-emoji emoji-id='5222300011366200403'>🇮🇳</tg-emoji>", "India"},
	"92":  {"PK", "<tg-emoji emoji-id='5224637061985742245'>🇵🇰</tg-emoji>", "Pakistan"},
	"880": {"BD", "<tg-emoji emoji-id='5224615956516450654'>🇧🇩</tg-emoji>", "Bangladesh"},
	"886": {"TW", "<tg-emoji emoji-id='5224322352552096671'>🇹🇼</tg-emoji>", "Taiwan"},
	"93":  {"AF", "<tg-emoji emoji-id='5222096009009575868'>🇦🇫</tg-emoji>", "Afghanistan"},
	"94":  {"LK", "<tg-emoji emoji-id='5224277294050192388'>🇱🇰</tg-emoji>", "Sri Lanka"},
	"95":  {"MM", "<tg-emoji emoji-id='5224393700548814960'>🇲🇲</tg-emoji>", "Myanmar"},
	"98":  {"IR", "<tg-emoji emoji-id='5224374154152653367'>🇮🇷</tg-emoji>", "Iran"},
	"850": {"KP", "<tg-emoji emoji-id='5222345550904439270'>🇰🇵</tg-emoji>", "North Korea"},
	"852": {"HK", "<tg-emoji emoji-id='5224435456220868088'>🇭🇰</tg-emoji>", "Hong Kong"},
	"853": {"MO", "<tg-emoji emoji-id='5224435456220868088'>🇲🇴</tg-emoji>", "Macau"},
	"855": {"KH", "<tg-emoji emoji-id='5224638530864556281'>🇰🇭</tg-emoji>", "Cambodia"},
	"856": {"LA", "<tg-emoji emoji-id='5224638530864556281'>🇱🇦</tg-emoji>", "Laos"},
	"960": {"MV", "<tg-emoji emoji-id='5224393700548814960'>🇲🇻</tg-emoji>", "Maldives"},
	"961": {"LB", "🇱🇧", "Lebanon"},
	"962": {"JO", "<tg-emoji emoji-id='5222229234600130045'>🇯🇴</tg-emoji>", "Jordan"},
	"963": {"SY", "<tg-emoji emoji-id='5224601903383457698'>🇸🇾</tg-emoji>", "Syria"},
	"964": {"IQ", "<tg-emoji emoji-id='5221980268230882832'>🇮🇶</tg-emoji>", "Iraq"},
	"965": {"KW", "<tg-emoji emoji-id='5222225596762830469'>🇰🇼</tg-emoji>", "Kuwait"},
	"966": {"SA", "<tg-emoji emoji-id='5224698145010624573'>🇸🇦</tg-emoji>", "Saudi Arabia"},
	"967": {"YE", "<tg-emoji emoji-id='5222300655611294950'>🇾🇪</tg-emoji>", "Yemen"},
	"968": {"OM", "<tg-emoji emoji-id='5222396686785066306'>🇴🇲</tg-emoji>", "Oman"},
	"970": {"PS", "<tg-emoji emoji-id='5222041677673282461'>🇵🇸</tg-emoji>", "Palestine"},
	"971": {"AE", "<tg-emoji emoji-id='5224565851427976312'>🇦🇪</tg-emoji>", "UAE"},
	"972": {"IL", "<tg-emoji emoji-id='5224720599099648709'>🇮🇱</tg-emoji>", "Israel"},
	"973": {"BH", "<tg-emoji emoji-id='5222225596762830469'>🇧🇭</tg-emoji>", "Bahrain"},
	"974": {"QA", "<tg-emoji emoji-id='5222225596762830469'>🇶🇦</tg-emoji>", "Qatar"},
	"975": {"BT", "<tg-emoji emoji-id='5222444378101925267'>🇧🇹</tg-emoji>", "Bhutan"},
	"976": {"MN", "<tg-emoji emoji-id='5224192257992701543'>🇲🇳</tg-emoji>", "Mongolia"},
	"977": {"NP", "<tg-emoji emoji-id='5222444378101925267'>🇳🇵</tg-emoji>", "Nepal"},
	"992": {"TJ", "<tg-emoji emoji-id='5222217865821696536'>🇹🇯</tg-emoji>", "Tajikistan"},
	"993": {"TM", "<tg-emoji emoji-id='5224256935905208951'>🇹🇲</tg-emoji>", "Turkmenistan"},
	"994": {"AZ", "<tg-emoji emoji-id='5224426544163728284'>🇦🇿</tg-emoji>", "Azerbaijan"},
	"995": {"GE", "<tg-emoji emoji-id='5222152195771742239'>🇬🇪</tg-emoji>", "Georgia"},
	"996": {"KG", "<tg-emoji emoji-id='5224426544163728284'>🇰🇬</tg-emoji>", "Kyrgyzstan"},
	"998": {"UZ", "<tg-emoji emoji-id='5222404546575219535'>🇺🇿</tg-emoji>", "Uzbekistan"},

	// ================= EUROPE =================
	"30":  {"GR", "<tg-emoji emoji-id='5222463490706389920'>🇬🇷</tg-emoji>", "Greece"},
	"31":  {"NL", "<tg-emoji emoji-id='5224516489368841614'>🇳🇱</tg-emoji>", "Netherlands"},
	"32":  {"BE", "<tg-emoji emoji-id='5224520754271366661'>🇧🇪</tg-emoji>", "Belgium"},
	"33":  {"FR", "<tg-emoji emoji-id='5222029789203804982'>🇫🇷</tg-emoji>", "France"},
	"34":  {"ES", "<tg-emoji emoji-id='5222024776976970940'>🇪🇸</tg-emoji>", "Spain"},
	"36":  {"HU", "<tg-emoji emoji-id='5224691998912427164'>🇭🇺</tg-emoji>", "Hungary"},
	"39":  {"IT", "<tg-emoji emoji-id='5222460101977190141'>🇮🇹</tg-emoji>", "Italy"},
	"40":  {"RO", "<tg-emoji emoji-id='5222273794885826118'>🇷🇴</tg-emoji>", "Romania"},
	"41":  {"CH", "<tg-emoji emoji-id='5224707263226194753'>🇨🇭</tg-emoji>", "Switzerland"},
	"43":  {"AT", "<tg-emoji emoji-id='5224520754271366661'>🇦🇹</tg-emoji>", "Austria"},
	"44":  {"GB", "<tg-emoji emoji-id='5224518800061245598'>🇬🇧</tg-emoji>", "United Kingdom"},
	"45":  {"DK", "<tg-emoji emoji-id='5224245902134226386'>🇩🇰</tg-emoji>", "Denmark"},
	"46":  {"SE", "<tg-emoji emoji-id='5222201098269373561'>🇸🇪</tg-emoji>", "Sweden"},
	"47":  {"NO", "<tg-emoji emoji-id='5224465228934163949'>🇳🇴</tg-emoji>", "Norway"},
	"48":  {"PL", "<tg-emoji emoji-id='5224670399521892983'>🇵🇱</tg-emoji>", "Poland"},
	"49":  {"DE", "<tg-emoji emoji-id='5222165617544542414'>🇩🇪</tg-emoji>", "Germany"},
	"350": {"GI", "<tg-emoji emoji-id='5224518800061245598'>🇬🇮</tg-emoji>", "Gibraltar"},
	"351": {"PT", "<tg-emoji emoji-id='5224404094369672274'>🇵🇹</tg-emoji>", "Portugal"},
	"352": {"LU", "<tg-emoji emoji-id='5224499567197700690'>🇱🇺</tg-emoji>", "Luxembourg"},
	"353": {"IE", "<tg-emoji emoji-id='5222233374948602940'>🇮🇪</tg-emoji>", "Ireland"},
	"354": {"IS", "<tg-emoji emoji-id='5222063229819172521'>🇮🇸</tg-emoji>", "Iceland"},
	"355": {"AL", "<tg-emoji emoji-id='5224312057515486246'>🇦🇱</tg-emoji>", "Albania"},
	"356": {"MT", "<tg-emoji emoji-id='5224312057515486246'>🇲🇹</tg-emoji>", "Malta"},
	"357": {"CY", "<tg-emoji emoji-id='5224601903383457698'>🇨🇾</tg-emoji>", "Cyprus"},
	"358": {"FI", "<tg-emoji emoji-id='5224282903277482188'>🇫🇮</tg-emoji>", "Finland"},
	"359": {"BG", "<tg-emoji emoji-id='5224670399521892983'>🇧🇬</tg-emoji>", "Bulgaria"},
	"370": {"LT", "<tg-emoji emoji-id='5224245902134226386'>🇱🇹</tg-emoji>", "Lithuania"},
	"371": {"LV", "<tg-emoji emoji-id='5224245902134226386'>🇱🇻</tg-emoji>", "Latvia"},
	"372": {"EE", "<tg-emoji emoji-id='5224245902134226386'>🇪🇪</tg-emoji>", "Estonia"},
	"373": {"MD", "<tg-emoji emoji-id='5222273794885826118'>🇲🇩</tg-emoji>", "Moldova"},
	"374": {"AM", "<tg-emoji emoji-id='5224369957969603463'>🇦🇲</tg-emoji>", "Armenia"},
	"375": {"BY", "<tg-emoji emoji-id='5280820319458707404'>🇧🇾</tg-emoji>", "Belarus"},
	"376": {"AD", "<tg-emoji emoji-id='5221987861733061751'>🇦🇩</tg-emoji>", "Andorra"},
	"377": {"MC", "<tg-emoji emoji-id='5221937224068640464'>🇲🇨</tg-emoji>", "Monaco"},
	"378": {"SM", "<tg-emoji emoji-id='5224312057515486246'>🇸🇲</tg-emoji>", "San Marino"},
	"380": {"UA", "<tg-emoji emoji-id='5222250679371839695'>🇺🇦</tg-emoji>", "Ukraine"},
	"381": {"RS", "<tg-emoji emoji-id='5222145396838512729'>🇷🇸</tg-emoji>", "Serbia"},
	"382": {"ME", "<tg-emoji emoji-id='5224463399278096980'>🇲🇪</tg-emoji>", "Montenegro"},
	"383": {"XK", "<tg-emoji emoji-id='5222145396838512729'>🇽🇰</tg-emoji>", "Kosovo"},
	"385": {"HR", "<tg-emoji emoji-id='5224660718665607511'>🇭🇷</tg-emoji>", "Croatia"},
	"386": {"SI", "<tg-emoji emoji-id='5224660718665607511'>🇸🇮</tg-emoji>", "Slovenia"},
	"387": {"BA", "<tg-emoji emoji-id='5224660718665607511'>🇧🇦</tg-emoji>", "Bosnia"},
	"389": {"MK", "<tg-emoji emoji-id='5222470435668505656'>🇲🇰</tg-emoji>", "North Macedonia"},
	"420": {"CZ", "<tg-emoji emoji-id='5224499567197700690'>🇨🇿</tg-emoji>", "Czech Republic"},
	"421": {"SK", "<tg-emoji emoji-id='5222401879400528047'>🇸🇰</tg-emoji>", "Slovakia"},
	"423": {"LI", "<tg-emoji emoji-id='5224520754271366661'>🇱🇮</tg-emoji>", "Liechtenstein"},

	// ================= AMERICAS =================
	"1":   {"US", "<tg-emoji emoji-id='5224321781321442532'>🇺🇸</tg-emoji>", "United States"},
	"51":  {"PE", "<tg-emoji emoji-id='5224482026551258766'>🇵🇪</tg-emoji>", "Peru"},
	"52":  {"MX", "<tg-emoji emoji-id='5224482026551258766'>🇲🇽</tg-emoji>", "Mexico"},
	"53":  {"CU", "<tg-emoji emoji-id='5224482026551258766'>🇨🇺</tg-emoji>", "Cuba"},
	"54":  {"AR", "<tg-emoji emoji-id='5224482026551258766'>🇦🇷</tg-emoji>", "Argentina"},
	"55":  {"BR", "<tg-emoji emoji-id='5224688610183228070'>🇧🇷</tg-emoji>", "Brazil"},
	"56":  {"CL", "<tg-emoji emoji-id='5224482026551258766'>🇨🇱</tg-emoji>", "Chile"},
	"57":  {"CO", "<tg-emoji emoji-id='5224455152940886669'>🇨🇴</tg-emoji>", "Colombia"},
	"58":  {"VE", "<tg-emoji emoji-id='5434009132753499322'>🇻🇪</tg-emoji>", "Venezuela"},
	"501": {"BZ", "<tg-emoji emoji-id='5224482026551258766'>🇧🇿</tg-emoji>", "Belize"},
	"502": {"GT", "<tg-emoji emoji-id='5222128302868672826'>🇬??</tg-emoji>", "Guatemala"},
	"503": {"SV", "<tg-emoji emoji-id='5222128302868672826'>🇸🇻</tg-emoji>", "El Salvador"},
	"504": {"HN", "<tg-emoji emoji-id='5222229234600130045'>🇭🇳</tg-emoji>", "Honduras"},
	"505": {"NI", "<tg-emoji emoji-id='5222128302868672826'>🇳🇮</tg-emoji>", "Nicaragua"},
	"506": {"CR", "<tg-emoji emoji-id='5222128302868672826'>🇨🇷</tg-emoji>", "Costa Rica"},
	"507": {"PA", "<tg-emoji emoji-id='5222111719999945107'>🇵🇦</tg-emoji>", "Panama"},
	"509": {"HT", "<tg-emoji emoji-id='5224683146984831315'>🇭🇹</tg-emoji>", "Haiti"},
	"591": {"BO", "<tg-emoji emoji-id='5224482026551258766'>🇧🇴</tg-emoji>", "Bolivia"},
	"592": {"GY", "<tg-emoji emoji-id='5224570532942329532'>🇬🇾</tg-emoji>", "Guyana"},
	"593": {"EC", "<tg-emoji emoji-id='5224191188545840926'>🇪🇨</tg-emoji>", "Ecuador"},
	"595": {"PY", "<tg-emoji emoji-id='5222152565138929235'>🇵🇾</tg-emoji>", "Paraguay"},
	"597": {"SR", "<tg-emoji emoji-id='5224567367551428669'>🇸🇷</tg-emoji>", "Suriname"},
	"598": {"UY", "<tg-emoji emoji-id='5222466849370813232'>🇺🇾</tg-emoji>", "Uruguay"},

	// ================= OCEANIA =================
	"61":  {"AU", "<tg-emoji emoji-id='5224573595254009705'>🇦🇺</tg-emoji>", "Australia"},
	"670": {"TL", "<tg-emoji emoji-id='5224515905253291409'>🇹🇱</tg-emoji>", "Timor-Leste"},
	"673": {"BN", "<tg-emoji emoji-id='5224312886444174057'>🇧🇳</tg-emoji>", "Brunei"},
	"674": {"NR", "<tg-emoji emoji-id='5224573595254009705'>🇳🇷</tg-emoji>", "Nauru"},
	"675": {"PG", "<tg-emoji emoji-id='5224500164198149905'>🇵🇬</tg-emoji>", "Papua New Guinea"},
	"676": {"TO", "<tg-emoji emoji-id='5224573595254009705'>🇹🇴</tg-emoji>", "Tonga"},
	"677": {"SB", "<tg-emoji emoji-id='5222290588207954120'>🇸🇧</tg-emoji>", "Solomon Islands"},
	"678": {"VU", "<tg-emoji emoji-id='5222126748090512778'>🇻🇺</tg-emoji>", "Vanuatu"},
	"679": {"FJ", "<tg-emoji emoji-id='5221962676044838178'>🇫🇯</tg-emoji>", "Fiji"},
	"680": {"PW", "<tg-emoji emoji-id='5224573595254009705'>🇵🇼</tg-emoji>", "Palau"},
	"685": {"WS", "<tg-emoji emoji-id='5224660353593387686'>🇼🇸</tg-emoji>", "Samoa"},
	"686": {"KI", "<tg-emoji emoji-id='5224573595254009705'>🇰🇮</tg-emoji>", "Kiribati"},
	"687": {"NC", "<tg-emoji emoji-id='5224573595254009705'>🇳🇨</tg-emoji>", "New Caledonia"},
	"688": {"TV", "<tg-emoji emoji-id='5224573595254009705'>🇹🇻</tg-emoji>", "Tuvalu"},
	"689": {"PF", "<tg-emoji emoji-id='5224573595254009705'>🇵🇫</tg-emoji>", "French Polynesia"},
	"691": {"FM", "<tg-emoji emoji-id='5224573595254009705'>🇫🇲</tg-emoji>", "Micronesia"},
	"692": {"MH", "<tg-emoji emoji-id='5224573595254009705'>🇲🇭</tg-emoji>", "Marshall Islands"},

	// ================= SPECIAL FLAGS =================
	"scotland": {"SCT", "<tg-emoji emoji-id='5224580312582861623'>🏴</tg-emoji>", "Scotland"},
	"wales":    {"WLS", "<tg-emoji emoji-id='5224431333052264232'>🏴</tg-emoji>", "Wales"},
	"eu":       {"EU", "<tg-emoji emoji-id='5222108911091331711'>🇪🇺</tg-emoji>", "European Union"},
	"un":       {"UN", "<tg-emoji emoji-id='5451772687993031127'>🇺🇳</tg-emoji>", "United Nations"},
}

var PlatformEmojiMap = map[string]string{
	"whatsapp":  "5334998226636390258",
	"facebook":  "5323261730283863478",
	"telegram":  "5330237710655306682",
	"instagram": "5319160079465857105",
	"tiktok":    "5327982530702359565",
	"imo":       "5821117155072020062",
}

func GetPlatformEmoji(name string) string {
	id, ok := PlatformEmojiMap[strings.ToLower(name)]
	if !ok {
		return ""
	}
	return fmt.Sprintf("<tg-emoji emoji-id='%s'>💠</tg-emoji>", id)
}

func GetPlatformEmojiID(name string) string {
	return PlatformEmojiMap[strings.ToLower(name)]
}

func GetFlagByName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	// Strip HTML if present
	if strings.Contains(name, ">") {
		parts := strings.Split(name, ">")
		name = strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))
	}

	for _, info := range CountryMap {
		target := strings.ToLower(info.Name)
		if target == name || strings.HasSuffix(name, " "+target) || strings.HasSuffix(name, "	"+target) {
			return info.Flag
		}
	}
	return "🌍"
}

// GetFlagEmojiByName returns just the emoji character without HTML tags (for buttons)
func GetFlagEmojiByName(name string) string {
	html := GetFlagByName(name)
	if !strings.Contains(html, "<tg-emoji") {
		return html
	}
	// Extract emoji between > and <
	start := strings.Index(html, ">")
	end := strings.LastIndex(html, "<")
	if start != -1 && end != -1 && end > start+1 {
		return html[start+1 : end]
	}
	return html
}

// GetFlagEmojiIDByName returns just the emoji ID (for buttons)
func GetFlagEmojiIDByName(name string) string {
	html := GetFlagByName(name)
	if !strings.Contains(html, "emoji-id='") {
		return ""
	}
	start := strings.Index(html, "emoji-id='") + 10
	end := strings.Index(html[start:], "'")
	if start != -1 && end != -1 {
		return html[start : start+end]
	}
	return ""
}

func DetectCountry(number string) (string, string, string) {
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
			return info.Code, info.Name, info.Flag
		}
	}
	return "UN", "Unknown", "🌍"
}

func GetServiceAnimation(service string) string {
	s := strings.ToLower(service)
	switch {
	case containsAny(s, "whatsapp", "ws", "wa", "واتساب", "واتس"):
		return "📞"
	case containsAny(s, "facebook", "fb", "فيسبوك"):
		return "💬"
	case containsAny(s, "telegram", "tg", "تيليجرام", "تلي"):
		return "✈️"
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
	case containsAny(s, "imo", "ايمو"):
		return "📱"
	}
	return "📱"
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
