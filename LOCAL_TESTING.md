# Advanced Local Testing Guide (Webhooks + Nginx)

To simulate your VPS environment on your local machine, follow these steps:

### 1. Install & Start Ngrok
Ngrok creates a public "tunnel" to your local machine so Telegram can reach you.
1.  **Install**: `sudo snap install ngrok` (or follow [ngrok.com](https://ngrok.com))
2.  **Start it**: 
    ```bash
    ngrok http 80
    ```
    *Copy the "Forwarding" URL (e.g., `https://a1b2-c3d4.ngrok-free.app`).*

### 2. Configure Local Nginx
1.  **Install**: `sudo apt update && sudo apt install nginx`
2.  **Edit Config**: `sudo nano /etc/nginx/sites-available/default`
3.  **Paste the proxy part**:
    ```nginx
    server {
        listen 80;
        location / {
            proxy_pass http://localhost:8080; # To your Go Bot
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
    ```
4.  **Restart Nginx**: `sudo systemctl restart nginx`

### 3. Update your `.env`
Set the `WEBHOOK_URL` to your **ngrok** address:
```bash
WEBHOOK_URL=https://a1b2-c3d4.ngrok-free.app/
LISTEN_PORT=8080
```

### 4. Run the Bot
```bash
./hm
```

### What to Look For:
- **Webhook**: You should see "bot starting in webhook mode" in the logs.
- **Nginx**: Messages will flow from Telegram -> Ngrok -> Nginx -> Go Bot.
- **Jitter**: High-traffic logs will show that `polling AJAX source` happens at slightly different times (16s, 19s, 17s, etc.), confirming the shock-absorber logic is working.
