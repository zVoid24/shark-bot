import os
import sys
import time
import requests
from bs4 import BeautifulSoup
from datetime import datetime, timedelta
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import re
import phonenumbers
from phonenumbers import geocoder
import json
import sqlite3

def load_env():
    """Simple parser for .env file to load variables into os.environ"""
    env_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".env")
    if not os.path.exists(env_path):
        print("[⚠️] .env file not found. Using default values.")
        return

    with open(env_path, "r") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, value = line.split("=", 1)
                os.environ[key.strip()] = value.strip()

# Load environment variables
load_env()

# ================= CONFIGURATION =================
BOT_TOKEN = os.getenv("BOT_TOKEN", "8670441836:AAF3h0mbS3-bGwz41SQsZOrkUO2S3WN-hzU")
OWNER_IDS = os.getenv("OWNER_IDS", "").split(",")
OWNER_ID = OWNER_IDS[0] if OWNER_IDS and OWNER_IDS[0] else None

# Get target group IDs from env, use provided default if not found
TARGET_GROUP_IDS = os.getenv("TARGET_GROUP_IDS", "-1003422191454").split(",")
CHAT_ID = TARGET_GROUP_IDS[0] if TARGET_GROUP_IDS and TARGET_GROUP_IDS[0] else "-1003422191454"

LOGIN_URL = "http://185.2.83.39/ints/login"
SMS_URL = "http://185.2.83.39/ints/agent/SMSCDRReports"
DB_FILE = "otp_numbers.db"

# ================= ANIMATED SERVICE ICONS =================
def get_service_animation(service):
    s = service.lower() if service else ""
    
    if any(x in s for x in ["whatsapp", "ws", "wa", "واتساب", "واتس"]):
        return "📞"
    
    if any(x in s for x in ["facebook", "fb", "فيسبوك"]):
        return "💬"
    
    if any(x in s for x in ["telegram", "tg", "تيليجرام", "تلي"]):
        return "👉"
    
    if any(x in s for x in ["instagram", "ig", "انستقرام", "انستا"]):
        return "📷"
    
    if any(x in s for x in ["twitter", "x", "تويتر"]):
        return "🐦"
    
    if any(x in s for x in ["tiktok", "تيك توك", "تيك"]):
        return "🎵"
    
    if any(x in s for x in ["snapchat", "snap", "سناب"]):
        return "👻"
    
    if any(x in s for x in ["google", "gmail", "جوجل", "جميل"]):
        return "🔍"
    
    return ""

# ================= POWER ICONS =================
POWER_ICON = "⛩️"
POWER_EYE = "👁"

