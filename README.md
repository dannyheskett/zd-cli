# Zendesk CLI (zd)

A powerful command-line interface for managing Zendesk instances. Manage tickets, users, organizations, and groups with ease.

## Features

- **Multi-Instance Support** - Manage multiple Zendesk accounts and switch between them
- **Dual Authentication** - API Token and OAuth 2.0 support
- **Smart Caching** - 15-minute TTL cache reduces API calls by 52%
- **Multiple Output Formats** - Table (human), JSON, and CSV output
- **Shell Completion** - Tab completion for bash, zsh, fish, and PowerShell
- **Comprehensive Commands** - 47+ commands covering users, tickets, organizations, and groups

## Installation

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/dannyheskett/zd-cli/releases):

**Linux (AMD64):**
```bash
wget https://github.com/dannyheskett/zd-cli/releases/latest/download/zd-linux-amd64
chmod +x zd-linux-amd64
sudo mv zd-linux-amd64 /usr/local/bin/zd
```

**macOS (Intel):**
```bash
wget https://github.com/dannyheskett/zd-cli/releases/latest/download/zd-darwin-amd64
chmod +x zd-darwin-amd64
sudo mv zd-darwin-amd64 /usr/local/bin/zd
```

**macOS (Apple Silicon):**
```bash
wget https://github.com/dannyheskett/zd-cli/releases/latest/download/zd-darwin-arm64
chmod +x zd-darwin-arm64
sudo mv zd-darwin-arm64 /usr/local/bin/zd
```

**Windows:**
Download `zd-windows-amd64.exe` from releases and add to your PATH.

### Build from Source

**Prerequisites:** Go 1.24 or higher

```bash
git clone https://github.com/dannyheskett/zd-cli.git
cd zd-cli
go build -o zd ./cmd/zd
```

### Install to System

After building:
```bash
sudo ./zd install
```

Or manually:
```bash
sudo cp zd /usr/local/bin/
```

### Shell Completion

```bash
zd completion
```

Automatically detects your shell and installs tab completion.

---

## Quick Start

### 1. Initialize

```bash
zd init
```

**Example Session:**
```
Welcome to Zendesk CLI!
Let's set up your first Zendesk instance.

Instance name: production
Zendesk subdomain: mycompany
Authentication method:
  1. API Token
  2. OAuth
> Select: 1

Email address: admin@mycompany.com
API Token: ********

✓ Configuration initialized successfully!
Instance 'production' is now active.
Run 'zd test' to verify your connection.
```

### 2. Test Connection

```bash
zd test
```

**Output:**
```
Testing connection to 'production' (mycompany.zendesk.com)...
✓ Connection successful!
  Authenticated as: John Doe
  Email: admin@mycompany.com
  Role: admin
```

### 3. Start Using Commands

```bash
# List tickets
zd ticket list --per-page 10

# Search users
zd user search "john"

# View ticket details
zd ticket show 12345
```

---

## Configuration

### Config File Location

`~/.zd/config` (INI format)

**Example:**
```ini
[core]
current = production

[instance "production"]
subdomain = mycompany
auth_type = token
email     = admin@mycompany.com
api_token = your_api_token_here

[instance "staging"]
subdomain = mycompany-staging
auth_type = oauth
oauth_client_id  = zd_cli_abc123
oauth_secret     = secret_here
oauth_token      = access_token
oauth_refresh    = refresh_token
oauth_expiry     = 2026-03-01T12:00:00Z
```

### Managing Instances

```bash
# Add another instance
zd instance add

# List all instances
zd instance list
```

**Output:**
```
    NAME                 SUBDOMAIN                      AUTH TYPE  EMAIL
--------------------------------------------------------------------------------
*   production           mycompany.zendesk.com          token      admin@mycompany.com
    staging              mycompany-staging.zendesk.com  oauth      (OAuth)
```

```bash
# Switch between instances
zd instance switch staging

# View current instance
zd instance current
```

**Output:**
```
Current instance: staging
  Subdomain: mycompany-staging.zendesk.com
  Auth Type: oauth
```

```bash
# Remove an instance
zd instance remove staging
```

---

## Authentication

### API Token (Recommended for Personal Use)

**Getting Your API Token:**
1. Log in to Zendesk
2. Go to Admin → Channels → API
3. Enable token access
4. Click "Add API token"
5. Copy the token

**Setup:**
```bash
zd init
# Select "API Token" and enter your email and token
```

