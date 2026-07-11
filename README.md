# eRecruit — Recruitment & Job Portal

A server-rendered recruitment platform built in Go with the [Gin](https://github.com/gin-gonic/gin) web framework. It supports three user roles — **Super Admin**, **Recruiter**, and **Applicant** — covering job posting, applications, skill matching, interview scheduling, and AI-assisted resume parsing/validation via the Gemini API.

## Features

### Authentication & Access Control
- Email/password signup and login with **bcrypt** password hashing.
- **Google OAuth2** login/signup as an alternative to email/password.
- **JWT**-based session cookies (`Authorization` cookie, HS256, 30-day expiry).
- **CSRF protection** middleware that issues and validates per-session tokens on state-changing forms.
- **Role-based access control** (Super Admin / Recruiter / Applicant) enforced via middleware, with new users prompted to choose a group on first login.
- Recruiter accounts require **Super Admin approval** before gaining full access; unapproved recruiters are redirected to a restricted-access page.

### Super Admin
- Dashboard listing recruiters pending approval.
- View a recruiter's profile, approve, or decline their account.

### Recruiter
- Company profile management (name, description).
- Create, edit, and manage job postings with a required-skills list (add/remove skills with live search).
- View all job postings and browse the list of applicants per posting.
- View an applicant's profile, skills, and uploaded resume.
- **Download** an applicant's resume, or **AI-extract** key resume details (name, education, skills, projects, work experience) using Google Gemini.
- Schedule interviews with applicants (date/time), with duplicate-request detection.

### Applicant
- Profile management (name, school, college, age) with a skills list (add/remove via live search).
- **Resume upload** with AI-powered validation: the uploaded PDF is parsed and cross-checked against the applicant's profile/skills using Gemini before being accepted; mismatches are reported back to the user.
- Resume export/download.
- Browse all job postings and view posting details (including whether they've already applied).
- Apply to job postings.
- View incoming interview requests, view details, decline, or request an alternate date/time.

### Data & Infrastructure
- **PostgreSQL** database with schema for users, roles, recruiter/applicant profiles, skills, job postings, applications, and interviews.
- Type-safe database access generated via **sqlc** (`home/products`) from `query.sql` + `schema.sql`.
- Server-rendered HTML via Go's `html/template`, with templates per role under `home/templates`.

## Tech Stack

| Layer          | Technology                                  |
|----------------|----------------------------------------------|
| Language       | Go 1.23                                       |
| Web framework  | [Gin](https://github.com/gin-gonic/gin)       |
| Database       | PostgreSQL                                    |
| DB access      | [sqlc](https://sqlc.dev/) generated queries (`lib/pq` driver) |
| Auth           | JWT (`golang-jwt`), bcrypt, Google OAuth2 (`golang.org/x/oauth2`) |
| PDF parsing    | `github.com/ledongthuc/pdf`                   |
| AI integration | Google Gemini API (resume parsing & validation) |
| Templating     | Go `html/template`                            |

## Project Structure

```
home/
├── main.go                  # Route definitions and app bootstrap
├── controllers/             # Request handlers (auth, applicant, recruiter, super admin)
├── middleware/               # Auth, RBAC, CSRF, recruiter-approval gating
├── initializers/             # Env loading, DB connection setup
├── products/                 # sqlc-generated DB models & query code
├── templates/                # HTML templates per role/page
├── resume/                   # Uploaded applicant resumes (PDF)
├── schema.sql                 # Database schema
├── query.sql                  # SQL queries used to generate products/
└── sqlc.yaml                  # sqlc codegen config
```

## Local Setup

### Prerequisites
- Go 1.23+
- PostgreSQL (running locally or accessible remotely)
- (Optional) [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) if you plan to regenerate `home/products` after editing `query.sql`/`schema.sql`
- A Google Cloud OAuth 2.0 Client ID/Secret (for Google login)
- A Google Gemini API key (for resume parsing/validation features)

### 1. Clone and install dependencies
```bash
git clone <repo-url>
cd task2/home
go mod download
```

### 2. Set up the database
Create a PostgreSQL database and load the schema:
```bash
createdb erecruit_db
psql -d erecruit_db -f schema.sql
```

The app also needs at least a few role rows in `role_master` (ids used in the app: `1 = Super Admin`, `2 = Recruiter`, `3 = Applicant`):
```sql
INSERT INTO role_master (name) VALUES ('Super Admin'), ('Recruiter'), ('Applicant');
```

### 3. Configure environment variables
Create a `home/.env` file (this file is git-ignored — never commit real secrets) with the following keys:

```
PORT=8000
DB_URL="host=localhost user=<db_user> password=<db_password> dbname=erecruit_db port=5432 sslmode=disable"
SECRET=<a long random string used to sign JWTs>
GEMINI_API_KEY=<your Google Gemini API key>
OAUTH_CLIENT_ID=<your Google OAuth client ID>
OAUTH_CLIENT_SECRET=<your Google OAuth client secret>
```

Notes:
- `SECRET` should be a long, random value — used to sign/verify JWT session tokens.
- `GEMINI_API_KEY` is required for resume upload validation and the recruiter's resume-extract feature; without it those specific endpoints will fail, but the rest of the app still runs.
- The Google OAuth redirect URL is hardcoded to `http://127.0.0.1:8000/auth/google/callback` (see `home/controllers/usersControllers.go`) — configure this as an authorized redirect URI in your Google Cloud OAuth client, and adjust it in code if you deploy elsewhere.

### 4. Create the resume upload directory
```bash
mkdir -p home/resume
```
Uploaded applicant resumes are stored here as `<user_id>_resume.pdf`.

### 5. Run the app
```bash
cd home
go run main.go
```
The server starts on the port defined by `PORT` (default `8000`). Visit `http://localhost:8000`.

### 6. (Optional) Regenerate DB code after changing SQL
If you modify `schema.sql` or `query.sql`, regenerate the `home/products` package:
```bash
cd home
sqlc generate
```

## Security Notes
- The `.env` file must never be committed — it is already listed in `.gitignore`. Rotate any credentials that may have been previously exposed in version history.
- Cookies are currently set with `Secure: false` for local HTTP development; set `Secure: true` and serve over HTTPS in any production deployment.