# ================= COMPLETE ANIMATED COUNTRY FLAGS (Lebanon Normal) =================
COUNTRY_MAP = {
    # Africa
    "20": ("EG", "🇪🇬"),  # Egypt
    "27": ("ZA", "🇿🇦"),  # South Africa
    "211": ("SS", "🇸🇸"),  # South Sudan
    "212": ("MA", "🇲🇦"),  # Morocco
    "213": ("DZ", "🇩🇿"),  # Algeria
    "216": ("TN", "🇹🇳"),  # Tunisia
    "218": ("LY", "🇱🇾"),  # Libya
    "220": ("GM", "🇬🇲"),  # Gambia
    "221": ("SN", "🇸🇳"),  # Senegal
    "222": ("MR", "🇲🇷"),  # Mauritania
    "223": ("ML", "🇲🇱"),  # Mali
    "224": ("GN", "🇬🇳"),  # Guinea
    "225": ("CI", "🇨🇮"),  # Ivory Coast - নরমাল ফ্ল্যাগ
    "226": ("BF", "🇧🇫"),  # Burkina Faso
    "227": ("NE", "🇳🇪"),  # Niger
    "228": ("TG", "🇹🇬"),  # Togo
    "229": ("BJ", "🇧🇯"),  # Benin
    "230": ("MU", "🇲🇺"),  # Mauritius
    "231": ("LR", "🇱🇷"),  # Liberia
    "232": ("SL", "🇸🇱"),  # Sierra Leone
    "233": ("GH", "🇬🇭"),  # Ghana
    "234": ("NG", "🇳🇬"),  # Nigeria
    "235": ("TD", "🇹🇩"),  # Chad
    "236": ("CF", "🇨🇫"),  # Central African Republic
    "237": ("CM", "🇨🇲"),  # Cameroon
    "238": ("CV", "🇨🇻"),  # Cape Verde
    "239": ("ST", "🇸🇹"),  # Sao Tome
    "240": ("GQ", "🇬🇶"),  # Equatorial Guinea
    "241": ("GA", "🇬🇦"),  # Gabon
    "242": ("CG", "🇨🇬"),  # Congo
    "243": ("CD", "🇨🇩"),  # DR Congo
    "244": ("AO", "🇦🇴"),  # Angola
    "245": ("GW", "🇬🇼"),  # Guinea-Bissau
    "248": ("SC", "🇸🇨"),  # Seychelles
    "249": ("SD", "🇸🇩"),  # Sudan
    "250": ("RW", "🇷🇼"),  # Rwanda
    "251": ("ET", "🇪🇹"),  # Ethiopia
    "252": ("SO", "🇸🇴"),  # Somalia
    "253": ("DJ", "🇩🇯"),  # Djibouti
    "254": ("KE", "🇰🇪"),  # Kenya
    "255": ("TZ", "🇹🇿"),  # Tanzania
    "256": ("UG", "🇺🇬"),  # Uganda
    "257": ("BI", "🇧🇮"),  # Burundi
    "258": ("MZ", "🇲🇿"),  # Mozambique
    "260": ("ZM", "🇿🇲"),  # Zambia
    "261": ("MG", "🇲🇬"),  # Madagascar
    "262": ("RE", "🇷🇪"),  # Reunion
    "263": ("ZW", "🇿🇼"),  # Zimbabwe
    "264": ("NA", "🇳🇦"),  # Namibia
    "265": ("MW", "🇲🇼"),  # Malawi
    "266": ("LS", "🇱🇸"),  # Lesotho
    "267": ("BW", "🇧🇼"),  # Botswana
    "268": ("SZ", "🇸🇿"),  # Eswatini
    "269": ("KM", "🇰🇲"),  # Comoros

    # Asia
    "60": ("MY", "🇲🇾"),  # Malaysia
    "62": ("ID", "🇮🇩"),  # Indonesia
    "63": ("PH", "🇵🇭"),  # Philippines
    "64": ("NZ", "🇳🇿"),  # New Zealand
    "65": ("SG", "🇸🇬"),  # Singapore
    "66": ("TH", "🇹🇭"),  # Thailand
    "81": ("JP", "🇯🇵"),  # Japan
    "82": ("KR", "🇰🇷"),  # South Korea
    "84": ("VN", "🇻🇳"),  # Vietnam
    "86": ("CN", "🇨🇳"),  # China
    "90": ("TR", "🇹🇷"),  # Turkey
    "91": ("IN", "🇮🇳"),  # India
    "92": ("PK", "🇵🇰"),  # Pakistan
    "93": ("AF", "🇦🇫"),  # Afghanistan
    "94": ("LK", "🇱🇰"),  # Sri Lanka
    "95": ("MM", "🇲🇲"),  # Myanmar
    "98": ("IR", "🇮🇷"),  # Iran
    "850": ("KP", "🇰🇵"),  # North Korea
    "852": ("HK", "🇭🇰"),  # Hong Kong
    "853": ("MO", "🇲🇴"),  # Macau
    "855": ("KH", "🇰🇭"),  # Cambodia
    "856": ("LA", "🇱🇦"),  # Laos
    "960": ("MV", "🇲🇻"),  # Maldives
    "961": ("LB", "🇱🇧"),  # Lebanon - NORMAL FLAG (as requested)
    "962": ("JO", "🇯🇴"),  # Jordan
    "963": ("SY", "🇸🇾"),  # Syria
    "964": ("IQ", "🇮🇶"),  # Iraq
    "965": ("KW", "🇰🇼"),  # Kuwait
    "966": ("SA", "🇸🇦"),  # Saudi Arabia
    "967": ("YE", "🇾🇪"),  # Yemen
    "968": ("OM", "🇴🇲"),  # Oman
    "970": ("PS", "🇵🇸"),  # Palestine
    "971": ("AE", "🇦🇪"),  # UAE
    "972": ("IL", "🇮🇱"),  # Israel
    "973": ("BH", "🇧🇭"),  # Bahrain
    "974": ("QA", "🇶🇦"),  # Qatar
    "975": ("BT", "🇧🇹"),  # Bhutan
    "976": ("MN", "🇲🇳"),  # Mongolia
    "977": ("NP", "🇳🇵"),  # Nepal
    "992": ("TJ", "🇹🇯"),  # Tajikistan
    "993": ("TM", "🇹🇲"),  # Turkmenistan
    "994": ("AZ", "🇦🇿"),  # Azerbaijan
    "995": ("GE", "🇬🇪"),  # Georgia
    "996": ("KG", "🇰🇬"),  # Kyrgyzstan
    "998": ("UZ", "🇺🇿"),  # Uzbekistan

    # Europe
    "30": ("GR", "🇬🇷"),  # Greece
    "31": ("NL", "🇳🇱"),  # Netherlands
    "32": ("BE", "🇧🇪"),  # Belgium
    "33": ("FR", "🇫🇷"),  # France
    "34": ("ES", "🇪🇸"),  # Spain
    "36": ("HU", "🇭🇺"),  # Hungary
    "39": ("IT", "🇮🇹"),  # Italy
    "40": ("RO", "🇷🇴"),  # Romania
    "41": ("CH", "🇨🇭"),  # Switzerland
    "43": ("AT", "🇦🇹"),  # Austria
    "44": ("GB", "🇬🇧"),  # United Kingdom
    "45": ("DK", "🇩🇰"),  # Denmark
    "46": ("SE", "🇸🇪"),  # Sweden
    "47": ("NO", "🇳🇴"),  # Norway
    "48": ("PL", "🇵🇱"),  # Poland
    "49": ("DE", "🇩🇪"),  # Germany
    "350": ("GI", "🇬🇮"),  # Gibraltar
    "351": ("PT", "🇵🇹"),  # Portugal
    "352": ("LU", "🇱🇺"),  # Luxembourg
    "353": ("IE", "🇮🇪"),  # Ireland
    "354": ("IS", "🇮🇸"),  # Iceland
    "355": ("AL", "🇦🇱"),  # Albania
    "356": ("MT", "🇲🇹"),  # Malta
    "357": ("CY", "🇨🇾"),  # Cyprus
    "358": ("FI", "🇫🇮"),  # Finland
    "359": ("BG", "🇧🇬"),  # Bulgaria
    "370": ("LT", "🇱🇹"),  # Lithuania
    "371": ("LV", "🇱🇻"),  # Latvia
    "372": ("EE", "🇪🇪"),  # Estonia
    "373": ("MD", "🇲🇩"),  # Moldova
    "374": ("AM", "🇦🇲"),  # Armenia
    "375": ("BY", "🇧🇾"),  # Belarus
    "376": ("AD", "🇦🇩"),  # Andorra
    "377": ("MC", "🇲🇨"),  # Monaco
    "378": ("SM", "🇸🇲"),  # San Marino
    "380": ("UA", "🇺🇦"),  # Ukraine
    "381": ("RS", "🇷🇸"),  # Serbia
    "382": ("ME", "🇲🇪"),  # Montenegro
    "383": ("XK", "🇽🇰"),  # Kosovo
    "385": ("HR", "🇭🇷"),  # Croatia
    "386": ("SI", "🇸🇮"),  # Slovenia
    "387": ("BA", "🇧🇦"),  # Bosnia
    "389": ("MK", "🇲🇰"),  # North Macedonia
    "420": ("CZ", "🇨🇿"),  # Czech Republic
    "421": ("SK", "🇸🇰"),  # Slovakia
    "423": ("LI", "🇱🇮"),  # Liechtenstein

    # Americas
    "1": ("US", "🇺🇸"),  # United States
    "51": ("PE", "🇵🇪"),  # Peru
    "52": ("MX", "🇲🇽"),  # Mexico
    "53": ("CU", "🇨🇺"),  # Cuba
    "54": ("AR", "🇦🇷"),  # Argentina
    "55": ("BR", "🇧🇷"),  # Brazil
    "56": ("CL", "🇨🇱"),  # Chile
    "57": ("CO", "🇨🇴"),  # Colombia
    "58": ("VE", "🇻🇪"),  # Venezuela
    "501": ("BZ", "🇧🇿"),  # Belize
    "502": ("GT", "🇬🇹"),  # Guatemala
    "503": ("SV", "🇸🇻"),  # El Salvador
    "504": ("HN", "🇭🇳"),  # Honduras
    "505": ("NI", "🇳🇮"),  # Nicaragua
    "506": ("CR", "🇨🇷"),  # Costa Rica
    "507": ("PA", "🇵🇦"),  # Panama
    "509": ("HT", "🇭🇹"),  # Haiti
    "591": ("BO", "🇧🇴"),  # Bolivia
    "592": ("GY", "🇬🇾"),  # Guyana
    "593": ("EC", "🇪🇨"),  # Ecuador
    "595": ("PY", "🇵🇾"),  # Paraguay
    "597": ("SR", "🇸🇷"),  # Suriname
    "598": ("UY", "🇺🇾"),  # Uruguay

    # Oceania
    "61": ("AU", "🇦🇺"),  # Australia
    "64": ("NZ", "🇳🇿"),  # New Zealand
    "670": ("TL", "🇹🇱"),  # Timor-Leste
    "673": ("BN", "🇧🇳"),  # Brunei
    "674": ("NR", "🇳🇷"),  # Nauru
    "675": ("PG", "🇵🇬"),  # Papua New Guinea
    "676": ("TO", "🇹🇴"),  # Tonga
    "677": ("SB", "🇸🇧"),  # Solomon Islands
    "678": ("VU", "🇻🇺"),  # Vanuatu
    "679": ("FJ", "🇫🇯"),  # Fiji
    "680": ("PW", "🇵🇼"),  # Palau
    "685": ("WS", "🇼🇸"),  # Samoa
    "686": ("KI", "🇰🇮"),  # Kiribati
    "687": ("NC", "🇳🇨"),  # New Caledonia
    "688": ("TV", "🇹🇻"),  # Tuvalu
    "689": ("PF", "🇵🇫"),  # French Polynesia
    "691": ("FM", "🇫🇲"),  # Micronesia
    "692": ("MH", "🇲🇭"),  # Marshall Islands

    # Special Flags
    "scotland": ("SCT", "🏴󠁧󠁢󠁳󠁣󠁴󠁿"),  # Scotland
    "wales": ("WLS", "🏴󠁧󠁢󠁷󠁬󠁳󠁿"),  # Wales
    "eu": ("EU", "🇪🇺"),  # European Union
    "un": ("UN", "🇺🇳"),  # United Nations
}