### OAuth (Recommended for Organizations)

**Admin Setup Required:**
1. Go to Admin Center → Apps and integrations → APIs → Zendesk API → OAuth Clients
2. Click "Add OAuth Client"
3. Set Client Kind to "Confidential"
4. Add Redirect URL: `http://localhost:8080/callback`
5. Save and copy Client ID and Secret

**User Setup:**
```bash
zd init
# Select "OAuth"
# Enter Client ID and Secret
# Browser opens for authorization
# Click "Allow"
```

See `docs/oauth-setup.md` for detailed OAuth instructions.

---

## Command Reference

### User Commands

#### List Users

```bash
zd user list --per-page 10
```

**Output:**
```
Users (Page 1, showing 10 of 1234 total)
────────────────────────────────────────────────────────────────────────────────

#1   John Doe | john.doe@company.com | admin | ID: 123456789 | ✓
#2   Jane Smith | jane.smith@company.com | agent | ID: 987654321 | ✓
#3   Bob Johnson | bob@example.com | end-user | ID: 456789123
#4   Alice Williams | alice.w@company.com | agent | ID: 789123456 | ✓
#5   Charlie Brown | charlie@example.com | end-user | ID: 321654987 | ✓
...
```

**With Filters:**
```bash
zd user list --page 2 --per-page 50
```

#### Search Users

```bash
zd user search "john"
```

**Output:**
```
Found 5 user(s)
────────────────────────────────────────────────────────────────────────────────

#1   John Doe | john.doe@company.com | admin | ID: 123456789 | ✓
#2   John Smith | john.smith@example.com | end-user | ID: 234567890
#3   Johnny Walker | johnny.w@company.com | agent | ID: 345678901 | ✓
#4   John Johnson | jjohnson@example.com | end-user | ID: 456789012
#5   Johnathan Lee | jlee@company.com | agent | ID: 567890123 | ✓
```

#### Show User Details

```bash
zd user show 123456789
```

**Output:**
```
User: John Doe
────────────────────────────────────────────────────────────────────────────────
ID:           123456789
Email:        john.doe@company.com
Role:         admin
Phone:        555-1234
Time Zone:    Eastern Time (US & Canada)
Locale:       en-US

Status:
  ✓ Active
  ✓ Verified

Dates:
  Created:      2023-01-15 10:30:00 EST
  Last Updated: 2026-02-01 14:22:00 EST
  Last Login:   2026-02-07 09:15:00 EST

URL:          https://mycompany.zendesk.com/api/v2/users/123456789.json
```

#### Show Current User

```bash
zd user me
```

**Output:**
```
User: John Doe
────────────────────────────────────────────────────────────────────────────────
ID:           123456789
Email:        john.doe@company.com
Role:         admin
Phone:        555-1234
Time Zone:    Eastern Time (US & Canada)
Locale:       en-US

Status:
  ✓ Active
  ✓ Verified
  ✓ 2FA Enabled

Dates:
  Created:      2023-01-15 10:30:00 EST
  Last Updated: 2026-02-01 14:22:00 EST
  Last Login:   2026-02-07 09:15:00 EST

URL:          https://mycompany.zendesk.com/api/v2/users/123456789.json
```

#### Create User

```bash
zd user create --name "New User" --email "newuser@example.com" --role end-user
```

**Output:**
```
✓ User created successfully!
User ID: 999888777
Name: New User
Email: newuser@example.com
Role: end-user
```

**Interactive Mode:**
```bash
zd user create
```

**Prompts:**
```
Name: New User
Email: newuser@example.com

✓ User created successfully!
User ID: 999888777
```

#### Update User

```bash
zd user update 999888777 --name "Updated Name" --phone "555-9999"
```

**Output:**
```
✓ User #999888777 updated successfully!
User: Updated Name
────────────────────────────────────────────────────────────────────────────────
ID:           999888777
Email:        newuser@example.com
Role:         end-user
Phone:        555-9999
...
```

#### Suspend/Unsuspend User

```bash
zd user suspend 999888777
```

**Output:**
```
✓ User #999888777 suspended
Name: Updated Name
```

```bash
zd user unsuspend 999888777
```

**Output:**
```
✓ User #999888777 unsuspended
Name: Updated Name
```

#### Delete User

```bash
zd user delete 999888777
```

