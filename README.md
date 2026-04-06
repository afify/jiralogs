# LogS

TUI app for viewing and managing JIRA worklogs.

## Setup

Create a `.env` file:

```
JIRA_BASE_URL=https://your-instance.atlassian.net
JIRA_EMAIL=your.email@domain.com
JIRA_API_TOKEN=your_api_token
```

Create a `leaves.txt` file (or at `~/.jiralogs/leaves.txt`) with leave dates:

```
# One date per line, YYYY-MM-DD
2026-01-01
2026-04-15
```

## Run

```
make
```

## Build

```
make build
```
