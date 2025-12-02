import time
import logging
from google import genai
from . import config

logger = logging.getLogger(__name__)

# Initialize Client
client = genai.Client(
    api_key=config.API_KEY,
    http_options={'timeout': config.HTTP_TIMEOUT_MS}
)

def generate_with_fallback(prompt):
    """Tries primary model, falls back to stable models on 503/Timeout."""
    models = [
        "gemini-3-pro-preview",         # Primary
        "gemini-2.5-pro"                # Fallback
    ]

    last_exception = None

    for attempt in range(config.MAX_RETRIES):
        model_name = models[attempt % len(models)]
        
        try:
            logger.info(f"Generating with {model_name} (Attempt {attempt+1})")
            response = client.models.generate_content(
                model=model_name,
                contents=prompt
            )
            return response.text

        except Exception as e:
            last_exception = e
            error_msg = str(e)
            logger.error(f"Failed with {model_name}: {error_msg}")
            
            sleep_time = 15 * (attempt + 1)
            if "503" in error_msg or "overloaded" in error_msg.lower():
                logger.warning("Server overloaded. Waiting 30s...")
                sleep_time = 30
            
            if attempt < config.MAX_RETRIES - 1:
                time.sleep(sleep_time)

    raise last_exception