**Output:**
```
WARNING: This will permanently delete user 999888777
Type 'yes' to confirm: yes
✓ User #999888777 deleted
```

**Skip Confirmation:**
```bash
zd user delete 999888777 --force
```

---

### Ticket Commands

#### List Tickets

```bash
zd ticket list --per-page 10
```

**Output:**
```
Tickets (Page 1, showing 10 of 5678 total)
────────────────────────────────────────────────────────────────────────────────

#1    new      | | Login issues on mobile app | ID: 12345
#2    open     | | Cannot access dashboard | ID: 12346
#3    ↑pending | | High priority - System down | ID: 12347
#4    !open    | | URGENT: Payment processing broken | ID: 12348
#5    solved   | | Password reset request | ID: 12349
#6    closed   | | Feature request: Dark mode | ID: 12350
#7    pending  | | Slow loading times | ID: 12351
#8    open     | | Email notifications not working | ID: 12352
#9    new      | | Question about pricing | ID: 12353
#10   solved   | | Account locked | ID: 12354

More results available. Use --page 2 to see next page.
```

**Filter by Status:**
```bash
zd ticket list --status open --per-page 5
```

**Output:**
```
Tickets (Page 1, showing 5 of 234 total)
────────────────────────────────────────────────────────────────────────────────

#1    open     | | Cannot access dashboard | ID: 12346
#2    !open    | | URGENT: Payment processing broken | ID: 12348
#3    open     | | Email notifications not working | ID: 12352
#4    open     | | Integration sync failing | ID: 12400
#5    ↑open    | | Data export not completing | ID: 12455
```

**Legend:**
- `!` = Urgent priority (red)
- `↑` = High priority (yellow)
- Status colors: new (cyan), open (blue), pending (yellow), solved (green), closed (gray)

#### Show Ticket Details

```bash
zd ticket show 12345
```

**Output:**
```
Ticket #12345: Login issues on mobile app
────────────────────────────────────────────────────────────────────────────────
Status:       open
Priority:     normal
Type:         incident

People:
  Requester ID: 123456789
  Submitter ID: 123456789
  Assignee ID:  987654321
  Group:        12355006972955

Dates:
  Created:      2026-02-01 10:30:00 EST
  Updated:      2026-02-07 09:15:00 EST

Tags: mobile, login, bug

Description:
Users are reporting they cannot log in to the mobile app. The login
button appears to be unresponsive after entering credentials. This
started happening after the latest app update.

URL: https://mycompany.zendesk.com/api/v2/tickets/12345.json
```

#### View Ticket Comments

```bash
zd ticket comments 12345
```

**Output:**
```
Comments for Ticket #12345 (4 total)
────────────────────────────────────────────────────────────────────────────────

#1   [Public] Author ID: 123456789 | 2026-02-01 10:30:00 EST
Users are reporting they cannot log in to the mobile app. The login
button appears to be unresponsive after entering credentials.

#2   [Public] Author ID: 987654321 | 2026-02-01 11:15:00 EST
Thanks for reporting this. I've escalated to the development team.
We're investigating the issue now.

#3   [Private] Author ID: 987654321 | 2026-02-01 14:30:00 EST
Found the bug - it's related to the OAuth token refresh. Deploying fix now.

#4   [Public] Author ID: 987654321 | 2026-02-07 09:15:00 EST
This has been fixed in version 2.1.5. Please update your app.
```

#### Search Tickets

```bash
zd ticket search "login"
```

**Output:**
```
Found 25 ticket(s)
────────────────────────────────────────────────────────────────────────────────

#1    closed   | | Login page not loading | ID: 11111
#2    solved   | | Cannot login after password reset | ID: 11112
#3    open     | | Login issues on mobile app | ID: 12345
#4    closed   | | Login credentials incorrect | ID: 11113
#5    pending  | | SSO login failing | ID: 11114
...
```

**Advanced Search:**
```bash
zd ticket search "status:open priority:urgent"
zd ticket search "assignee:me status:pending"
```

#### Create Ticket

```bash
zd ticket create \
  --subject "Website down" \
  --description "The main website is not responding" \
  --priority urgent \
  --tags outage,website
```

**Output:**
```
✓ Ticket created successfully!
Ticket ID: 12999
Status: new
URL: https://mycompany.zendesk.com/api/v2/tickets/12999.json
```

**Interactive Mode:**
```bash
zd ticket create
```