# ================= YOUR EXACT MESSAGE AND KEYBOARD FORMAT =================
def send_to_telegram(short_code, flag_emoji, service, service_icon, custom_number, otp_code):
    """Send formatted message to Telegram with YOUR EXACT design"""
    url = f"https://api.telegram.org/bot{BOT_TOKEN}/sendMessage"
    
    # YOUR EXACT MESSAGE FORMAT
    msg = f"""
{flag_emoji} #{short_code} {service_icon} <code>{custom_number}</code>

{POWER_ICON} 𝙿𝙾𝚆𝙴𝚁𝙴𝙳 𝙱𝚈  <a href="https://t.me/tamim_amv">𝙏𝘼𝙈𝙄𝙈</a> {POWER_EYE}
"""
    
    # YOUR EXACT KEYBOARD FORMAT (EXACTLY AS YOU PROVIDED, FIXED FOR STANDARD BOT API)
    keyboard = {
        "inline_keyboard": [
            [
                {
                    "text": otp_code,
                    "callback_data": f"copy_{otp_code}"
                }
            ],
            [
                {
                    "text": "Number Bot",
                    "url": "https://t.me/sharknumber2bot"
                },
                {
                    "text": "Method",
                    "url": "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz"
                }
            ]
        ]
    }
    
    payload = {
        "chat_id": OWNER_ID if OWNER_ID else CHAT_ID,
        "text": msg,
        "parse_mode": "HTML",
        "reply_markup": json.dumps(keyboard),
        "disable_web_page_preview": True
    }
    
    try:
        res = requests.post(url, data=payload, timeout=10)
        if res.status_code == 200:
            return True
        else:
            print(f"[❌] Telegram error: {res.status_code}")
            return False
    except requests.exceptions.RequestException as e:
        print(f"[❌] Telegram request error: {e}")
        return False

