import time
import re
import requests
import sqlite3
import hashlib
import threading
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.by import By
from webdriver_manager.chrome import ChromeDriverManager
from bs4 import BeautifulSoup
from selenium.common.exceptions import UnexpectedAlertPresentException, NoAlertPresentException

# --- Configuration ---
BOT_TOKEN = "7999875640:AAH8VDaq0cNAft4xMCQsyuMv-FdTiVUuNrE"
TARGET_CHAT_ID = "-1003422191454"
BASE_URL = "http://185.2.83.39/ints/agent/SMSCDRReports"
LOGIN_URL = "http://185.2.83.39/ints/login"

# --- Panel Login Credentials ---
PANEL_USERNAME = "rizvi1"
PANEL_PASSWORD = "rizvi20030"

# --- Country List (E.164 calling codes) ---
COUNTRY_CODES = {
    '1': ('USA/Canada', '🇺🇸'), '7': ('Russia/KZ', '🇷🇺'), '20': ('Egypt', '🇪🇬'),
    '27': ('South Africa', '🇿🇦'), '30': ('Greece', '🇬🇷'), '31': ('Netherlands', '🇳🇱'),
    '32': ('Belgium', '🇧🇪'), '33': ('France', '🇫🇷'), '34': ('Spain', '🇪🇸'),
    '36': ('Hungary', '🇭🇺'), '39': ('Italy', '🇮🇹'), '40': ('Romania', '🇷🇴'),
    '41': ('Switzerland', '🇨🇭'), '43': ('Austria', '🇦🇹'), '44': ('UK', '🇬🇧'),
    '45': ('Denmark', '🇩🇰'), '46': ('Sweden', '🇸🇪'), '47': ('Norway', '🇳🇴'),
    '48': ('Poland', '🇵🇱'), '49': ('Germany', '🇩🇪'), '51': ('Peru', '🇵🇪'),
    '52': ('Mexico', '🇲🇽'), '53': ('Cuba', '🇨🇺'), '54': ('Argentina', '🇦🇷'),
    '55': ('Brazil', '🇧🇷'), '56': ('Chile', '🇨🇱'), '57': ('Colombia', '🇨🇴'),
    '58': ('Venezuela', '🇻🇪'), '60': ('Malaysia', '🇲🇾'), '61': ('Australia', '🇦🇺'),
    '62': ('Indonesia', '🇮🇩'), '63': ('Philippines', '🇵🇭'), '64': ('New Zealand', '🇳🇿'),
    '65': ('Singapore', '🇸🇬'), '66': ('Thailand', '🇹🇭'), '81': ('Japan', '🇯🇵'),
    '82': ('South Korea', '🇰🇷'), '84': ('Vietnam', '🇻🇳'), '86': ('China', '🇨🇳'),
    '90': ('Turkey', '🇹🇷'), '91': ('India', '🇮🇳'), '92': ('Pakistan', '🇵🇰'),
    '93': ('Afghanistan', '🇦🇫'), '94': ('Sri Lanka', '🇱🇰'), '95': ('Myanmar', '🇲🇲'),
    '98': ('Iran', '🇮🇷'), '212': ('Morocco', '🇲🇦'), '213': ('Algeria', '🇩🇿'),
    '216': ('Tunisia', '🇹🇳'), '218': ('Libya', '🇱🇾'), '220': ('Gambia', '🇬🇲'),
    '221': ('Senegal', '🇸🇳'), '222': ('Mauritania', '🇲🇷'), '223': ('Mali', '🇲🇱'),
    '224': ('Guinea', '🇬🇳'), '225': ('Ivory Coast', '🇨🇮'), '226': ('Burkina Faso', '🇧🇫'),
    '227': ('Niger', '🇳🇪'), '228': ('Togo', '🇹🇬'), '229': ('Benin', '🇧🇯'),
    '230': ('Mauritius', '🇲🇺'), '231': ('Liberia', '🇱🇷'), '232': ('Sierra Leone', '🇸🇱'),
    '233': ('Ghana', '🇬🇭'), '234': ('Nigeria', '🇳🇬'), '235': ('Chad', '🇹🇩'),
    '236': ('Central African Rep', '🇨🇫'), '237': ('Cameroon', '🇨🇲'), '238': ('Cape Verde', '🇨🇻'),
    '239': ('São Tomé', '🇸🇹'), '240': ('Equatorial Guinea', '🇬🇶'), '241': ('Gabon', '🇬🇦'),
    '242': ('Congo', '🇨🇬'), '243': ('DR Congo', '🇨🇩'), '244': ('Angola', '🇦🇴'),
    '245': ('Guinea-Bissau', '🇬🇼'), '246': ('Diego Garcia', '🇮🇴'), '248': ('Seychelles', '🇸🇨'),
    '249': ('Sudan', '🇸🇩'), '250': ('Rwanda', '🇷🇼'), '251': ('Ethiopia', '🇪🇹'),
    '252': ('Somalia', '🇸🇴'), '253': ('Djibouti', '🇩🇯'), '254': ('Kenya', '🇰🇪'),
    '255': ('Tanzania', '🇹🇿'), '256': ('Uganda', '🇺🇬'), '257': ('Burundi', '🇧🇮'),
    '258': ('Mozambique', '🇲🇿'), '260': ('Zambia', '🇿🇲'), '261': ('Madagascar', '🇲🇬'),
    '262': ('Réunion', '🇷🇪'), '263': ('Zimbabwe', '🇿🇼'), '264': ('Namibia', '🇳🇦'),
    '265': ('Malawi', '🇲🇼'), '266': ('Lesotho', '🇱🇸'), '267': ('Botswana', '🇧🇼'),
    '268': ('Eswatini', '🇸🇿'), '269': ('Comoros', '🇰🇲'), '290': ('Saint Helena', '🇸🇭'),
    '291': ('Eritrea', '🇪🇷'), '297': ('Aruba', '🇦🇼'), '298': ('Faroe Islands', '🇫🇴'),
    '299': ('Greenland', '🇬🇱'), '350': ('Gibraltar', '🇬🇮'), '351': ('Portugal', '🇵🇹'),
    '352': ('Luxembourg', '🇱🇺'), '353': ('Ireland', '🇮🇪'), '354': ('Iceland', '🇮🇸'),
    '355': ('Albania', '🇦🇱'), '356': ('Malta', '🇲🇹'), '357': ('Cyprus', '🇨🇾'),
    '358': ('Finland', '🇫🇮'), '359': ('Bulgaria', '🇧🇬'), '370': ('Lithuania', '🇱🇹'),
    '371': ('Latvia', '🇱🇻'), '372': ('Estonia', '🇪🇪'), '373': ('Moldova', '🇲🇩'),
    '374': ('Armenia', '🇦🇲'), '375': ('Belarus', '🇧🇾'), '376': ('Andorra', '🇦🇩'),
    '377': ('Monaco', '🇲🇨'), '378': ('San Marino', '🇸🇲'), '379': ('Vatican', '🇻🇦'),
    '380': ('Ukraine', '🇺🇦'), '381': ('Serbia', '🇷🇸'), '382': ('Montenegro', '🇲🇪'),
    '383': ('Kosovo', '🇽🇰'), '385': ('Croatia', '🇭🇷'), '386': ('Slovenia', '🇸🇮'),
    '387': ('Bosnia', '🇧🇦'), '389': ('North Macedonia', '🇲🇰'), '420': ('Czechia', '🇨🇿'),
    '421': ('Slovakia', '🇸🇰'), '423': ('Liechtenstein', '🇱🇮'), '500': ('Falkland Islands', '🇫🇰'),
    '501': ('Belize', '🇧🇿'), '502': ('Guatemala', '🇬🇹'), '503': ('El Salvador', '🇸🇻'),
    '504': ('Honduras', '🇭🇳'), '505': ('Nicaragua', '🇳🇮'), '506': ('Costa Rica', '🇨🇷'),
    '507': ('Panama', '🇵🇦'), '508': ('Saint Pierre', '🇵🇲'), '509': ('Haiti', '🇭🇹'),
    '590': ('Guadeloupe', '🇬🇵'), '591': ('Bolivia', '🇧🇴'), '592': ('Guyana', '🇬🇾'),
    '593': ('Ecuador', '🇪🇨'), '594': ('French Guiana', '🇬🇫'), '595': ('Paraguay', '🇵🇾'),
    '596': ('Martinique', '🇲🇶'), '597': ('Suriname', '🇸🇷'), '598': ('Uruguay', '🇺🇾'),
    '599': ('Caribbean NL', '🇧🇶'), '670': ('Timor-Leste', '🇹🇱'), '672': ('Antarctica', '🇦🇶'),
    '673': ('Brunei', '🇧🇳'), '674': ('Nauru', '🇳🇷'), '675': ('Papua New Guinea', '🇵🇬'),
    '676': ('Tonga', '🇹🇴'), '677': ('Solomon Islands', '🇸🇧'), '678': ('Vanuatu', '🇻🇺'),
    '679': ('Fiji', '🇫🇯'), '680': ('Palau', '🇵🇼'), '681': ('Wallis and Futuna', '🇼🇫'),
    '682': ('Cook Islands', '🇨🇰'), '683': ('Niue', '🇳🇺'), '685': ('Samoa', '🇼🇸'),
    '686': ('Kiribati', '🇰🇮'), '687': ('New Caledonia', '🇳🇨'), '688': ('Tuvalu', '🇹🇻'),
    '689': ('French Polynesia', '🇵🇫'), '690': ('Tokelau', '🇹🇰'), '691': ('Micronesia', '🇫🇲'),
    '692': ('Marshall Islands', '🇲🇭'), '850': ('North Korea', '🇰🇵'), '852': ('Hong Kong', '🇭🇰'),
    '853': ('Macau', '🇲🇴'), '855': ('Cambodia', '🇰🇭'), '856': ('Laos', '🇱🇦'),
    '880': ('Bangladesh', '🇧🇩'), '886': ('Taiwan', '🇹🇼'), '960': ('Maldives', '🇲🇻'),
    '961': ('Lebanon', '🇱🇧'), '962': ('Jordan', '🇯🇴'), '963': ('Syria', '🇸🇾'),
    '964': ('Iraq', '🇮🇶'), '965': ('Kuwait', '🇰🇼'), '966': ('Saudi Arabia', '🇸🇦'),
    '967': ('Yemen', '🇾🇪'), '968': ('Oman', '🇴🇲'), '970': ('Palestine', '🇵🇸'),
    '971': ('UAE', '🇦🇪'), '972': ('Israel', '🇮🇱'), '973': ('Bahrain', '🇧🇭'),
    '974': ('Qatar', '🇶🇦'), '975': ('Bhutan', '🇧🇹'), '976': ('Mongolia', '🇲🇳'),
    '977': ('Nepal', '🇳🇵'), '992': ('Tajikistan', '🇹🇯'), '993': ('Turkmenistan', '🇹🇲'),
    '994': ('Azerbaijan', '🇦🇿'), '995': ('Georgia', '🇬🇪'), '996': ('Kyrgyzstan', '🇰🇬'),
    '998': ('Uzbekistan', '🇺🇿'),
}