**Prompts:**
```
Subject: Website down
Description: The main website is not responding

✓ Ticket created successfully!
Ticket ID: 12999
```

#### Update Ticket

```bash
zd ticket update 12999 --status open --priority high --assignee 987654321
```

**Output:**
```
✓ Ticket #12999 updated successfully!
Ticket #12999: Website down
────────────────────────────────────────────────────────────────────────────────
Status:       open
Priority:     high
Type:         incident

People:
  Requester ID: 123456789
  Submitter ID: 123456789
  Assignee ID:  987654321
...
```

#### Add Comment to Ticket

```bash
zd ticket comment 12999 --message "Website is back up now" --public
```

**Output:**
```
✓ Added public comment to ticket #12999
```

**Private Comment:**
```bash
zd ticket comment 12999 --message "Internal note: server restart fixed it" --private
```

**Interactive Mode:**
```bash
zd ticket comment 12999
```

**Prompts:**
```
Comment: Issue has been resolved
✓ Added public comment to ticket #12999
```

#### Assign Ticket

```bash
zd ticket assign 12999 987654321
```

**Output:**
```
✓ Ticket #12999 assigned to user 987654321
```

#### Close Ticket

```bash
zd ticket close 12999 --comment "Issue resolved"
```

**Output:**
```
✓ Ticket #12999 closed
```

---

### Organization Commands

#### List Organizations

```bash
zd org list --per-page 5
```

**Output:**
```
Organizations (Page 1, showing 5 of 123 total)
────────────────────────────────────────────────────────────────────────────────

#1   Acme Corporation | ID: 11111111 | shared tickets
#2   Tech Startup Inc | ID: 22222222
#3   Global Solutions | ID: 33333333 | shared tickets | shared comments
#4   Small Business LLC | ID: 44444444
#5   Enterprise Co | ID: 55555555 | shared tickets

More results available. Use --page 2 to see next page.
```

#### Show Organization

```bash
zd org show 11111111
```

**Output:**
```
Organization: Acme Corporation
────────────────────────────────────────────────────────────────────────────────
ID:           11111111
Domains:      acme.com, acmecorp.com

Sharing:
  ✓ Shared Tickets
  ○ Private Comments

Dates:
  Created:      2023-01-01 00:00:00 UTC
  Last Updated: 2026-01-15 10:30:00 UTC

Tags: enterprise, premium

Notes:
VIP customer - handle with priority

URL: https://mycompany.zendesk.com/api/v2/organizations/11111111.json
```

#### Search Organizations

```bash
zd org search "Acme"
```

**Output:**
```
Found 3 organization(s)
────────────────────────────────────────────────────────────────────────────────

#1   Acme Corporation | ID: 11111111 | shared tickets
#2   Acme Holdings | ID: 66666666
#3   Acme Industries | ID: 77777777 | shared tickets
```

#### List Users in Organization

```bash
zd org users 11111111 --per-page 5
```

**Output:**
```
Users (Page 1, showing 5 of 45 total)
────────────────────────────────────────────────────────────────────────────────

#1   John Doe | john.doe@acme.com | admin | ID: 123456789 | ✓
#2   Jane Smith | jane.smith@acme.com | agent | ID: 987654321 | ✓
#3   Bob Wilson | bob.wilson@acme.com | end-user | ID: 111222333 | ✓
#4   Sarah Connor | sarah.c@acme.com | agent | ID: 444555666 | ✓
#5   Mike Ross | mike.ross@acme.com | end-user | ID: 777888999
```

#### List Tickets for Organization

```bash
zd org tickets 11111111 --per-page 5
```

**Output:**
```
Tickets (Page 1, showing 5 of 234 total)
────────────────────────────────────────────────────────────────────────────────

#1    open     | | Dashboard access issue | ID: 12345
#2    pending  | | API integration question | ID: 12400
#3    solved   | | Billing inquiry | ID: 12450
#4    new      | | Feature request | ID: 12500
#5    closed   | | Login problem | ID: 12550
```

---

### Group Commands

#### List Groups

```bash
zd group list
```

**Output:**
```
Groups (Page 1, showing 5 of 15 total)
────────────────────────────────────────────────────────────────────────────────

#1   Support Team | ID: 12345678
#2   Sales Team | ID: 23456789
#3   Engineering | ID: 34567890 | default
#4   Customer Success | ID: 45678901
#5   Management | ID: 56789012
```

#### Show Group

```bash
zd group show 12345678
```