def notify_owner_startup():
    """Send a startup message to the owner for testing"""
    if not OWNER_ID:
        print("[⚠️] OWNER_ID not found in .env. Skipping startup notification.")
        return
    
    url = f"https://api.telegram.org/bot{BOT_TOKEN}/sendMessage"
    payload = {
        "chat_id": OWNER_ID,
        "text": f"🚀 <b>Shark Bot Scraper Started!</b>\n\n📅 Date: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n🤖 Bot: @sharknumber2bot\n\nMonitoring <b>{CHAT_ID}</b> for new SMS.",
        "parse_mode": "HTML"
    }
    
    try:
        res = requests.post(url, data=payload, timeout=10)
        if res.status_code == 200:
            print(f"[✅] Startup message sent to owner ID: {OWNER_ID}")
        else:
            print(f"[❌] Failed to notify owner: {res.status_code} - {res.text}")
    except Exception as e:
        print(f"[❌] Error sending startup message: {e}")

# ================= HELPER FUNCTIONS =================
def detect_country_from_number(number):
    """Detect country from phone number using COUNTRY_MAP"""
    n = str(number).replace("+", "").replace(" ", "").replace("-", "")
    for code in sorted(COUNTRY_MAP.keys(), key=len, reverse=True):
        if n.startswith(code):
            return COUNTRY_MAP[code]
    return ("UN", "🌍")

