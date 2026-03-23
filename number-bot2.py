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

# ================= CONFIGURATION =================
BOT_TOKEN = "8156968237:AAFe_oMzD8E64-oBbGUKX4v5yx7dIETES0o"
CHAT_ID = "-1003422191454"
LOGIN_URL = "http://185.2.83.39/ints/login"
SMS_URL = "http://185.2.83.39/ints/agent/SMSCDRReports"
DB_FILE = "otp_numbers.db"

# ================= ANIMATED SERVICE ICONS =================
def get_service_animation(service):
    s = service.lower() if service else ""
    
    if any(x in s for x in ["whatsapp", "ws", "wa", "واتساب", "واتس"]):
        return "<tg-emoji emoji-id='5334998226636390258'>📞</tg-emoji>"
    
    if any(x in s for x in ["facebook", "fb", "فيسبوك"]):
        return "<tg-emoji emoji-id='5323261730283863478'>💬</tg-emoji>"
    
    if any(x in s for x in ["telegram", "tg", "تيليجرام", "تلي"]):
        return "<tg-emoji emoji-id='5330237710655306682'>👉</tg-emoji>"
    
    if any(x in s for x in ["instagram", "ig", "انستقرام", "انستا"]):
        return "<tg-emoji emoji-id='5319160079465857105'>📷</tg-emoji>"
    
    if any(x in s for x in ["twitter", "x", "تويتر"]):
        return "<tg-emoji emoji-id='5224499567197700690'>🐦</tg-emoji>"
    
    if any(x in s for x in ["tiktok", "تيك توك", "تيك"]):
        return "<tg-emoji emoji-id='5224601903383457698'>🎵</tg-emoji>"
    
    if any(x in s for x in ["snapchat", "snap", "سناب"]):
        return "<tg-emoji emoji-id='5222345550904439270'>👻</tg-emoji>"
    
    if any(x in s for x in ["google", "gmail", "جوجل", "جميل"]):
        return "<tg-emoji emoji-id='5222029789203804982'>🔍</tg-emoji>"
    
    return ""

# ================= POWER ICONS =================
POWER_ICON = "<tg-emoji emoji-id='6010280017437661523'>⛩️</tg-emoji>"
POWER_EYE = "<tg-emoji emoji-id='5888704237910627502'>👁</tg-emoji>"

