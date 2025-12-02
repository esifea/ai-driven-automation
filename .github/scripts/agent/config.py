import os

# API Settings
API_KEY = os.getenv("GEMINI_API_KEY")
MAX_RETRIES = int(os.getenv("MAX_RETRIES", 5))

# Task
TASK_ID = os.getenv("TASK_ID", "01")
PR_QUESTION = os.getenv("PR_QUESTION")
COMMENT_PATH = os.getenv("COMMENT_PATH", "")
COMMENT_START_LINE = os.getenv("COMMENT_START_LINE", "")
COMMENT_END_LINE = os.getenv("COMMENT_END_LINE", "")

# Timeout (10 minutes)
HTTP_TIMEOUT_MS = 600000 

# Context filter
EXCLUDE_DIRS = {".git", "node_modules", "vendor", ".github", "dist", "bin"}
SKIP_FILES = {"go.sum", "go.mod", "package-lock.json"}
