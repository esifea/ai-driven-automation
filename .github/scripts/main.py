import logging
import os
import sys
import subprocess
from agent import config, context, roles, utils

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def run_coder_mode():
    """
    Coder Mode:
    1. Reads instruction (Task ID)
    2. Reads Feedback (from 'FEEDBACK' env var, populated by PR review body)
    3. Generates Code
    4. Writes to Disk
    """
    task_id = os.getenv("TASK_ID", "01")
    feedback = os.getenv("FEEDBACK", "")
    
    logger.info(f"--- CODER AGENT STARTED (Task: {task_id}) ---")
    
    if feedback:
        logger.info(f"Feedback received from Reviewer: {feedback}")

    # Load Context
    try:
        overview_doc = context.get_overview_doc()
        instruction = context.get_refactoring_doc()
        repo_context = context.get_codebase_context()
    except Exception as e:
        logger.error(f"Context load failed: {e}")
        sys.exit(1)

    # Generate Code
    generated_code = roles.run_coder(instruction, repo_context, feedback=feedback, overview=overview_doc)

    # Write Files
    files = utils.parse_files(generated_code)
    if files:
        count = utils.write_files_to_disk(files)
        logger.info(f"Coder wrote {count} files to disk.")
    else:
        logger.warning("Coder generated no file output.")

def run_reviewer_mode():
    """
    Reviewer Mode:
    1. Reads instruction & current codebase
    2. Generates Critique
    3. Submits FORMAL REVIEW via 'gh' CLI (Approve / Request Changes)
    """
    logger.info("--- REVIEWER AGENT STARTED ---")
    pr_number = os.getenv("PR_NUMBER")
    
    if not pr_number:
        logger.error("PR_NUMBER is missing.")
        sys.exit(1)

    try:
        instruction = context.get_refactoring_doc() 
        repo_context = context.get_codebase_context()
    except Exception as e:
        logger.error(f"Context load failed: {e}")
        sys.exit(1)

    # Generate Review
    review_result = roles.run_reviewer(instruction, repo_context)
    logger.info("Review generated. Submitting formal review to GitHub...")

    # Determine Status & Format Body
    if "STATUS: PASS" in review_result:
        event_type = "APPROVE"
        body = f"## AI Review: PASS ✅\n\n{review_result}"
    else:
        event_type = "REQUEST_CHANGES"
        body = f"## AI Review: CHANGES REQUESTED ❌\n\n{review_result}"

    # Submit via gh CLI
    try:
        subprocess.run([
            "gh", "pr", "review", pr_number,
            "--" + event_type.lower().replace("_", "-"), # --approve or --request-changes
            "--body", body
        ], check=True)
        logger.info(f"Submitted review: {event_type}")
    except subprocess.CalledProcessError as e:
        logger.error(f"Failed to submit review: {e}")
        sys.exit(1)

def main():
    mode = os.getenv("MODE", "coder")
    
    # Handle Q&A mode (e.g. /ask in comment)
    if config.PR_QUESTION:
        logger.info(f"Processing Q&A: {config.PR_QUESTION}")
        repo_context = context.get_codebase_context()
        overview_doc = context.get_overview_doc()
        answer = roles.run_coder(None, repo_context, overview=overview_doc)
        with open("answer.md", "w") as f:
            f.write(answer)
        return

    if mode == 'reviewer':
        run_reviewer_mode()
    else:
        run_coder_mode()

if __name__ == "__main__":
    main()
