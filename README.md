# Litterbox

A simple idea submission and review system built with PocketBase and Go. Perfect for convention panels, feedback collection, or any scenario where you need to collect and review text submissions.

## Features

- **Public submission page** - Anyone can submit ideas via a simple form
- **Protected review dashboard** - Admin-only interface to review and manage submissions
- **Status management** - Organize submissions by status (New, Approved, Done, Hidden, Deleted)
- **Automatic timestamps** - Track when submissions were created and updated
- **Collapsible groups** - Keep the review interface clean and organized

## Prerequisites

- Go 1.21 or higher
- Git

## Setup

### 1. Clone the repository

```bash
cd litterbox
```

### 2. Initialize Go module and install dependencies

```bash
go mod init litterbox
go get github.com/pocketbase/pocketbase
go get github.com/pocketbase/pocketbase/plugins/migratecmd
```

### 3. Run the application

```bash
go run . serve
```

The server will start on `http://127.0.0.1:8090`

### 4. Create an admin account

1. Visit `http://127.0.0.1:8090/_/` in your browser
2. Create your admin account (email and password)
3. The `submissions` collection will be automatically created via migrations

## Usage

### Public submission page
- Visit `http://127.0.0.1:8090/`
- Users can submit ideas anonymously

### Review dashboard
- Visit `http://127.0.0.1:8090/review`
- Log in with your admin credentials
- Review submissions and change their status

### Status options
- **Approved** - Ideas you want to use or move forward with
- **New** - Newly submitted ideas (default)
- **Done** - Completed ideas
- **Hidden** - Ideas you want to keep but not display prominently
- **Deleted** - Soft-deleted ideas (still in database)

## Deployment

### Building for production

```bash
go build -o litterbox
```

This creates a single executable with everything embedded.

### Running in production

```bash
./litterbox serve --http="0.0.0.0:8090"
```

### Custom domain setup

#### Option 1: Reverse proxy with Nginx

1. Install Nginx on your server

2. Create an Nginx configuration file (`/etc/nginx/sites-available/litterbox`):

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

3. Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/litterbox /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

4. Set up HTTPS with Let's Encrypt:

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

#### Option 2: Direct HTTPS with PocketBase

PocketBase can serve HTTPS directly:

```bash
./litterbox serve --http="yourdomain.com:443" --https
```

You'll need to place your SSL certificates in the `pb_data` directory:
- `pb_data/cert.pem`
- `pb_data/key.pem`

#### Option 3: Using a service like Caddy

Create a `Caddyfile`:

```
yourdomain.com {
    reverse_proxy localhost:8090
}
```

Run Caddy:

```bash
caddy run
```

Caddy will automatically handle HTTPS certificates via Let's Encrypt.

### Running as a system service

Create a systemd service file (`/etc/systemd/system/litterbox.service`):

```ini
[Unit]
Description=Litterbox Submission System
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/litterbox
ExecStart=/path/to/litterbox/litterbox serve
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable litterbox
sudo systemctl start litterbox
sudo systemctl status litterbox
```

## Customization

### Editing templates

HTML templates are located in the `templates/` directory:
- `submission_form.html` - Public submission page
- `thank_you.html` - Thank you confirmation page
- `review.html` - Admin review dashboard

Edit these files to customize the look and feel of your application.

### Database location

PocketBase stores all data in the `pb_data/` directory. Make sure to back this up regularly!

## Development

### Creating additional migrations

```bash
go run . migrate create "migration_name"
```

### Viewing logs

Logs are stored in `pb_data/logs/`

### Accessing the PocketBase admin UI

Visit `http://127.0.0.1:8090/_/` to access the built-in PocketBase admin interface where you can:
- View and edit collections directly
- Manage admin accounts
- View logs
- Configure settings

## Troubleshooting

### Port already in use

If port 8090 is already taken, specify a different port:

```bash
go run . serve --http="127.0.0.1:8091"
```

### Migration issues

If you need to reset the database:

```bash
rm -rf pb_data
go run . serve
```

This will recreate everything from scratch (you'll need to create a new admin account).

### Email validation

Admin accounts require valid email formats. Use real TLDs (`.com`, `.org`, etc.) - test domains like `.test` or `.asdf` won't work.

## License

See LICENSE file for details.