def mask_number(phone):
    num = str(phone).replace('+', '').strip()
    if len(num) > 7:
        return f"{num[:3]}SRK{num[-4:]}"
    return num

def get_country_info(phone_number):
    num = str(phone_number).replace('+', '').strip()
    for i in range(4, 0, -1):
        prefix = num[:i]
        if prefix in COUNTRY_CODES: return COUNTRY_CODES[prefix]
    return ('UN', '❓')

def extract_otp(message):
    """Extract FULL OTP from message — prefers longest match and keyword context. Used for display + duplicate check."""
    if not message or not message.strip():
        return None
    text = message.strip()
    # 1) Right after OTP/code/pin keywords — capture full value (3–8 digits or hyphenated)
    keyword_patterns = [
        r'(?:otp|code|pin|password|verification\s*code)\s*:?\s*(\d{3,8}(?:[-\s]?\d{3,8})?)',
        r'(?:is|:)\s*(\d{3,8}(?:[-\s]?\d{3,8})?)',
        r'(\d{3,8}(?:[-\s]?\d{3,8})?)\s*(?:is your|is the|as your)\s*(?:otp|code)',
    ]
    for pat in keyword_patterns:
        m = re.search(pat, text, re.IGNORECASE)
        if m:
            full = re.sub(r'\s+', '', m.group(1).strip())  # 191 284 -> 191284
            if full:
                return full
    # 2) Split OTP like "191 284" or "191 - 284" -> combine to one (full OTP)
    m = re.search(r'\b(\d{3,8})\s*[-]?\s*(\d{3,8})\b', text)
    if m:
        return m.group(1) + '-' + m.group(2)  # 191 284 -> 191-284
    # 3) Any hyphenated group (e.g. 732-366)
    m = re.search(r'\b(\d{3,8}-\d{3,8})\b', text)
    if m:
        return m.group(1)
    # 4) All digit groups — return LONGEST so we get full OTP (e.g. 191284 not 191)
    candidates = re.findall(r'\b(\d{3,8})\b', text)
    if candidates:
        return max(candidates, key=len)
    return None