def extract_otp_code(message):
    """Extract OTP code from message"""
    if not message:
        return "N/A"
    
    t = str(message).replace("nn", " ").replace("n", " ")
    
    # Pattern: two groups of digits separated by space/hyphen
    split = re.findall(r'(\d{2,4})[\s\-:]+(\d{2,4})', t)
    for a, b in split:
        combo = a + b
        if 4 <= len(combo) <= 8:
            return combo
    
    # Pattern: standalone 4-8 digit number
    normal = re.findall(r'\b\d{4,8}\b', t)
    if normal:
        return normal[0]
    
    # Fallback: any digits
    anynum = re.findall(r'\d+', t)
    if anynum:
        return anynum[0]
    
    return "N/A"

def mask_phone_number(number):
    """Mask phone number with SHARK in middle"""
    n = str(number)
    if len(n) < 8:
        return n
    return n[:5] + "SHARK" + n[-3:]

class NumberTracker:
    """Database to track phone numbers - each number processed only once"""
    
    def __init__(self, db_file: str = DB_FILE):
        self.db_file = db_file
        self.init_database()
        self.session_numbers = set()  # Track in current session
        
    def init_database(self):
        """Initialize SQLite database"""
        conn = sqlite3.connect(self.db_file)
        cursor = conn.cursor()
        
        cursor.execute('''
        CREATE TABLE IF NOT EXISTS processed_numbers (
            phone_number TEXT PRIMARY KEY,
            first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
            last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
            otp_code TEXT,
            service_name TEXT,
            posted INTEGER DEFAULT 1
        )
        ''')
        
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_phone ON processed_numbers(phone_number)')
        conn.commit()
        conn.close()
        print(f"[✅] Database initialized: {self.db_file}")
    
    def is_number_seen(self, phone_number: str) -> bool:
        """Check if number has been processed before"""
        # Check session memory first (faster)
        if phone_number in self.session_numbers:
            return True
        
        # Check database
        conn = sqlite3.connect(self.db_file)
        cursor = conn.cursor()
        
        cursor.execute(
            "SELECT COUNT(*) FROM processed_numbers WHERE phone_number = ?",
            (phone_number,)
        )
        
        count = cursor.fetchone()[0]
        conn.close()
        
        return count > 0
    
    def add_number(self, phone_number: str, otp_code: str, service_name: str):
        """Add number to database and session memory"""
        # Add to session memory
        self.session_numbers.add(phone_number)
        
        # Add to database
        conn = sqlite3.connect(self.db_file)
        cursor = conn.cursor()
        
        try:
            cursor.execute('''
            INSERT INTO processed_numbers (phone_number, otp_code, service_name)
            VALUES (?, ?, ?)
            ''', (phone_number, otp_code, service_name))
        except sqlite3.IntegrityError:
            # Already exists, update last_seen
            cursor.execute('''
            UPDATE processed_numbers 
            SET last_seen = CURRENT_TIMESTAMP 
            WHERE phone_number = ?
            ''', (phone_number,))
        
        conn.commit()
        conn.close()
    
    def get_stats(self) -> dict:
        """Get statistics"""
        conn = sqlite3.connect(self.db_file)
        cursor = conn.cursor()
        
        cursor.execute("SELECT COUNT(*) FROM processed_numbers")
        total = cursor.fetchone()[0]
        
        cursor.execute("SELECT MIN(first_seen), MAX(last_seen) FROM processed_numbers")
        first_date, last_date = cursor.fetchone()
        
        conn.close()
        
        return {
            'total_numbers': total,
            'first_date': first_date,
            'last_date': last_date,
            'session_count': len(self.session_numbers)
        }

