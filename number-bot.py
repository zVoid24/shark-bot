import os
import re
import time
import json
import asyncio
import logging
import aiosqlite
from dotenv import load_dotenv
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup, ReplyKeyboardMarkup, KeyboardButton, ReplyKeyboardRemove
from telegram.ext import (
    Application, CommandHandler, MessageHandler, CallbackQueryHandler, 
    ConversationHandler, ContextTypes, filters
)
# Import BaseFilter for creating custom filter class
from telegram.ext.filters import BaseFilter 
from telegram.constants import ParseMode
from telegram.error import BadRequest  # 🔥 ADDED: For error handling

# Load environment variables from .env file
load_dotenv()

# --- LOGGING SETUP ---
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)
# Suppress httpx logs (Too verbose)
logging.getLogger("httpx").setLevel(logging.WARNING)
logger = logging.getLogger(__name__)

# --- CONFIGURATION ---
BOT_TOKEN = os.getenv("BOT_TOKEN")

# Initial owner IDs from environment variable
INITIAL_OWNER_IDS = os.getenv("OWNER_IDS", "").split(',')
target_group_ids_str = os.getenv("TARGET_GROUP_IDS", "")
TARGET_GROUP_IDS = [int(group_id) for group_id in target_group_ids_str.split(',') if group_id]

# Check if essential variables are set
if not BOT_TOKEN or not (INITIAL_OWNER_IDS and INITIAL_OWNER_IDS[0]) or not TARGET_GROUP_IDS:
    logger.error("Essential environment variables (BOT_TOKEN, OWNER_IDS, TARGET_GROUP_IDS) are not set.")
    raise ValueError("Variables missing in .env")

# Cooldown time for getting a phone number (in seconds)
COOLDOWN_SECONDS = 10
DB_PATH = "bot_data.db"

# --- GLOBAL REGEX PATTERNS ---
OTP_PATTERNS = [
    re.compile(r"(?:Your Code|Code|OTP|Codigo|verification|OTP Code)\s*(?:➡️|:|\s)\s*([\d\s-]+)", re.IGNORECASE),
    re.compile(r"G-([\d]+) is your Google verification code", re.IGNORECASE),
    re.compile(r"#\s*([\d]+)\s*is your Facebook code", re.IGNORECASE),
    re.compile(r"Your WhatsApp(?: Business)? code\s*([\d\s-]+)", re.IGNORECASE),
    re.compile(r"\b(\d{3}[-\s]\d{3,4})\b", re.IGNORECASE),
    re.compile(r"code is\s*[:\s]*(\d{4,8})", re.IGNORECASE),
    re.compile(r"code:\s*(\d{4,8})", re.IGNORECASE),
    re.compile(r"\b(\d{4,8})\b", re.IGNORECASE)
]

NUMBER_PATTERNS = [
    re.compile(r"(?:Number|Mobile|Phone|📱|☎️|📞)\s*[:\s]*(\+?[\d•\*xX⁕\s-]{7,})", re.IGNORECASE),
    re.compile(r"(\b[\d]*[\*xX•⁕]+[\d]{3,}\b|\b\d{10,}\b)", re.IGNORECASE)
]

SERVICE_PATTERN = re.compile(r"(?:Service|🔥 Service|Code)\s*(?:WhatsApp|Telegram|Google|Facebook|:|\s)\s*(\w+)", re.IGNORECASE)

# --- ASYNC QUEUE INITIALIZATION ---
# Initially None, will be initialized in main()
OTP_QUEUE = None
START_QUEUE = None  # NEW: For /start command

# --- RESTORED FILE PATHS (Legacy Support) ---
PLATFORM_DATA_FILE = "platform_data.json"
USER_LIST_FILE = "user_list.json"
BLOCKED_USERS_FILE = "blocked_users.json"
ADMIN_IDS_FILE = "admin_ids.json"

# States for Conversation Handler
CHOOSE_PLATFORM, ASK_PLATFORM_NAME, CHOOSE_COUNTRY, ASK_COUNTRY_NAME, HANDLE_NUMBER_FILE = range(5)
ASK_PLATFORM_REMOVE, ASK_COUNTRY_REMOVE = range(5, 7)
ASK_CHAT_ID = range(7, 8)

# --- DATA FUNCTIONS ---
def load_json_data(file_path, default_type=dict):
    if os.path.exists(file_path):
        with open(file_path, 'r', encoding='utf-8') as f:
            try: return json.load(f)
            except: return default_type()
    return default_type()

def save_json_data(data, file_path):
    with open(file_path, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=4, ensure_ascii=False)

# Initialize OWNER_IDS
def initialize_owner_ids():
    admin_ids = load_json_data(ADMIN_IDS_FILE, default_type=list)
    if not admin_ids:
        admin_ids = [uid.strip() for uid in INITIAL_OWNER_IDS if uid.strip()]
        save_json_data(admin_ids, ADMIN_IDS_FILE)
    return admin_ids

OWNER_IDS = initialize_owner_ids()

