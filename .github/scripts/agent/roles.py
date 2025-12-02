from .llm import generate_with_fallback
from . import config

def run_coder(instruction, context, feedback="", overview=""):
    """
    The Coder persona. 
    Handles both Refactoring (File I/O) and Q&A (Explanation).
    """
    
    # Mode 1: Q&A / Quick Fix
    if config.PR_QUESTION:
        user_request = config.PR_QUESTION.replace('/ask', '').strip()

        if config.COMMENT_PATH and config.COMMENT_END_LINE:
            line_info = f"TARGET LINE: ${config.COMMENT_END_LINE}" # default single line

            # Handle multi-line selected comment
            if config.COMMENT_START_LINE and config.COMMENT_START_LINE != "None" and config.COMMENT_START_LINE != config.COMMENT_END_LINE:
                line_info = f"TARGET LINES: {config.COMMENT_START_LINE}-{config.COMMENT_END_LINE}"

            user_request = (
                f"TARGET FILE: {config.COMMENT_PATH}\n"
                f"{line_info}\n"
                f"QUESTION: {user_request}"
            )

        prompt = f"""
        You are a Helpful Senior Engineer Assistant.
        GLOBAL PROJECT RULES: {overview}
        CONTEXT: {context}
        USER REQUEST: {user_request}
        
        INSTRUCTIONS:
        - If asking for explanation, explain clearly.
        - If asking for code changes, output the FULL file content.
        - Use format: ### File: path/to/file.go
        """
        return generate_with_fallback(prompt)

    # Mode 2: Strict Refactoring Loop
    prompt = f"""
    You are a Senior Go Engineer. Implement the following task.
   
    GLOBAL PROJECT RULES (MUST FOLLOW):
    {overview}

    CONTEXT:
    {context}
    
    TASK INSTRUCTIONS:
    {instruction}
    
    REQUIREMENTS:
    1. Output the FULL content of any file you create or modify.
    2. Format: 
       ### File: path/to/file
       ```
       // content
       ```
    3. FIX issues from the FEEDBACK below (if any).
    
    PREVIOUS REVIEWER FEEDBACK:
    {feedback if feedback else "None."}
    """
    return generate_with_fallback(prompt)

def run_reviewer(instruction, generated_code):
    """The Reviewer persona."""
    prompt = f"""
    You are a Strict Code Reviewer (Principal Engineer).
    
    Verify the code below against instructions:
    {instruction}
    
    GENERATED CODE:
    {generated_code}
    
    OUTPUT FORMAT:
    First line: STATUS: [PASS or FAIL]
    Subsequent lines: Bullet points of critique.
    """
    return generate_with_fallback(prompt)