# Initialize tracker
tracker = NumberTracker()

def detect_service_from_message(message):
    """Detect service name from message content"""
    message_lower = message.lower()
    
    service_patterns = {
        'WHATSAPP': ['whatsapp', 'wa', 'واتساب', 'واتس'],
        'FACEBOOK': ['facebook', 'fb', 'فيسبوك'],
        'TELEGRAM': ['telegram', 'tg', 'تيليجرام', 'تلي'],
        'INSTAGRAM': ['instagram', 'ig', 'انستقرام', 'انستا'],
        'TWITTER': ['twitter', 'x.com', 'تويتر'],
        'TIKTOK': ['tiktok', 'تيك توك', 'تيك'],
        'SNAPCHAT': ['snapchat', 'سناب'],
        'GOOGLE': ['google', 'gmail', 'جوجل', 'جميل'],
        'AMAZON': ['amazon', 'امازون'],
        'NETFLIX': ['netflix', 'نتفلكس'],
        'PAYPAL': ['paypal', 'باي بال'],
        'APPLE': ['apple', 'icloud', 'ابل'],
        'MICROSOFT': ['microsoft', 'outlook', 'مايكروسوفت'],
        'UBER': ['uber', 'اوبر'],
        'BINANCE': ['binance'],
        'COINBASE': ['coinbase'],
        'SPOTIFY': ['spotify', 'سبوتيفاي'],
        'YOUTUBE': ['youtube', 'يوتيوب'],
        'LINKEDIN': ['linkedin', 'لينكد'],
        'DISCORD': ['discord', 'ديسكورد'],
    }
    
    for service, keywords in service_patterns.items():
        for keyword in keywords:
            if keyword in message_lower:
                return service
    
    # If no service detected, try to extract from message structure
    lines = message.split('\n')
    if lines:
        first_line = lines[0].strip()
        for service in service_patterns.keys():
            if service.lower() in first_line.lower():
                return service
    
    return "UNKNOWN"