# --- DATABASE CLASS ---
class Database:
    @staticmethod
    async def setup():
        """Sets up the SQLite database tables."""
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("PRAGMA journal_mode=WAL;")
            await db.execute("PRAGMA synchronous=NORMAL;")
            
            await db.execute('''CREATE TABLE IF NOT EXISTS users 
                                (user_id TEXT PRIMARY KEY, full_name TEXT, is_blocked INTEGER DEFAULT 0)''')
            
            await db.execute('''CREATE TABLE IF NOT EXISTS platform_numbers 
                                (id INTEGER PRIMARY KEY AUTOINCREMENT, platform TEXT, country TEXT, number TEXT)''')
            
            # Clean duplicates first
            await db.execute('''
                DELETE FROM platform_numbers 
                WHERE id NOT IN (
                    SELECT MIN(id) 
                    FROM platform_numbers 
                    GROUP BY platform, country, number
                )
            ''')
            
            # Unique Index
            await db.execute("CREATE UNIQUE INDEX IF NOT EXISTS idx_pn_unique ON platform_numbers(platform, country, number)")
            
            await db.execute("CREATE INDEX IF NOT EXISTS idx_pn_plat_coun ON platform_numbers(platform, country)")
            await db.execute("CREATE INDEX IF NOT EXISTS idx_pn_number ON platform_numbers(number)")

            await db.execute('''CREATE TABLE IF NOT EXISTS active_numbers 
                                (number TEXT PRIMARY KEY, user_id TEXT, timestamp REAL, message_id INTEGER, platform TEXT, country TEXT)''')
            await db.execute("CREATE INDEX IF NOT EXISTS idx_an_user ON active_numbers(user_id)")
            
            await db.execute('''CREATE TABLE IF NOT EXISTS otp_stats 
                                (country TEXT PRIMARY KEY, count INTEGER DEFAULT 0)''')
            
            await db.execute('''CREATE TABLE IF NOT EXISTS user_otp_stats 
                                (user_id TEXT, country TEXT, count INTEGER DEFAULT 0, PRIMARY KEY (user_id, country))''')
            
            await db.execute('''CREATE TABLE IF NOT EXISTS seen_numbers 
                                (user_id TEXT, number TEXT, country TEXT)''')
            await db.execute("CREATE INDEX IF NOT EXISTS idx_seen_user ON seen_numbers(user_id, country)")

            await db.execute('''CREATE TABLE IF NOT EXISTS admins 
                                (user_id TEXT PRIMARY KEY)''')
            
            await db.execute('''CREATE TABLE IF NOT EXISTS settings 
                                (key TEXT PRIMARY KEY, value TEXT)''')
            
            # Seed Admins
            for uid in OWNER_IDS:
                if uid.strip():
                    await db.execute("INSERT OR IGNORE INTO admins (user_id) VALUES (?)", (uid.strip(),))
            
            await db.execute("INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)", 
                             ("group_link", "https://t.me/tgwscreatebdotp"))
            
            await db.commit()
            logger.info("Database setup complete.")

    @staticmethod
    async def cleanup_duplicates():
        """Removes duplicate numbers."""
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute('''
                DELETE FROM platform_numbers 
                WHERE id NOT IN (
                    SELECT MIN(id) 
                    FROM platform_numbers 
                    GROUP BY platform, country, number
                )
            ''')
            await db.commit()

    @staticmethod
    async def migrate_from_json():
        """Migrates legacy JSON data to SQLite."""
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT COUNT(*) FROM platform_numbers")
            count = (await cursor.fetchone())[0]
            
            if count == 0:
                p_data = load_json_data(PLATFORM_DATA_FILE)
                for p, c_dict in p_data.items():
                    for c, nums in c_dict.items():
                        for n in nums:
                            await db.execute("INSERT OR IGNORE INTO platform_numbers (platform, country, number) VALUES (?, ?, ?)", (p, c, n))
                
                u_list = load_json_data(USER_LIST_FILE, default_type=list)
                for u in u_list: await db.execute("INSERT OR IGNORE INTO users (user_id, full_name) VALUES (?, ?)", (u, "Unknown"))
                
                blk = load_json_data(BLOCKED_USERS_FILE, default_type=list)
                for u in blk: await db.execute("INSERT OR REPLACE INTO users (user_id, is_blocked) VALUES (?, 1)", (u,))
                
                await db.commit()
                logger.info("JSON Migration checked/completed.")

    @staticmethod
    async def is_admin(user_id):
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT 1 FROM admins WHERE user_id = ?", (str(user_id),))
            return await cursor.fetchone() is not None

    @staticmethod
    async def is_blocked(user_id):
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT 1 FROM users WHERE user_id = ? AND is_blocked = 1", (str(user_id),))
            return await cursor.fetchone() is not None
    
    @staticmethod
    async def get_all_admins():
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT user_id FROM admins")
            return [row[0] for row in await cursor.fetchall()]

# --- HELPER FUNCTIONS ---
async def get_available_count(platform=None, country=None):
    async with aiosqlite.connect(DB_PATH) as db:
        if platform and country:
            query = "SELECT COUNT(*) FROM platform_numbers WHERE platform = ? AND country = ? AND number NOT IN (SELECT number FROM active_numbers)"
            params = (platform, country)
        elif platform:
            query = "SELECT COUNT(*) FROM platform_numbers WHERE platform = ? AND number NOT IN (SELECT number FROM active_numbers)"
            params = (platform,)
        else:
            return 0
        cursor = await db.execute(query, params)
        res = await cursor.fetchone()
        return res[0] if res else 0

async def permanently_remove_number(number_to_remove: str):
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT platform, country FROM platform_numbers WHERE number = ?", (number_to_remove,))
        row = await cursor.fetchone()
        
        await db.execute("DELETE FROM platform_numbers WHERE number = ?", (number_to_remove,))
        await db.execute("DELETE FROM active_numbers WHERE number = ?", (number_to_remove,))
        await db.commit()
        
        if row:
            logger.info(f"Removed number: {number_to_remove}")
            return True
        return False

# --- UI HELPER FUNCTIONS ---

# 🔥 ADDED: Safe Edit Function to fix "Message is not modified" error
async def safe_edit_message_text(update_obj, text, reply_markup=None, parse_mode=ParseMode.HTML, **kwargs):
    """Safely edits a message and ignores 'Message is not modified' errors."""
    try:
        await update_obj.edit_message_text(text=text, reply_markup=reply_markup, parse_mode=parse_mode, **kwargs)
    except BadRequest as e:
        if "Message is not modified" in str(e):
            pass  # Ignore this error
        else:
            logger.error(f"Error editing message: {e}")
            # Don't raise, just log
    except Exception as e:
        logger.error(f"Unexpected error in safe_edit: {e}")

async def show_platform_list(update_obj, is_callback=False):
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT DISTINCT platform FROM platform_numbers")
        platforms = await cursor.fetchall()

    if not platforms:
        msg = "<b>Sorry, no platforms are available right now.</b>"
        if is_callback: await safe_edit_message_text(update_obj, msg) # 🔥 Use Safe Edit
        else: await update_obj.reply_text(msg, parse_mode=ParseMode.HTML)
        return

    keyboard = []
    for row in platforms:
        plat = row[0]
        count = await get_available_count(platform=plat)
        keyboard.append([InlineKeyboardButton(f"{plat} ({count})", callback_data=f"select_platform::{plat}")])

    reply_markup = InlineKeyboardMarkup(keyboard)
    msg = "<b>🔧 Select the platform you need to access:</b>"
    
    if is_callback:
        await safe_edit_message_text(update_obj, msg, reply_markup=reply_markup) # 🔥 Use Safe Edit
    else:
        await update_obj.reply_text(msg, reply_markup=reply_markup, parse_mode=ParseMode.HTML)

async def show_country_list(query, platform):
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT DISTINCT country FROM platform_numbers WHERE platform = ?", (platform,))
        countries = await cursor.fetchall()

    if not countries:
        await safe_edit_message_text(query, f"<b>Sorry, no countries available for {platform}.</b>") # 🔥 Use Safe Edit
        return

    keyboard = []
    for row in countries:
        country = row[0]
        count = await get_available_count(platform=platform, country=country)
        keyboard.append([InlineKeyboardButton(f"{country} ({count})", callback_data=f"select_country::{platform}::{country}")])
    
    keyboard.append([InlineKeyboardButton("⬅️ Back to Platforms", callback_data="back_to_platforms")])
    await safe_edit_message_text(query, f"<b>Select your country for {platform}:</b>", reply_markup=InlineKeyboardMarkup(keyboard)) # 🔥 Use Safe Edit

