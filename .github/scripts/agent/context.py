import os
import glob
import logging
from . import config

logger = logging.getLogger(__name__)

def get_overview_doc():
    """Reads the global project overview and migration rules."""
    path = "docs/refactoring/00_overview.md"
    if os.path.exists(path):
        with open(path, 'r') as f:
            return f.read()
    return ""

def get_refactoring_doc():
    """Reads the specific refactoring task markdown file."""
    if config.PR_QUESTION:
        return None # No doc needed for Q&A mode

    search_pattern = f"docs/refactoring/{config.TASK_ID}_*.md"
    files = glob.glob(search_pattern)
    if not files:
        raise FileNotFoundError(f"No instruction file found for Task {config.TASK_ID}")
    
    with open(files[0], 'r') as f:
        return f.read()

def get_codebase_context():
    """Reads relevant file headers/structures."""
    context = "Current Project Structure & Key Files:\n"
    
    for root, dirs, files in os.walk("."):
        dirs[:] = [d for d in dirs if d not in config.EXCLUDE_DIRS]
        
        for file in files:
            if file.endswith(('.go', '.md', '.tf')):
                if file in config.SKIP_FILES: continue
                if "00_overview.md" in file: continue

                path = os.path.join(root, file)
                try:
                    with open(path, 'r', errors='ignore') as f:
                        lines = f.readlines()
                        # Context Limit: 80 lines per file
                        context += f"\n--- File: {path} ---\n"
                        context += "".join(lines[:80]) 
                        if len(lines) > 80: context += "\n... (truncated)\n"
                except Exception as e:
                    logger.warning(f"Could not read {path}: {e}")

    return context