def extract_sms(driver):
    """Extract SMS data from the website"""
    try:
        driver.get(SMS_URL)
        time.sleep(5)  # Wait for page to load
        
        # Click on the "Show Report" button to load data
        try:
            show_buttons = driver.find_elements(By.XPATH, "//input[@type='submit' and contains(@value, 'Show Report')]")
            for button in show_buttons:
                if button.get_attribute('data') == '2':
                    button.click()
                    time.sleep(3)
                    break
        except:
            print("[ℹ️] Show Report button not found or not clickable")
        
        # Get page source and parse with BeautifulSoup
        soup = BeautifulSoup(driver.page_source, 'html.parser')
        
        # Find the table with id='dt'
        table = soup.find('table', {'id': 'dt'})
        if not table:
            print("[⚠️] Table with id='dt' not found")
            return
        
        # Find all rows in tbody
        tbody = table.find('tbody')
        if not tbody:
            print("[⚠️] Table body not found")
            return
            
        rows = tbody.find_all('tr')
        
        if not rows or len(rows) == 0:
            print("[ℹ️] No data rows found in table")
            return
        
        new_numbers = 0
        duplicate_numbers = 0
        
        for row in rows:
            cols = row.find_all('td')
            if len(cols) < 9:  # Based on HTML, there should be 9 columns
                continue
                
            # Extract data from columns based on HTML structure
            date_time = cols[0].get_text(strip=True)
            number = cols[2].get_text(strip=True)
            service_from_column = cols[3].get_text(strip=True)
            message = cols[5].get_text(strip=True)
            
            # Skip empty rows or summary rows
            if not message or not number or number == "" or "Total" in date_time:
                continue
            
            # ===========================================
            # ✅ CRITICAL CHECK: Has this number been processed before?
            # ===========================================
            if tracker.is_number_seen(number):
                duplicate_numbers += 1
                continue  # SKIP - এই নাম্বার আগেই process হয়েছে
            # ===========================================
            
            # ✅ NEW NUMBER FOUND
            new_numbers += 1
            
            # Parse date/time
            try:
                # Use current local time with Cairo/Dhaka offset logic if needed, 
                # but better to just use current time for processing
                timestamp = datetime.now()
            except:
                timestamp = datetime.now()

            # Extract OTP code using improved function
            otp_code = extract_otp_code(message)

            # Detect country using COUNTRY_MAP
            short_code, flag_emoji = detect_country_from_number(number)
            
            # Mask phone number
            custom_number = mask_phone_number(number)

            # Detect service from message content
            detected_service = detect_service_from_message(message)
            
            # If service detection failed, use column value
            if detected_service == "UNKNOWN" and service_from_column and service_from_column != "0":
                detected_service = service_from_column.upper().replace('_', ' ').strip()
            
            # Get service animation
            service_icon = get_service_animation(detected_service)
            
            # Send to Telegram with YOUR EXACT DESIGN
            if send_to_telegram(short_code, flag_emoji, detected_service, service_icon, custom_number, otp_code):
                # ===========================================
                # ✅ MARK AS PROCESSED - একবারই mark করবে
                # ===========================================
                tracker.add_number(number, otp_code, detected_service)
                # ===========================================
                print(f"[✅] NEW: {short_code} | {custom_number} | OTP: {otp_code} | Service: {detected_service}")
            else:
                print(f"[❌] Failed to send: {custom_number}")

        # Print summary
        if new_numbers > 0:
            print(f"[📊] This check: {new_numbers} new numbers, {duplicate_numbers} duplicates skipped")
        elif duplicate_numbers > 0:
            print(f"[📊] All numbers duplicate - skipped {duplicate_numbers}")
        else:
            print(f"[{datetime.now().strftime('%H:%M:%S')}] No SMS found")
            
    except Exception as e:
        print(f"[❌] Failed to extract SMS: {e}")
        import traceback
        traceback.print_exc()