# --- SUPERFAST BATCH BROADCAST TASK ---
async def run_broadcast_task(bot, user_ids, message_text):
    logger.info(f"Starting Superfast Broadcast to {len(user_ids)} users...")
    success_count = 0
    failure_count = 0
    
    # Helper to send a single message (without waiting)
    async def send_msg(uid):
        try:
            await bot.send_message(chat_id=uid, text=message_text, parse_mode=ParseMode.HTML)
            return True
        except Exception:
            return False

    # Process in batches of 20 concurrently to respect Telegram Limits (~30/sec)
    batch_size = 20
    for i in range(0, len(user_ids), batch_size):
        batch = user_ids[i:i + batch_size]
        # Create concurrent tasks for the batch
        tasks = [send_msg(uid) for uid in batch]
        # Run batch concurrently
        results = await asyncio.gather(*tasks)
        
        success_count += results.count(True)
        failure_count += results.count(False)
        
        # Wait 1.1s to avoid hitting flood limits (20 msgs / 1.1s = ~18 msgs/sec safe rate)
        await asyncio.sleep(1.1)

    logger.info(f"Broadcast Complete! Success: {success_count}, Failed: {failure_count}")

# --- BACKGROUND WORKER FOR OTP PROCESSING ---
# This worker processes OTP messages from the queue independently
async def otp_processor_worker(bot):
    logger.info("🚀 OTP Processor Worker Started in Background")
    
    while True:
        # Wait for queue initialization
        if OTP_QUEUE is None:
            await asyncio.sleep(1)
            continue
            
        # Get data from queue
        update_data = await OTP_QUEUE.get()
        
        try:
            # Unpack data
            message_text, chat_id, message_id = update_data
            
            # --- Logic from original handle_otp_message moved here ---
            otp_code = None
            for pattern in OTP_PATTERNS:
                otp_match = pattern.search(message_text)
                if otp_match:
                    otp_code = re.sub(r'[\s-]', '', otp_match.group(1))
                    if len(otp_code) >= 4: break

            detected_number_str = None
            for pattern in NUMBER_PATTERNS:
                number_match = pattern.search(message_text)
                if number_match:
                    detected_number_str = number_match.group(1).replace(" ", "").replace("-", "")
                    if any(char in detected_number_str for char in ['•', '*', '⁕', 'x', 'X']) or len(re.sub(r'\D', '', detected_number_str)) >= 7:
                        break

            if otp_code and detected_number_str:
                logger.info(f"Worker Found Match: OTP={otp_code} Num={detected_number_str}")
                
                clean_detected = re.sub(r'[^\d\*•xX⁕]', '', detected_number_str)
                parts = re.split(r'[\*•xX⁕]+', clean_detected)
                clean_prefix = parts[0] if parts[0] else ""
                clean_suffix = parts[-1] if parts[-1] else ""

                async with aiosqlite.connect(DB_PATH) as db:
                    cursor = await db.execute("SELECT number, user_id, message_id, platform, country FROM active_numbers")
                    active_rows = await cursor.fetchall()

                    match = None
                    for row in active_rows:
                        full_number = row[0]
                        clean_active = re.sub(r'\D', '', full_number)
                        if clean_active.startswith(clean_prefix) and clean_active.endswith(clean_suffix):
                            match = row
                            break

                    if match:
                        full_number_found, found_user_id, menu_message_id, platform_found, country_found = match
                        
                        # Auto Change Logic
                        query = "SELECT number FROM platform_numbers WHERE platform = ? AND country = ? AND number != ? AND number NOT IN (SELECT number FROM active_numbers) ORDER BY RANDOM() LIMIT 1"
                        cursor = await db.execute(query, (platform_found, country_found, full_number_found))
                        new_row = await cursor.fetchone()

                        if new_row:
                            new_number = new_row[0]
                            try:
                                await db.execute('''INSERT INTO active_numbers (number, user_id, timestamp, message_id, platform, country) 
                                                    VALUES (?, ?, ?, ?, ?, ?)''', 
                                                    (new_number, found_user_id, time.time(), menu_message_id, platform_found, country_found))
                                await db.execute("INSERT INTO seen_numbers (user_id, number, country) VALUES (?, ?, ?)", 
                                                 (found_user_id, new_number, country_found))
                            except aiosqlite.IntegrityError: pass

                        # 🔥 REMOVE FROM ACTIVE ONLY (Release the number)
                        await db.execute("DELETE FROM active_numbers WHERE number = ?", (full_number_found,))
                        
                        # Stats
                        await db.execute("INSERT OR IGNORE INTO otp_stats (country) VALUES (?)", (country_found,))
                        await db.execute("UPDATE otp_stats SET count = count + 1 WHERE country = ?", (country_found,))
                        await db.execute("INSERT OR IGNORE INTO user_otp_stats (user_id, country) VALUES (?, ?)", (found_user_id, country_found))
                        await db.execute("UPDATE user_otp_stats SET count = count + 1 WHERE user_id = ? AND country = ?", (found_user_id, country_found))
                        await db.commit()

                        # UI Update Logic
                        if found_user_id:
                            try:
                                # Update Menu
                                if menu_message_id:
                                    cursor = await db.execute("SELECT number FROM active_numbers WHERE user_id = ?", (found_user_id,))
                                    current_user_numbers = [r[0] for r in await cursor.fetchall()]
                                    
                                    number_display = ""
                                    for num in current_user_numbers:
                                        number_display += f"<code>{num}</code>\n"
                                    
                                    new_text = (
                                        f"<b>{country_found} ({platform_found}) Number(s) Assigned:</b>\n"
                                        f"{number_display}\n"
                                        f"<b>Waiting for OTP...</b>"
                                    )
                                    
                                    cursor = await db.execute("SELECT value FROM settings WHERE key = 'group_link'")
                                    res = await cursor.fetchone()
                                    group_link = res[0] if res else "https://t.me/tgwscreatebdotp"

                                    kb = [
                                        [InlineKeyboardButton("🔄 Change Number", callback_data=f"change_number::{platform_found}::{country_found}")],
                                        [InlineKeyboardButton("OTP Groupe 👥", url=group_link)],
                                        [InlineKeyboardButton("⬅️ Back", callback_data=f"back_to_countries::{platform_found}")]
                                    ]
                                    
                                    # Use bot object passed to worker
                                    await safe_edit_message_text( # 🔥 Use Safe Edit
                                        bot, # Need to pass bot manually slightly different for edit_message_text on bot instance vs update
                                        # But here we use bot.edit_message_text, so we can't use safe_edit helper directly as it expects update obj.
                                        # Let's just use try-except here manually for safety.
                                        text=new_text, 
                                        chat_id=found_user_id,
                                        message_id=menu_message_id,
                                        reply_markup=InlineKeyboardMarkup(kb), 
                                        parse_mode=ParseMode.HTML
                                    )

                                # Send OTP
                                service_match = SERVICE_PATTERN.search(message_text)
                                display_service = service_match.group(1) if service_match else (platform_found if platform_found else "OTP")
                                
                                msg = (
                                    f"<b>✅ OTP Received for</b> <code>{full_number_found}</code>\n\n"
                                    f"<b>🔑 Your {display_service} Code:</b> <code>{otp_code}</code>"
                                )
                                await bot.send_message(chat_id=found_user_id, text=msg, parse_mode=ParseMode.HTML)
                                logger.info(f"OTP sent to user {found_user_id}")
                                
                                # 🔥 DELETE PERMANENTLY
                                await permanently_remove_number(full_number_found) 
                                logger.info(f"Number {full_number_found} deleted permanently after OTP.")

                            except Exception as e:
                                logger.error(f"Error sending OTP to user: {e}")
                                
        except Exception as e:
            logger.error(f"Worker Error: {e}")
        finally:
            OTP_QUEUE.task_done()