**Output:**
```
Group: Support Team
────────────────────────────────────────────────────────────────────────────────
ID:           12345678
Description:  First-line support team for customer inquiries

Status:
  ✓ Active

Dates:
  Created:      2023-01-01 00:00:00 UTC
  Last Updated: 2024-06-15 10:00:00 UTC

URL: https://mycompany.zendesk.com/api/v2/groups/12345678.json
```

#### List Users in Group

```bash
zd group users 12345678 --per-page 5
```

**Output:**
```
Users (Page 1, showing 5 of 12 total)
────────────────────────────────────────────────────────────────────────────────

#1   John Doe | john.doe@company.com | agent | ID: 123456789 | ✓
#2   Jane Smith | jane.smith@company.com | agent | ID: 987654321 | ✓
#3   Bob Johnson | bob.j@company.com | agent | ID: 456789123 | ✓
#4   Alice Williams | alice.w@company.com | agent | ID: 789123456 | ✓
#5   Charlie Brown | charlie.b@company.com | agent | ID: 321654987 | ✓
```

#### List Group Memberships

```bash
zd group memberships 12345678 --per-page 5
```

**Output:**
```
Group Memberships (Page 1, showing 5 of 12 total)
────────────────────────────────────────────────────────────────────────────────

#1   User ID: 123456789 | default
#2   User ID: 987654321
#3   User ID: 456789123
#4   User ID: 789123456 | default
#5   User ID: 321654987
```

---

### Output Formats

All commands support multiple output formats:

#### JSON Output

```bash
zd user show 123456789 -o json
```

**Output:**
```json
{
  "id": 123456789,
  "name": "John Doe",
  "email": "john.doe@company.com",
  "role": "admin",
  "verified": true,
  "active": true,
  "suspended": false,
  "phone": "555-1234",
  "time_zone": "Eastern Time (US & Canada)",
  "locale": "en-US",
  "created_at": "2023-01-15T10:30:00Z",
  "updated_at": "2026-02-01T14:22:00Z",
  ...
}
```

#### CSV Output

```bash
zd user list --per-page 5 -o csv
```

**Output:**
```csv
id,name,email,role,active,verified,suspended,organization_id,phone,time_zone,created_at,updated_at
123456789,John Doe,john.doe@company.com,admin,true,true,false,11111111,555-1234,Eastern Time (US & Canada),2023-01-15T10:30:00Z,2026-02-01T14:22:00Z
987654321,Jane Smith,jane.smith@company.com,agent,true,true,false,11111111,555-5678,Eastern Time (US & Canada),2023-02-20T15:45:00Z,2026-01-30T11:20:00Z
456789123,Bob Johnson,bob@example.com,end-user,true,false,false,,555-9012,Pacific Time (US & Canada),2023-03-10T09:00:00Z,2025-12-15T16:30:00Z
```

**Pipe to File:**
```bash
zd ticket list --status solved --per-page 1000 -o csv > solved_tickets.csv
zd user list --per-page 1000 -o csv > all_users.csv
```

---

### Cache Management

#### View Cache Info

```bash
zd cache info
```

**Output:**
```
Cache Information
─────────────────
Location:     ~/.zd/cache
Entries:      32
Total size:   740.44 KB
Default TTL:  10 minutes
```

#### Clear Cache

```bash
zd cache clear
```

**Output:**
```
✓ Cache cleared successfully!
```

#### Bypass Cache

All read commands support `--refresh` flag:

```bash
zd ticket show 12345 --refresh
zd user list --refresh
zd org show 11111111 --refresh
```

---

## Advanced Usage

### Multiple Instances

```bash
# Add staging instance
zd instance add
# Name: staging, subdomain: mycompany-staging, ...

# List instances
zd instance list
```

**Output:**
```
    NAME                 SUBDOMAIN                      AUTH TYPE  EMAIL
--------------------------------------------------------------------------------
*   production           mycompany.zendesk.com          token      admin@mycompany.com
    staging              mycompany-staging.zendesk.com  oauth      (OAuth)
    dev                  mycompany-dev.zendesk.com      token      dev@mycompany.com
```

```bash
# Switch to staging
zd instance switch staging

# Run commands against staging
zd ticket list --per-page 5

# Switch back to production
zd instance switch production
```

### Override Instance for Single Command

```bash
# Run against staging without switching
zd --instance staging ticket list

# Current instance remains unchanged
zd instance current
```