def detect_service_tag(message_text):
    text = message_text.lower()
    if 'whatsapp' in text or 'wa ' in text or ' wa' in text: return "#WHATSAPP"
    if 'telegram' in text or ' tg ' in text or 'tg.' in text: return "#TELEGRAM"
    if 'facebook' in text or ' fb ' in text or 'fb.' in text or 'messenger' in text: return "#FACEBOOK"
    if 'imo' in text: return "#IMO"
    if 'tiktok' in text: return "#TIKTOK"
    if 'instagram' in text or ' ig ' in text or 'ig.' in text: return "#INSTAGRAM"
    if 'google' in text: return "#GOOGLE"
    if 'chatgpt' in text or 'chat gpt' in text: return "#CHATGPT"
    if 'kimi' in text: return "#KIMI"
    if 'payoneer' in text: return "#PAYONEER"
    if 'outlook' in text or 'microsoft' in text: return "#OUTLOOK"
    return "#SMS"

def live_timer_and_delete(message_id, service_tag, country_name, flag, masked_num, otp, kb):
    seconds_left = 600  # 10 minutes
    while seconds_left > 0:
        time.sleep(30)
        seconds_left -= 30
        if seconds_left <= 0: break
        mins = seconds_left // 60
        new_text = (
            f"{service_tag} #{country_name.upper()} {flag} <code>+{masked_num}</code>\n\n"
            f"<pre>OTP: {otp}</pre>\n"
            f"⏳ Auto-delete in {mins}m..."
        )
        try:
            requests.post(f"https://api.telegram.org/bot{BOT_TOKEN}/editMessageText", json={
                'chat_id': TARGET_CHAT_ID, 'message_id': message_id,
                'text': new_text, 'parse_mode': 'HTML', 'reply_markup': kb
            })
        except: pass

    try:
        requests.post(f"https://api.telegram.org/bot{BOT_TOKEN}/deleteMessage", 
                      data={'chat_id': TARGET_CHAT_ID, 'message_id': message_id})
    except: pass