# --- BACKGROUND WORKER FOR /START COMMAND (NEW) ---
async def start_command_worker(bot):
    logger.info("🚀 START Command Worker Started")
    while True:
        if START_QUEUE is None:
            await asyncio.sleep(1)
            continue
        
        # Get request
        chat_id, user_obj = await START_QUEUE.get()
        
        try:
            # --- Start Logic ---
            user = user_obj
            if await Database.is_blocked(user.id): 
                START_QUEUE.task_done()
                continue

            async with aiosqlite.connect(DB_PATH) as db:
                cursor = await db.execute("SELECT 1 FROM users WHERE user_id = ?", (str(user.id),))
                is_exist = await cursor.fetchone()
                await db.execute("INSERT OR IGNORE INTO users (user_id, full_name) VALUES (?, ?)", (str(user.id), user.full_name))
                await db.commit()
                
                # Notification
                if not is_exist:
                    notification_text = (
                        f"<b>👤 New User Joined!</b>\n\n"
                        f"<b>Name:</b> {user.full_name}\n"
                        f"<b>ID:</b> <code>{user.id}</code>\n"
                        f"<b>Username:</b> @{user.username if user.username else 'N/A'}"
                    )
                    owners_to_notify = [uid.strip() for uid in INITIAL_OWNER_IDS if uid.strip()]
                    for owner_id in owners_to_notify:
                        try:
                            await bot.send_message(chat_id=owner_id, text=notification_text, parse_mode=ParseMode.HTML)
                        except: pass

            keyboard = [[KeyboardButton("Get a Phone Number ☎️")], [KeyboardButton("📊 My Status")]]
            reply_markup = ReplyKeyboardMarkup(keyboard, resize_keyboard=True)
            await bot.send_message(
                chat_id=chat_id,
                text=f"<b>Hi {user.mention_html()}!</b>\n\n<b>You can get a phone number by clicking the \"Get a Phone Number ☎️\" button below.</b>",
                parse_mode=ParseMode.HTML, reply_markup=reply_markup
            )
            
        except Exception as e:
            logger.error(f"Start Worker Error: {e}")
        finally:
            START_QUEUE.task_done()