**Output:**
```
Current instance: production
```

### Scripting & Automation

```bash
# Export all users to CSV
zd user list --per-page 1000 -o csv > users_backup.csv

# Export all tickets to JSON
zd ticket list --per-page 1000 -o json > tickets.json

# Search and process with jq
zd user search "admin" -o json | jq '.[] | {id, name, email}'

# Bulk operations
for ticket_id in $(cat ticket_ids.txt); do
  zd ticket close $ticket_id --comment "Bulk closure"
done
```

---

## Configuration Files

### Directory Structure

```
~/.zd/
├── config              # Main configuration (INI format)
└── cache/             # API response cache
    └── *.json         # Cached responses (auto-managed)
```

### Permissions

- Config directory: `0700` (rwx------)
- Config file: `0600` (rw-------)
- Cache files: `0600` (rw-------)

**Security:** Only your user can read/write these files.

---

## Troubleshooting

### "no configuration found"

**Solution:**
```bash
zd init
```

### "no current instance set"

**Solution:**
```bash
zd instance list
zd instance switch <instance-name>
```

### "Authentication Failed"

**Solution:**
```bash
# Test connection
zd test

# For API token auth
# 1. Verify token in Zendesk (Admin → API → Tokens)
# 2. Regenerate if needed
# 3. Update: zd instance add

# For OAuth auth
zd reauth
```

### "Rate Limit Exceeded"