def send_telegram_otp(number, message):
    country_name, flag = get_country_info(number)
    service_tag = detect_service_tag(message)
    masked_num = mask_number(number)
    
    otp = extract_otp(message)
    otp = otp if otp else "N/A"

    # GUI: Screenshot style — header, copyable OTP block, auto-delete notice
    final_text = (
        f"{service_tag} #{country_name.upper()} {flag} <code>+{masked_num}</code>\n\n"
        f"<pre>OTP: {otp}</pre>\n"
        f"⏳ OTP auto-deletes after 10 minutes."
    )
    
    kb = {'inline_keyboard': [[
        {'text': '🤖 Number BoT', 'url': 'https://t.me/sharknumber2bot'},
        {'text': '⚡ YOUTUBE', 'url': 'https://youtube.com/@sharkmethod'}
    ]]}
    
    try:
        r = requests.post(f"https://api.telegram.org/bot{BOT_TOKEN}/sendMessage", json={
            'chat_id': TARGET_CHAT_ID, 'text': final_text, 'parse_mode': 'HTML', 'reply_markup': kb
        }).json()
        
        if r.get('ok'):
            msg_id = r['result']['message_id']
            threading.Thread(target=live_timer_and_delete, args=(msg_id, service_tag, country_name, flag, masked_num, otp, kb), daemon=True).start()
    except: pass