# --- OTP GROUP MESSAGE HANDLER (UPDATED) ---
async def handle_otp_message(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    message = (update.channel_post or update.edited_channel_post or 
               update.message or update.edited_message)
    if not message or (not message.text and not message.caption): return
    if message.chat.type == 'private': return
    message_text = message.text or message.caption 

    # Filter for target groups and ensure it's not a command
    if message.chat_id not in TARGET_GROUP_IDS or message_text.startswith('/'): return 

    # 🔥 UPDATED: Push to Queue instead of processing
    if OTP_QUEUE is not None:
        await OTP_QUEUE.put((message_text, message.chat_id, message.message_id))
    else:
        logger.error("OTP Queue is not initialized!")

# --- GENERAL COMMANDS ---
async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    # 🔥 UPDATED: Offload to Background Queue
    if START_QUEUE is not None:
        await START_QUEUE.put((update.effective_chat.id, update.effective_user))
    else:
        # Fallback if queue not ready
        await update.message.reply_text("Bot starting up...")

async def my_status(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user_id = str(update.effective_user.id)
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT country, count FROM user_otp_stats WHERE user_id = ? ORDER BY count DESC", (user_id,))
        rows = await cursor.fetchall()

    if not rows:
        await update.message.reply_text("<b>No OTP status available yet.</b>", parse_mode=ParseMode.HTML)
        return

    msg = "<b>📊 My OTP Usage 📊</b>\n\n"
    total = 0
    for country, count in rows:
        msg += f"<b>{country}:</b> <code>{count}</code> <b>OTPs</b>\n"
        total += count
    msg += f"\n<b>Total:</b> <code>{total}</code>"
    await update.message.reply_text(msg, parse_mode=ParseMode.HTML)

async def handle_get_number_button(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if await Database.is_blocked(update.effective_user.id): return
    # Use helper
    is_callback = True if update.callback_query else False
    target = update.callback_query if is_callback else update.message
    await show_platform_list(target, is_callback)

async def set_group_link(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        new_link = context.args[0]
        if not new_link.startswith("http"):
            await update.message.reply_text("<b>Invalid link. Please include http/https.</b>", parse_mode=ParseMode.HTML)
            return
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("INSERT OR REPLACE INTO settings (key, value) VALUES ('group_link', ?)", (new_link,))
            await db.commit()
        await update.message.reply_text(f"<b>OTP Group link updated successfully to:</b>\n{new_link}", parse_mode=ParseMode.HTML)
    except:
        await update.message.reply_text("<b>Usage: /setgroupe [link]</b>", parse_mode=ParseMode.HTML)

# --- NEW COMMAND: TOGGLE REMOVE POLICY (CASE INSENSITIVE FIX) ---
async def toggle_remove_policy(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        if len(context.args) != 3:
            raise ValueError
        plat, coun, status = context.args
        status = status.lower()
        if status not in ['on', 'off']:
            await update.message.reply_text("<b>Status must be 'on' or 'off'.</b>", parse_mode=ParseMode.HTML)
            return

        # Store with lowercase keys
        key = f"remove_policy::{plat.lower()}::{coun.lower()}"
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", (key, status))
            await db.commit()
        
        await update.message.reply_text(
            f"<b>✅ Configuration Updated!</b>\n\n"
            f"<b>Platform:</b> {plat}\n"
            f"<b>Country:</b> {coun}\n"
            f"<b>Remove on Change:</b> {status.upper()}",
            parse_mode=ParseMode.HTML
        )
    except:
        await update.message.reply_text("<b>Usage: /numberremove [Platform] [Country] [on/off]</b>", parse_mode=ParseMode.HTML)

# --- NEW COMMAND: SET NUMBER LIMIT (CASE INSENSITIVE FIX) ---
async def set_number_limit(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        if len(context.args) != 3:
            raise ValueError
        plat, coun, limit = context.args
        limit = int(limit)
        
        # Store with lowercase keys
        key = f"limit::{plat.lower()}::{coun.lower()}"
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", (key, str(limit)))
            await db.commit()
        
        await update.message.reply_text(
            f"<b>✅ Limit Updated!</b>\n\n"
            f"<b>Platform:</b> {plat}\n"
            f"<b>Country:</b> {coun}\n"
            f"<b>Max Numbers:</b> {limit}",
            parse_mode=ParseMode.HTML
        )
    except:
        await update.message.reply_text("<b>Usage: /numberlimit [Platform] [Country] [Limit]</b>", parse_mode=ParseMode.HTML)

# --- CALLBACK HANDLER (FIXED FOR DELETE ON CHANGE) ---
async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    query = update.callback_query
    user_id = str(query.from_user.id)
    data = query.data

    # 🔥 FIX: Answer immediately to stop "Query is too old"
    try: await query.answer()
    except: pass

    # 1. Block Check
    if await Database.is_blocked(user_id):
        try: await query.answer("You are blocked.", show_alert=True)
        except: pass
        return

    # 2. Cooldown Check (Change Number)
    is_change = data.startswith("change_number::")
    if is_change:
        last_request_time = context.user_data.get('last_request_time', 0)
        if time.time() - last_request_time < COOLDOWN_SECONDS:
            remaining = int(COOLDOWN_SECONDS - (time.time() - last_request_time))
            try:
                await query.answer(f"Please wait {remaining} seconds.", show_alert=True)
            except Exception as e:
                logger.error(f"Alert error: {e}")
            return
    
    # Routing
    if data == "back_to_platforms":
        await show_platform_list(query, is_callback=True)
        return

    if data.startswith("select_platform::"):
        platform = data.split('::')[1]
        context.user_data['selected_platform'] = platform
        await show_country_list(query, platform)
        return

    if data.startswith("back_to_countries::"):
        platform = data.split('::')[1]
        await show_country_list(query, platform)
        return

    # Number Assignment Logic
    if data.startswith("select_country::") or data.startswith("change_number::"):
        parts = data.split('::')
        platform, country = parts[1], parts[2]
        
        async with aiosqlite.connect(DB_PATH) as db:
            # Capture numbers to exclude (the ones currently held/released)
            cursor = await db.execute("SELECT number FROM active_numbers WHERE user_id = ?", (user_id,))
            old_nums = await cursor.fetchall()
            excluded_numbers = [r[0] for r in old_nums]

            # ALWAYS PERMANENTLY DELETE
            for onum in excluded_numbers:
                await db.execute("DELETE FROM platform_numbers WHERE number = ?", (onum,))
                logger.info(f"Number {onum} permanently deleted as it was assigned to user.")

            # Clear active numbers for user (release slot)
            await db.execute("DELETE FROM active_numbers WHERE user_id = ?", (user_id,))
            await db.commit()

            # 🔥 NEW: Check Limit BEFORE Assigning (Case Insensitive Lookup)
            limit_key = f"limit::{platform.lower()}::{country.lower()}"
            cursor = await db.execute("SELECT value FROM settings WHERE key = ?", (limit_key,))
            limit_res = await cursor.fetchone()
            max_limit = int(limit_res[0]) if limit_res else 2  # Default limit changed to 2

            if is_change:
                 # Clean up specific numbers logic was handled above (policy check)
                 # Now actually remove from active_numbers (which was done globally before)
                 pass 
            else:
                 # Fresh request. Check limit.
                 cursor = await db.execute("SELECT COUNT(*) FROM active_numbers WHERE user_id = ? AND platform = ? AND country = ?", (user_id, platform, country))
                 curr_cnt = (await cursor.fetchone())[0]
                 
                 if curr_cnt >= max_limit:
                     msg = f"<b>🚫 Limit Reached!</b>\n\nYou can only hold {max_limit} number(s) for {country} - {platform}."
                     await safe_edit_message_text(query, msg) # 🔥 Use Safe Edit
                     return

            # Retry logic for picking number
            assignable = []
            for _ in range(3):
                try:
                    # Q1: Strict uniqueness (checks seen_numbers)
                    q = '''SELECT number FROM platform_numbers WHERE platform=? AND country=? 
                           AND number NOT IN (SELECT number FROM active_numbers)
                           AND number NOT IN (SELECT number FROM seen_numbers WHERE user_id=?) 
                           ORDER BY RANDOM() LIMIT 2'''
                    cursor = await db.execute(q, (platform, country, user_id))
                    assignable = [r[0] for r in await cursor.fetchall()]
                    
                    if not assignable:
                         # Q2: Fallback (allows recycling BUT excludes current/just-released numbers)
                         if excluded_numbers:
                             # Dynamic query to exclude specific numbers
                             placeholders = ','.join(['?'] * len(excluded_numbers))
                             q2 = f'''SELECT number FROM platform_numbers WHERE platform=? AND country=? 
                                      AND number NOT IN (SELECT number FROM active_numbers)
                                      AND number NOT IN ({placeholders})
                                      ORDER BY RANDOM() LIMIT 2'''
                             params = (platform, country, *excluded_numbers)
                             cursor = await db.execute(q2, params)
                         else:
                             # Standard fallback if no exclusions
                             q2 = '''SELECT number FROM platform_numbers WHERE platform=? AND country=? 
                                     AND number NOT IN (SELECT number FROM active_numbers) 
                                     ORDER BY RANDOM() LIMIT 2'''
                             cursor = await db.execute(q2, (platform, country))
                             
                         assignable = [r[0] for r in await cursor.fetchall()]

                    if not assignable: break

                    msg_id = query.message.message_id
                    now = time.time()
                    for num in assignable:
                        await db.execute("INSERT INTO active_numbers (number, user_id, timestamp, message_id, platform, country) VALUES (?, ?, ?, ?, ?, ?)", (num, user_id, now, msg_id, platform, country))
                        await db.execute("INSERT INTO seen_numbers (user_id, number, country) VALUES (?, ?, ?)", (user_id, num, country))
                    await db.commit()
                    context.user_data['last_request_time'] = now
                    break 
                except aiosqlite.IntegrityError:
                    continue 
            
            if not assignable:
                msg = f"<b>Sorry, no numbers are currently available for {country}.</b>"
                await safe_edit_message_text(query, msg) # 🔥 Use Safe Edit
                return

            cursor = await db.execute("SELECT value FROM settings WHERE key='group_link'")
            res = await cursor.fetchone()
            g_link = res[0] if res else "https://t.me/+6DfK0boauLAzNWZk"

            num_disp = "".join([f"<code>{n}</code>\n" for n in assignable])
            kb = [[InlineKeyboardButton("🔄 Change Number", callback_data=f"change_number::{platform}::{country}")],
                  [InlineKeyboardButton("OTP Groupe 👥", url=g_link)],
                  [InlineKeyboardButton("⬅️ Back", callback_data=f"back_to_countries::{platform}")]]
            
            new_text = (
                f"<b>{country} ({platform}) Number(s) Assigned:</b>\n"
                f"{num_disp}\n"
                f"<b>Waiting for OTP...</b>"
            )
            await safe_edit_message_text(query, new_text, reply_markup=InlineKeyboardMarkup(kb)) # 🔥 Use Safe Edit

# --- ADMIN COMMANDS ---
async def show_admin_commands(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    admin_help_text = (
        "<b>--- Admin Command Panel ---</b>\n\n"
        "<b><u>Admin Management</u></b>\n"
        "<code>/addadmin [user_id]</code> - Add a user to the admin list.\n"
        "<code>/rmvadmin [user_id]</code> - Remove a user from the admin list.\n"
        "<code>/adminlist</code> - Show the list of all current admins.\n\n"
        "<b><u>Number Management</u></b>\n"
        "<code>/addnumber</code> - Add numbers to a platform and country.\n"
        "<code>/removenumber</code> - Remove a country's numbers from a platform.\n"
        "<code>/numberlimit [Plat] [Coun] [Limit]</code> - Set limit per user.\n"
        "<code>/numberremove [Plat] [Coun] [on/off]</code> - Set delete policy on change.\n"
        "<code>/cancel</code> - Cancel the current operation (add/remove).\n\n"
        "<b><u>Configuration</u></b>\n"
        "<code>/addchatid</code> - Add new Target Group Chat IDs for OTP monitoring.\n"
        "<code>/setgroupe [link]</code> - Set the OTP Group link.\n\n"
        "<b><u>User Management</u></b>\n"
        "<code>/block [user_id]</code> - Block a user from the bot.\n"
        "<code>/unblock [user_id]</code> - Unblock a user.\n"
        "<code>/unblockall</code> - Unblock all users at once.\n"
        "<code>/blocklist</code> - Show the list of all blocked users.\n"
        "<code>/seestatus [user_id]</code> - See a user's status and their OTP stats.\n\n"
        "<b><u>Broadcast & Stats</u></b>\n"
        "<code>/all [message]</code> - Send a message to all users of the bot.\n"
        "<code>/statusall</code> - Show total OTP stats for all countries.\n\n"
        "<b><u>Bot Management</u></b>\n"
        "<code>/resetall</code> - Reset all user data (active numbers, all stats).\n"
        "<code>/admin</code> or <code>/cmd</code> - Show this help message.\n"
    )
    await update.message.reply_text(admin_help_text, parse_mode=ParseMode.HTML)

async def show_admin_list(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    admins = await Database.get_all_admins()
    msg = "<b>--- Current Admin List ---</b>\n\n" + "\n".join([f"• <code>{uid}</code>" for uid in admins])
    await update.message.reply_text(msg, parse_mode=ParseMode.HTML)

async def add_admin(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        user_to_add = context.args[0]
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("INSERT OR IGNORE INTO admins (user_id) VALUES (?)", (user_to_add,))
            await db.commit()
        if user_to_add not in OWNER_IDS: OWNER_IDS.append(user_to_add)
        await update.message.reply_text(f"<b>User <code>{user_to_add}</code> has been added as an admin.</b>", parse_mode=ParseMode.HTML)
    except:
        await update.message.reply_text("<b>Usage: /addadmin [user_id]</b>", parse_mode=ParseMode.HTML)

async def remove_admin(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        user_to_remove = context.args[0]
        if user_to_remove in INITIAL_OWNER_IDS:
             await update.message.reply_text("<b>Error: Principal Owner cannot be removed.</b>", parse_mode=ParseMode.HTML)
             return
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT COUNT(*) FROM admins")
            count = (await cursor.fetchone())[0]
            if count <= 1:
                await update.message.reply_text("<b>Cannot remove the last admin.</b>", parse_mode=ParseMode.HTML)
                return
            await db.execute("DELETE FROM admins WHERE user_id = ?", (user_to_remove,))
            await db.commit()
        if user_to_remove in OWNER_IDS: OWNER_IDS.remove(user_to_remove)
        await update.message.reply_text(f"<b>User <code>{user_to_remove}</code> removed.</b>", parse_mode=ParseMode.HTML)
    except:
        await update.message.reply_text("<b>Usage: /rmvadmin [user_id]</b>", parse_mode=ParseMode.HTML)

async def reset_all_data(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    async with aiosqlite.connect(DB_PATH) as db:
        await db.execute("DELETE FROM active_numbers")
        await db.execute("DELETE FROM otp_stats")
        await db.execute("DELETE FROM user_otp_stats")
        await db.execute("DELETE FROM seen_numbers")
        await db.commit()
    await update.message.reply_text("<b>All user data has been successfully reset.</b>", parse_mode=ParseMode.HTML)

async def block_user(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        uid = context.args[0]
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("INSERT OR REPLACE INTO users (user_id, is_blocked) VALUES (?, 1)", (uid,))
            await db.commit()
        await update.message.reply_text(f"<b>User <code>{uid}</code> has been blocked.</b>", parse_mode=ParseMode.HTML)
    except: pass

async def unblock_user(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        uid = context.args[0]
        async with aiosqlite.connect(DB_PATH) as db:
            await db.execute("UPDATE users SET is_blocked = 0 WHERE user_id = ?", (uid,))
            await db.commit()
        await update.message.reply_text(f"<b>User <code>{uid}</code> has been unblocked.</b>", parse_mode=ParseMode.HTML)
    except: pass

async def unblock_all_users(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    async with aiosqlite.connect(DB_PATH) as db:
        await db.execute("UPDATE users SET is_blocked = 0")
        await db.commit()
    await update.message.reply_text("<b>Successfully unblocked all users.</b>", parse_mode=ParseMode.HTML)

async def see_user_status(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    try:
        uid = context.args[0]
        status_msg = f"<b>Status for User ID:</b> <code>{uid}</code>\n\n"
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT is_blocked FROM users WHERE user_id = ?", (uid,))
            row = await cursor.fetchone()
            status_msg += f"<b>Status:</b> {'🔴 Blocked' if row and row[0] else '🟢 Active'}\n"
            cursor = await db.execute("SELECT number FROM active_numbers WHERE user_id = ?", (uid,))
            row = await cursor.fetchone()
            status_msg += f"<b>Holding:</b> <code>{row[0] if row else 'None'}</code>\n"
            status_msg += "\n<b>--- User OTP Stats ---</b>\n"
            cursor = await db.execute("SELECT country, count FROM user_otp_stats WHERE user_id = ?", (uid,))
            rows = await cursor.fetchall()
            if rows:
                for c, count in rows: status_msg += f"<b>{c}:</b> <code>{count}</code> <b>OTPs</b>\n"
            else: status_msg += "<b>No OTPs received by this user yet.</b>"
        await update.message.reply_text(status_msg, parse_mode=ParseMode.HTML)
    except: 
        await update.message.reply_text("<b>Usage: /seestatus [user_id]</b>", parse_mode=ParseMode.HTML)

async def show_blocklist(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT user_id FROM users WHERE is_blocked = 1")
        blocked = [r[0] for r in await cursor.fetchall()]
    msg = "<b>--- Blocked Users List ---</b>\n\n" + "\n".join([f"<b>ID:</b> <code>{u}</code>" for u in blocked]) if blocked else "<b>There are no blocked users.</b>"
    await update.message.reply_text(msg, parse_mode=ParseMode.HTML)

async def broadcast_message(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    msg = " ".join(context.args)
    if not msg:
        await update.message.reply_text("<b>Usage: /all [your message]</b>", parse_mode=ParseMode.HTML)
        return
    try: await context.bot.send_message(chat_id=update.effective_user.id, text=f"<b>--- Broadcast Preview ---</b>\n\n{msg}", parse_mode=ParseMode.HTML)
    except: return
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT user_id FROM users")
        user_ids = [r[0] for r in await cursor.fetchall()]
    await update.message.reply_text(f"<b>Preview sent. Starting broadcast to {len(user_ids)} users...</b>", parse_mode=ParseMode.HTML)
    asyncio.create_task(run_broadcast_task(context.bot, user_ids, msg))

async def status_all(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    if not await Database.is_admin(update.effective_user.id): return
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT country, count FROM otp_stats ORDER BY count DESC")
        rows = await cursor.fetchall()
    msg = "<b>📊 Total OTP Status 📊</b>\n\n"
    total = 0
    for c, count in rows:
        msg += f"<b>{c}:</b> <code>{count}</code> <b>OTPs</b>\n"
        total += count
    msg += f"\n<b>Total All OTPs:</b> <code>{total}</code>"
    await update.message.reply_text(msg, parse_mode=ParseMode.HTML)

# --- CONVERSATION FOR ADDING NUMBERS (REPLY KEYBOARD) ---

async def add_number(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    if not await Database.is_admin(update.effective_user.id): 
        await update.message.reply_text("<b>Sorry, this command is for the owner only.</b>", parse_mode=ParseMode.HTML)
        return ConversationHandler.END
    
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT DISTINCT platform FROM platform_numbers")
        plats = [r[0] for r in await cursor.fetchall()]
    
    kb = [[KeyboardButton(p)] for p in plats]
    kb.append([KeyboardButton("➕ New Platform")])
    
    await update.message.reply_text(
        "<b>Select a platform to add numbers to:</b>", 
        reply_markup=ReplyKeyboardMarkup(kb, one_time_keyboard=True, resize_keyboard=True), 
        parse_mode=ParseMode.HTML
    )
    return CHOOSE_PLATFORM

async def handle_choose_platform(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    text = update.message.text
    if text == "➕ New Platform":
        await update.message.reply_text("<b>Enter the new platform name (e.g., WhatsApp):</b>", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
        return ASK_PLATFORM_NAME
    else:
        context.user_data['platform_to_add'] = text
        async with aiosqlite.connect(DB_PATH) as db:
            cursor = await db.execute("SELECT DISTINCT country FROM platform_numbers WHERE platform = ?", (text,))
            countries = [row[0] for row in await cursor.fetchall()]
        kb = [[KeyboardButton(c)] for c in countries]
        kb.append([KeyboardButton("➕ New Country")])
        await update.message.reply_text(
            f"<b>Platform: {text}</b>\nSelect a country or add new:",
            reply_markup=ReplyKeyboardMarkup(kb, one_time_keyboard=True, resize_keyboard=True),
            parse_mode=ParseMode.HTML
        )
        return CHOOSE_COUNTRY

async def ask_platform_name(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    context.user_data['platform_to_add'] = update.message.text.strip()
    await update.message.reply_text("<b>Enter country name (with flag):</b>", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
    return ASK_COUNTRY_NAME

async def handle_choose_country(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    text = update.message.text
    if text == "➕ New Country":
        await update.message.reply_text("<b>Enter new country name (with flag):</b>", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
        return ASK_COUNTRY_NAME
    context.user_data['country_to_add'] = text
    await update.message.reply_text(f"<b>Selected: {context.user_data['platform_to_add']} - {text}</b>\nUpload .txt file:", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
    return HANDLE_NUMBER_FILE

async def ask_country_name(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    context.user_data['country_to_add'] = update.message.text.strip()
    await update.message.reply_text(
        f"<b>Adding numbers for: {context.user_data['platform_to_add']} - {context.user_data['country_to_add']}</b>\nUpload .txt file:", 
        reply_markup=ReplyKeyboardRemove(), 
        parse_mode=ParseMode.HTML
    )
    return HANDLE_NUMBER_FILE

async def handle_number_file(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    if not update.message.document:
        await update.message.reply_text("<b>Please upload a .txt file only.</b>", parse_mode=ParseMode.HTML)
        return HANDLE_NUMBER_FILE
    
    file = await context.bot.get_file(update.message.document.file_id)
    content = await file.download_as_bytearray()
    numbers = [n.strip() for n in content.decode('utf-8').split('\n') if n.strip()]
    
    plat = context.user_data['platform_to_add']
    coun = context.user_data['country_to_add']
    
    async with aiosqlite.connect(DB_PATH) as db:
        count = 0
        for num in numbers:
            try:
                await db.execute("INSERT INTO platform_numbers (platform, country, number) VALUES (?, ?, ?)", (plat, coun, num))
                count += 1
            except: pass
        await db.commit()
    
    await update.message.reply_text(f"<b>Successfully added {count} numbers to {plat} ({coun}).</b>", parse_mode=ParseMode.HTML)
    
    # Broadcast
    broadcast_msg = (
        f"<b>🚀 New Numbers Added!</b>\n\n"
        f"<b>Platform:</b> <code>{plat}</code>\n"
        f"<b>Country:</b> <code>{coun}</code>\n"
        f"<b>Quantity:</b> <code>{count}</code> <b>numbers</b>\n\n"
        f"<i>Get your number now using the button below!</i>"
    )
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT user_id FROM users")
        uids = [r[0] for r in await cursor.fetchall()]
    asyncio.create_task(run_broadcast_task(context.bot, uids, broadcast_msg))
    
    context.user_data.pop('platform_to_add', None)
    context.user_data.pop('country_to_add', None)
    return ConversationHandler.END

async def remove_number(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    if not await Database.is_admin(update.effective_user.id): 
        await update.message.reply_text("<b>Sorry, this command is for the owner only.</b>", parse_mode=ParseMode.HTML)
        return ConversationHandler.END
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT DISTINCT platform FROM platform_numbers")
        plats = [r[0] for r in await cursor.fetchall()]
    if not plats:
        await update.message.reply_text("<b>There are no platforms to remove.</b>", parse_mode=ParseMode.HTML)
        return ConversationHandler.END
    kb = [[KeyboardButton(p)] for p in plats]
    await update.message.reply_text("<b>Select the platform:</b>", reply_markup=ReplyKeyboardMarkup(kb, one_time_keyboard=True), parse_mode=ParseMode.HTML)
    return ASK_PLATFORM_REMOVE

async def ask_platform_remove(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    plat = update.message.text
    context.user_data['platform_to_remove'] = plat
    async with aiosqlite.connect(DB_PATH) as db:
        cursor = await db.execute("SELECT DISTINCT country FROM platform_numbers WHERE platform = ?", (plat,))
        countries = [r[0] for r in await cursor.fetchall()]
    kb = [[KeyboardButton(c)] for c in countries]
    await update.message.reply_text("<b>Select the country to remove:</b>", reply_markup=ReplyKeyboardMarkup(kb, one_time_keyboard=True), parse_mode=ParseMode.HTML)
    return ASK_COUNTRY_REMOVE

async def ask_country_remove(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    country = update.message.text
    plat = context.user_data['platform_to_remove']
    async with aiosqlite.connect(DB_PATH) as db:
        await db.execute("DELETE FROM platform_numbers WHERE platform = ? AND country = ?", (plat, country))
        await db.commit()
    await update.message.reply_text(f"<b>{country} has been successfully removed from {plat}.</b>", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
    context.user_data.pop('platform_to_remove', None)
    return ConversationHandler.END

async def add_chat_id(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    if not await Database.is_admin(update.effective_user.id): 
        await update.message.reply_text("<b>Sorry, this command is for the owner only.</b>", parse_mode=ParseMode.HTML)
        return ConversationHandler.END
    await update.message.reply_text("<b>Enter the new Target Group Chat ID(s), one per line.</b>", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
    return ASK_CHAT_ID

async def handle_chat_id(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    global TARGET_GROUP_IDS
    text = update.message.text
    added = 0
    new_ids = []
    for line in text.split('\n'):
        try:
            cid = int(line.strip())
            if cid not in TARGET_GROUP_IDS:
                TARGET_GROUP_IDS.append(cid)
                new_ids.append(cid)
                added += 1
        except: pass
    if new_ids:
        await update.message.reply_text(f"<b>Successfully added {added} new Chat ID(s).</b>", parse_mode=ParseMode.HTML, reply_markup=ReplyKeyboardRemove())
    else:
        await update.message.reply_text("<b>No new valid Chat IDs were added.</b>", parse_mode=ParseMode.HTML, reply_markup=ReplyKeyboardRemove())
    return ConversationHandler.END

async def cancel(update: Update, context: ContextTypes.DEFAULT_TYPE) -> int:
    await update.message.reply_text("<b>Operation cancelled.</b>", reply_markup=ReplyKeyboardRemove(), parse_mode=ParseMode.HTML)
    return ConversationHandler.END

async def error_handler(update: object, context: ContextTypes.DEFAULT_TYPE) -> None:
    logger.error(f"Exception while handling an update: {context.error}")

# --- MAIN ---
def main() -> None:
    # Use global queue to initialize it inside the loop
    global OTP_QUEUE, START_QUEUE

    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    
    # Initialize Queues here
    OTP_QUEUE = asyncio.Queue()
    START_QUEUE = asyncio.Queue()

    loop.run_until_complete(Database.setup())
    loop.run_until_complete(Database.migrate_from_json())
    loop.run_until_complete(Database.cleanup_duplicates()) # One-time cleanup

    # Function to start background workers on app init
    async def post_init(app: Application):
        asyncio.create_task(otp_processor_worker(app.bot))
        asyncio.create_task(start_command_worker(app.bot))

    # Add post_init to builder
    application = Application.builder().token(BOT_TOKEN).post_init(post_init).read_timeout(30).write_timeout(30).build()

    add_conv = ConversationHandler(
        entry_points=[CommandHandler("addnumber", add_number)],
        states={
            CHOOSE_PLATFORM: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_choose_platform)],
            ASK_PLATFORM_NAME: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_platform_name)],
            CHOOSE_COUNTRY: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_choose_country)],
            ASK_COUNTRY_NAME: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_country_name)],
            HANDLE_NUMBER_FILE: [MessageHandler(filters.Document.ALL, handle_number_file)]
        },
        fallbacks=[CommandHandler("cancel", cancel)]
    )
    remove_conv = ConversationHandler(
        entry_points=[CommandHandler("removenumber", remove_number)],
        states={
            ASK_PLATFORM_REMOVE: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_platform_remove)],
            ASK_COUNTRY_REMOVE: [MessageHandler(filters.TEXT & ~filters.COMMAND, ask_country_remove)]
        },
        fallbacks=[CommandHandler("cancel", cancel)]
    )
    chat_id_conv = ConversationHandler(
        entry_points=[CommandHandler("addchatid", add_chat_id)],
        states={ASK_CHAT_ID: [MessageHandler(filters.TEXT & ~filters.COMMAND, handle_chat_id)]},
        fallbacks=[CommandHandler("cancel", cancel)]
    )

    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("getnumber", handle_get_number_button))
    application.add_handler(MessageHandler(filters.Regex("^Get a Phone Number ☎️$"), handle_get_number_button))
    application.add_handler(MessageHandler(filters.Regex("^📊 My Status$"), my_status))
    
    application.add_handler(CallbackQueryHandler(button_callback))
    
    application.add_handler(CommandHandler("addadmin", add_admin))
    application.add_handler(CommandHandler("rmvadmin", remove_admin))
    application.add_handler(CommandHandler("adminlist", show_admin_list))
    application.add_handler(CommandHandler("cmd", show_admin_commands))
    application.add_handler(CommandHandler("admin", show_admin_commands))
    application.add_handler(CommandHandler("setgroupe", set_group_link))
    application.add_handler(CommandHandler("numberremove", toggle_remove_policy)) # NEW COMMAND
    application.add_handler(CommandHandler("numberlimit", set_number_limit)) # NEW COMMAND
    application.add_handler(CommandHandler("resetall", reset_all_data))
    application.add_handler(CommandHandler("block", block_user))
    application.add_handler(CommandHandler("unblock", unblock_user))
    application.add_handler(CommandHandler("unblockall", unblock_all_users))
    application.add_handler(CommandHandler("seestatus", see_user_status))
    application.add_handler(CommandHandler("blocklist", show_blocklist))
    application.add_handler(CommandHandler("all", broadcast_message))
    application.add_handler(CommandHandler("statusall", status_all))
    
    application.add_handler(add_conv)
    application.add_handler(remove_conv)
    application.add_handler(chat_id_conv)

    application.add_handler(MessageHandler(filters.ALL, handle_otp_message))
    application.add_error_handler(error_handler)

    print("Bot is running with SQLite and Background Worker Tasks...")
    application.run_polling(allowed_updates=Update.ALL_TYPES)

if __name__ == "__main__":
    main()