**Solution:**
- Wait 1 minute before retrying
- Use cache (don't use --refresh unnecessarily)
- The tool automatically retries with backoff

### "Resource Not Found"

**Solution:**
```bash
# Verify the ID exists
zd user search "name"
zd ticket search "keyword"
```

### Connection Issues

**Solution:**
```bash
# Check network
ping mycompany.zendesk.com

# Verify subdomain is correct
zd instance current

# Test with verbose errors
zd test
```

---

## Command Cheat Sheet

### Common Operations

```bash
# Configuration
zd init                           # First-time setup
zd instance add                   # Add another Zendesk account
zd instance switch <name>         # Switch accounts
zd test                          # Verify connection

# Users
zd user me                        # Show current user
zd user list                      # List all users
zd user search "john"            # Search by name
zd user show 123456              # View user details
zd user create                    # Create user (interactive)
zd user update 123456 --role agent # Promote to agent
zd user suspend 123456            # Suspend user
zd user delete 123456             # Delete user

# Tickets
zd ticket list                    # List all tickets
zd ticket list --status open      # Filter by status
zd ticket search "login"         # Search tickets
zd ticket show 12345             # View ticket
zd ticket comments 12345         # View conversation
zd ticket create                  # Create ticket (interactive)
zd ticket update 12345 --priority high # Update ticket
zd ticket comment 12345          # Add comment (interactive)
zd ticket assign 12345 987654   # Assign ticket
zd ticket close 12345            # Close ticket

# Organizations
zd org list                       # List organizations
zd org show 11111                # View organization
zd org search "acme"             # Search organizations
zd org users 11111               # Users in org
zd org tickets 11111             # Tickets for org

# Groups
zd group list                     # List groups
zd group show 12345              # View group
zd group users 12345             # Users in group
zd group memberships 12345       # Group memberships

# Cache
zd cache info                     # Cache statistics
zd cache clear                    # Clear cache

# Utilities
zd install                        # Install to /usr/local/bin
zd completion                     # Install shell completion
zd reauth                         # Re-authorize OAuth
```

### Output Formats

```bash
# Default table format (human-readable)
zd user list

# JSON format (for scripts/jq)
zd user list -o json

# CSV format (for spreadsheets)
zd user list -o csv > users.csv
```

### Pagination

```bash
# Specify page and results per page
zd ticket list --page 2 --per-page 50

# Maximum per page is 100
zd user list --per-page 100
```

---

## Best Practices

### 1. Use Cache Wisely

```bash
# Let cache work (default behavior)
zd ticket list

# Only use --refresh when you need real-time data
zd ticket show 12345 --refresh
```

### 2. Use Appropriate Output Formats

```bash
# Interactive use - table format (default)
zd user list

# Scripting - JSON or CSV
zd user list -o json | jq '.[].email'
```

### 3. Leverage Search

```bash
# Find before operating
zd user search "john" -o json | jq '.[0].id'
ticket_id=$(zd ticket search "login issue" -o json | jq -r '.[0].id')
zd ticket update $ticket_id --status solved
```

### 4. Use Instance Switching for Multi-Environment

```bash
# Test in staging first
zd instance switch staging
zd ticket create --subject "Test"

# Then promote to production
zd instance switch production
zd ticket create --subject "Production Issue"
```

---

## Examples by Use Case

### Daily Support Workflow

```bash
# Check your assigned tickets
zd ticket search "assignee:me status:open"

# View a ticket
zd ticket show 12345

# Read the conversation
zd ticket comments 12345

# Add a response
zd ticket comment 12345 --message "I'm looking into this now"

# Update and close
zd ticket update 12345 --status solved
zd ticket close 12345 --comment "Issue resolved"
```

### User Management

```bash
# Find a user
zd user search "john.doe@company.com"

# Create new user
zd user create \
  --name "New Employee" \
  --email "new.employee@company.com" \
  --role agent

# Update user role
zd user update 123456789 --role admin

# Suspend problematic user
zd user suspend 987654321
```

### Reporting & Export

```bash
# Export all users
zd user list --per-page 1000 -o csv > all_users_$(date +%Y%m%d).csv

# Export open tickets
zd ticket list --status open --per-page 1000 -o json > open_tickets.json

# Export organization data
zd org list --per-page 1000 -o csv > organizations.csv
```

### Bulk Ticket Updates

```bash
# Get list of ticket IDs
zd ticket search "status:new tag:urgent" -o json | jq -r '.[].id' > urgent_tickets.txt

# Assign them all
while read ticket_id; do
  zd ticket assign $ticket_id 987654321
  echo "Assigned ticket $ticket_id"
done < urgent_tickets.txt
```

---

---

## Development

### Project Structure

```
zd-cli/
├── cmd/zd/main.go              # CLI entry point
├── internal/
│   ├── auth/                   # Authentication (token, OAuth)
│   ├── cache/                  # Response caching
│   ├── client/                 # Zendesk API client
│   ├── commands/               # CLI commands
│   ├── config/                 # Configuration management
│   ├── output/                 # Output formatting (JSON/CSV/table)
│   └── progress/               # Progress indicators
└── go.mod                      # Dependencies
```

### Building

```bash
go build -o zd ./cmd/zd
```

### Running Tests

```bash
# Manual testing workflow
./zd init
./zd test
./zd user list --per-page 5
./zd ticket list --per-page 5
./zd org list --per-page 5
```

---

## API Coverage

### Implemented Endpoints

**Users (9 endpoints):**
- GET /users/me.json
- GET /users.json
- GET /users/{id}.json
- GET /users/search.json
- POST /users.json
- PUT /users/{id}.json
- DELETE /users/{id}.json

**Tickets (10 endpoints):**
- GET /tickets.json
- GET /tickets/{id}.json
- GET /tickets/{id}/comments.json
- GET /search.json (tickets)
- POST /tickets.json
- PUT /tickets/{id}.json

**Organizations (5 endpoints):**
- GET /organizations.json
- GET /organizations/{id}.json
- GET /organizations/search.json
- GET /organizations/{id}/users.json
- GET /organizations/{id}/tickets.json

**Groups (4 endpoints):**
- GET /groups.json
- GET /groups/{id}.json
- GET /groups/{id}/users.json
- GET /groups/{id}/memberships.json

**Total:** 28+ API endpoints

---

## Contributing

Contributions welcome! The architecture is designed for easy extension:

### Adding a New Command

1. Create API method in `internal/client/`
2. Create command in `internal/commands/`
3. Register in `cmd/zd/main.go`
4. Add output formatting (JSON/CSV/table)
5. Add caching support
6. Update documentation

See `docs/implementation-roadmap.md` for planned features.

---

## License

MIT License

---

## Support

- **Command Help:** `zd <command> --help`
- **Issues:** https://github.com/dannyheskett/zd-cli/issues
- **Discussions:** https://github.com/dannyheskett/zd-cli/discussions

---

## Changelog

### v0.3.0 (Current)
- Initial public release
- User management (full CRUD)
- Ticket management (full CRUD + comments)
- Organization management
- Group management
- API Token + OAuth authentication
- Smart caching (10-minute TTL)
- JSON/CSV/Table output formats
- Shell completion (bash/zsh/fish/powershell)
- Multi-instance support
- Error handling with retry logic
- 47+ commands