# ================= COMPLETE ANIMATED COUNTRY FLAGS (Lebanon Normal) =================
COUNTRY_MAP = {
    # Africa
    "20": ("EG", "<tg-emoji emoji-id='5222161185138292290'>🇪🇬</tg-emoji>"),  # Egypt
    "27": ("ZA", "<tg-emoji emoji-id='5224696216570309138'>🇿🇦</tg-emoji>"),  # South Africa
    "211": ("SS", "<tg-emoji emoji-id='5224618146949773268'>🇸🇸</tg-emoji>"),  # South Sudan
    "212": ("MA", "<tg-emoji emoji-id='5224530035695693965'>🇲🇦</tg-emoji>"),  # Morocco
    "213": ("DZ", "<tg-emoji emoji-id='5224260376174015500'>🇩🇿</tg-emoji>"),  # Algeria
    "216": ("TN", "<tg-emoji emoji-id='5221991375016310330'>🇹🇳</tg-emoji>"),  # Tunisia
    "218": ("LY", "<tg-emoji emoji-id='5222194286451242896'>🇱🇾</tg-emoji>"),  # Libya
    "220": ("GM", "<tg-emoji emoji-id='5221949872747330159'>🇬🇲</tg-emoji>"),  # Gambia
    "221": ("SN", "<tg-emoji emoji-id='5224358988623130949'>🇸🇳</tg-emoji>"),  # Senegal
    "222": ("MR", "<tg-emoji emoji-id='5224269666188274723'>🇲🇷</tg-emoji>"),  # Mauritania
    "223": ("ML", "<tg-emoji emoji-id='5224322352552096671'>🇲🇱</tg-emoji>"),  # Mali
    "224": ("GN", "<tg-emoji emoji-id='5222337588035073000'>🇬🇳</tg-emoji>"),  # Guinea
    "225": ("CI", "🇨🇮"),  # Ivory Coast - নরমাল ফ্ল্যাগ
    "226": ("BF", "🇧🇫"),  # Burkina Faso
    "227": ("NE", "<tg-emoji emoji-id='5222099049846420864'>🇳🇪</tg-emoji>"),  # Niger
    "228": ("TG", "<tg-emoji emoji-id='5222408051268532030'>🇹🇬</tg-emoji>"),  # Togo
    "229": ("BJ", "<tg-emoji emoji-id='5224515905253291409'>🇧🇯</tg-emoji>"),  # Benin
    "230": ("MU", "<tg-emoji emoji-id='5224393700548814960'>🇲🇺</tg-emoji>"),  # Mauritius
    "231": ("LR", "<tg-emoji emoji-id='5224420995065983217'>🇱🇷</tg-emoji>"),  # Liberia
    "232": ("SL", "<tg-emoji emoji-id='5224420995065983217'>🇸🇱</tg-emoji>"),  # Sierra Leone
    "233": ("GH", "<tg-emoji emoji-id='5224511339703056124'>🇬🇭</tg-emoji>"),  # Ghana
    "234": ("NG", "<tg-emoji emoji-id='5224723614166691638'>🇳🇬</tg-emoji>"),  # Nigeria
    "235": ("TD", "<tg-emoji emoji-id='5222060468155204001'>🇹🇩</tg-emoji>"),  # Chad
    "236": ("CF", "<tg-emoji emoji-id='5222060468155204001'>🇨🇫</tg-emoji>"),  # Central African Republic
    "237": ("CM", "<tg-emoji emoji-id='5222234560359577687'>🇨🇲</tg-emoji>"),  # Cameroon
    "238": ("CV", "<tg-emoji emoji-id='5224567367551428669'>🇨🇻</tg-emoji>"),  # Cape Verde
    "239": ("ST", "<tg-emoji emoji-id='5221953304426198315'>🇸🇹</tg-emoji>"),  # Sao Tome
    "240": ("GQ", "<tg-emoji emoji-id='5224455152940886669'>🇬🇶</tg-emoji>"),  # Equatorial Guinea
    "241": ("GA", "<tg-emoji emoji-id='5222152195771742239'>🇬🇦</tg-emoji>"),  # Gabon
    "242": ("CG", "<tg-emoji emoji-id='5224490444687158452'>🇨🇬</tg-emoji>"),  # Congo
    "243": ("CD", "<tg-emoji emoji-id='5224490444687158452'>🇨🇩</tg-emoji>"),  # DR Congo
    "244": ("AO", "<tg-emoji emoji-id='5224379767674907895'>🇦🇴</tg-emoji>"),  # Angola
    "245": ("GW", "<tg-emoji emoji-id='5224705704153066489'>🇬🇼</tg-emoji>"),  # Guinea-Bissau
    "248": ("SC", "<tg-emoji emoji-id='5224467496676896871'>🇸🇨</tg-emoji>"),  # Seychelles
    "249": ("SD", "<tg-emoji emoji-id='5224372990216514135'>🇸🇩</tg-emoji>"),  # Sudan
    "250": ("RW", "<tg-emoji emoji-id='5222449197055227754'>🇷🇼</tg-emoji>"),  # Rwanda
    "251": ("ET", "<tg-emoji emoji-id='5224467805914542024'>🇪🇹</tg-emoji>"),  # Ethiopia
    "252": ("SO", "<tg-emoji emoji-id='5222370504664428325'>🇸🇴</tg-emoji>"),  # Somalia
    "253": ("DJ", "<tg-emoji emoji-id='5221991375016310330'>🇩🇯</tg-emoji>"),  # Djibouti
    "254": ("KE", "<tg-emoji emoji-id='5222089648163009103'>🇰🇪</tg-emoji>"),  # Kenya
    "255": ("TZ", "<tg-emoji emoji-id='5224397364155923150'>🇹🇿</tg-emoji>"),  # Tanzania
    "256": ("UG", "<tg-emoji emoji-id='5222464040462200940'>🇺🇬</tg-emoji>"),  # Uganda
    "257": ("BI", "<tg-emoji emoji-id='5224490444687158452'>🇧🇮</tg-emoji>"),  # Burundi
    "258": ("MZ", "<tg-emoji emoji-id='5222470388423864826'>🇲🇿</tg-emoji>"),  # Mozambique
    "260": ("ZM", "<tg-emoji emoji-id='5224646626877911277'>🇿🇲</tg-emoji>"),  # Zambia
    "261": ("MG", "<tg-emoji emoji-id='5222042605386217334'>🇲🇬</tg-emoji>"),  # Madagascar
    "262": ("RE", "<tg-emoji emoji-id='5222042605386217334'>🇷🇪</tg-emoji>"),  # Reunion
    "263": ("ZW", "<tg-emoji emoji-id='5222060442385397848'>🇿🇼</tg-emoji>"),  # Zimbabwe
    "264": ("NA", "<tg-emoji emoji-id='5224690826386351746'>🇳🇦</tg-emoji>"),  # Namibia
    "265": ("MW", "<tg-emoji emoji-id='5222470435668505656'>🇲🇼</tg-emoji>"),  # Malawi
    "266": ("LS", "<tg-emoji emoji-id='5224660718665607511'>🇱🇸</tg-emoji>"),  # Lesotho
    "267": ("BW", "<tg-emoji emoji-id='5224570532942329532'>🇧🇼</tg-emoji>"),  # Botswana
    "268": ("SZ", "<tg-emoji emoji-id='5224269666188274723'>🇸🇿</tg-emoji>"),  # Eswatini
    "269": ("KM", "<tg-emoji emoji-id='5222398735484466247'>🇰🇲</tg-emoji>"),  # Comoros

    # Asia
    "60": ("MY", "<tg-emoji emoji-id='5224312886444174057'>🇲🇾</tg-emoji>"),  # Malaysia
    "62": ("ID", "<tg-emoji emoji-id='5224405893960969756'>🇮🇩</tg-emoji>"),  # Indonesia
    "63": ("PH", "<tg-emoji emoji-id='5222065042295376892'>🇵🇭</tg-emoji>"),  # Philippines
    "64": ("NZ", "<tg-emoji emoji-id='5224573595254009705'>🇳🇿</tg-emoji>"),  # New Zealand
    "65": ("SG", "<tg-emoji emoji-id='5224194023224257181'>🇸🇬</tg-emoji>"),  # Singapore
    "66": ("TH", "<tg-emoji emoji-id='5224638530864556281'>🇹🇭</tg-emoji>"),  # Thailand
    "81": ("JP", "<tg-emoji emoji-id='5222390089715299207'>🇯🇵</tg-emoji>"),  # Japan
    "82": ("KR", "<tg-emoji emoji-id='5222345550904439270'>🇰🇷</tg-emoji>"),  # South Korea
    "84": ("VN", "<tg-emoji emoji-id='5222359651282071925'>🇻🇳</tg-emoji>"),  # Vietnam
    "86": ("CN", "<tg-emoji emoji-id='5224435456220868088'>🇨🇳</tg-emoji>"),  # China
    "90": ("TR", "<tg-emoji emoji-id='5224601903383457698'>🇹🇷</tg-emoji>"),  # Turkey
    "91": ("IN", "<tg-emoji emoji-id='5222300011366200403'>🇮🇳</tg-emoji>"),  # India
    "92": ("PK", "<tg-emoji emoji-id='5224637061985742245'>🇵🇰</tg-emoji>"),  # Pakistan
    "93": ("AF", "<tg-emoji emoji-id='5222096009009575868'>🇦🇫</tg-emoji>"),  # Afghanistan
    "94": ("LK", "<tg-emoji emoji-id='5224277294050192388'>🇱🇰</tg-emoji>"),  # Sri Lanka
    "95": ("MM", "<tg-emoji emoji-id='5224393700548814960'>🇲🇲</tg-emoji>"),  # Myanmar
    "98": ("IR", "<tg-emoji emoji-id='5224374154152653367'>🇮🇷</tg-emoji>"),  # Iran
    "850": ("KP", "<tg-emoji emoji-id='5222345550904439270'>🇰🇵</tg-emoji>"),  # North Korea
    "852": ("HK", "<tg-emoji emoji-id='5224435456220868088'>🇭🇰</tg-emoji>"),  # Hong Kong
    "853": ("MO", "<tg-emoji emoji-id='5224435456220868088'>🇲🇴</tg-emoji>"),  # Macau
    "855": ("KH", "<tg-emoji emoji-id='5224638530864556281'>🇰🇭</tg-emoji>"),  # Cambodia
    "856": ("LA", "<tg-emoji emoji-id='5224638530864556281'>🇱🇦</tg-emoji>"),  # Laos
    "960": ("MV", "<tg-emoji emoji-id='5224393700548814960'>🇲🇻</tg-emoji>"),  # Maldives
    "961": ("LB", "🇱🇧"),  # Lebanon - NORMAL FLAG (as requested)
    "962": ("JO", "<tg-emoji emoji-id='5222229234600130045'>🇯🇴</tg-emoji>"),  # Jordan
    "963": ("SY", "<tg-emoji emoji-id='5224601903383457698'>🇸🇾</tg-emoji>"),  # Syria
    "964": ("IQ", "<tg-emoji emoji-id='5221980268230882832'>🇮🇶</tg-emoji>"),  # Iraq
    "965": ("KW", "<tg-emoji emoji-id='5222225596762830469'>🇰🇼</tg-emoji>"),  # Kuwait
    "966": ("SA", "<tg-emoji emoji-id='5224698145010624573'>🇸🇦</tg-emoji>"),  # Saudi Arabia
    "967": ("YE", "<tg-emoji emoji-id='5222300655611294950'>🇾🇪</tg-emoji>"),  # Yemen
    "968": ("OM", "<tg-emoji emoji-id='5222396686785066306'>🇴🇲</tg-emoji>"),  # Oman
    "970": ("PS", "<tg-emoji emoji-id='5222041677673282461'>🇵🇸</tg-emoji>"),  # Palestine
    "971": ("AE", "<tg-emoji emoji-id='5224565851427976312'>🇦🇪</tg-emoji>"),  # UAE
    "972": ("IL", "<tg-emoji emoji-id='5224720599099648709'>🇮🇱</tg-emoji>"),  # Israel
    "973": ("BH", "<tg-emoji emoji-id='5222225596762830469'>🇧🇭</tg-emoji>"),  # Bahrain
    "974": ("QA", "<tg-emoji emoji-id='5222225596762830469'>🇶🇦</tg-emoji>"),  # Qatar
    "975": ("BT", "<tg-emoji emoji-id='5222444378101925267'>🇧🇹</tg-emoji>"),  # Bhutan
    "976": ("MN", "<tg-emoji emoji-id='5224192257992701543'>🇲🇳</tg-emoji>"),  # Mongolia
    "977": ("NP", "<tg-emoji emoji-id='5222444378101925267'>🇳🇵</tg-emoji>"),  # Nepal
    "992": ("TJ", "<tg-emoji emoji-id='5222217865821696536'>🇹🇯</tg-emoji>"),  # Tajikistan
    "993": ("TM", "<tg-emoji emoji-id='5224256935905208951'>🇹🇲</tg-emoji>"),  # Turkmenistan
    "994": ("AZ", "<tg-emoji emoji-id='5224426544163728284'>🇦🇿</tg-emoji>"),  # Azerbaijan
    "995": ("GE", "<tg-emoji emoji-id='5222152195771742239'>🇬🇪</tg-emoji>"),  # Georgia
    "996": ("KG", "<tg-emoji emoji-id='5224426544163728284'>🇰🇬</tg-emoji>"),  # Kyrgyzstan
    "998": ("UZ", "<tg-emoji emoji-id='5222404546575219535'>🇺🇿</tg-emoji>"),  # Uzbekistan

    # Europe
    "30": ("GR", "<tg-emoji emoji-id='5222463490706389920'>🇬🇷</tg-emoji>"),  # Greece
    "31": ("NL", "<tg-emoji emoji-id='5224516489368841614'>🇳🇱</tg-emoji>"),  # Netherlands
    "32": ("BE", "<tg-emoji emoji-id='5224520754271366661'>🇧🇪</tg-emoji>"),  # Belgium
    "33": ("FR", "<tg-emoji emoji-id='5222029789203804982'>🇫🇷</tg-emoji>"),  # France
    "34": ("ES", "<tg-emoji emoji-id='5222024776976970940'>🇪🇸</tg-emoji>"),  # Spain
    "36": ("HU", "<tg-emoji emoji-id='5224691998912427164'>🇭🇺</tg-emoji>"),  # Hungary
    "39": ("IT", "<tg-emoji emoji-id='5222460101977190141'>🇮🇹</tg-emoji>"),  # Italy
    "40": ("RO", "<tg-emoji emoji-id='5222273794885826118'>🇷🇴</tg-emoji>"),  # Romania
    "41": ("CH", "<tg-emoji emoji-id='5224707263226194753'>🇨🇭</tg-emoji>"),  # Switzerland
    "43": ("AT", "<tg-emoji emoji-id='5224520754271366661'>🇦🇹</tg-emoji>"),  # Austria
    "44": ("GB", "<tg-emoji emoji-id='5224518800061245598'>🇬🇧</tg-emoji>"),  # United Kingdom
    "45": ("DK", "<tg-emoji emoji-id='5224245902134226386'>🇩🇰</tg-emoji>"),  # Denmark
    "46": ("SE", "<tg-emoji emoji-id='5222201098269373561'>🇸🇪</tg-emoji>"),  # Sweden
    "47": ("NO", "<tg-emoji emoji-id='5224465228934163949'>🇳🇴</tg-emoji>"),  # Norway
    "48": ("PL", "<tg-emoji emoji-id='5224670399521892983'>🇵🇱</tg-emoji>"),  # Poland
    "49": ("DE", "<tg-emoji emoji-id='5222165617544542414'>🇩🇪</tg-emoji>"),  # Germany
    "350": ("GI", "<tg-emoji emoji-id='5224518800061245598'>🇬🇮</tg-emoji>"),  # Gibraltar
    "351": ("PT", "<tg-emoji emoji-id='5224404094369672274'>🇵🇹</tg-emoji>"),  # Portugal
    "352": ("LU", "<tg-emoji emoji-id='5224499567197700690'>🇱🇺</tg-emoji>"),  # Luxembourg
    "353": ("IE", "<tg-emoji emoji-id='5222233374948602940'>🇮🇪</tg-emoji>"),  # Ireland
    "354": ("IS", "<tg-emoji emoji-id='5222063229819172521'>🇮🇸</tg-emoji>"),  # Iceland
    "355": ("AL", "<tg-emoji emoji-id='5224312057515486246'>🇦🇱</tg-emoji>"),  # Albania
    "356": ("MT", "<tg-emoji emoji-id='5224312057515486246'>🇲🇹</tg-emoji>"),  # Malta
    "357": ("CY", "<tg-emoji emoji-id='5224601903383457698'>🇨🇾</tg-emoji>"),  # Cyprus
    "358": ("FI", "<tg-emoji emoji-id='5224282903277482188'>🇫🇮</tg-emoji>"),  # Finland
    "359": ("BG", "<tg-emoji emoji-id='5224670399521892983'>🇧🇬</tg-emoji>"),  # Bulgaria
    "370": ("LT", "<tg-emoji emoji-id='5224245902134226386'>🇱🇹</tg-emoji>"),  # Lithuania
    "371": ("LV", "<tg-emoji emoji-id='5224245902134226386'>🇱🇻</tg-emoji>"),  # Latvia
    "372": ("EE", "<tg-emoji emoji-id='5224245902134226386'>🇪🇪</tg-emoji>"),  # Estonia
    "373": ("MD", "<tg-emoji emoji-id='5222273794885826118'>🇲🇩</tg-emoji>"),  # Moldova
    "374": ("AM", "<tg-emoji emoji-id='5224369957969603463'>🇦🇲</tg-emoji>"),  # Armenia
    "375": ("BY", "<tg-emoji emoji-id='5280820319458707404'>🇧🇾</tg-emoji>"),  # Belarus
    "376": ("AD", "<tg-emoji emoji-id='5221987861733061751'>🇦🇩</tg-emoji>"),  # Andorra
    "377": ("MC", "<tg-emoji emoji-id='5221937224068640464'>🇲🇨</tg-emoji>"),  # Monaco
    "378": ("SM", "<tg-emoji emoji-id='5224312057515486246'>🇸🇲</tg-emoji>"),  # San Marino
    "380": ("UA", "<tg-emoji emoji-id='5222250679371839695'>🇺🇦</tg-emoji>"),  # Ukraine
    "381": ("RS", "<tg-emoji emoji-id='5222145396838512729'>🇷🇸</tg-emoji>"),  # Serbia
    "382": ("ME", "<tg-emoji emoji-id='5224463399278096980'>🇲🇪</tg-emoji>"),  # Montenegro
    "383": ("XK", "<tg-emoji emoji-id='5222145396838512729'>🇽🇰</tg-emoji>"),  # Kosovo
    "385": ("HR", "<tg-emoji emoji-id='5224660718665607511'>🇭🇷</tg-emoji>"),  # Croatia
    "386": ("SI", "<tg-emoji emoji-id='5224660718665607511'>🇸🇮</tg-emoji>"),  # Slovenia
    "387": ("BA", "<tg-emoji emoji-id='5224660718665607511'>🇧🇦</tg-emoji>"),  # Bosnia
    "389": ("MK", "<tg-emoji emoji-id='5222470435668505656'>🇲🇰</tg-emoji>"),  # North Macedonia
    "420": ("CZ", "<tg-emoji emoji-id='5224499567197700690'>🇨🇿</tg-emoji>"),  # Czech Republic
    "421": ("SK", "<tg-emoji emoji-id='5222401879400528047'>🇸🇰</tg-emoji>"),  # Slovakia
    "423": ("LI", "<tg-emoji emoji-id='5224520754271366661'>🇱🇮</tg-emoji>"),  # Liechtenstein

    # Americas
    "1": ("US", "<tg-emoji emoji-id='5224321781321442532'>🇺🇸</tg-emoji>"),  # United States
    "51": ("PE", "<tg-emoji emoji-id='5224482026551258766'>🇵🇪</tg-emoji>"),  # Peru
    "52": ("MX", "<tg-emoji emoji-id='5224482026551258766'>🇲🇽</tg-emoji>"),  # Mexico
    "53": ("CU", "<tg-emoji emoji-id='5224482026551258766'>🇨🇺</tg-emoji>"),  # Cuba
    "54": ("AR", "<tg-emoji emoji-id='5224482026551258766'>🇦🇷</tg-emoji>"),  # Argentina
    "55": ("BR", "<tg-emoji emoji-id='5224688610183228070'>🇧🇷</tg-emoji>"),  # Brazil
    "56": ("CL", "<tg-emoji emoji-id='5224482026551258766'>🇨🇱</tg-emoji>"),  # Chile
    "57": ("CO", "<tg-emoji emoji-id='5224455152940886669'>🇨🇴</tg-emoji>"),  # Colombia
    "58": ("VE", "<tg-emoji emoji-id='5434009132753499322'>🇻🇪</tg-emoji>"),  # Venezuela
    "501": ("BZ", "<tg-emoji emoji-id='5224482026551258766'>🇧🇿</tg-emoji>"),  # Belize
    "502": ("GT", "<tg-emoji emoji-id='5222128302868672826'>🇬🇹</tg-emoji>"),  # Guatemala
    "503": ("SV", "<tg-emoji emoji-id='5222128302868672826'>🇸🇻</tg-emoji>"),  # El Salvador
    "504": ("HN", "<tg-emoji emoji-id='5222229234600130045'>🇭🇳</tg-emoji>"),  # Honduras
    "505": ("NI", "<tg-emoji emoji-id='5222128302868672826'>🇳🇮</tg-emoji>"),  # Nicaragua
    "506": ("CR", "<tg-emoji emoji-id='5222128302868672826'>🇨🇷</tg-emoji>"),  # Costa Rica
    "507": ("PA", "<tg-emoji emoji-id='5222111719999945107'>🇵🇦</tg-emoji>"),  # Panama
    "509": ("HT", "<tg-emoji emoji-id='5224683146984831315'>🇭🇹</tg-emoji>"),  # Haiti
    "591": ("BO", "<tg-emoji emoji-id='5224482026551258766'>🇧🇴</tg-emoji>"),  # Bolivia
    "592": ("GY", "<tg-emoji emoji-id='5224570532942329532'>🇬🇾</tg-emoji>"),  # Guyana
    "593": ("EC", "<tg-emoji emoji-id='5224191188545840926'>🇪🇨</tg-emoji>"),  # Ecuador
    "595": ("PY", "<tg-emoji emoji-id='5222152565138929235'>🇵🇾</tg-emoji>"),  # Paraguay
    "597": ("SR", "<tg-emoji emoji-id='5224567367551428669'>🇸🇷</tg-emoji>"),  # Suriname
    "598": ("UY", "<tg-emoji emoji-id='5222466849370813232'>🇺🇾</tg-emoji>"),  # Uruguay

    # Oceania
    "61": ("AU", "<tg-emoji emoji-id='5224573595254009705'>🇦🇺</tg-emoji>"),  # Australia
    "64": ("NZ", "<tg-emoji emoji-id='5224573595254009705'>🇳🇿</tg-emoji>"),  # New Zealand
    "670": ("TL", "<tg-emoji emoji-id='5224515905253291409'>🇹🇱</tg-emoji>"),  # Timor-Leste
    "673": ("BN", "<tg-emoji emoji-id='5224312886444174057'>🇧🇳</tg-emoji>"),  # Brunei
    "674": ("NR", "<tg-emoji emoji-id='5224573595254009705'>🇳🇷</tg-emoji>"),  # Nauru
    "675": ("PG", "<tg-emoji emoji-id='5224500164198149905'>🇵🇬</tg-emoji>"),  # Papua New Guinea
    "676": ("TO", "<tg-emoji emoji-id='5224573595254009705'>🇹🇴</tg-emoji>"),  # Tonga
    "677": ("SB", "<tg-emoji emoji-id='5222290588207954120'>🇸🇧</tg-emoji>"),  # Solomon Islands
    "678": ("VU", "<tg-emoji emoji-id='5222126748090512778'>🇻🇺</tg-emoji>"),  # Vanuatu
    "679": ("FJ", "<tg-emoji emoji-id='5221962676044838178'>🇫🇯</tg-emoji>"),  # Fiji
    "680": ("PW", "<tg-emoji emoji-id='5224573595254009705'>🇵🇼</tg-emoji>"),  # Palau
    "685": ("WS", "<tg-emoji emoji-id='5224660353593387686'>🇼🇸</tg-emoji>"),  # Samoa
    "686": ("KI", "<tg-emoji emoji-id='5224573595254009705'>🇰🇮</tg-emoji>"),  # Kiribati
    "687": ("NC", "<tg-emoji emoji-id='5224573595254009705'>🇳🇨</tg-emoji>"),  # New Caledonia
    "688": ("TV", "<tg-emoji emoji-id='5224573595254009705'>🇹🇻</tg-emoji>"),  # Tuvalu
    "689": ("PF", "<tg-emoji emoji-id='5224573595254009705'>🇵🇫</tg-emoji>"),  # French Polynesia
    "691": ("FM", "<tg-emoji emoji-id='5224573595254009705'>🇫🇲</tg-emoji>"),  # Micronesia
    "692": ("MH", "<tg-emoji emoji-id='5224573595254009705'>🇲🇭</tg-emoji>"),  # Marshall Islands

    # Special Flags
    "scotland": ("SCT", "<tg-emoji emoji-id='5224580312582861623'>🏴󠁧󠁢󠁳󠁣󠁴󠁿</tg-emoji>"),  # Scotland
    "wales": ("WLS", "<tg-emoji emoji-id='5224431333052264232'>🏴󠁧󠁢󠁷󠁬󠁳󠁿</tg-emoji>"),  # Wales
    "eu": ("EU", "<tg-emoji emoji-id='5222108911091331711'>🇪🇺</tg-emoji>"),  # European Union
    "un": ("UN", "<tg-emoji emoji-id='5451772687993031127'>🇺🇳</tg-emoji>"),  # United Nations
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
    
    # YOUR EXACT KEYBOARD FORMAT (EXACTLY AS YOU PROVIDED)
    keyboard = {
        "inline_keyboard": [
            [
                {
                    "text": otp_code,
                    "icon_custom_emoji_id": "6176966310920983412",
                    "copy_text": {"text": otp_code},
                    "style": "primary"
                }
            ],
            [
                {
                    "text": "Number Bot",
                    "icon_custom_emoji_id": "5231197925178089666",
                    "url": "https://t.me/sharknumber2bot",
                    "style": "danger"
                },
                {
                    "text": "Method",
                    "icon_custom_emoji_id": "5942902988564600402",
                    "url": "https://youtube.com/@sharkmethod?si=q2WqPvrY4iK77avz",
                    "style": "success"
                }
            ]
        ]
    }
    
    payload = {
        "chat_id": CHAT_ID,
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
                timestamp = datetime.strptime(date_time, '%Y-%m-%d %H:%M:%S')
            except:
                timestamp = datetime.utcnow() + timedelta(hours=6)  # Dhaka time

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