def wait_for_login(driver, timeout=180):
    """Wait for manual login"""
    print("[*] Waiting for manual login...")
    start = time.time()
    while time.time() - start < timeout:
        time.sleep(5)
        try:
            current_url = driver.current_url
            page_source = driver.page_source
            
            # Check if logged in
            if "login" not in current_url.lower() or "Logout" in page_source or "logout" in page_source:
                print("[✅] Login successful!")
                return True
        except Exception as e:
            print(f"[⚠️] Login check error: {e}")
    print("[❌] Login timeout!")
    return False

def launch_browser():
    """Launch Chrome browser with optimized settings"""
    print("[*] Launching Chrome browser...")

    chrome_options = Options()
    chrome_options.add_argument("--disable-extensions")
    chrome_options.add_argument("--disable-gpu")
    chrome_options.add_argument("--no-sandbox")
    chrome_options.add_argument("--start-maximized")
    chrome_options.add_argument("--disable-dev-shm-usage")
    chrome_options.add_argument("--log-level=3")
    chrome_options.add_argument("--disable-logging")
    chrome_options.add_argument("--disable-usb")
    chrome_options.add_experimental_option("excludeSwitches", ["enable-logging"])
    
    # For better performance
    chrome_options.add_argument("--disable-blink-features=AutomationControlled")
    chrome_options.add_experimental_option("useAutomationExtension", False)
    
    # Uncomment below line for headless mode
    # chrome_options.add_argument("--headless=new")

    try:
        service = Service(log_path='NUL')  # For Windows, use '/dev/null' on Linux/Mac
        driver = webdriver.Chrome(service=service, options=chrome_options)
        
        # Hide automation
        driver.execute_script("Object.defineProperty(navigator, 'webdriver', {get: () => undefined})")
        
        return driver
    except Exception as e:
        print(f"[❌] Browser launch failed: {e}")
        sys.exit(1)

def print_stats():
    """Print statistics"""
    stats = tracker.get_stats()
    print("\n" + "="*50)
    print("📊 NUMBER TRACKER STATISTICS")
    print("="*50)
    print(f"📁 Database: {DB_FILE}")
    print(f"📞 Total Unique Numbers: {stats['total_numbers']}")
    print(f"🔄 Current Session: {stats['session_count']} numbers")
    if stats['first_date']:
        print(f"📅 First Number: {stats['first_date']}")
    if stats['last_date']:
        print(f"📅 Last Number: {stats['last_date']}")
    print("="*50 + "\n")

def main():
    """Main function"""
    # Suppress unnecessary logs
    os.environ['TF_CPP_MIN_LOG_LEVEL'] = '3'
    os.environ['WDM_LOG_LEVEL'] = '0'
    
    # Print initial stats
    print_stats()
    
    driver = launch_browser()
    
    try:
        print(f"[*] Opening login page: {LOGIN_URL}")
        driver.get(LOGIN_URL)
        
        if not wait_for_login(driver):
            print("[❌] Login failed. Exiting...")
            return

        # Notify owner that bot has started
        notify_owner_startup()

        print("[✅] Login verified. Starting OTP monitoring...")
        print("[🔄] Checking for new SMS every 16 seconds...")
        print("[🚫] IMPORTANT: Each phone number processed ONLY ONCE!")
        print("[🚫] Duplicate numbers will be completely ignored.")
        print("[✅] Using YOUR EXACT message and keyboard format")
        
        check_count = 0
        while True:
            check_count += 1
            print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Check #{check_count}")
            
            try:
                extract_sms(driver)
                
                # Print stats every 10 checks
                if check_count % 10 == 0:
                    print_stats()
                    
            except Exception as e:
                print(f"[❌] Error during SMS extraction: {e}")
            
            # Wait for 16 seconds before next check
            print(f"[⏳] Waiting 16 seconds for next check...")
            time.sleep(16)
            
    except KeyboardInterrupt:
        print("\n[🛑] Stopped by user.")
        print_stats()
    except Exception as e:
        print(f"[❌] Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        print_stats()
    finally:
        print("[*] Closing browser...")
        driver.quit()
        print("[✅] Browser closed. Goodbye!")

if __name__ == "__main__":
    main()