def main():
    options = Options()
    # options.add_argument("--headless") # Uncomment if you want to hide the browser
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    
    driver = webdriver.Chrome(service=Service(ChromeDriverManager().install()), options=options)
    
    # --- Auto Login ---
    try:
        print("[*] Attempting Automatic Login...")
        driver.get(LOGIN_URL)
        time.sleep(3)
        
        driver.find_element(By.NAME, "username").send_keys(PANEL_USERNAME)
        driver.find_element(By.NAME, "password").send_keys(PANEL_PASSWORD)
        driver.find_element(By.NAME, "password").send_keys("\n") 
        time.sleep(5)
        print("[+] Login Successful!")
    except Exception as e:
        print(f"[-] Auto-Login issue: {e}")

    conn = sqlite3.connect("otp_secure.db")
    conn.execute("CREATE TABLE IF NOT EXISTS logs (h TEXT PRIMARY KEY)")
    conn.execute("CREATE TABLE IF NOT EXISTS sent_otps (number TEXT, otp TEXT, PRIMARY KEY (number, otp))")
    
    print("[*] Monitoring SMS CDR Reports for OTPs...")
    while True:
        try:
            driver.get(BASE_URL)
            time.sleep(4) 
            
            # --- Handle Ajax/DataTables Alerts ---
            try:
                alert = driver.switch_to.alert
                print(f"[*] Closing Panel Alert: {alert.text}")
                alert.accept()
            except NoAlertPresentException:
                pass

            soup = BeautifulSoup(driver.page_source, 'html.parser')
            table = soup.find('table', {'id': 'dt'})
            if not table: table = soup.find('table') 
            
            if table:
                rows = table.find_all('tr')[1:]
                for row in rows:
                    cols = [c.text.strip() for c in row.find_all('td')]
                    if len(cols) < 6: continue
                    
                    # Column 2 = Number, Column 5 = Message
                    number = "".join(filter(str.isdigit, cols[2]))
                    message = cols[5]
                    
                    if number and message:
                        h = hashlib.md5(f"{number}{message}".encode()).hexdigest()
                        if not conn.execute("SELECT h FROM logs WHERE h=?", (h,)).fetchone():
                            otp_val = extract_otp(message) or "N/A"
                            # এক নম্বরে একই OTP বারবার না পাঠানো
                            if conn.execute("SELECT 1 FROM sent_otps WHERE number=? AND otp=?", (number, otp_val)).fetchone():
                                conn.execute("INSERT INTO logs VALUES (?)", (h,))
                                conn.commit()
                                continue
                            print(f"[+] Processing new OTP for {number}")
                            send_telegram_otp(number, message)
                            conn.execute("INSERT INTO logs VALUES (?)", (h,))
                            conn.execute("INSERT OR IGNORE INTO sent_otps VALUES (?, ?)", (number, otp_val))
                            conn.commit()
            
            time.sleep(8)
        except Exception as e:
            print(f"Error during monitoring: {e}")
            time.sleep(10)

if __name__ == "__main__":
    main()